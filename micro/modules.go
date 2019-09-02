package main

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"strings"
)


type QueueModule interface {
	Module
	GetQueueConfig() *kafka.ConfigMap
	Send(envelope *Envelope)  error
	GetRecieveChannel()  *kafka.Consumer
	Recieve(m *kafka.Message) error
}

type Module interface {
	Init(app *App)
	Name() string
	DisConnect()
	Connect()
	GetCommands() map[string]Decoder
	GetTermChannel() chan interface{}
}

func moduleFactory(app *App,moduleName string)  {
	switch  strings.ToUpper(moduleName){
	case "USERS":
		log.Printf("Creating userModule")
		u := new(userModule)
		u.Init(app)
	case "FRONT":
		log.Printf("Creating frontModule")
		f := new(frontModule)
		f.Init(app)
	default:
		ErrExit("initModules", "Unkown module: " + moduleName)
	}
}

func Listen(m QueueModule,topic string, Consumer *kafka.Consumer,dispatch func(m *kafka.Message) error) {
	log.Printf("Module %s Listens for topic for topic %s...",m.Name(),topic)
	counter := 0
	go func() {
		run := true
		for run == true {
			select {
			case sig := <- m.GetTermChannel():
				fmt.Printf("Caught signal %v: terminating\n", sig)
				run = false
			default:
				log.Printf("%s dispatch pulling topic %s....",m.Name(),topic)
				run = pull(topic,Consumer,dispatch)
			}
			counter += 1
			if counter % 100 == 0 {
				log.Printf("Pulling %s %d times", m.Name(), counter)
			}
		}
		log.Printf(m.Name() + " Aj Aj Exiting listener loop...")
	}()
}


func InitCommand(m Module,name string,cmd Decoder)  {
	_,ok := m.GetCommands()[name]
	if !ok {
		m.GetCommands()[name] = cmd
	} else {
		ErrExit("InitCommand","Command confilict")
	}
}



/*
func registerWithConsule(m Module) {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		panic(err)
	}

	// Get a handle to the KV API
	kv := client.KV()

	// PUT a new KV pair
	for k,_  := range m.Commands() {
		channel := &api.KVPair{Key: k, Value: []byte(m.Name())}
		_, err = kv.Put(channel, nil)
		if err != nil {
			panic(err)
		}
	}

	for k,_  := range m.Commands() {
		pair, _, err := kv.Get(k, nil)
		if err != nil || k != string(pair.Value) {
			panic(err)
		}
		fmt.Printf("KV: %v %s\n", pair.Key, pair.Value)

	}

}

 */
