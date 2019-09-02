package main

import (
	"encoding/json"
	"fmt"
	"log"
	"github.com/nsqio/go-nsq"
	"strings"
)

func initDiscover(consumers *nsq.Consumer, discoverHosts []string) {
	fmt.Print("Initializing nsqlookupd ....")

	if err := consumers.ConnectToNSQLookupds(discoverHosts); err != nil {
		fmt.Print("OBS,ConnectToNSQLookupds failed... Exiting....")
		ErrExit("initDiscover", err.Error())
	}
}

func initTopics(m Module,topics []string, host, port string, prefix string) []*nsq.Consumer {
	config := nsq.NewConfig()
	id := getId("MSG")
	fmt.Printf("Id: %s\n", id)
	fmt.Printf("Topics....: %+v\n", topics)
	var subscriptions [] *nsq.Consumer
	for topic := range topics {
		tp := topics[topic]
		fmt.Printf("subscribing to topic....: %s\n", tp)
		q, _ := nsq.NewConsumer(tp, id, config)
		q.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
			err := decodeNsq(m,message)
			return err
		}))
		err := q.ConnectToNSQD(host + ":" + port)
		if err != nil {
			log.Fatal("Could not connect.....\nExiting.....", err)
		}
		subscriptions = append(subscriptions, q)
	}
	return subscriptions
}

func initIncoming(m Module,host,port string) *nsq.Consumer {
	config := nsq.NewConfig()
	topic := m.Name()
	log.Printf("Connecting incomming channel to NSQ....%s\n", topic)
	q, err := nsq.NewConsumer(topic, getId(m.Name()), config)
	if err != nil {
		log.Fatal(err)
	}
	q.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		err := decodeNsq(m,message)
		return err
	}))
	err = q.ConnectToNSQD(host + ":" + port)
	if err != nil {
		log.Fatal("Could not connect.....\nExiting.....", err)
	}
	return q
}

func initOutgoing(m Module,host,port string) *nsq.Producer {
	config := nsq.NewConfig()
	log.Printf("Connecting outgoing channel to NSQ....%s\n",m.Name())
	p, err := nsq.NewProducer(host+":"+ port, config)
	if err != nil {
		ErrExit("initOutgoing", m.Name() + " error creating nsq producer")
	}
	return p
}

func decodeNsq(m Module,message *nsq.Message) error {
	var env Envelope
	if err := json.Unmarshal(message.Body, &env); err != nil {
		return err
	}
	fmt.Printf("Recieved message from NSQ: %+v\n", env)
	fmt.Printf("Message size is: %+v\n", len(message.Body))
	return decode(m,&env)
}


func decode(m Module,env *Envelope) error {
	fmt.Printf("Decoding message: %+v\n",env)
	Decoder := m.GetCommands()[strings.ToUpper(env.Command)]
	if (Decoder) == nil {
		return Err("decodeNsq","Decoder for this Instruction is missing: " + env.Command)
	}
	return Decoder(env)
}

func decodeKafka() error {
	log.Print("decodeKafka not implemented....")
	return nil
}



