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

CI runs:

```sh
make generate-check
```

That command regenerates the client and fails if `internal/generated/intercom/client.gen.go` is stale.
