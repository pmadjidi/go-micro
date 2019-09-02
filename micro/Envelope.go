
package main

import (
	"github.com/jinzhu/copier"
	"log"
)

type Envelope struct {
	Source      string `json:"source"`
	Destination string  `json:"destination"`
	Command     string `json:"command"`
	Stamps      []int64 `json:"stamps"`
	Id          string  `json:"id"`
	resp        chan interface{}
	Payload     interface{} `json:"payload"`
}

func newEnvelope(payload interface{},resp bool, source, dest, command string) (*Envelope, error) {
	envId := command + "-" + getId("ENVELOP")
	env := &Envelope{source, dest, command, []int64{stamp()}, envId, nil, payload}
	if resp {
		env.resp = make(chan interface{})
	}

	//noinspection SyntaxError
	err := validateEnvelope(env)
	if err != nil {
		log.Printf(err.Error())
		return nil,err
	}
	return env,nil
}

func replicateEnvelope(env *Envelope,resp bool) *Envelope {
	newEnv  := Envelope{}
	copier.Copy(&newEnv,env)
	if resp {
		newEnv.resp = make(chan interface{})
	}
	return &newEnv
}

func validateEnvelope(env *Envelope) error {

	if env == nil {
		return Err("validateEnvelope", "Envlope is nil")
	} else if env.Source == "" {
		return Err("validateEnvelope", "No source")
	} else if env.Destination == "" {
		return Err("validateEnvelope", "No destination")
	} else if env.Source == env.Destination {
		return Err("validateEnvelope", "source and destination equal loop risk")
	} else if env.Command == "" {
		return Err("validateEnvelope", "unkown command")
	}

	return nil
}
