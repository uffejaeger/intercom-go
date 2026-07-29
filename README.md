# intercom-go

Idiomatic Go SDK for the [Intercom API](https://developers.intercom.com/).

`intercom-go` wraps Intercom's published OpenAPI spec with a hand-shaped Go API. Generated OpenAPI code is kept internal, while callers use stable service wrappers such as `client.Admins.Me(ctx)`, `client.Contacts.Search(ctx, ...)`, and `client.Conversations.Reply(ctx, ...)`.

## Install

```sh
go get github.com/uffejaeger/intercom-go@latest
```

```go
import intercom "github.com/uffejaeger/intercom-go"
```

Go records the selected release in your module files. Review SDK upgrades like
other production dependency changes rather than tracking an unreviewed branch.

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
	log.Printf("intercom attempt=%d/%d status=%d request_id=%s remaining=%s",
		info.Attempt, info.MaxAttempts, info.StatusCode, info.RequestID, info.RateLimitRemaining)
}))
```

Override headers, query parameters, or retry behavior for one call:

```go
ctx, err = intercom.WithRequestOptions(ctx, intercom.RequestOptions{
	Headers: http.Header{"X-Correlation-Id": []string{correlationID}},
	Query:   url.Values{"custom": []string{"value"}},
	Retry:   &intercom.RetryConfig{MaxAttempts: 1}, // Disable retries for this call.
})
contact, err := client.Contacts.Get(ctx, "contact_id")
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
	if errors.Is(err, intercom.ErrPaginationStalled) {
		// Intercom returned a cursor that was already requested.
	}
	return err
}
```

Handle API errors:

```go
contact, err := client.Contacts.Get(ctx, "missing")
if err != nil {
	if intercom.IsNotFound(err) {
		return nil
	}
	var apiErr *intercom.ErrorResponse
	if errors.As(err, &apiErr) {
		log.Printf("intercom status=%d request_id=%s retry_after=%s",
			apiErr.StatusCode, apiErr.RequestID, apiErr.Headers.Get("Retry-After"))
	}
	return err
}
```

Parse webhook notifications:

```go
event, err := intercom.ParseAndVerifyWebhook(r, clientSecret, 0)
if err != nil {
	return err
}

log.Printf("intercom webhook topic=%s id=%s", event.Topic, event.ID)
```

`ParseAndVerifyWebhook` limits the body to 1 MiB by default, verifies
`X-Hub-Signature` over the exact bytes received, and then parses the event.
Use `VerifyWebhookSignature` and `ParseWebhookPayload` separately when the
application already owns the raw payload bytes.

Runnable examples:

- [`examples/identify_admin`](examples/identify_admin)
- [`examples/search_contacts`](examples/search_contacts)

## API Coverage

The SDK targets Intercom API version `2.15`, pinned in [`spec/intercom.openapi.yaml`](spec/intercom.openapi.yaml). Public root-package services cover the pinned spec while generated client code stays internal under [`internal/generated/intercom`](internal/generated/intercom).

See [`docs/coverage.md`](docs/coverage.md) for the current public SDK coverage audit.
See [`docs/generation.md`](docs/generation.md) for the generation workflow.
See [`docs/production.md`](docs/production.md) for production HTTP, deadline, retry, rate-limit, and observability guidance.
See [`docs/compatibility.md`](docs/compatibility.md) for Go support, semantic versioning, deprecation, and v1 criteria.

Security reports belong in [GitHub's private vulnerability reporting flow](SECURITY.md).
For usage questions and maintenance expectations, see [support guidance](SUPPORT.md).

## License

MIT
