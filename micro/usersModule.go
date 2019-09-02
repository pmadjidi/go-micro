package main

import (
	"crypto/rsa"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mitchellh/mapstructure"
	"github.com/nsqio/go-nsq"
	"log"
	"os"
	"sync"
)

type Queues struct {
	incoming  *nsq.Consumer
	outgoing  *nsq.Producer
	kincoming *kafka.Consumer
	userDB    *kafka.Consumer
	koutgoing *kafka.Producer
}

type DBs struct {
	DLock  sync.RWMutex
	unames UserDb
	dnames UserDb
	uids   UserDb
	verify UserDb
}

type SQL struct {
	Db    *sql.DB
	SLock sync.RWMutex
}

type ModuleName struct {
	name string
}

type ModuleKeys struct {
	Key    *rsa.PrivateKey
	KeysDB keyDB
}

type userModule struct {
	app *App
	Queues
	DBs
	SQL
	ModuleName
	ModuleKeys
	commands Commands
	Kconf    *kafka.ConfigMap
}

func (m *userModule) addUser(aUser USER) {
	m.DLock.Lock()
	defer m.DLock.Unlock()

	m.unames[aUser.Uname()] = aUser
	m.dnames[aUser.DisplayName()] = aUser
	m.verify[aUser.Uname()] = aUser

}

func (m *userModule) exitsUserByName(name string) bool {
	m.DLock.RLock()
	defer m.DLock.RUnlock()

	_, ok := m.unames[name]
	return ok
}

func (m *userModule) getUserbyName(name string) (USER, error) {
	m.DLock.RLock()
	defer m.DLock.RUnlock()

	aUser, ok := m.unames[name]
	if !ok {
		return nil, Err("userModule:getUserbyName", "Unkown User")
	}
	return aUser, nil
}

func (m *userModule) getUserbyId(id string) (USER, error) {
	m.DLock.RLock()
	defer m.DLock.RUnlock()

	aUser, ok := m.uids[id]
	if !ok {
		return nil, Err("userModule:getUserbyId", "Unkown User")
	}
	return aUser, nil
}

func (m *userModule) getUserbyDisplayName(dname string) (USER, error) {
	m.DLock.RLock()
	defer m.DLock.RUnlock()

	aUser, ok := m.dnames[dname]
	if !ok {
		return nil, Err("userModule:getUserbyDisplayName", "Unkown User")
	}
	return aUser, nil
}

func (m *userModule) testUser() {
//	aUser := Person{"Payam", "Madjidi", "payam", "payam",
//		"payam@gmail.com", "Sweden", "", getNewKeys(), stamp(), 0, "simsalabim", getId("USER")}
	//PrettyPrint(aUser)


	aUser := User{"Payam","Onacci Man","abrakatabra",getNewKeysString(),stamp(),0,getId("USER")}
	aPerson := Person{&aUser,"Payam","Madjidi","payam@gmail.com","Sweden",""}
	PrettyPrint(aPerson)
	var rsa rsa.PrivateKey

	err := json.Unmarshal([]byte(aUser.Key),&rsa)
	if err != nil {
		log.Printf(err.Error())
		panic(err)
	}

	enc := encrypt(&rsa, "hello this is a test of message encryption")
	dec := decrypt(&rsa, string(enc))
	fmt.Println(dec)
	sigature := sign(&rsa, "signing this nice message")
	verifySignature(&rsa, "signing this nice message", sigature)
	m.addUser(&aPerson)
}

func (m *userModule) dispatch() {
	log.Printf("%s Start dispatching messages", m.name)
	Listen(m, m.name, m.kincoming, m.Recieve)
}

func (m *userModule) dispatchUserDb() {
	name := m.name + "-DB"
	log.Printf("%s Start dispatching messages", name)
	Listen(m, name, m.userDB, m.Recieve)
}

func (m *userModule) validateUser(u USER) bool {
	return u.Uname() != ""
}

