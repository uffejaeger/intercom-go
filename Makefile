.PHONY: coverage fix format generate generate-check lint pre-push test

OAPI_CODEGEN := go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.6.0
GO_FILES := $(shell git ls-files --cached --others --exclude-standard '*.go')
COVERAGE_THRESHOLD ?= 80

fix:
	go fix ./...

format:
	gofmt -w $(GO_FILES)

lint:
	go vet ./...

coverage:
	go build ./...
	go test -covermode=atomic -coverprofile=coverage.out .
	go tool cover -func=coverage.out
	@coverage=$$(go tool cover -func=coverage.out | awk '/^total:/ { gsub("%", "", $$3); print $$3 }'); \
	awk -v coverage="$$coverage" -v threshold="$(COVERAGE_THRESHOLD)" 'BEGIN { \
		if (coverage + 0 < threshold + 0) { \
			printf "coverage %.1f%% is below required %.1f%%\n", coverage, threshold; \
			exit 1; \
		} \
		printf "coverage %.1f%% meets required %.1f%%\n", coverage, threshold; \
	}'

generate:
	go run ./internal/tools/normalize-spec spec/intercom.openapi.yaml spec/intercom.codegen.yaml
	$(OAPI_CODEGEN) -config oapi-codegen.yaml spec/intercom.codegen.yaml

generate-check: generate
	git diff --exit-code -- internal/generated/intercom/client.gen.go

test:
	go test ./...

pre-push: fix format lint coverage generate-check
