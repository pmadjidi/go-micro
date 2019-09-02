package main

import (
	"fmt"
	"log"
	"time"
)

func Err(funcName,errorMsg string) error {
	log.Printf("!!! Error in %s: %s",funcName,errorMsg)
	return fmt.Errorf("%s Error in function %s: %s\n",time.Now(),funcName,errorMsg)
}

func ErrExit(funcName,errorMsg string) {
	fmt.Printf("%s Error in function %s: %s\n",time.Now(),funcName,errorMsg)
	log.Fatal("Exiting.....")
}