func (m *userModule) authTrail(env *Envelope) error {
	var c Credentials
	mapstructure.Decode(env.Payload, &c)
	log.Printf("Got Credentials %+v", c)

	aUser, err := m.getUserbyName(c.Username)
	// If a password exists for the given user
	// AND, if it is the same as the password we received, the we can move ahead
	// if NOT, then we return an "Unauthorized" status
	if err != nil || aUser.Secret() != c.Password {
		log.Printf("Login failed for %s", c.Username)
		env.Command = "ATF"
	} else {
		log.Printf("Login success for %s", c.Username)
		c.Token = getId("Token")
		env.Command = "ATS"
	}
	env.Payload = c
	replayReverseAddress(env)
	env.Stamps = append(env.Stamps, stamp())
	m.Send(env)
	return nil
}


func (m *userModule) registerTrail(env *Envelope) error {
	var u USER
	mapstructure.Decode(env.Payload, &u)
	log.Printf("Got User %+v", u)
	_, err := m.getUserbyName(u.Uname())
	if err == nil  {
		log.Printf("%s: Creating new user for %s", m.name, u.Uname())
		env.Command = "REG"
		UserActivate(u)
		env.Payload = u
		env.Destination = m.Name() + "-DB"
		m.Send(env)
	} else {
		log.Printf("%s: User exists %s", m.name, u.Uname())
		replayReverseAddress(env)
		env.Command = "RTF"
		m.Send(env)
	}
	return nil
}

func (m *userModule) registerTrailSuccess(env *Envelope) error {
	return nil
}

func (m *userModule) registerTrailFail(env *Envelope) error {
	return nil
}

func (m *userModule) register(env *Envelope) error {
	var u PERSON
	err := mapstructure.Decode(env.Payload, &u)
	if err == nil {
		m.addUser(u)
		log.Printf("Got User %+v", u)
	}
	return err
}

func (m *userModule) update(env *Envelope) error {
	fmt.Println("userUpdate: not implemented")
	return nil
}

func (m *userModule) login(env *Envelope) error {
	fmt.Println("userLogin: not implemented")
	return nil
}

func (m *userModule) logout(env *Envelope) error {
	fmt.Println("userLogout: not implemented")
	return nil
}

func (m *userModule) delete(env *Envelope) error {
	fmt.Println("userDelete: not implemented")
	return nil
}

func (m *userModule) lock(env *Envelope) error {
	fmt.Println("userLock: not implemented")
	return nil
}

func (m *userModule) GetTermChannel() chan interface{} {
	return m.app.quit
}

func (m *userModule) activateUserDecoders(app *App) {
	InitCommand(m, "AT", m.authTrail)
	InitCommand(m, "RT", m.registerTrail)
	InitCommand(m, "REG", m.register)
}

func (m *userModule) deactivateUserDecoders(app *App) {
	m.commands = make(Commands)
}

func (m *userModule) Init(app *App) {

	m.app = app
	m.name = app.Name + "-USERS"
	var discoverAddr []string
	m.initQueues(app)
	m.commands = make(Commands)
	m.initDbs(app)
	discoverAddr = append(discoverAddr, app.discoverHost+":"+app.discoverPort)
	initDiscover(m.incoming, discoverAddr)
	log.Printf("Created module....%s", m.name)
	m.initSql(app)
	app.Modules[m.Name()] = m
	m.activateUserDecoders(app)
	m.testUser()
	m.Key,_ = getNewKeys()
	m.KeysDB = make(keyDB)
	m.dispatch()
	m.dispatchUserDb()
	m.publishKey()
}

func (m *userModule) initSql(app *App) {
	dbName := m.Name() + ".db"
	tableDefinitations := "CREATE TABLE IF NOT EXISTS users (id integer not null primary key, name text,sirname text,displayname text,avatar text);"
	m.Db = configureDb(dbName, tableDefinitations)
}

