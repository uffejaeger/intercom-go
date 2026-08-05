// verify_webhook runs a minimal HTTP endpoint that verifies Intercom webhooks
// before processing their payloads.
package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	intercom "github.com/uffejaeger/intercom-go"
)

func main() {
	clientSecret := os.Getenv("INTERCOM_CLIENT_SECRET")
	if clientSecret == "" {
		log.Fatal("INTERCOM_CLIENT_SECRET is required")
	}

	http.HandleFunc("/webhooks/intercom", func(w http.ResponseWriter, r *http.Request) {
		event, err := intercom.ParseAndVerifyWebhook(r, clientSecret, 0)
		if err != nil {
			switch {
			case errors.Is(err, intercom.ErrWebhookPayloadTooLarge):
				http.Error(w, "payload too large", http.StatusRequestEntityTooLarge)
			case errors.Is(err, intercom.ErrWebhookSignatureMissing),
				errors.Is(err, intercom.ErrWebhookSignatureInvalid),
				errors.Is(err, intercom.ErrWebhookSignatureUnsupported):
				http.Error(w, "invalid webhook signature", http.StatusUnauthorized)
			default:
				http.Error(w, "invalid webhook", http.StatusBadRequest)
			}
			return
		}

		// Decode event.Data.Item only for topics this endpoint handles.
		log.Printf("verified webhook topic=%s id=%s", event.Topic, event.ID)
		fmt.Fprintln(w, "ok")
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
