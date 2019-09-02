package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/auth0/go-jwt-middleware"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/hashicorp/consul/api"
	"log"
	"os"
)

type App struct {
	Name             string
	Services        []string
	Prefix           string
	quit                chan interface{}
	Router              *mux.Router
	jwt                 *jwtmiddleware.JWTMiddleware
	port                string
	host                string
	discoverHost        string
	discoverPort        string
	Storage             *Storage
	Modules				Modules
	Kconf               *kafka.ConfigMap
	Reset bool
}



func newApp(AppName, port, host, discoverHost, discoverPort string,reset bool, Services []string) *App {

	prefix := fmt.Sprintf("%x", sha256.Sum256([]byte(AppName)))

	 config := &kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("kafkabootstrapServers"),
		//"group.id":          os.Getenv("kafkaGroupId"),
		"group.id":     RandStringRunes(10),
		"auto.offset.reset": os.Getenv("kafkaAutoOffsetReset"),
	}

	newApp := &App{
		Name:             AppName,
		Services:            Services,
		Prefix:           prefix,
		port:                port,
		host:                host,
		discoverHost:        discoverHost,
		discoverPort:        discoverPort,
		quit:                make(chan interface{}),
		Router:              mux.NewRouter(),
		jwt: jwtmiddleware.New(jwtmiddleware.Options{
			ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
				return []byte("My Secret"), nil
			},
			// When set, the middleware verifies that tokens are signed with the specific signing algorithm
			// If the signing method is not constant the ValidationKeyGetter callback can be used to implement additional checks
			// Important to avoid security issues described here: https://auth0.com/blog/2015/03/31/critical-vulnerabilities-in-json-web-token-libraries/
			SigningMethod: jwt.SigningMethodHS256,
		}),
		Storage:  newStorage(AppName),
		Modules:  make(Modules),
		Kconf: config,
		Reset: reset,
	}

	initModules(newApp)
	registerApp(newApp)
	PrettyPrint(newApp)
	return newApp
}


func initModules(app *App) {
	log.Printf("Init modules for App %s....",app.Name)
	log.Printf("Modules are %+v....",app.Services)
	for _, moduleName := range app.Services {
			moduleFactory(app,moduleName)
	}
}



func stop(app *App,err error, source string) {
	for _, mod := range app.Modules {
		fmt.Printf("Stop called from:%s\n", source)
		mod.DisConnect()
	}
}



func registerApp(app *App) {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		panic(err)
	}

	// Get a handle to the KV API
	kv := client.KV()

	// PUT a new KV pair
	name := &api.KVPair{Key: app.Name, Value: []byte(app.Prefix)}
	_, err = kv.Put(name, nil)
	if err != nil {
		panic(err)
	}

	// Lookup the pair
	pair, _, err := kv.Get(app.Name, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("KV: %v %s\n", pair.Key, pair.Value)
}



