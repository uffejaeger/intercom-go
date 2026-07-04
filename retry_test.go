package intercom

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestRetryTransportBehavior(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		body         string
		config       RetryConfig
		withRetry    bool
		responses    []retryTestResult
		wantAttempts int
		wantErr      bool
	}{
		{
			name:      "retries 429 with retry-after",
			method:    http.MethodGet,
			config:    RetryConfig{MaxAttempts: 2},
			withRetry: true,
			responses: []retryTestResult{
				{status: http.StatusTooManyRequests, headers: http.Header{"Retry-After": []string{"0"}}},
				{status: http.StatusOK, body: `{"ok":true}`},
			},
			wantAttempts: 2,
		},
		{
			name:      "retries 429 with x-ratelimit-reset",
			method:    http.MethodGet,
			config:    RetryConfig{MaxAttempts: 2},
			withRetry: true,
			responses: []retryTestResult{
				{
					status: http.StatusTooManyRequests,
					headers: http.Header{
						rateLimitResetHeader: []string{strconv.FormatInt(time.Now().Add(-time.Second).Unix(), 10)},
					},
				},
				{status: http.StatusOK, body: `{"ok":true}`},
			},
			wantAttempts: 2,
		},
		{
			name:      "retries selected 5xx",
			method:    http.MethodGet,
			config:    RetryConfig{MaxAttempts: 2, InitialBackoff: time.Nanosecond, MaxBackoff: time.Nanosecond},
			withRetry: true,
			responses: []retryTestResult{
				{status: http.StatusServiceUnavailable},
				{status: http.StatusOK, body: `{"ok":true}`},
			},
			wantAttempts: 2,
		},
		{
			name:      "retries transient network error",
			method:    http.MethodGet,
			config:    RetryConfig{MaxAttempts: 2, InitialBackoff: time.Nanosecond, MaxBackoff: time.Nanosecond},
			withRetry: true,
			responses: []retryTestResult{
				{err: temporaryNetError{}},
				{status: http.StatusOK, body: `{"ok":true}`},
			},
			wantAttempts: 2,
		},
		{
			name:   "does not retry without option",
			method: http.MethodGet,
			responses: []retryTestResult{
				{status: http.StatusServiceUnavailable},
				{status: http.StatusOK, body: `{"ok":true}`},
			},
			wantAttempts: 1,
			wantErr:      true,
		},
		{
			name:      "does not retry non-retryable status",
			method:    http.MethodGet,
			config:    RetryConfig{MaxAttempts: 2},
			withRetry: true,
			responses: []retryTestResult{
				{status: http.StatusBadRequest},
				{status: http.StatusOK, body: `{"ok":true}`},
			},
			wantAttempts: 1,
			wantErr:      true,
		},
		{
			name:      "does not retry unsafe method by default",
			method:    http.MethodPost,
			body:      `{"name":"Acme"}`,
			config:    RetryConfig{MaxAttempts: 2},
			withRetry: true,
			responses: []retryTestResult{
				{status: http.StatusServiceUnavailable},
				{status: http.StatusOK, body: `{"ok":true}`},
			},
			wantAttempts: 1,
			wantErr:      true,
		},
		{
			name:      "retries unsafe method when enabled",
			method:    http.MethodPost,
			body:      `{"name":"Acme"}`,
			config:    RetryConfig{MaxAttempts: 2, AllowUnsafeMethods: true, InitialBackoff: time.Nanosecond, MaxBackoff: time.Nanosecond},
			withRetry: true,
			responses: []retryTestResult{
				{status: http.StatusServiceUnavailable},
				{status: http.StatusOK, body: `{"ok":true}`},
			},
			wantAttempts: 2,
		},
		{
			name:      "uses custom status codes",
			method:    http.MethodGet,
			config:    RetryConfig{MaxAttempts: 2, StatusCodes: []int{http.StatusTeapot}, InitialBackoff: time.Nanosecond, MaxBackoff: time.Nanosecond},
			withRetry: true,
			responses: []retryTestResult{
				{status: http.StatusTeapot},
				{status: http.StatusOK, body: `{"ok":true}`},
			},
			wantAttempts: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attempts := 0
			bodies := make([]string, 0, len(tt.responses))
			transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.Body != nil {
					body, err := io.ReadAll(req.Body)
					if err != nil {
						t.Fatalf("read body: %v", err)
					}
					bodies = append(bodies, string(body))
				}
				result := tt.responses[attempts]
				attempts++
				if result.err != nil {
					return nil, result.err
				}
				return retryTestResponse(req, result), nil
			})

			options := []Option{
				WithBaseURL("https://example.test"),
				WithHTTPClient(&http.Client{Transport: transport}),
			}
			if tt.withRetry {
				options = append(options, WithRetry(tt.config))
			}
			client, err := NewClient("token", options...)
			if err != nil {
				t.Fatalf("NewClient returned error: %v", err)
			}

			var body io.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			}
			req, err := client.NewRequest(context.Background(), tt.method, "/test", body)
			if err != nil {
				t.Fatalf("NewRequest returned error: %v", err)
			}
			_, err = client.Do(req)
			if tt.wantErr && err == nil {
				t.Fatal("Do returned nil error, want error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("Do returned error: %v", err)
			}
			if attempts != tt.wantAttempts {
				t.Fatalf("attempts = %d, want %d", attempts, tt.wantAttempts)
			}
			if tt.body != "" && tt.wantAttempts > 1 {
				for _, got := range bodies {
					if got != tt.body {
						t.Fatalf("body = %q, want %q", got, tt.body)
					}
				}
			}
		})
	}
}

