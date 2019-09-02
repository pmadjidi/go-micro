package main

import (
	"crypto/rsa"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	"github.com/nsqio/go-nsq"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
	"golang.org/x/crypto/acme/autocert"
)

type webData struct {
	WsAddr string
	Action string
	Title string
}

func (f *frontModule) StartListenToEndPoint(app *App) {
	log.Printf("Starting to listen to port 8080")
	prefix := "/" + app.Name
	mux := http.NewServeMux()
	mux.HandleFunc(prefix + "/echo", f.echo)
	mux.HandleFunc(prefix + "/login", f.loginPage)
	mux.HandleFunc(prefix +"/auth", f.login)
	mux.HandleFunc(prefix +"/slackSalesForce", f.registerSlackSalesForceIngegrationPage)
	mux.HandleFunc(prefix +"/regslacksforce", f.registerSlackSalesForceIngegrationPage)
	mux.HandleFunc(prefix +"/registerUser", f.registerUserPage)
	mux.HandleFunc(prefix +"/reguser", f.registerPerson)
	mux.HandleFunc(prefix +"/", f.home)
	f.startWebDevelopment(mux)
}

func (f *frontModule) startWebDevelopment(mux *http.ServeMux) {

	log.Fatal(http.ListenAndServe(":8080", mux))
}

func (f *frontModule) startWebProduction(mux *http.ServeMux) {
	certManager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache("certs"),
	}

	server := &http.Server{
		Addr:    ":443",
		Handler: mux,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}

	go http.ListenAndServe(":80", certManager.HTTPHandler(nil))
	server.ListenAndServeTLS("", "")

}


func (f *frontModule) home(w http.ResponseWriter, r *http.Request) {

	wsAddr := "ws://" + r.Host + "/" + f.app.Name + "/echo"
	d := webData {wsAddr,"","Echo"}
	log.Printf("Listening to addrs: %s\n", wsAddr)
	err := homeTemplate.Execute(w, d)
	if err != nil {
		log.Printf("Error executing template homeTemplate %s\n", err.Error())
	}
}

func (f *frontModule) registerUserPage(w http.ResponseWriter, r *http.Request) {
	Addr := "http://" + r.Host + "/" + f.app.Name + "/reguser"
	d := webData {"",Addr,"User registration page"}
	log.Printf("Listening to addrs: %s\n", Addr)
	err := registerPersonTemplate.Execute(w, d)
	if err != nil {
		log.Printf("Error executing template registerTemplate %s\n", err.Error())
	}
}

func (f *frontModule) registerSlackSalesForceIngegrationPage(w http.ResponseWriter, r *http.Request) {
	Addr := "http://" + r.Host + "/" + f.app.Name + "/reg"
	action :=  "/" + f.app.Prefix + "/regslacksforce"
	d := webData {Addr,action,"Register Slack & SalesForce integration page"}
	log.Printf("Listening to addrs: %s\n", Addr)
	err := regSlackSalesForceIntegrationTemplate.Execute(w, d)
	if err != nil {
		log.Printf("Error executing template registerTemplate %s\n", err.Error())
	}
}

func (f *frontModule) loginPage(w http.ResponseWriter, r *http.Request) {
	wsAddr := "http://" + r.Host +  "/" + f.app.Name + "/auth"
	log.Printf("Listening to addrs: %s\n", wsAddr)
	err := passwordTemplate.Execute(w, wsAddr)
	if err != nil {
		log.Printf("Error executing template passwordTemplate\n")
	}
}

