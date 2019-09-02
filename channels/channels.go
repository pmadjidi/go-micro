package main

import (
	"fmt"
	"github.com/hashicorp/consul/api"
)


var CHANNELS = make(map[string]string)

func userCommands() {
	C := "USERS"
	CHANNELS["AUTH"] = C
	CHANNELS["UPDATE"] = C
	CHANNELS["LOGIN"] = C
	CHANNELS["LOGOUT"] = C
	CHANNELS["UREGISTER"] = C
	CHANNELS["LOCK"] = C
}



func registerWithConsule() {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		panic(err)
	}

	// Get a handle to the KV API
	kv := client.KV()

	// PUT a new KV pair
	for k,v  := range CHANNELS {
		channel := &api.KVPair{Key: k, Value: []byte(v)}
		_, err = kv.Put(channel, nil)
		if err != nil {
			panic(err)
		}
	}

	for k,v  := range CHANNELS {
		pair, _, err := kv.Get(k, nil)
		if err != nil || v != string(pair.Value) {
			panic(err)
		}
		fmt.Printf("KV: %v %s\n", pair.Key, pair.Value)

	}

}

func main() {
	userCommands()
	registerWithConsule()
}

