# Launch Checklist

This checklist prepares the public `v0.2.2` release and retains the proposed
community announcements for explicit approval. Do not submit either draft
without approval of its final text, target, links, and account identity.

## Repository settings

- [x] Repository description: `Modern, production-ready Go SDK for Intercom API 2.15 (unofficial, community-maintained).`
- [x] Homepage: <https://pkg.go.dev/github.com/uffejaeger/intercom-go>
- [x] Topics: `api-client`, `go`, `golang`, `intercom`, `intercom-api`, `openapi`, `sdk`.
- [ ] Upload [`docs/assets/social-preview.png`](assets/social-preview.png), a 1280×640 PNG under 1 MB, in GitHub's **Settings → General → Social preview**.

## Release checks

- [x] Run `make pre-push`.
- [ ] Merge the release-preparation pull request into `main`.
- [ ] Publish immutable tag and GitHub release `v0.2.2`.
- [ ] Verify `GOPROXY=proxy.golang.org go list -m github.com/uffejaeger/intercom-go@v0.2.2`.
- [ ] Verify <https://pkg.go.dev/github.com/uffejaeger/intercom-go> shows `v0.2.2`.

Installation command:

```sh
go get github.com/uffejaeger/intercom-go@v0.2.2
```

Release summary:

> v0.2.2 adds a migration guide from the legacy v2 client and runnable examples
> for verified webhooks, conversation iteration, and response/rate-limit
> observability. It makes no public Go API changes.

## Draft: Go Forum Releases

**Title:** `intercom-go v0.2.2 — a community-maintained Go SDK for Intercom API 2.15`

> I released `intercom-go` v0.2.2, an unofficial, community-maintained Go SDK
> for Intercom API 2.15.
>
> It keeps generated OpenAPI code internal while exposing idiomatic services,
> opt-in conservative retries, regional endpoints, request/rate-limit metadata,
> verified webhooks, and local HTTP test support.
>
> Install: `go get github.com/uffejaeger/intercom-go@v0.2.2`
>
> The release includes a migration guide for
> `gopkg.in/intercom/intercom-go.v2` and runnable webhook, conversation, and
> observability examples: <https://github.com/uffejaeger/intercom-go>
>
> I would especially value feedback from teams replacing the legacy client:
> which workflow or migration helper would make evaluation easier?

## Draft: Intercom Community

**Title:** `Community Go SDK for Intercom API 2.15 — looking for migration feedback`

> I maintain `intercom-go`, an unofficial, community-maintained Go SDK for
> Intercom API 2.15. It is not affiliated with or endorsed by Intercom.
>
> The project offers typed services, optional conservative retries, regional
> endpoints, response/rate-limit metadata, verified webhooks, and local HTTP
> test support. The latest release also includes a migration guide for the
> legacy `gopkg.in/intercom/intercom-go.v2` client:
> <https://github.com/uffejaeger/intercom-go/blob/main/docs/migrating-from-intercom-go-v2.md>
>
> If you are using Go with Intercom, which legacy-client workflow would you
> want documented or supported first?

## Approval gate

Before posting either draft, confirm the exact final text, destination URL,
account identity, and every linked page with the repository owner. Do not
cross-post or submit to newsletters, curated lists, Reddit, or Intercom
documentation without separate approval.
