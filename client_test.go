package intercom

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestNewClientRequiresToken(t *testing.T) {
	_, err := NewClient(" ")
	if err == nil {
		t.Fatal("expected error")
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
