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

Runnable examples:

- [`examples/identify_admin`](examples/identify_admin)
- [`examples/search_contacts`](examples/search_contacts)

## API Coverage

The SDK targets Intercom API version `2.15`, pinned in [`spec/intercom.openapi.yaml`](spec/intercom.openapi.yaml). Public root-package services cover the pinned spec while generated client code stays internal under [`internal/generated/intercom`](internal/generated/intercom).

See [`docs/coverage.md`](docs/coverage.md) for the current public SDK coverage audit.
See [`docs/generation.md`](docs/generation.md) for the generation workflow.

## Development

```sh
go test ./...
make coverage
make generate-check
```

Before opening a PR, run:

```sh
make pre-push
```

## License

MIT
