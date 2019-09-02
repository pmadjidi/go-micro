package main

import (
	"crypto/rsa"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/mitchellh/mapstructure"
	"log"
	"os"
	"sync"
	"github.com/nsqio/go-nsq"
	"time"
)

type messageModule struct {
	app       *App
	name      string
	Db        *sql.DB
	Lock      sync.RWMutex
	unames    UserDb
	dnames    UserDb
	uids      UserDb
	commands  Commands
	incoming  *nsq.Consumer
	outgoing  *nsq.Producer
	kincoming *kafka.Consumer
	koutgoing *kafka.Producer
	Key       *rsa.PrivateKey
	KeysDB    keyDB
	Kconf               *kafka.ConfigMap
}

func (m *messageModule) publishKey() error {
	k := Keys{m.name, stamp(), m.Key.E,getId("KEY")}
	env, err := m.newEnvelope(k, "KEYS", "ANNOUNCE")
	if err != nil {
		log.Printf("userModule:publishKey Error creating  envelope...")
		return err
	}

	return  m.send(env)
}

func (m *messageModule) send(env *Envelope) error {
	err := send(env,m.koutgoing)
	return err
}

func (m *messageModule) getCommands() map[string]Decoder {
	return m.commands
}


func (m *messageModule) recieve(msg *kafka.Message) {


	fmt.Printf("%% Message on %s:\n%s\n",
		msg.TopicPartition, string(msg.Value))
	if msg.Headers != nil {
		fmt.Printf("%% Headers: %v\n", msg.Headers)
	}

	key := string(msg.Key)
	var env Envelope
	e := json.Unmarshal(msg.Value,&env)
	if e != nil {
		log.Printf("Can not decode envelope %s %s ",msg.Key,msg.Value)
		return
	}


	log.Printf("userModule:recieve decoded envelope %+v",env)

	if key  == "ANNOUNCE" {
		var k Keys
		mapstructure.Decode(env.Payload, &k)
		log.Printf("Got Keys %+v",k)

	}
}

func (m *messageModule) newEnvelope(payload interface{},dest,command string) (*Envelope,error) {
destination := m.app.Name + "-" + dest
return &Envelope{m.name,destination,command,[]int64{stamp()},getId("ENVELOP"),make(chan interface{}),payload},nil
}

func (m *messageModule) dispatch() {
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		run := true
		for run == true {
			select {
			case sig := <- m.app.quit:
				fmt.Printf("Caught signal %v: terminating\n", sig)
				run = false
			case <-ticker.C :
				log.Printf("%s tick tack...",m.name)
			default:
				ev := m.kincoming.Poll(100)
				if ev == nil {
					continue
				}
				switch e := ev.(type) {
				case *kafka.Message:
					log.Printf("got event %s",e.Key)
					m.recieve(e)
				case kafka.Error:
					fmt.Fprintf(os.Stderr, "%% Error: %v: %v\n", e.Code(), e)
					if e.Code() == kafka.ErrAllBrokersDown {
						run = false
					}
				default:
					fmt.Printf("Ignored %v\n", e)
				}
			}
		}
		log.Printf(m.name + " Aj Aj Exiting listener loop...")
	}()
}




