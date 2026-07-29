# Compatibility And Release Policy

## Go versions

The SDK supports Go 1.24, Go 1.25, and Go 1.26. The `go` directive declares Go
1.24 as the consumer minimum, while the `toolchain` directive selects Go 1.26.5
for repository development when automatic toolchain selection is enabled.

CI tests all three release families with `GOTOOLCHAIN=local`, ensuring each
matrix entry uses the selected compiler instead of silently switching to the
development toolchain. Go 1.22 and Go 1.23 are not supported because the latest
`github.com/oapi-codegen/runtime` dependency requires Go 1.24.

The project supports the widest contiguous Go-version window permitted by its
current production dependencies, up to a target of five release families.
Dependency correctness and security fixes take priority over extending that
window with an outdated dependency. Patch-level Go updates may be adopted in
any SDK release; dropping a supported release family requires an SDK minor
release and release-note notice.

## Semantic versioning

Releases follow [Semantic Versioning](https://semver.org/). The public API
includes exported identifiers in the root `intercom` package and the
`intercomtest` package. Code under `internal/`, repository tools, tests, and
examples are not part of the compatibility promise.

Before v1.0.0:

- Patch releases contain backward-compatible fixes, documentation, and
  maintenance updates.
- Minor releases may add functionality and may contain a breaking public API
  change when the design cannot be corrected compatibly. Breaking changes must
  be called out in release notes with migration guidance.

Starting with v1.0.0, backward-incompatible public API changes require a new
major release. Changes required by Intercom API behavior may still affect
runtime results without changing Go signatures; those changes are documented
in release notes when known.

## Deprecation

When practical, an API scheduled for removal is marked with a Go `Deprecated:`
doc comment and a replacement is documented. Before v1.0.0, deprecated APIs
remain available for at least two minor releases and 90 days. After v1.0.0,
removal waits for the next major release.

The deprecation window may be shortened for a security vulnerability, data-loss
risk, upstream API removal, or behavior that cannot be preserved safely.
Release notes explain any exception and the recommended migration.

## Release notes and installation

Every release uses a `vMAJOR.MINOR.PATCH` tag and includes categorized release
notes plus an exact `go get` command. Applications should commit `go.mod` and
`go.sum`, review release notes, and test dependency upgrades before deployment.

Install the newest published release with:

```sh
go get github.com/uffejaeger/intercom-go@latest
```

Install a reviewed exact release with:

```sh
go get github.com/uffejaeger/intercom-go@vX.Y.Z
```

## v1.0.0 exit criteria

The project is ready for v1.0.0 when all of the following are true:

- The public wrapper coverage check accounts for every operation in the pinned
  Intercom OpenAPI specification.
- The public API has completed a compatibility review and has no known
  high-impact design changes pending.
- CI passes formatting, generated-code freshness, vet, Staticcheck, unit and
  integration tests, the race detector, the supported-Go matrix, coverage, and
  known-vulnerability scanning.
- A scheduled, read-only live canary validates authentication, regional routing,
  API-version headers, and representative pagination against a dedicated
  Intercom test workspace.
- Production, security, support, contribution, release, and deprecation
  guidance is published and current.
- There are no unresolved known critical security issues or release-blocking
  correctness defects.

Meeting these criteria establishes the v1 compatibility promise; it does not
mean every possible Intercom workflow has a high-level convenience API.
