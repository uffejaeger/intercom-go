# Contributing

Thanks for considering a contribution to `intercom-go`.

## Pull Requests

Keep pull requests focused and include a short summary of the SDK behavior being changed. Link the relevant GitHub issue when there is one.

Before opening a PR, run:

```sh
make pre-push
```

That runs formatting, linting, coverage, and generated-code freshness checks.

## Local Checks

Useful individual commands:

```sh
go test ./...
make coverage
make generate-check
```

Run fuzz targets manually with:

```sh
go test -run=Fuzz -fuzz=Fuzz -fuzztime=30s .
```

Do not commit Intercom access tokens or real customer data. Tests should stay offline unless a future issue explicitly defines a safe live-test setup.
