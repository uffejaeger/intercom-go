# Migrating From `intercom-go.v2`

This SDK is an independent, community-maintained alternative to
[`gopkg.in/intercom/intercom-go.v2`](https://github.com/intercom/intercom-go).
It is not a drop-in replacement: its public API is shaped around the current
Intercom API and requires a `context.Context` for every request.

Start a migration on a branch, upgrade one workflow at a time, and retain the
legacy client until its replacement has been exercised against a test
workspace. Never run both clients with the same request path in production
unless the operation is safe to duplicate.

## Install and construct a client

Replace the legacy module:

```sh
go get github.com/uffejaeger/intercom-go@v0.2.2
```

```go
// Before
import intercom "gopkg.in/intercom/intercom-go.v2"

legacy := intercom.NewClient("access-token", "")

// After
import intercom "github.com/uffejaeger/intercom-go"

client, err := intercom.NewClient("access-token")
if err != nil {
	return err
}
```

For applications that already keep credentials in the environment, use
`intercom.NewClientFromEnv()`. Select a non-US Intercom region at construction
time with `intercom.WithRegion(intercom.EU)` or `intercom.WithRegion(intercom.AU)`.

## Add a context and deadline

Legacy calls do not take a context. Give each replacement request a deadline
that covers its expected work, including any retry waits:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

contact, err := client.Contacts.Get(ctx, contactID)
```

## Map common contact workflows

| Legacy v2 workflow | `intercom-go` workflow |
| --- | --- |
| `ic.Contacts.FindByID(id)` | `client.Contacts.Get(ctx, id)` |
| `ic.Contacts.FindByUserID(id)` | `client.Contacts.GetByExternalID(ctx, id)` |
| `ic.Contacts.ListByEmail(email, ...)` | `client.Contacts.Search(ctx, intercom.ContactSearch{Field: "email", Operator: intercom.ContactSearchEquals, Value: email})` |
| `ic.Contacts.List(...)` | `client.Contacts.List(ctx)` for the first page, or `SearchIter` for cursor-paginated search results |
| `ic.Contacts.Create(...)` | `client.Contacts.Create(ctx, intercom.ContactCreate{...})` |
| `ic.Contacts.Update(...)` | `client.Contacts.Update(ctx, contactID, intercom.ContactUpdate{...})` |

The older SDK distinguishes users and contacts in places where the current
Intercom API uses contacts. Review each legacy user call against the current
API before moving it; do not assume a mechanical method rename is correct.

## Replace client options and HTTP tracing

Pass options when constructing the client instead of calling `ic.Option` later:

```go
client, err := intercom.NewClient("access-token",
	intercom.WithHTTPClient(http.DefaultClient),
	intercom.WithRetry(intercom.RetryConfig{MaxAttempts: 3}),
	intercom.WithResponseHook(func(info intercom.ResponseInfo) {
		log.Printf("status=%d request_id=%s remaining=%s",
			info.StatusCode, info.RequestID, info.RateLimitRemaining)
	}),
)
```

`WithResponseHook` is the replacement for ad-hoc transport tracing when the
goal is observability. Keep hooks fast and never log access tokens, request
bodies, or customer data.

## Paginate conversations safely

Use an iterator rather than manually retaining a cursor:

```go
iter := client.Conversations.ListIter(ctx, intercom.CursorPageOptions{PerPage: 25})
for iter.Next() {
	conversation := iter.Conversation()
	// Process the item without logging customer data.
	_ = conversation
}
if err := iter.Err(); err != nil {
	return err
}
```

The iterator stops with `intercom.ErrPaginationStalled` if Intercom repeats a
cursor, preventing an accidental infinite loop.

## Handle errors and webhooks

Use `intercom.IsNotFound(err)` for missing resources and `errors.As` with
`*intercom.ErrorResponse` when an application needs response status, headers,
or Intercom's request ID. For webhooks, use `ParseAndVerifyWebhook` so the
signature is checked against the exact request bytes before parsing.

See the runnable [webhook](../examples/verify_webhook),
[conversation](../examples/list_conversations), and
[observability](../examples/observe_rate_limits) examples. Consult the
[production guide](production.md) before enabling retries for an existing
write path.
