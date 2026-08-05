// observe_rate_limits records request metadata without logging credentials or payloads.
package main

import (
	"context"
	"log"

	intercom "github.com/uffejaeger/intercom-go"
)

func main() {
	client, err := intercom.NewClientFromEnv(
		intercom.WithResponseHook(func(info intercom.ResponseInfo) {
			log.Printf("attempt=%d/%d status=%d request_id=%s remaining=%s reset=%s duration=%s",
				info.Attempt,
				info.MaxAttempts,
				info.StatusCode,
				info.RequestID,
				info.RateLimitRemaining,
				info.RateLimitReset,
				info.Duration,
			)
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := client.Admins.Me(context.Background()); err != nil {
		log.Fatal(err)
	}
}