func (f *frontModule) echo(w http.ResponseWriter, r *http.Request) {
	wsAddr := "http://" + r.Host + "/" + f.app.Name + "/auth"
	t, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			log.Printf("frontModule,echo: session_token missing, StatusUnauthorized")
			http.Redirect(w, r, wsAddr, http.StatusTemporaryRedirect)
			return
		}
		// For any other type of error, return a bad request status
		log.Printf("frontModule,echo: error: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionToken := t.Value

	// We then get the name of the user from our cache, where we set the session token
	response, ok := f.Sessions[sessionToken]
	if !ok {
		http.Redirect(w, r, "http://localhost/login", http.StatusBadRequest)
		// If there is an error fetching from cache, return an internal server error status
		//w.WriteHeader(http.StatusUnauthorized)
		//	w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("User %s connected...", response)

	var upgrader = websocket.Upgrader{}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		err = f.decodeWebSocket(message)
		if err != nil {
			log.Println("Decode:", err)
			c.WriteMessage(mt, []byte(err.Error()))
		} else {
			err = c.WriteMessage(mt, message)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}
}


func (f *frontModule) regResponse(env *Envelope) error {
	req := f.getRequestFromcache(env)
	if req != nil {
		f.deleteRequest(env)
		var u User
		mapstructure.Decode(env.Payload, &u)
		log.Printf("frontModule:REG decodedd user to: %+v", u)
		req.resp <- u
		return nil
	}
	return Err("frontModule:regResponse","Request not found, allready processed?")
}

func (f *frontModule) authResponse(env *Envelope) error {
	req := f.getRequestFromcache(env)
	if req != nil {
		f.deleteRequest(env)
		var c Credentials
		mapstructure.Decode(env.Payload, &c)
		log.Printf("frontModule:REG decodedd credentials to: %+v", c)
		req.resp <- c
		return nil
	}
	return Err("frontModule:authResponse","Request not found, allready processed?")
}

func (f *frontModule) lock(env *Envelope) error {
	fmt.Println("userLock: not implemented")
	return nil
}



func (f *frontModule) activateUserDecoders(app *App) {
	InitCommand(f, "REG", f.regResponse)
	InitCommand(f, "RS", f.regResponse)
	InitCommand(f, "AF", f.authResponse)
	InitCommand(f, "AS", f.authResponse)
}


func (f *frontModule) GetCommands() map[string]Decoder {
	return f.commands
}

func (f *frontModule) newEnvelope(payload interface{},dest,command string) (*Envelope,error) {
	destination := f.app.Name + "-" + dest
	env,err :=  newEnvelope(payload,true,f.name,destination,command)
	//return &Envelope{f.name,destination,command,[]int64{stamp()},getId("ENVELOP"),make(chan interface{}),payload},nil
	return env,err

}


func (f *frontModule) publishKey()  {
	k := Keys{f.name,stamp(),f.Key.E,getId("KEY")}
	env ,err := f.newEnvelope(k,"KEYS","ANNOUNCE")
	if err != nil {
		log.Printf("userModule:publishKey Error creating  envelope...")
	}

	f.Send(env)
}

type Sessions map[string]*Credentials
type Requests map[string]*Envelope

type frontModule struct {
	app          *App
	name         string
	Db           *sql.DB
	commands     Commands
	users        UserDb
	Sessions     Sessions
	RLock        sync.RWMutex
	Requests     Requests
	incoming     *nsq.Consumer
	outgoing     *nsq.Producer
	kincoming    *kafka.Consumer
	koutgoing    *kafka.Producer
	Key    *rsa.PrivateKey
	KeysDB  keyDB
	Kconf               *kafka.ConfigMap
}

func (f *frontModule) Name() string {
	return f.name
}

func (f *frontModule) Commands() Commands {
	return f.commands
}

func (f *frontModule) DB(db *sql.DB) {
}

func (f *frontModule) GetQueueConfig()   *kafka.ConfigMap {
	return f.Kconf
}

func (f *frontModule) AddUser(aUser USER) bool {
	_, ok := f.users[aUser.Uname()]
	if ok {
		log.Printf("User %s allready in database", aUser.Uname())
		return false
	}

	f.users[aUser.Uname()] = aUser
	log.Printf("Number of users: %n",len(f.users))
	return true

}

func (f frontModule) DisConnect() {
	f.outgoing.Stop()
	f.incoming.Stop()
	f.kincoming.Close()
	f.koutgoing.Close()
}

func (f *frontModule) Connect() {
}



func (f *frontModule) cacheRequest(env *Envelope) {
	f.RLock.Lock()
	f.Requests[env.Id] = env
	f.RLock.Unlock()
}

func (f *frontModule) getRequestFromcache(env *Envelope) *Envelope {
	f.RLock.RLock()
	c, ok := f.Requests[env.Id]
	f.RLock.RUnlock()
	if !ok {
		log.Printf("Requesting id: %s is missing, mabye time out.....",env.Id)
		return nil
	}
	return c
}

func (f *frontModule) deleteRequest(env *Envelope) bool {
	if env == nil {
		return false
	}
	_, ok := f.Requests[env.Id]
	if !ok {
		fmt.Printf("Requesting id: %s is missing, mabye time out.....", env.Id)
		return false
	} else {
		delete(f.Requests, env.Id)
		return true
	}
}

func (f *frontModule) dispatch() {
	Listen(f,f.name,f.kincoming,f.Recieve)
}


func (f *frontModule) Poll(d int) kafka.Event{
	return f.kincoming.Poll(d)
}


func (f *frontModule) initQeues(app *App){
	var discoverAddr []string
	f.incoming = initIncoming(f, app.host,app.port)
	f.outgoing = initOutgoing(f,app.host,app.port)
	f.kincoming = initKafkaIncoming(f,false,f.name,app.Name +"-USERS-DB")
	f.koutgoing = initKafkaOutgoing()
	discoverAddr = append(discoverAddr, app.discoverHost+":"+app.discoverPort)
	initDiscover(f.incoming, discoverAddr)
}

func (f *frontModule) initDbs(app *App) {
	f.Sessions = make(Sessions)
	f.Requests = make(Requests)
	f.commands = make(Commands)
	f.KeysDB = make(keyDB)
}



func (f *frontModule) initSQL(app *App) {
	dbName := f.name + ".db"
	tableDefinitations := ""
	f.Db = configureDb(dbName, tableDefinitations)
}


func (f *frontModule) initKeys(app *App) error {
	key,err  := getNewKeys()
	if err != nil {
		return Err("frontModule","Key generation faild")
	}
	f.Key = key
	return nil
}





func (f *frontModule) Init(app *App) {
	log.Printf("Created module....%s", f.name)
	f.Kconf= &kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("kafkabootstrapServers"),
		"group.id":    app.Name,
		"auto.offset.reset": os.Getenv("kafkaAutoOffsetReset"),
	}
	f.name = app.Name + "-FORNT"
	f.app = app
	f.initQeues(app)
	f.users = make(UserDb)
	f.initDbs(app)
	f.initSQL(app)
	f.initKeys(app)
	app.Modules[f.name] = f
	f.activateUserDecoders(app)
	f.dispatch()
	f.StartListenToEndPoint(app)
}

