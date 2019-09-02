package main

import (
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		ErrExit("init","Dot env file (.env) is missing")
	}
}
