package intercom

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestNewClientRequiresToken(t *testing.T) {
	_, err := NewClient(" ")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewClientNilOption(t *testing.T) {
	client, err := NewClient("token", nil, nil)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestWithRegion(t *testing.T) {
	tests := []struct {
		name    string
		region  Region
		wantURL string
		wantErr bool
	}{
		{name: "us", region: US, wantURL: "https://api.intercom.io"},
		{name: "eu", region: EU, wantURL: "https://api.eu.intercom.io"},
		{name: "au", region: AU, wantURL: "https://api.au.intercom.io"},
		{name: "unknown", region: Region("moon"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient("token", WithRegion(tt.region))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("NewClient returned error: %v", err)
			}

			if got := client.BaseURL(); got != tt.wantURL {
				t.Fatalf("BaseURL() = %q, want %q", got, tt.wantURL)
			}
		})
	}
}

func TestNewClientOptionErrors(t *testing.T) {
	tests := []struct {
		name string
		opt  Option
	}{
		{name: "nil http client", opt: WithHTTPClient(nil)},
		{name: "nil response hook", opt: WithResponseHook(nil)},
		{name: "empty base URL", opt: WithBaseURL(" ")},
		{name: "empty API version", opt: WithAPIVersion(" ")},
		{name: "empty user agent", opt: WithUserAgent(" ")},
		{name: "custom option", opt: func(*Client) error { return errors.New("boom") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient("token", tt.opt)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestResponseHookObservesGeneratedServiceResponse(t *testing.T) {
	var infos []ResponseInfo
	headers := make(http.Header)
	headers.Set(requestIDHeader, "req-123")
	headers.Set(rateLimitLimitHeader, "10000")
	headers.Set(rateLimitRemainingHeader, "9999")
	headers.Set(rateLimitResetHeader, "1735689600")
	headers.Set("Retry-After", "1")
	headers.Set("Content-Type", "application/json")
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     headers,
			Body:       io.NopCloser(strings.NewReader(`{"type":"admin","id":"1","email":"admin@example.com"}`)),
			Request:    req,
		}, nil
	})
	client, err := NewClient(
		"token",
		WithBaseURL("https://example.test"),
		WithHTTPClient(&http.Client{Transport: transport}),
		WithResponseHook(func(info ResponseInfo) {
			infos = append(infos, info)
		}),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	if _, err := client.Admins.Me(context.Background()); err != nil {
		t.Fatalf("Admins.Me returned error: %v", err)
	}

	if len(infos) != 1 {
		t.Fatalf("hook calls = %d, want 1", len(infos))
	}
	info := infos[0]
	if info.StatusCode != http.StatusOK || info.RequestID != "req-123" {
		t.Fatalf("response info = %#v", info)
	}
	if info.RateLimitLimit != "10000" || info.RateLimitRemaining != "9999" || info.RateLimitReset != "1735689600" || info.RetryAfter != "1" {
		t.Fatalf("rate limit info = %#v", info)
	}
	if info.Duration < 0 {
		t.Fatalf("Duration = %s, want non-negative", info.Duration)
	}

	headers.Set(requestIDHeader, "changed")
	if got := info.Headers.Get(requestIDHeader); got != "req-123" {
		t.Fatalf("Headers = %q, want req-123", got)
	}
}

func TestResponseHookObservesClientDoErrorResponse(t *testing.T) {
	var info ResponseInfo
	headers := make(http.Header)
	headers.Set(requestIDHeader, "req-rate-limit")
	headers.Set(rateLimitResetHeader, "1735689600")
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Header:     headers,
			Body:       io.NopCloser(strings.NewReader(`{"type":"error.list","errors":[{"code":"rate_limit_exceeded"}]}`)),
			Request:    req,
		}, nil
	})
	client, err := NewClient(
		"token",
		WithBaseURL("https://example.test"),
		WithHTTPClient(&http.Client{Transport: transport}),
		WithResponseHook(func(got ResponseInfo) {
			info = got
		}),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	req, err := client.NewRequest(context.Background(), http.MethodGet, "/me", nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}
	if _, err := client.Do(req); err == nil {
		t.Fatal("expected API error")
	}

	if info.StatusCode != http.StatusTooManyRequests || info.RequestID != "req-rate-limit" || info.RateLimitReset != "1735689600" {
		t.Fatalf("response info = %#v", info)
	}
}

func TestResponseHookObservesTransportError(t *testing.T) {
	wantErr := errors.New("network down")
	var info ResponseInfo
	client, err := NewClient(
		"token",
		WithBaseURL("https://example.test"),
		WithHTTPClient(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, wantErr
		})}),
		WithResponseHook(func(got ResponseInfo) {
			info = got
		}),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	req, err := client.NewRequest(context.Background(), http.MethodGet, "/me", nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}
	if _, err := client.Do(req); !errors.Is(err, wantErr) {
		t.Fatalf("Do error = %v, want %v", err, wantErr)
	}

	if !errors.Is(info.Err, wantErr) || info.StatusCode != 0 || info.Headers != nil {
		t.Fatalf("response info = %#v", info)
	}
}