func (f *frontModule) decodeWebSocket(m []byte) error {
	var env Envelope
	if err := json.Unmarshal(m, &env); err != nil {
		return err
	}
	fmt.Printf("Recieved message from WebSocket: %+v\n", env)
	fmt.Printf("Message size is: %+v\n", len(m))
	//return decode(&env,f.app)
	return nil
}

type Credentials struct {
	Username string `json:"Username"`
	Password string `json:"Password"`
	Token  string `json:"Token"`
}

func (f *frontModule) GetRecieveChannel() *kafka.Consumer {
	return f.kincoming
}

func (f *frontModule) GetTermChannel() chan interface{} {
	return f.app.quit
}


func (f *frontModule) newCredentials(uname,password string) *Credentials {
	return &Credentials{uname,password,""}
}

func (f *frontModule) login(w http.ResponseWriter, r *http.Request) {
	/* Get the JSON body and decode into credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		log.Printf("frontModule.Signin:  Can not decode creds %+v",creds)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	*/
	e := r.ParseForm()
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}


	c := f.newCredentials(r.FormValue("uname"),r.FormValue("password"))

	log.Printf("%+v", c)
	// Get the expected password from our in memory map

	env,e := f.newEnvelope(c, "USERS","AT")
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("login envelope %+v", env)


	//f.Out <- env

	f.Send(env)


	ticker := time.NewTicker(5 * time.Second)


	var ans Credentials

	select {
	case d := <- env.resp:
		ans = d.(Credentials)
	case <-ticker.C:
		w.Write([]byte("TimeOut..."))
		return
	}

	if ans.Token != "" {
		f.Sessions[c.Token] = c
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   c.Token,
			Expires: time.Now().Add(120 * time.Second),
		})
		w.Write([]byte("LOGIN SUCCESS..."))
		return
	} else {
		//w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("LOGIN FAILED..."))
		return
	}

}

func (f *frontModule) registerPerson(w http.ResponseWriter, r *http.Request) {
	e := r.ParseForm()
	if e != nil {
		panic(e)

	}
	log.Printf("Form is: %+v", r.PostForm)
	var aUser Person
	aUser.Name = r.FormValue("uname")
	aUser.Password = r.FormValue("password")
	aUser.Displayname = r.FormValue("dname")
	aUser.Sirname = r.FormValue("sname")
	aUser.Givenname = r.FormValue("gname")
	aUser.Avatar = r.FormValue("avatar")
	aUser.Email = r.FormValue("email")
	aUser.Flag = 0


	env,e := f.newEnvelope(aUser,"USERS","RT")
	if e != nil {
		log.Printf(e.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}


	e = f.Send(env)
	if e != nil {
		log.Printf(e.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ticker := time.NewTicker(10 * time.Second)
	var ans Person


	select {
	case d := <- env.resp:
		ans = d.(Person)
	case <-ticker.C:
		w.Write([]byte("TimeOut..."))
		return
	}

	if HasFlag(&ans,ACTIVE) {
		w.Write([]byte("Registered thanx....."))
	} else {
		w.Write([]byte("Failed to register Already exists....."))
	}
}

func (f *frontModule) Send(env *Envelope)  error {
	if env == nil {
		return Err("frontModule:Send","Envelope is nil")
	}
	err := send(env, f.koutgoing)
	if err == nil {
		f.cacheRequest(env)
		return nil
	}

	return Err("frontModule:Send",err.Error())
}

func (f *frontModule) GetSendChannel() map[string]Decoder {
	return f.commands
}



func (f *frontModule) Recieve(msg *kafka.Message) error {
	return recieve(msg,f)
}



