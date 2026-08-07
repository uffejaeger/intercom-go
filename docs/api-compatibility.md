# Public API Compatibility Audit

This audit records the public API decisions made before the v1.0.0 compatibility
promise. It was completed on 2026-07-29 against release `v0.2.0`.

## Scope and result

The compatibility surface consists of every exported identifier in:

- `github.com/uffejaeger/intercom-go`
- `github.com/uffejaeger/intercom-go/intercomtest`

Repository tools, examples, and tests are excluded. Packages under `internal/`
are also excluded except for the generated model package described below. Go's
`gorelease` tool reports that the current module is a valid semantic version
candidate for `v1.0.0`. A manual review found no high-impact public API redesign
that should delay the v1 compatibility promise.

The review does not claim that every runtime behavior is immutable or that every
Intercom workflow has a high-level helper. Behavioral changes, upstream API
changes, and operational verification remain governed by the release policy,
tests, and live-canary criterion.

## Decisions

### Generated model aliases

The root package intentionally exposes aliases for selected generated request
and response models. This keeps generated code internal while avoiding a large,
lossy hand-maintained model layer.

Those aliases and the exported shapes they expose are part of the public API.
Specification regeneration may add compatible API, but removal, renaming, or
type changes must not bypass compatibility review.

Go's API export data does not expand aliases into an `internal` package during a
normal public-module comparison. To avoid a blind spot, the automated gate also
compares the generated model package directly. This is deliberately
conservative: it can flag a changed generated type that is not currently
reachable through a root-package alias. Maintainers should confirm reachability
during review, but must preserve any shape that is publicly exposed.

### Request-scoped options

`RequestOptions` and `WithRequestOptions` remain the extension point for
per-request headers, query values, and retry overrides. Carrying these options
through `context.Context` preserves the service method signatures while keeping
the options request-local.

The exported fields and validation behavior are part of the compatibility
promise. Future request-scoped features should be added compatibly or introduced
through a separate API when adding a field would itself be incompatible.

### Iterators

`ContactIterator`, `ConversationIterator`, and `TicketIterator` consistently use
the `Next`, typed-value accessor, and `Err` pattern. They provide a small
convenience layer over cursor pagination without hiding terminal errors.

The iterator names, method signatures, and `ErrPaginationStalled` error contract
are part of the compatibility promise. New resources should follow the same
pattern when an iterator adds material value.

### Errors

`ErrorResponse` preserves HTTP status, Intercom error details, request ID,
response headers, and a bounded response body. `IsStatus` is the general
classifier, with helpers for common status classes.

The error types, wrapped-error classification through `errors.As`, and existing
predicate semantics are part of the compatibility promise. New predicates may
be added without changing the general classification contract.

### Variant request interfaces

Public request interfaces with unexported marker methods intentionally form a
closed set of supported variants. This prevents callers from supplying values
that the SDK cannot serialize correctly. New supported variants may be added
compatibly; opening or otherwise redesigning these interfaces is not required
for v1.

### Low-level escape hatch

`Client.NewRequest` and `Client.Do` remain available for Intercom endpoints or
features that do not yet have a high-level wrapper. Their presence lets wrapper
coverage evolve without forcing consumers to wait for an SDK release.

## Automated compatibility gate

`make api-compatibility` exports the public module API from the current working
tree and compares it with the pinned released baseline using Go's `apidiff`
tool. The command fails when an incompatible compile-time change is reported.
CI runs it in the quality job, and `make pre-push` runs it locally.

The generated OpenAPI client is under Go's `internal/` boundary, so downstream
SDK consumers cannot import it. It is verified for reproducibility by `make
generate-check`, rather than treated as a second public API surface.

The checker has a narrow source-compatibility allowance for the reviewed
Contact and Ticket boundary models: API 2.16 changed Intercom's wire
representation of `Contact.owner_id` from an integer to a string and Ticket
assignee IDs from strings to integers. The SDK preserves the historical public
field types and converts the wire values at the boundary. `apidiff` reports
the required alias-to-struct change and its dependent methods and iterators as
incompatible, so the checker permits only those exact entries. External
consumer compile and response-conversion regression tests cover the preserved
contracts. No other `apidiff` finding is suppressed.

The current baseline is `v0.2.0`. After publishing a release that expands the
public API, maintainers advance `API_BASELINE` in `Makefile` to that release so
later removals of newly added API are also detected. The `apidiff` version is
pinned because compatibility classification changes are themselves reviewable
tooling changes.

The gate is deliberately strict, but it is not a substitute for review:

- It detects compile-time API incompatibilities, not behavior changes.
- It does not prove that Intercom's live API still honors an endpoint contract.
- An intentional incompatible change must be documented and released under the
  semantic-versioning policy; disabling the check is not an acceptable shortcut.
