<div align="center">
  <img src="./docs/assets/intercom-go-header.svg" alt="intercom-go — an unofficial, community-maintained Go SDK compatible with the Intercom API" />

  <p>
    <a href="https://github.com/uffejaeger/intercom-go/actions/workflows/test.yml"><img alt="Test status" src="https://github.com/uffejaeger/intercom-go/actions/workflows/test.yml/badge.svg?branch=main" /></a>
  </p>

  <p>
    <a href="https://pkg.go.dev/github.com/uffejaeger/intercom-go">Go Reference</a> •
    <a href="https://github.com/uffejaeger/intercom-go/releases/latest">Releases</a> •
    Go 1.24+ •
    <a href="./LICENSE">MIT License</a>
  </p>

  <p>
    <a href="#quick-start">Quick start</a> •
    <a href="#why-intercom-go">Why intercom-go?</a> •
    <a href="#everyday-api">Examples</a> •
    <a href="#api-coverage">API coverage</a> •
    <a href="./docs/production.md">Production</a>
  </p>
</div>

<p align="center"><sub><strong>About the name:</strong> The company formerly known as Intercom <a href="https://www.intercom.com/blog/today-intercom-becomes-fin/">became Fin</a> in May 2026. Intercom remains the helpdesk and <a href="https://developers.intercom.com/docs/build-an-integration/learn-more/rest-apis/api-changelog">versioned API</a>, so this module remains <code>intercom-go</code>.</sub></p>

## Why intercom-go?

Generated where it should be. Hand-shaped where it matters.

| Capability | What you get |
| --- | --- |
| **Idiomatic Go API** | Use focused services such as `client.Admins.Me(ctx)`, `client.Contacts.Search(ctx, ...)`, and `client.Conversations.Reply(ctx, ...)` instead of generated operation names and unions. |
| **Production-minded** | Opt into conservative retries, inspect request IDs and rate limits, customize individual requests, and verify webhook signatures. |
| **Complete and current** | Public services account for every operation in the pinned Intercom API `2.16` specification, with an automated coverage audit. |
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

The SDK targets Intercom API version `2.16`, pinned in
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

## Trademark Notice

The Fin and Intercom names and logos are trademarks or service marks of
Intercom, Inc. or its affiliates in the U.S. and other countries. The Go
trademark and Go logo are trademarks of Google LLC. Their inclusion does not
imply affiliation, sponsorship, or endorsement.

## License

[MIT](LICENSE)
