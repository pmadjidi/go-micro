package main

import (
	"bytes"
	"fmt"
	"github.com/minio/minio-go"
	"log"
	"os"

)

var storage *Storage

type Storage struct {
	EndPoint string
	EndpointNum string
	EndpointPort string
	AccessKeyID string
	SecretAccessKey string
	Location string
	BucketName string
	SSL bool
	Client *minio.Client
}

func newStorage(bucketName string) *Storage {
	log.Printf("Created storage for %s\n", bucketName)

	endpointNum := os.Getenv("endpointNum")
	endpointHost := os.Getenv("endpointHost")
	endpointPort := os.Getenv("endpointPort")
	accessKeyID := os.Getenv("accessKeyID")
	secretAccessKey := os.Getenv("secretAccessKey")
	SSL :=  os.Getenv("useSSL")
	geoLocation := os.Getenv("location")
	useSSL := true
	if SSL== "" || SSL=="false" {
		useSSL = false
	}

	minioClient, err := minio.New(endpointHost,accessKeyID , secretAccessKey, useSSL)

	if err != nil {
		log.Fatal(err)
	}

	err = minioClient.MakeBucket(bucketName, geoLocation)

	if err != nil {
		log.Printf("Error creating bucket %s, %s\n", bucketName,err)
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, err := minioClient.BucketExists(bucketName)
		if err == nil && exists {
			log.Printf("We already own storage for %s\n", bucketName)
		} else {
			fmt.Printf("Panicing.....%s, %s\n",bucketName,exists)
			log.Fatal(err)
		}
	} else {
		log.Printf("Successfully created storage for %s\n", bucketName)
	}

	return  &Storage{endpointHost,endpointNum,endpointPort,accessKeyID,secretAccessKey,geoLocation,bucketName,useSSL,
	minioClient}

}





func save(app *App,fileName string,content []byte) {
	ioreader := bytes.NewReader(content)
	size := int64(len(content))

	var opts minio.PutObjectOptions
	n, err := app.Storage.Client.PutObject(app.Storage.BucketName,fileName,ioreader,size,opts)
	if err != nil {
		log.Fatalln(err)
	}

	if (n < size) {
		log.Printf("Varning byte stored to bucket is smaller then object size...%s,%d,%d",fileName,size,n)
	} else {
		log.Printf("Successfully uploaded %s of size %d\n", fileName, n)
	}
}


