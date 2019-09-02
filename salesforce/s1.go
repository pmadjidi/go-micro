package main

import (
	"github.com/simpleforce/simpleforce"
	"log"
)

var (
	sfURL      = "http://login.salesforce.com"
	sfUser     = "simon@vainu.io"
	sfPassword = "Onacci123"
	sfToken    = "UmMeS14iDXikmtdLge3zzPOJ"
)

func creteClient() *simpleforce.Client {
	client := simpleforce.NewClient(sfURL, simpleforce.DefaultClientID, simpleforce.DefaultAPIVersion)
	if client == nil {
		panic("Client allocation failed")
	}
    
	err := client.LoginPassword(sfUser, sfPassword, sfToken)
	if err != nil {
		log.Printf("Auth failed with %s",err.Error())
		panic("Authentication failed")
	}
    
	// Do some other stuff with the client instance if needed.

	return client
}

func main() {
	c := creteClient()
	log.Printf("Got client %+v",c)
}
