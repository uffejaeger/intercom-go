# Production Use

This guide covers the HTTP configuration and operational behavior of `intercom-go` in production services. It is intentionally about the SDK's behavior; choose timeout and retry values that fit the latency and duplicate-work tolerance of your application.

## Recommended starting configuration

Use an explicit HTTP client, opt in to conservative retries for read operations, and capture response metadata. This is a starting point, not a universal timeout policy:

```go
httpClient := &http.Client{Timeout: 10 * time.Second}

client, err := intercom.NewClient("access-token",
	intercom.WithHTTPClient(httpClient),
	intercom.WithRetry(intercom.RetryConfig{MaxAttempts: 3}),
	intercom.WithResponseHook(func(info intercom.ResponseInfo) {
		log.Printf("intercom status=%d request_id=%s duration=%s remaining=%s",
			info.StatusCode, info.RequestID, info.Duration, info.RateLimitRemaining)
	}),
)
if err != nil {
	return err
}
```

For each operation, create a context with a deadline:

```go
ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
defer cancel()

contact, err := client.Contacts.Get(ctx, "contact_id")
```

## Bound every call with a context deadline

Every service method accepts a `context.Context`. Give each request a deadline that is appropriate for the caller's latency budget. A deadline covers the complete operation—including connection time, the response, and any retry waits—and stops a retry wait when it expires.

Derive the context from the inbound request or job context (`parentCtx` above), rather than using `context.Background()` directly for a request in a long-running service. That preserves caller cancellation as well as a deadline.

## Set an HTTP client timeout

Use `WithHTTPClient` to configure transport-level limits, including a timeout. `http.Client.Timeout` applies to the whole HTTP exchange; the request context bounds the individual SDK call. Whichever expires first cancels the request.

```go
client, err := intercom.NewClient("access-token",
	intercom.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
)
if err != nil {
	return err
}
```

Do not modify `http.DefaultClient` globally. Supplying an explicit client keeps this SDK's behavior isolated from unrelated outbound HTTP calls. Set the client timeout slightly above the usual per-call context deadline when the context is your primary latency budget; this still gives the client a bounded fallback for calls whose context was accidentally left without a deadline.

## Retries

Retries are opt-in. Enable them only when your application can tolerate repeated attempts:

```go
client, err := intercom.NewClient("access-token", intercom.WithRetry(intercom.RetryConfig{
	MaxAttempts: 3, // Total attempts, including the initial request.
}))
```

By default, the retry policy retries `429`, `500`, `502`, `503`, and `504`, plus transient network errors. It uses exponential backoff with jitter: 100 ms initial backoff, 2 s maximum backoff, 20% jitter, and three total attempts. Configure `RetryConfig` if your workload needs different status codes or timing. `MaxAttempts: 1` disables retrying while retaining a single attempt.

Only safe HTTP methods (`GET`, `HEAD`, `OPTIONS`, and `TRACE`) are retried by default. This protects operations that can create, update, or delete Intercom data from being replayed after an ambiguous failure.

`AllowUnsafeMethods` permits retries for `POST`, `PUT`, `PATCH`, and `DELETE`:

```go
intercom.WithRetry(intercom.RetryConfig{
	MaxAttempts:         3,
	AllowUnsafeMethods: true,
})
```

Enable it only when the operation is idempotent, or your application can safely handle duplicate side effects. In particular, an HTTP failure can happen after Intercom has received and processed a mutating request. Requests with bodies are retried only when their body can be recreated.

The retry configuration does not override a context deadline. Budget for all attempts and their waits inside the operation deadline; otherwise a retry may be cancelled before it can run.

## Rate limits and server retry hints

For a retried response, the SDK uses `Retry-After` first. It supports both seconds and HTTP-date values. For `429 Too Many Requests` responses without `Retry-After`, it waits until the Unix timestamp in `X-RateLimit-Reset`. Otherwise it uses the configured exponential backoff.

The SDK exposes rate-limit metadata through `ResponseHook`, including `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`, and `Retry-After`. Use the remaining count and reset time to reduce concurrency or defer non-urgent work before requests begin failing. Do not try to parse these headers from a service method's return value; the hook makes them available for both successful and unsuccessful HTTP responses.

For high-volume workloads, use a shared limiter or queue ahead of the SDK. On a low remaining count, reduce concurrency; when a reset time is available, defer non-urgent work until after it. Retries help an individual request recover, but they do not replace admission control across concurrent workers.

## Safe observability

Use `WithResponseHook` to record status, duration, request ID, and rate-limit information. The hook runs synchronously once per HTTP attempt, including retries, so keep it fast and avoid making network calls from it.

```go
client, err := intercom.NewClient("access-token", intercom.WithResponseHook(func(info intercom.ResponseInfo) {
	log.Printf("intercom status=%d request_id=%s duration=%s remaining=%s reset=%s",
		info.StatusCode, info.RequestID, info.Duration, info.RateLimitRemaining, info.RateLimitReset)
}))
```

Include Intercom's `X-Request-Id` in logs and support requests; it helps Intercom correlate a failing request. For API error responses, the same value is available as `ErrorResponse.RequestID`. A transport error has no HTTP response, so `ResponseInfo.StatusCode` is zero, `RequestID` is empty, and `ResponseInfo.Err` contains the transport error.

Never log the access token, `Authorization` header, request body, or raw customer data unless your logging policy and retention controls explicitly permit it. Prefer structured metadata such as status, request ID, duration, and aggregate rate-limit values. If you log identifiers for troubleshooting, treat them as customer data and apply the same access and retention controls.
