.PHONY: generate test

generate:
	go run ./internal/tools/normalize-spec spec/intercom.openapi.yaml spec/intercom.codegen.yaml
	oapi-codegen -config oapi-codegen.yaml spec/intercom.codegen.yaml

test:
	go test ./...
