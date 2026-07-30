<div align="center">
  <img src="./docs/assets/intercom-go-mark.svg" width="144" height="144" alt="intercom-go project mark" />

  <h1>intercom-go</h1>

  <p><strong>An idiomatic, production-ready Go SDK for the Intercom API.</strong></p>

  <p>
    Generated from Intercom's pinned OpenAPI specification.<br />
    Hand-shaped for Go applications.
  </p>

  <p>
    <a href="https://github.com/uffejaeger/intercom-go/actions/workflows/test.yml"><img alt="Test status" src="https://github.com/uffejaeger/intercom-go/actions/workflows/test.yml/badge.svg?branch=main" /></a>
    <a href="https://pkg.go.dev/github.com/uffejaeger/intercom-go"><img alt="Go Reference" src="https://pkg.go.dev/badge/github.com/uffejaeger/intercom-go.svg" /></a>
    <a href="https://github.com/uffejaeger/intercom-go/releases/latest"><img alt="Latest release" src="https://img.shields.io/github/v/release/uffejaeger/intercom-go" /></a>
    <img alt="Go 1.24 or newer" src="https://img.shields.io/badge/Go-1.24%2B-00ADD8?logo=go&amp;logoColor=white" />
    <a href="./LICENSE"><img alt="MIT license" src="https://img.shields.io/github/license/uffejaeger/intercom-go" /></a>
  </p>

  <p>
    <a href="#quick-start">Quick start</a> •
    <a href="#why-intercom-go">Why intercom-go?</a> •
    <a href="#everyday-api">Examples</a> •
    <a href="#api-coverage">API coverage</a> •
    <a href="./docs/production.md">Production guide</a>
  </p>
</div>

## Why intercom-go?

Generated where it should be. Hand-shaped where it matters.

| | |
| --- | --- |
| **Idiomatic Go API** | Use focused services such as `client.Admins.Me(ctx)`, `client.Contacts.Search(ctx, ...)`, and `client.Conversations.Reply(ctx, ...)` instead of generated operation names and unions. |
| **Production-minded** | Opt into conservative retries, inspect request IDs and rate limits, customize individual requests, and verify webhook signatures. |
| **Complete and current** | Public services account for every operation in the pinned Intercom API `2.15` specification, with an automated coverage audit. |
| **Stable public surface** | Generated OpenAPI code stays internal while compatibility checks protect the SDK API your application imports. |
| **Easy to test** | The public [`intercomtest`](https://pkg.go.dev/github.com/uffejaeger/intercom-go/intercomtest) package scripts local Intercom responses and captures outgoing requests without calling Intercom. |

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

Set `INTERCOM_ACCESS_TOKEN`, then create a client and call a service:

```go
package main

import (
	"context"
	"fmt"
	"log"

	intercom "github.com/uffejaeger/intercom-go"
)

func main() {
	client, err := intercom.NewClientFromEnv()
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

Pass a token directly or select the EU or Australian Intercom region with
client options:

```go
client, err := intercom.NewClient("access-token", intercom.WithRegion(intercom.EU))
```

## Everyday API

### Retrieve and search contacts

```go
contact, err := client.Contacts.Get(ctx, "contact_id")

contacts, err := client.Contacts.Search(ctx, intercom.ContactSearch{
	Field:    "email",
	Operator: intercom.ContactSearchEquals,
	Value:    "customer@example.com",
	PerPage:  25,
})
```

### Paginate results

Use explicit page options:

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

Or iterate through cursor-paginated results lazily:

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

### Handle API errors

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

## Production-ready HTTP

### Conservative retries

Retries are opt-in:

```go
client, err := intercom.NewClient("access-token", intercom.WithRetry(intercom.RetryConfig{
	MaxAttempts: 3,
}))
```

The retry policy honors Intercom's `X-RateLimit-Reset` header for rate limits,
falls back to `Retry-After`, and retries selected transient failures. Mutating
requests are not retried unless `AllowUnsafeMethods` is set.

### Response metadata

Observe attempts, request IDs, and rate-limit information without changing
service method signatures:

```go
client, err := intercom.NewClient("access-token", intercom.WithResponseHook(func(info intercom.ResponseInfo) {
	log.Printf("intercom attempt=%d/%d status=%d request_id=%s remaining=%s",
		info.Attempt, info.MaxAttempts, info.StatusCode, info.RequestID, info.RateLimitRemaining)
}))
```

### Per-request options

Override headers, query parameters, or retry behavior for one call:

```go
ctx, err = intercom.WithRequestOptions(ctx, intercom.RequestOptions{
	Headers: http.Header{"X-Correlation-Id": []string{correlationID}},
	Query:   url.Values{"custom": []string{"value"}},
	Retry:   &intercom.RetryConfig{MaxAttempts: 1}, // Disable retries for this call.
})
contact, err := client.Contacts.Get(ctx, "contact_id")
```

See the [production guide](docs/production.md) for HTTP client configuration,
deadlines, retry safety, rate limits, and observability.

## Webhooks

Parse and verify webhook notifications in one step:

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

## API Coverage

The SDK targets Intercom API version `2.15`, pinned in
[`spec/intercom.openapi.yaml`](spec/intercom.openapi.yaml). Public root-package
services cover the pinned specification while generated client code stays
internal under [`internal/generated/intercom`](internal/generated/intercom).

- [Public SDK coverage audit](docs/coverage.md)
- [Production usage](docs/production.md)
- [Go compatibility, releases, and v1 criteria](docs/compatibility.md)
- [Public API compatibility audit](docs/api-compatibility.md)
- [Spec normalization and client generation](docs/generation.md)

## Runnable Examples

- [`examples/identify_admin`](examples/identify_admin)
- [`examples/search_contacts`](examples/search_contacts)

## Support and Security

- Use the [support guidance](SUPPORT.md) for usage questions and maintenance expectations.
- Report security issues through [GitHub's private vulnerability reporting flow](SECURITY.md).
- Read [CONTRIBUTING.md](CONTRIBUTING.md) before proposing a change.

## License

[MIT](LICENSE)
