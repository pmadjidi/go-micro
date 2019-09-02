package main

type  ApiKey struct {
	Name  string `json:"name"`
	Key  string `json:"key"`
}

type BOT interface {
	USER
	IntegrationKeys() []ApiKey
}


type Bot struct {
	*User
	IntegrationKeys []ApiKey
}





