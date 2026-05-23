.PHONY: generate generate-check test

OAPI_CODEGEN := go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.6.0

generate:
	go run ./internal/tools/normalize-spec spec/intercom.openapi.yaml spec/intercom.codegen.yaml
	$(OAPI_CODEGEN) -config oapi-codegen.yaml spec/intercom.codegen.yaml

generate-check: generate
	git diff --exit-code -- internal/generated/intercom/client.gen.go

test:
	go test ./...
