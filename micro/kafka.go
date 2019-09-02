package main

import (
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"os"
)

func initKafkaIncoming(m QueueModule, reset bool, topicNames ...string) *kafka.Consumer {
	log.Printf("initKafkaIncoming: kafkaBootstrapServers is set to %s", os.Getenv("kafkabootstrapServers"))

	cfg := m.GetQueueConfig()
	if reset {
		randomGroupId := RandStringRunes(10)
		cfg.SetKey("group.id", randomGroupId)
		log.Printf("Using this kafka configuration...")
		PrettyPrint(cfg)
	}

	c, err := kafka.NewConsumer(cfg)

	if err != nil {
		ErrExit("initKafkaIncoming", err.Error())
	}

	log.Printf("Connected to Kafka incomming...." + os.Getenv("kafkabootstrapServers"))

	log.Printf("=== initKafkaIncoming: %s  Subscribing to topic/topics %+v....", m.Name(), topicNames)

	c.SubscribeTopics(topicNames, nil)
	//func(*Consumer, Event) error
	return c

}

func initKafkaOutgoing() *kafka.Producer {
	log.Printf("initKafkaOutgoing: kafkaBootstrapServers is set to %s", os.Getenv("kafkabootstrapServers"))
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": os.Getenv("kafkabootstrapServers"), "debug": "broker"})
	if err != nil {
		ErrExit("initKafkaOutgoing", err.Error())
	}

	log.Printf("Connecting to Kafka outgoing...." + os.Getenv("kafkabootstrapServers"))
	go func() {
		for e := range p.Events() {
			//fmt.Printf("Go event %+v\n",e)
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
			fmt.Printf("%+v", e)
		}
	}()

	return p
}

func pull(name string, consumer *kafka.Consumer, dispatch func(msg *kafka.Message) (error)) bool {
	ev := consumer.Poll(600)
	if ev == nil {
		return true
	}
	switch e := ev.(type) {
	case *kafka.Message:
		log.Printf("%s received kafka event %s", name, e.Key)
		dispatch(e)
		return true
	case kafka.Error:
		fmt.Fprintf(os.Stderr, "%% Error: %v: %v\n", e.Code(), e)
		if e.Code() == kafka.ErrAllBrokersDown {
			return false
		}
	default:
		fmt.Printf("Ignored %v\n", e)
		return true
	}
	return false
}

func send(env *Envelope, producer *kafka.Producer) error {
	key := env.Command + "-" + env.Id
	log.Printf("Sending message to topic %s command: %s", env.Destination, key)
	env.Stamps = append(env.Stamps, stamp())
	val, e := json.Marshal(env)
	if e != nil {
		return Err("frontModule:send", "Error mashaling Envelope:"+env.Id)
	}
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &env.Destination, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          val,
	}

	e = producer.Produce(msg, nil)
	if e != nil {
		log.Printf("Error sending payload to %s topic on kafka", env.Destination)
		return e
	}
	return nil
}

func recieve(msg *kafka.Message, m QueueModule) error {

	var env Envelope

	fmt.Printf("### recieve Message on %s:\n%s\n",
		msg.TopicPartition, string(msg.Value))
	if msg.Headers != nil {
		fmt.Printf("### recieve Headers: %v\n", msg.Headers)
	}

	err := json.Unmarshal(msg.Value, &env)
	if err != nil {
		errMsg := m.Name() + ": Can not decode envelope Key: " + string(msg.Key) + " Value: " + string(msg.Value)
		return Err("recieve", errMsg)
	}

	log.Printf("%s recieve decoded envelope %s", m.Name(), env.Id)

	f, ok := m.GetCommands()[env.Command]
	if !ok {
		return Err(m.Name()+" recieve:", "unkown Command: "+env.Command)
	}

	err = f(&env)
	if err != nil {
		log.Printf("%s Error recive %s", m.Name(), err.Error())
		return err
	}

	return nil
}
