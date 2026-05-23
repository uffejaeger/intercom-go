# intercom-go

Idiomatic Go SDK for the Intercom API.

This project is starting from Intercom's published OpenAPI description rather than hand-written endpoint coverage. The long-term goal is to provide a stable, community-friendly SDK while keeping generated code reproducible from the upstream API spec.

## Status

Early bootstrap work is in progress.

- The Intercom API `2.15` OpenAPI spec is pinned in [`spec/intercom.openapi.yaml`](spec/intercom.openapi.yaml).
- Spec source metadata is tracked in [`spec/metadata.json`](spec/metadata.json).
- The root package contains the initial public client primitives.
- OpenAPI-generated client stubs are committed under [`internal/generated/intercom`](internal/generated/intercom).
- CI runs `go test ./...`.

Generated code is kept internal while the public SDK surface is designed. Community-facing endpoint services will wrap the generated client instead of exposing generator-specific APIs directly.

See [`docs/generation.md`](docs/generation.md) for the generation workflow.

## Install

```sh
go get github.com/uffejaeger/intercom-go
```

## Usage

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	intercom "github.com/uffejaeger/intercom-go"
)

func main() {
	client, err := intercom.NewClient(os.Getenv("INTERCOM_ACCESS_TOKEN"), intercom.WithRegion(intercom.EU))
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

Contacts:

```go
contact, err := client.Contacts.Get(ctx, "contact_id")
contacts, err := client.Contacts.List(ctx)
contacts, err := client.Contacts.Search(ctx, intercom.ContactSearch{
	Field:    "email",
	Operator: intercom.ContactSearchEquals,
	Value:    "customer@example.com",
	PerPage:  25,
})
```

## Examples

- [`examples/identify_admin`](examples/identify_admin)
- [`examples/search_contacts`](examples/search_contacts)

## Design Principles

- Keep the upstream OpenAPI spec pinned and reviewable.
- Keep generated code internal until the public SDK shape is intentionally designed.
- Prefer small, idiomatic public APIs over exposing generator-specific types everywhere.
- Automate spec update detection, but require human review before generated changes are released.
