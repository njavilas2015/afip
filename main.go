package main

import (
	"log"

	internal "github.com/njavilas2015/afip/internal"
)

func main() {

	internal.GenerateKey()

	client, err := internal.GetClient("wsaa", false)

	if err != nil {
		log.Fatalf("Error getting client: %v", err)
	}

	log.Println("Client configured:", client)
}