func TestRetryRespectsContextCancellationDuringBackoff(t *testing.T) {
	attempts := 0
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		attempts++
		time.AfterFunc(time.Millisecond, cancel)
		return retryTestResponse(req, retryTestResult{
			status:  http.StatusServiceUnavailable,
			headers: http.Header{"Retry-After": []string{"3600"}},
		}), nil
	})
	client, err := NewClient(
		"token",
		WithBaseURL("https://example.test"),
		WithHTTPClient(&http.Client{Transport: transport}),
		WithRetry(RetryConfig{MaxAttempts: 2}),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	req, err := client.NewRequest(ctx, http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}

	_, err = client.Do(req)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Do error = %v, want context canceled", err)
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
}

func TestWithRetryWorksBeforeWithHTTPClient(t *testing.T) {
	attempts := 0
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		attempts++
		if attempts == 1 {
			return retryTestResponse(req, retryTestResult{status: http.StatusTooManyRequests, headers: http.Header{"Retry-After": []string{"0"}}}), nil
		}
		return retryTestResponse(req, retryTestResult{status: http.StatusOK, body: `{"ok":true}`}), nil
	})
	client, err := NewClient(
		"token",
		WithRetry(RetryConfig{MaxAttempts: 2}),
		WithBaseURL("https://example.test"),
		WithHTTPClient(&http.Client{Transport: transport}),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	req, err := client.NewRequest(context.Background(), http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}

	if _, err := client.Do(req); err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
}

func TestRetryAfterDelay(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://example.test", nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}
	res := retryTestResponse(req, retryTestResult{status: http.StatusTooManyRequests})

	tests := []struct {
		name string
		set  func(http.Header)
		want bool
	}{
		{
			name: "missing",
			want: false,
		},
		{
			name: "seconds",
			set: func(header http.Header) {
				header.Set("Retry-After", "1")
			},
			want: true,
		},
		{
			name: "http date",
			set: func(header http.Header) {
				header.Set("Retry-After", time.Now().UTC().Format(http.TimeFormat))
			},
			want: true,
		},
		{
			name: "invalid",
			set: func(header http.Header) {
				header.Set("Retry-After", "later")
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res.Header = make(http.Header)
			if tt.set != nil {
				tt.set(res.Header)
			}
			_, ok := retryAfterDelay(res)
			if ok != tt.want {
				t.Fatalf("retryAfterDelay ok = %v, want %v", ok, tt.want)
			}
		})
	}
}

