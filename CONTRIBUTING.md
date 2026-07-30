# Contributing

Thanks for considering a contribution to `intercom-go`.

## Pull Requests

Keep pull requests focused and include a short summary of the SDK behavior being changed. Link the relevant GitHub issue when there is one.

Before opening a PR, run:

```sh
make pre-push
```

That runs formatting, vet, Staticcheck, coverage, generated-code freshness,
public API compatibility, and known-vulnerability checks.

## CI And Branch Protection

The `test` GitHub Actions check runs for every pull request and push to `main`.
It verifies generated-code freshness, public API compatibility, `go vet`,
Staticcheck, coverage, race-enabled tests, and known-vulnerability scanning.
The compatibility matrix runs the complete package tests on Go 1.24, 1.25, and
1.26 with toolchain auto-switching disabled.

Repository administrators should require the `test` check before merging into `main`. Configure this in **Settings** > **Branches** > the `main` branch rule or ruleset.
The checked repository-security baseline and its review checklist are documented
in [`docs/maintainers.md`](docs/maintainers.md).

## Local Checks

Useful individual commands:

```sh
go test ./...
make coverage
make generate-check
make api-compatibility
make vuln
```

Run fuzz targets manually with:

```sh
go test -run=Fuzz -fuzz=Fuzz -fuzztime=30s .
```

Do not commit Intercom access tokens or real customer data. Tests should stay offline unless a future issue explicitly defines a safe live-test setup.

## Security And Support

Do not open a public issue for a suspected vulnerability. Follow
[`SECURITY.md`](SECURITY.md) to report it privately.

Use GitHub issues for reproducible SDK bugs and feature requests. See
[`SUPPORT.md`](SUPPORT.md) for the project's support scope and response
expectations.
