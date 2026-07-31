// list_conversations lists conversation pages lazily without retaining cursors.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	intercom "github.com/uffejaeger/intercom-go"
)

func main() {
	client, err := intercom.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	iter := client.Conversations.ListIter(ctx, intercom.CursorPageOptions{PerPage: 25})
	count := 0
	for iter.Next() {
		// Process the conversation here. Avoid logging customer data by default.
		count++
	}
	if err := iter.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("processed %d conversations\n", count)
}