func TestResponseHookObservesEachRetryResponse(t *testing.T) {
	responses := []*http.Response{
		{
			StatusCode: http.StatusServiceUnavailable,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("unavailable")),
		},
		{
			StatusCode: http.StatusNoContent,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("")),
		},
	}
	var infos []ResponseInfo
	client, err := NewClient(
		"token",
		WithBaseURL("https://example.test"),
		WithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			res := responses[0]
			responses = responses[1:]
			res.Request = req
			return res, nil
		})}),
		WithRetry(RetryConfig{MaxAttempts: 2, InitialBackoff: time.Nanosecond, MaxBackoff: time.Nanosecond, Jitter: 0}),
		WithResponseHook(func(info ResponseInfo) {
			infos = append(infos, info)
		}),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	req, err := client.NewRequest(context.Background(), http.MethodGet, "/me", nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}
	res, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	res.Body.Close()

	if len(infos) != 2 || infos[0].StatusCode != http.StatusServiceUnavailable || infos[1].StatusCode != http.StatusNoContent {
		t.Fatalf("response infos = %#v, want retry and final responses", infos)
	}
}

func TestResponseHookHTTPClientDefaultsTransport(t *testing.T) {
	client := responseHookHTTPClient(&http.Client{}, func(ResponseInfo) {})
	transport, ok := client.Transport.(*responseHookTransport)
	if !ok {
		t.Fatalf("Transport = %T, want *responseHookTransport", client.Transport)
	}
	if transport.base != http.DefaultTransport {
		t.Fatalf("base transport = %T, want http.DefaultTransport", transport.base)
	}
}

func TestNewClientFromEnv(t *testing.T) {
	t.Setenv(DefaultAccessTokenEnv, "token")

	client, err := NewClientFromEnv(WithRegion(EU))
	if err != nil {
		t.Fatalf("NewClientFromEnv returned error: %v", err)
	}

	if got, want := client.BaseURL(), "https://api.eu.intercom.io"; got != want {
		t.Fatalf("BaseURL() = %q, want %q", got, want)
	}
}

func TestNewClientFromEnvOverride(t *testing.T) {
	const tokenEnv = "CUSTOM_INTERCOM_ACCESS_TOKEN"
	t.Setenv(DefaultAccessTokenEnv, "")
	t.Setenv(tokenEnv, "token")

	client, err := NewClientFromEnvVar(tokenEnv)
	if err != nil {
		t.Fatalf("NewClientFromEnvVar returned error: %v", err)
	}

	if client == nil {
		t.Fatal("client is nil")
	}
}