func TestWithRetryValidation(t *testing.T) {
	tests := []struct {
		name   string
		config RetryConfig
	}{
		{name: "negative attempts", config: RetryConfig{MaxAttempts: -1}},
		{name: "negative initial backoff", config: RetryConfig{InitialBackoff: -1}},
		{name: "negative max backoff", config: RetryConfig{MaxBackoff: -1}},
		{name: "negative jitter", config: RetryConfig{Jitter: -1}},
		{name: "jitter too large", config: RetryConfig{Jitter: 1.1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient("token", WithRetry(tt.config))
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestRetryHTTPClientDefaults(t *testing.T) {
	client := retryHTTPClient(nil, RetryConfig{})
	if client == nil {
		t.Fatal("retryHTTPClient returned nil")
	}
	if _, ok := client.Transport.(*retryTransport); !ok {
		t.Fatalf("Transport = %T, want *retryTransport", client.Transport)
	}

	base := &http.Client{}
	clone := retryHTTPClient(base, RetryConfig{})
	if clone == base {
		t.Fatal("retryHTTPClient returned original client, want clone")
	}
	if _, ok := clone.Transport.(*retryTransport); !ok {
		t.Fatalf("clone Transport = %T, want *retryTransport", clone.Transport)
	}
}

func TestNormalizeRetryConfigDefaultsAndBounds(t *testing.T) {
	config := normalizeRetryConfig(RetryConfig{
		InitialBackoff: 2 * time.Second,
		MaxBackoff:     time.Second,
	})

	if config.maxAttempts != defaultRetryMaxAttempts {
		t.Fatalf("maxAttempts = %d, want %d", config.maxAttempts, defaultRetryMaxAttempts)
	}
	if config.maxBackoff != config.initialBackoff {
		t.Fatalf("maxBackoff = %s, want initialBackoff %s", config.maxBackoff, config.initialBackoff)
	}
	if !config.statusCodes[http.StatusTooManyRequests] || !config.statusCodes[http.StatusGatewayTimeout] {
		t.Fatalf("default status codes = %#v", config.statusCodes)
	}
}

func TestRetryTransportEdgeBranches(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://example.test", nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}

	t.Run("max attempts one", func(t *testing.T) {
		attempts := 0
		transport := &retryTransport{
			base: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				attempts++
				return retryTestResponse(req, retryTestResult{status: http.StatusServiceUnavailable}), nil
			}),
			config: normalizeRetryConfig(RetryConfig{MaxAttempts: 1}),
		}
		if _, err := transport.RoundTrip(req); err != nil {
			t.Fatalf("RoundTrip returned error: %v", err)
		}
		if attempts != 1 {
			t.Fatalf("attempts = %d, want 1", attempts)
		}
	})

	t.Run("nil request cannot retry", func(t *testing.T) {
		transport := &retryTransport{config: normalizeRetryConfig(RetryConfig{})}
		if transport.canRetryRequest(nil) {
			t.Fatal("canRetryRequest(nil) = true, want false")
		}
	})

	t.Run("unsafe body without get body cannot retry", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "https://example.test", io.NopCloser(strings.NewReader("{}")))
		if err != nil {
			t.Fatalf("NewRequest returned error: %v", err)
		}
		transport := &retryTransport{config: normalizeRetryConfig(RetryConfig{AllowUnsafeMethods: true})}
		if transport.canRetryRequest(req) {
			t.Fatal("canRetryRequest = true, want false")
		}
	})

	t.Run("context canceled does not retry", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.test", nil)
		if err != nil {
			t.Fatalf("NewRequest returned error: %v", err)
		}
		transport := &retryTransport{config: normalizeRetryConfig(RetryConfig{})}
		if transport.shouldRetry(req, retryTestResponse(req, retryTestResult{status: http.StatusServiceUnavailable}), nil) {
			t.Fatal("shouldRetry = true, want false")
		}
	})

	t.Run("nil response does not retry", func(t *testing.T) {
		transport := &retryTransport{config: normalizeRetryConfig(RetryConfig{})}
		if transport.shouldRetry(req, nil, nil) {
			t.Fatal("shouldRetry = true, want false")
		}
	})

	t.Run("clone error returns previous response", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "https://example.test", strings.NewReader("{}"))
		if err != nil {
			t.Fatalf("NewRequest returned error: %v", err)
		}
		req.GetBody = func() (io.ReadCloser, error) {
			return nil, errors.New("get body failed")
		}
		transport := &retryTransport{
			base: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return retryTestResponse(req, retryTestResult{
					status:  http.StatusServiceUnavailable,
					headers: http.Header{"Retry-After": []string{"0"}},
				}), nil
			}),
			config: normalizeRetryConfig(RetryConfig{MaxAttempts: 2, AllowUnsafeMethods: true}),
		}

		res, err := transport.RoundTrip(req)

		if err == nil {
			t.Fatal("RoundTrip returned nil error, want clone error")
		}
		if res == nil || res.StatusCode != http.StatusServiceUnavailable {
			t.Fatalf("response = %#v, want service unavailable response", res)
		}
	})
}

