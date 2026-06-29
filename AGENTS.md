# Repository Guidelines

## Project Structure & Module Organization

This repository is a Go module for an idiomatic Intercom SDK. Public SDK code lives in the root package, for example `client.go`, `options.go`, `admins.go`, `contacts.go`, and related `*_test.go` files. Generated OpenAPI client code is committed under `internal/generated/intercom` and should stay internal. The pinned upstream spec is in `spec/intercom.openapi.yaml`, with metadata in `spec/metadata.json`. Spec normalization tooling lives in `internal/tools/normalize-spec`. Examples are in `examples/identify_admin` and `examples/search_contacts`.

## Build, Test, and Development Commands

- `go fix ./...`: apply standard Go source rewrites.
- `gofmt -w $(git ls-files --cached --others --exclude-standard '*.go')`: format repo Go files.
- `go vet ./...`: run the standard Go lint checks.
- `go test ./...`: run the full test suite.
- `make coverage`: run root package coverage and enforce `COVERAGE_THRESHOLD` (default `99.9`).
- `make generate`: normalize the pinned spec, then regenerate `internal/generated/intercom/client.gen.go`.
- `make generate-check`: regenerate stubs and fail if committed generated files are stale.
- `make pre-push`: shorthand for `fix`, `format`, `lint`, `coverage`, and `generate-check`; use this before pushing to GitHub.

When local sandboxing blocks the default Go cache, use workspace-local cache paths:

```sh
GOCACHE=.cache/go-build GOMODCACHE=.cache/go-mod go test ./...
```

## Coding Style & Naming Conventions

Use standard Go style: tabs from `gofmt`, concise exported comments, and lowercase package names. Prefer small public service wrappers, such as `client.Admins.Me(ctx)` or `client.Contacts.Get(ctx, id)`, over exposing generated types directly. Keep generated changes reproducible through `make generate`.

## Testing Guidelines

Tests use Go’s standard `testing` package. Keep tests next to covered code in `*_test.go` files. Prefer table-driven tests for endpoint behavior, validation, and error mapping. Add regression tests for review fixes and public API changes. Before opening or updating a PR, run:

```sh
go fix ./...
gofmt -w $(git ls-files --cached --others --exclude-standard '*.go')
go vet ./...
make coverage
make generate-check
```

## Commit & Pull Request Guidelines

Commit history uses short, imperative summaries such as `Add public Contacts service`. Keep commits focused and preserve review history. **Never force-push or amend commits on PR branches.** Each round of changes (e.g. addressing review feedback) gets its own new commit, so the PR history shows what changed over time. Squashing happens only at merge to main — that is where the history gets cleaned up, not on the PR branch. Work on feature branches and open PRs into `main`; do not push directly to `main`. PRs should explain changed SDK behavior, list verification commands, and link issues. CI must pass before merge, especially the required `test` check.

## Security & Configuration Tips

Do not commit Intercom access tokens or real customer data. Examples should read credentials from `INTERCOM_ACCESS_TOKEN` via `NewClientFromEnv()` or explicit environment lookups.
