package intercom

import (
	"net/http"
	"time"
)

const (
	requestIDHeader          = "X-Request-Id"
	rateLimitLimitHeader     = "X-RateLimit-Limit"
	rateLimitRemainingHeader = "X-RateLimit-Remaining"
)

// ResponseInfo describes an HTTP response or transport error observed by a ResponseHook.
type ResponseInfo struct {
	// StatusCode is the HTTP response status code. It is zero when no response was received.
	StatusCode int
	// RequestID is Intercom's X-Request-Id response header.
	RequestID string
	// RateLimitLimit is Intercom's X-RateLimit-Limit response header.
	RateLimitLimit string
	// RateLimitRemaining is Intercom's X-RateLimit-Remaining response header.
	RateLimitRemaining string
	// RateLimitReset is Intercom's X-RateLimit-Reset response header.
	RateLimitReset string
	// RetryAfter is the standard Retry-After response header.
	RetryAfter string
	// Duration is the time spent waiting for the HTTP transport.
	Duration time.Duration
	// Headers is a copy of the received response headers. It is nil when no response was received.
	Headers http.Header
	// Err is the transport error returned for the response attempt.
	Err error
}

// ResponseHook observes metadata for each HTTP response and transport error without changing service method signatures.
// When retries are enabled, it is called for every attempt.
type ResponseHook func(ResponseInfo)

type responseHookTransport struct {
	base http.RoundTripper
	hook ResponseHook
}

func responseHookHTTPClient(httpClient *http.Client, hook ResponseHook) *http.Client {
	clone := *httpClient
	base := clone.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	clone.Transport = &responseHookTransport{base: base, hook: hook}
	return &clone
}

func (t *responseHookTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	started := time.Now()
	res, err := t.base.RoundTrip(req)

	info := ResponseInfo{
		Duration: time.Since(started),
		Err:      err,
	}
	if res != nil {
		info.StatusCode = res.StatusCode
		info.RequestID = res.Header.Get(requestIDHeader)
		info.RateLimitLimit = res.Header.Get(rateLimitLimitHeader)
		info.RateLimitRemaining = res.Header.Get(rateLimitRemainingHeader)
		info.RateLimitReset = res.Header.Get(rateLimitResetHeader)
		info.RetryAfter = res.Header.Get("Retry-After")
		info.Headers = res.Header.Clone()
	}

	t.hook(info)
	return res, err
}