func TestNewClientFromEnvRequiresToken(t *testing.T) {
	t.Setenv(DefaultAccessTokenEnv, "")

	_, err := NewClientFromEnv()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewClientFromEnvVarRequiresName(t *testing.T) {
	t.Setenv(DefaultAccessTokenEnv, "token")

	_, err := NewClientFromEnvVar(" ")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewClientFromEnvAppliesOptionsOnceToInitializedClient(t *testing.T) {
	t.Setenv(DefaultAccessTokenEnv, "token")

	calls := 0
	_, err := NewClientFromEnv(func(client *Client) error {
		calls++
		if client.httpClient == nil {
			t.Fatal("httpClient is nil")
		}
		if client.baseURL == "" {
			t.Fatal("baseURL is empty")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("NewClientFromEnv returned error: %v", err)
	}

	if calls != 1 {
		t.Fatalf("option calls = %d, want 1", calls)
	}
}

func TestDoAddsHeaders(t *testing.T) {
	var authorization, version, userAgent, accept string

	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		authorization = req.Header.Get("Authorization")
		version = req.Header.Get("Intercom-Version")
		userAgent = req.Header.Get("User-Agent")
		accept = req.Header.Get("Accept")
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})

	client, err := NewClient(
		"token",
		WithBaseURL("https://example.test"),
		WithHTTPClient(&http.Client{Transport: transport}),
		WithAPIVersion("2.14"),
		WithUserAgent("test-agent"),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	req, err := client.NewRequest(context.Background(), http.MethodGet, "/me", nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}

	res, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	defer res.Body.Close()

	if authorization != "Bearer token" {
		t.Fatalf("Authorization = %q", authorization)
	}
	if version != "2.14" {
		t.Fatalf("Intercom-Version = %q", version)
	}
	if userAgent != "test-agent" {
		t.Fatalf("User-Agent = %q", userAgent)
	}
	if accept != "application/json" {
		t.Fatalf("Accept = %q", accept)
	}
}

func TestDoErrors(t *testing.T) {
	tests := []struct {
		name      string
		req       *http.Request
		transport http.RoundTripper
	}{
		{
			name: "nil request",
		},
		{
			name: "transport error",
			req:  mustRequest(t, http.MethodGet, "https://example.test/me", nil),
			transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return nil, errors.New("network down")
			}),
		},
		{
			name: "error response read failure",
			req:  mustRequest(t, http.MethodGet, "https://example.test/me", nil),
			transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       errorReadCloser{},
					Request:    req,
				}, nil
			}),
		},
		{
			name: "api error response",
			req:  mustRequest(t, http.MethodGet, "https://example.test/me", nil),
			transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return jsonResponse(req, http.StatusUnauthorized, `{"type":"error.list","errors":[{"code":"unauthorized"}]}`), nil
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(
				"token",
				WithBaseURL("https://example.test"),
				WithHTTPClient(&http.Client{Transport: tt.transport}),
			)
			if err != nil {
				t.Fatalf("NewClient returned error: %v", err)
			}

			res, err := client.Do(tt.req)
			if err == nil {
				if res != nil {
					res.Body.Close()
				}
				t.Fatal("expected error")
			}
		})
	}
}

func TestNewRequest(t *testing.T) {
	tests := []struct {
		name            string
		path            string
		body            io.Reader
		wantURL         string
		wantContentType string
	}{
		{
			name:    "adds leading slash",
			path:    "contacts",
			wantURL: "https://example.test/contacts",
		},
		{
			name:            "sets json content type for body",
			path:            "/contacts/search",
			body:            strings.NewReader(`{}`),
			wantURL:         "https://example.test/contacts/search",
			wantContentType: "application/json",
		},
	}

	client, err := NewClient("token", WithBaseURL("https://example.test/"))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := client.NewRequest(context.Background(), http.MethodPost, tt.path, tt.body)
			if err != nil {
				t.Fatalf("NewRequest returned error: %v", err)
			}

			if got := req.URL.String(); got != tt.wantURL {
				t.Fatalf("URL = %q, want %q", got, tt.wantURL)
			}
			if got := req.Header.Get("Content-Type"); got != tt.wantContentType {
				t.Fatalf("Content-Type = %q, want %q", got, tt.wantContentType)
			}
		})
	}
}

func TestNewRequestInvalidURL(t *testing.T) {
	client := &Client{baseURL: "://"}

	_, err := client.NewRequest(context.Background(), http.MethodGet, "/me", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

type errorReadCloser struct{}

func (errorReadCloser) Read([]byte) (int, error) {
	return 0, errors.New("read failed")
}

func (errorReadCloser) Close() error {
	return nil
}

func mustRequest(t *testing.T, method, url string, body io.Reader) *http.Request {
	t.Helper()

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}
	return req
}
