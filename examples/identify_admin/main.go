package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	intercom "github.com/uffejaeger/intercom-go"
)

func main() {
	client, err := intercom.NewClient(os.Getenv("INTERCOM_ACCESS_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	req, err := client.NewRequest(context.Background(), http.MethodGet, "/me", nil)
	if err != nil {
		log.Fatal(err)
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	fmt.Println(res.Status)
}
