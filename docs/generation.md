# OpenAPI Generation

The SDK uses Intercom's published OpenAPI spec as the source of truth.

Generation currently uses `oapi-codegen` because it can generate a compact Go client and model package from the normalized Intercom spec. The generated package lives under `internal/generated/intercom` so generator-specific APIs do not become the public SDK contract.

The upstream spec is kept unchanged at `spec/intercom.openapi.yaml`. Before generation, `internal/tools/normalize-spec` writes `spec/intercom.codegen.yaml` with compatibility changes needed for Go generation:

- path placeholders are aligned with path parameters,
- component object schemas get deterministic Go names,
- properties get deterministic Go names,
- `$ref` properties with sibling metadata are wrapped using `allOf` so OpenAPI 3.0 tooling can see the metadata.

Run generation with:

```sh
make generate
```

Update the pinned upstream spec, then regenerate with:

```sh
make update-spec
make generate
```

CI runs:

```sh
make generate-check
```

That command regenerates the client and fails if `internal/generated/intercom/client.gen.go` is stale.

## Upstream Spec Updates

`make update-spec` reads `spec/metadata.json`, fetches the latest configured upstream spec from `intercom/Intercom-OpenAPI`, writes `spec/intercom.openapi.yaml`, and updates the pinned upstream commit in `spec/metadata.json`.

The scheduled `update-spec` GitHub Actions workflow runs weekly and can also be triggered manually. When upstream changes are detected, it:

- updates the pinned spec and metadata,
- regenerates `spec/intercom.codegen.yaml` and `internal/generated/intercom/client.gen.go`,
- formats, lints, tests with coverage, and verifies generated-code freshness,
- opens a pull request labeled `spec-update` and `generated`,
- records check outcomes in the pull request body so breaking upstream changes can still be reviewed,
- includes an OpenAPI diff summary with breaking candidates, additive changes, documentation-only candidates, and other schema changes.

Human review is still required before merging generated spec-update pull requests.
