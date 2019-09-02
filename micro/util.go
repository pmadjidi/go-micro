package main

import (
	"encoding/json"
	"fmt"
	"github.com/rs/xid"
	"math/rand"
	"regexp"
	"time"
)

func PrettyPrint(v interface{}) (err error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
	return
}


func getId(typ string) string {
	return xid.New().String() + "-"+ typ
}

func stamp() int64 {
	return time.Now().UnixNano()
}

func validateEmail(email string) bool {
	Re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return Re.MatchString(email)
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}


func replayReverseAddress(env *Envelope) {
	dest := env.Destination
	env.Destination = env.Source
	env.Source = dest
}

