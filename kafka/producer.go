package main

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"net"
	"strings"
)

func Err(e error) {
	if e != nil {
		fmt.Print(e)
	}
}

func ip() {
	interfaces, err := net.Interfaces()

	if err != nil {
		fmt.Print(err)
		return
	}

	for _, i := range interfaces {

		fmt.Printf("Name : %v \n", i.Name)

		byNameInterface, err := net.InterfaceByName(i.Name)

		if err != nil {
			fmt.Println(err)
		}

		//fmt.Println("Interface by Name : ", byNameInterface)

		addresses, err := byNameInterface.Addrs()

		for k, v := range addresses {

			fmt.Printf("Interface Address #%v : %v\n", k, v.String())
		}
		fmt.Println("------------------------------------")

	}
}

func ipByName(iname string) string {
	interfaces, err := net.Interfaces()

	if err != nil {
		fmt.Print(err)
		return ""
	}

	for _, i := range interfaces {

		//fmt.Printf("Name : %v \n", i.Name)

		byNameInterface, err := net.InterfaceByName(i.Name)

		if err != nil {
			fmt.Println(err)
		}

		if byNameInterface.Name == iname {

			//fmt.Println("Interface by Name : ", byNameInterface)

			addresses, err := byNameInterface.Addrs()
			Err(err)
			for _, v := range addresses {
				parts := strings.Split(v.String(), "/")
				IPAddress := net.ParseIP(parts[0])
				if IPAddress.To4() != nil {
					return parts[0]
				}
			}
		}
	}
	return ""
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func main() {

	ip := ipByName("en0") + ":9092"
	fmt.Printf("IP is set to %s \n",ip)
	//p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "130.229.155.151:9092","debug": "broker"})
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": ip,"debug": "broker"})
	if err != nil {
		panic(err)
	}

	defer p.Close()

	// Delivery report handler for produced messages
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

	// Produce messages to topic (asynchronously)
	topic := "test-USERS"
	for _, word := range []string{"Welcome", "to", "the", "Confluent", "Kafka", "Golang", "client"} {
		err = p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          []byte(word),
		}, nil)
		if err != nil {
			fmt.Printf("Obs Error: %s\n", err)
		} else {
			fmt.Printf("The world is: %s  ", word)
		}
	}

	// Wait for message deliveries before shutting down
	remain := p.Flush(15 * 1000)
	fmt.Printf("\nNumber of unsent messages is %d\n", remain)
}
