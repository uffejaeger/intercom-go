package main

import (
	"context"
	"fmt"
	"log"

	intercom "github.com/uffejaeger/intercom-go"
)

const (
	searchField   = "email"
	searchValue   = "customer@example.com"
	searchPerPage = 25
)

func main() {
	client, err := intercom.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	contacts, err := client.Contacts.Search(context.Background(), intercom.ContactSearch{
		Field:    searchField,
		Operator: intercom.ContactSearchEquals,
		Value:    searchValue,
		PerPage:  searchPerPage,
	})
	if err != nil {
		log.Fatal(err)
	}

	if contacts.TotalCount != nil {
		fmt.Println(*contacts.TotalCount)
	}
}