func TestRetryDelayBounds(t *testing.T) {
	transport := &retryTransport{config: normalizeRetryConfig(RetryConfig{
		InitialBackoff: time.Nanosecond,
		MaxBackoff:     time.Nanosecond,
		Jitter:         1,
	})}

	if delay := transport.retryDelay(40, nil); delay > time.Nanosecond {
		t.Fatalf("retryDelay = %s, want capped at %s", delay, time.Nanosecond)
	}

	transport.config.jitter = 0
	if delay := transport.retryDelay(1, nil); delay != time.Nanosecond {
		t.Fatalf("retryDelay without jitter = %s, want %s", delay, time.Nanosecond)
	}
}

func TestCloneRequestForRetryErrors(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "https://example.test", strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}
	req.GetBody = func() (io.ReadCloser, error) {
		return nil, errors.New("get body failed")
	}

	if _, err := cloneRequestForRetry(req); err == nil {
		t.Fatal("cloneRequestForRetry returned nil error, want error")
	}
}

func TestRetryAfterDelayMoreBranches(t *testing.T) {
	if _, ok := retryAfterDelay(nil); ok {
		t.Fatal("retryAfterDelay(nil) ok = true, want false")
	}

	req, err := http.NewRequest(http.MethodGet, "https://example.test", nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}
	res := retryTestResponse(req, retryTestResult{status: http.StatusTooManyRequests})
	res.Header.Set("Retry-After", time.Now().Add(-time.Second).UTC().Format(http.TimeFormat))
	delay, ok := retryAfterDelay(res)
	if !ok {
		t.Fatal("retryAfterDelay past date ok = false, want true")
	}
	if delay != 0 {
		t.Fatalf("retryAfterDelay past date = %s, want 0", delay)
	}

	res.Header.Set("Retry-After", time.Now().Add(time.Minute).UTC().Format(http.TimeFormat))
	delay, ok = retryAfterDelay(res)
	if !ok {
		t.Fatal("retryAfterDelay future date ok = false, want true")
	}
	if delay <= 0 {
		t.Fatalf("retryAfterDelay future date = %s, want positive delay", delay)
	}
}

