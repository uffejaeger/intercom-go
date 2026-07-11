# intercom-go

Idiomatic Go SDK for the [Intercom API](https://developers.intercom.com/).

`intercom-go` wraps Intercom's published OpenAPI spec with a hand-shaped Go API. Generated OpenAPI code is kept internal, while callers use stable service wrappers such as `client.Admins.Me(ctx)`, `client.Contacts.Search(ctx, ...)`, and `client.Conversations.Reply(ctx, ...)`.

## Install

```sh
go get github.com/uffejaeger/intercom-go
```

```go
import intercom "github.com/uffejaeger/intercom-go"
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	intercom "github.com/uffejaeger/intercom-go"
)

func main() {
	client, err := intercom.NewClientFromEnv(intercom.WithRegion(intercom.EU))
	if err != nil {
		log.Fatal(err)
	}

	admin, err := client.Admins.Me(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(*admin.Email)
}
```

`NewClientFromEnv` reads `INTERCOM_ACCESS_TOKEN`. To pass a token directly:

```go
client, err := intercom.NewClient("access-token")
```

Enable conservative opt-in retries:

```go
client, err := intercom.NewClient("access-token", intercom.WithRetry(intercom.RetryConfig{
	MaxAttempts: 3,
}))
```

Retries honor Intercom's `X-RateLimit-Reset` header for rate limits, fall back to `Retry-After` when present, and retry selected transient failures. Mutating requests are not retried unless `AllowUnsafeMethods` is set.

Observe response metadata without changing service method signatures:

```go
client, err := intercom.NewClient("access-token", intercom.WithResponseHook(func(info intercom.ResponseInfo) {
	log.Printf("intercom status=%d request_id=%s remaining=%s", info.StatusCode, info.RequestID, info.RateLimitRemaining)
}))
```

## Examples

Retrieve and search contacts:

```go
contact, err := client.Contacts.Get(ctx, "contact_id")

contacts, err := client.Contacts.Search(ctx, intercom.ContactSearch{
	Field:    "email",
	Operator: intercom.ContactSearchEquals,
	Value:    "customer@example.com",
	PerPage:  25,
})
```

Paginate list and search results:

```go
calls, err := client.Calls.ListWithOptions(ctx, intercom.PageOptions{
	Page:    2,
	PerPage: 25,
})

conversations, err := client.Conversations.ListWithOptions(ctx, intercom.CursorPageOptions{
	PerPage:       50,
	StartingAfter: "cursor",
})
```

Iterate through cursor-paginated results lazily:

```go
iter := client.Conversations.ListIter(ctx, intercom.CursorPageOptions{PerPage: 50})
for iter.Next() {
	conversation := iter.Conversation()
	log.Printf("conversation id=%s", *conversation.Id)
}
if err := iter.Err(); err != nil {
	return err
}
```

Handle API errors:

```go
contact, err := client.Contacts.Get(ctx, "missing")
if err != nil {
	var apiErr *intercom.ErrorResponse
	if errors.As(err, &apiErr) {
		log.Printf("intercom status=%d request_id=%s", apiErr.StatusCode, apiErr.RequestID)
	}
	return err
}
```

Parse webhook notifications:

```go
payload, err := io.ReadAll(r.Body)
if err != nil {
	return err
}

if err := intercom.VerifyWebhookSignature(clientSecret, payload, r.Header); err != nil {
	return err
}

event, err := intercom.ParseWebhookPayload(payload)
if err != nil {
	return err
}

log.Printf("intercom webhook topic=%s id=%s", event.Topic, event.ID)
```

`VerifyWebhookSignature` verifies `X-Hub-Signature` using the raw request body and your Intercom app client secret.

Runnable examples:

- [`examples/identify_admin`](examples/identify_admin)
- [`examples/search_contacts`](examples/search_contacts)

## API Coverage

The SDK targets Intercom API version `2.15`, pinned in [`spec/intercom.openapi.yaml`](spec/intercom.openapi.yaml). Public root-package services cover the pinned spec while generated client code stays internal under [`internal/generated/intercom`](internal/generated/intercom).

See [`docs/coverage.md`](docs/coverage.md) for the current public SDK coverage audit.
See [`docs/generation.md`](docs/generation.md) for the generation workflow.

## License

MIT
