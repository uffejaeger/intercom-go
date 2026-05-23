package main

import (
	"context"
	"fmt"
	"log"
	"os"

	intercom "github.com/uffejaeger/intercom-go"
)

func main() {
	client, err := intercom.NewClient(os.Getenv("INTERCOM_ACCESS_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	admin, err := client.Admins.Me(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(*admin.Email)
}