func TestRateLimitResetDelay(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://example.test", nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}

	tests := []struct {
		name      string
		response  *http.Response
		wantOK    bool
		wantDelay func(time.Duration) bool
	}{
		{
			name:   "nil response",
			wantOK: false,
		},
		{
			name: "non rate limit response",
			response: retryTestResponse(req, retryTestResult{
				status:  http.StatusServiceUnavailable,
				headers: http.Header{rateLimitResetHeader: []string{strconv.FormatInt(time.Now().Unix(), 10)}},
			}),
			wantOK: false,
		},
		{
			name:     "missing header",
			response: retryTestResponse(req, retryTestResult{status: http.StatusTooManyRequests}),
			wantOK:   false,
		},
		{
			name: "invalid header",
			response: retryTestResponse(req, retryTestResult{
				status:  http.StatusTooManyRequests,
				headers: http.Header{rateLimitResetHeader: []string{"soon"}},
			}),
			wantOK: false,
		},
		{
			name: "past reset timestamp",
			response: retryTestResponse(req, retryTestResult{
				status:  http.StatusTooManyRequests,
				headers: http.Header{rateLimitResetHeader: []string{strconv.FormatInt(time.Now().Add(-time.Second).Unix(), 10)}},
			}),
			wantOK: true,
			wantDelay: func(delay time.Duration) bool {
				return delay == 0
			},
		},
		{
			name: "future reset timestamp",
			response: retryTestResponse(req, retryTestResult{
				status:  http.StatusTooManyRequests,
				headers: http.Header{rateLimitResetHeader: []string{strconv.FormatInt(time.Now().Add(time.Minute).Unix(), 10)}},
			}),
			wantOK: true,
			wantDelay: func(delay time.Duration) bool {
				return delay > 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay, ok := rateLimitResetDelay(tt.response)
			if ok != tt.wantOK {
				t.Fatalf("rateLimitResetDelay ok = %v, want %v", ok, tt.wantOK)
			}
			if tt.wantDelay != nil && !tt.wantDelay(delay) {
				t.Fatalf("rateLimitResetDelay delay = %s", delay)
			}
		})
	}
}

func TestSleepWithContextImmediateCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.test", nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}

	if err := sleepWithContext(req, 0); !errors.Is(err, context.Canceled) {
		t.Fatalf("sleepWithContext = %v, want context canceled", err)
	}
}

func TestIsRetryableNetworkError(t *testing.T) {
	if !isRetryableNetworkError(io.EOF) {
		t.Fatal("io.EOF was not retryable")
	}
	if !isRetryableNetworkError(io.ErrUnexpectedEOF) {
		t.Fatal("io.ErrUnexpectedEOF was not retryable")
	}
	if isRetryableNetworkError(errors.New("permanent")) {
		t.Fatal("permanent error was retryable")
	}
}

func TestHeaderValue(t *testing.T) {
	tests := []struct {
		name   string
		header http.Header
		key    string
		want   string
	}{
		{
			name: "canonical header",
			header: http.Header{
				"Retry-After": []string{"1"},
			},
			key:  "Retry-After",
			want: "1",
		},
		{
			name: "non canonical intercom header",
			header: http.Header{
				rateLimitResetHeader: []string{"123"},
			},
			key:  rateLimitResetHeader,
			want: "123",
		},
		{
			name: "missing header",
			header: http.Header{
				"Other": []string{"value"},
			},
			key: rateLimitResetHeader,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := headerValue(tt.header, tt.key); got != tt.want {
				t.Fatalf("headerValue = %q, want %q", got, tt.want)
			}
		})
	}
}

type retryTestResult struct {
	status  int
	headers http.Header
	body    string
	err     error
}

func retryTestResponse(req *http.Request, result retryTestResult) *http.Response {
	if result.status == 0 {
		result.status = http.StatusOK
	}
	if result.body == "" {
		result.body = `{}`
	}
	if result.headers == nil {
		result.headers = make(http.Header)
	}
	return &http.Response{
		StatusCode: result.status,
		Header:     result.headers,
		Body:       io.NopCloser(strings.NewReader(result.body)),
		Request:    req,
	}
}

type temporaryNetError struct{}

func (temporaryNetError) Error() string {
	return "temporary network error"
}

func (temporaryNetError) Timeout() bool {
	return false
}

func (temporaryNetError) Temporary() bool {
	return true
}