func (m *userModule) initQueues(app *App) {
	m.Kconf = &kafka.ConfigMap{
		"bootstrap.servers":  os.Getenv("kafkabootstrapServers"),
		"group.id":           app.Name,
		"auto.offset.reset":  os.Getenv("kafkaAutoOffsetReset"),
		"enable.auto.commit": true,
	}

	m.incoming = initIncoming(m, app.host, app.port)
	m.outgoing = initOutgoing(m, app.host, app.port)
	m.kincoming = initKafkaIncoming(m, false, m.Name())
	m.koutgoing = initKafkaOutgoing()
	m.userDB = initKafkaIncoming(m, true, m.Name() + "-DB")
}

func (m *userModule) initDbs(app *App) {
	m.unames = make(UserDb)
	m.dnames = make(UserDb)
	m.uids = make(UserDb)
	m.verify = make(UserDb)
}


func (m *userModule) GetQueueConfig() *kafka.ConfigMap {
	return m.Kconf
}

func (m *userModule) publishKey() {
	k := Keys{m.Name(), stamp(), m.Key.E, getId("KEY")}
	env, err := m.newEnvelope(k, "KEYS", "ANNOUNCE")
	if err != nil {
		log.Printf("userModule:publishKey Error creating  envelope...")
	}

	m.Send(env)
}

func (m *userModule) Send(env *Envelope)  error {
	log.Printf("userModule:Send sending envelope %s  command: %s from: %s to: %s",m.Name(),env.Command,env.Source,env.Destination)
	err := send(env, m.koutgoing)
	return err
}

func (m *userModule) GetCommands() map[string]Decoder {
	return m.commands
}

func (m *userModule) GetSendChannel() map[string]Decoder {
	return m.commands
}

func (m *userModule) Recieve(msg *kafka.Message) error {
	log.Printf("userModule:Recieve %S Got kafka message",msg.String())
	return recieve(msg, m)
}

func (m *userModule) GetRecieveChannel() *kafka.Consumer {
	return m.kincoming
}

func (m *userModule) newEnvelope(payload interface{}, dest, command string) (*Envelope, error) {
	destination := m.app.Name + "-" + dest
	return &Envelope{m.Name(), destination, command, []int64{stamp()}, getId("ENVELOP"), make(chan interface{}), payload}, nil

}

func (m *userModule) Name() string {
	return m.name
}

func (m *userModule) Commands() Commands {
	return m.commands
}

func (m *userModule) DB(db *sql.DB) {
	m.Db = db
}

func (m *userModule) DisConnect() {
	m.outgoing.Stop()
	m.incoming.Stop()
	m.kincoming.Close()
	m.koutgoing.Close()
}

func (m *userModule) Connect() {
}

/*
func (m *userModule) getKafkaOffsetForUsers(topic string,partition int32)  ([]kafka.TopicPartition,error){
	consumer := m.kincoming
	topics := []kafka.TopicPartition{{Topic: &topic, Partition: partition}}

	offset,err := consumer.Committed(topics,5000)
	if err != nil {
		log.Printf("userModule:getKafkaOffset error: %s",err.Error())
		return []kafka.TopicPartition{},err
	}
	log.Printf("userModule:getKafkaOffsetForUsers Got offset %+v",offset)
	return offset,nil
}

func (m *userModule) resetKafkaOffsetForUsers() {
	consumer := m.kincoming
	topic := m.Name() + "-DB"
	metaData,err := consumer.GetMetadata(&topic,false,6000)
	PrettyPrint(metaData)
	//var partition int32 = 0
	//var offset kafka.Offset = 0
	topics := []kafka.TopicPartition{{Topic: &topic, Partition: 0, Offset: 0}}
	tp, err := consumer.CommitOffsets(topics)
	//err := consumer.Seek(topics,6000)
	//err := consumer.Assign(topics)
	if err != nil {
		log.Printf("userModule:resetKafkaOffsetForUsers Error: %s",err.Error())
		return
	}
	log.Printf("userModule:getKafkaOffsetForUsers Resetting offset for user db...%+v",tp)
	log.Printf("userModule:getKafkaOffsetForUsers Resetting offset for user db...%+v",tp)

}


*/
