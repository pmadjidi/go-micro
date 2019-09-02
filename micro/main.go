package main

import (
	"flag"
	"log"
)



func main() {
	var appName,serviceName, port, host,dhost,dport,reset string
	flag.StringVar(&appName, "app", "", "name of the app...")
	flag.StringVar(&port, "port", "4150", "port of the event bus...")
	flag.StringVar(&host, "host", "localhost", "hostname of the bus...")
	flag.StringVar(&dport, "dport", "4161", "port of the nsqlookupd...")
	flag.StringVar(&dhost, "dhost", "localhost", "hostname of nsqlookupd...")
	flag.StringVar(&serviceName, "service", "", "name of the microservice...")
	flag.StringVar(&reset, "reset", "false", "reset the microservice...")
	flag.Parse()
	if appName == "" {
		log.Fatal("App name, -app flag is mandatory....")
	}
	if serviceName == "" {
		log.Fatal("Service name, -service flag is mandatory....")
	}
	services := flag.Args()
	services = append(services, serviceName)
	log.Printf("Configuring the app for modules %+v....",services)
	var Reset = true
	if reset == "false" {
		Reset = false
	}
	app := newApp(appName, port, host, dhost,dport,Reset,services)
	catchTermSignal(app)
	<-app.quit

}
