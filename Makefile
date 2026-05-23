.PHONY: fix format generate generate-check lint pre-push test

OAPI_CODEGEN := go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.6.0
GO_FILES := $(shell git ls-files '*.go')

fix:
	go fix ./...

format:
	gofmt -w $(GO_FILES)

lint:
	go vet ./...

generate:
	go run ./internal/tools/normalize-spec spec/intercom.openapi.yaml spec/intercom.codegen.yaml
	$(OAPI_CODEGEN) -config oapi-codegen.yaml spec/intercom.codegen.yaml

generate-check: generate
	git diff --exit-code -- internal/generated/intercom/client.gen.go

test:
	go test ./...

pre-push: fix format lint test generate-check
