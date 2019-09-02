package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)


type Keys struct{
	serviceName string `json:"servicename"`
	stamp int64 `json:"stamp"`
	key int `json:"key"`
	Id string `json:"id"`
}

type keyDB map[string]*Keys


//*rsa.PrivateKey gets converted to string...
func getNewKeys() (*rsa.PrivateKey,error) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		log.Printf(err.Error())
		return nil,Err("getNewKeys","Key generation failed")
	}


	return key,nil
}


func getNewKeysString() (string) {
	key,err  := getNewKeys()
	if err != nil {
		return ""
	}

	skey, err := json.Marshal(key)
	if err != nil {
		Err("getNewKeys","Key mashalling failed")
		return ""
	}

	return string(skey)
}



func encrypt(keys *rsa.PrivateKey,msg string) []byte {
	publickey := &keys.PublicKey
	label := []byte("")
	sha1hash := sha1.New()
	encryptedmsg, err := rsa.EncryptOAEP(sha1hash, rand.Reader, publickey, []byte(msg), label)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Encrypted [%s] to \n[%x]\n", string(msg), encryptedmsg)
	fmt.Println()
	return encryptedmsg
}

func decrypt(keys *rsa.PrivateKey,secretMsg string)  []byte{
	label := []byte("")
	sha1hash := sha1.New()
	decryptedmsg, err := rsa.DecryptOAEP(sha1hash, rand.Reader, keys,[]byte(secretMsg), label)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Decrypted [%x] to \n[%s]\n", string(decryptedmsg), decryptedmsg)
	return decryptedmsg
}

func sign(keys *rsa.PrivateKey,msg string)  []byte{
	var h crypto.Hash
	hash := sha1.New()
	io.WriteString(hash, msg)
	hashed := hash.Sum(nil)
	signature, err := rsa.SignPKCS1v15(rand.Reader, keys, h, hashed)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("PKCS1v15 Signature : %x\n", signature)
	return signature
}

func verifySignature(keys *rsa.PrivateKey,msg string,signature []byte) bool{
	var h crypto.Hash
	publickey := &keys.PublicKey
	hash := sha1.New()
	io.WriteString(hash, msg)
	hashed := hash.Sum(nil)
	err := rsa.VerifyPKCS1v15(publickey, h, hashed, signature)
	if err != nil {
		fmt.Println("VerifyPKCS1v15 failed")
		return false
	} else {
		fmt.Println("VerifyPKCS1v15 successful")
		return true
	}
}