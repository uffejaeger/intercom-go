# intercom-go

Idiomatic Go SDK for the Intercom API.

This project is starting from Intercom's published OpenAPI description rather than hand-written endpoint coverage. The long-term goal is to provide a stable, community-friendly SDK while keeping generated code reproducible from the upstream API spec.

## Status

Early bootstrap work is in progress.

- The Intercom API `2.15` OpenAPI spec is pinned in [`spec/intercom.openapi.yaml`](spec/intercom.openapi.yaml).
- Spec source metadata is tracked in [`spec/metadata.json`](spec/metadata.json).
- The root package contains the initial public client primitives.
- CI runs `go test ./...`.
- Full generated endpoint coverage is being evaluated behind `internal/generated/...`.

## Install

```sh
go get github.com/uffejaeger/intercom-go
```

## Usage

```go
package main

import (
	"context"
	"log"
	"net/http"
	"os"

	intercom "github.com/uffejaeger/intercom-go"
)

func main() {
	client, err := intercom.NewClient(os.Getenv("INTERCOM_ACCESS_TOKEN"), intercom.WithRegion(intercom.EU))
	if err != nil {
		log.Fatal(err)
	}

	req, err := client.NewRequest(context.Background(), http.MethodGet, "/me", nil)
	if err != nil {
		log.Fatal(err)
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
}
```

## Design Principles

- Keep the upstream OpenAPI spec pinned and reviewable.
- Keep generated code internal until the public SDK shape is intentionally designed.
- Prefer small, idiomatic public APIs over exposing generator-specific types everywhere.
- Automate spec update detection, but require human review before generated changes are released.
