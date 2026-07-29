package intercom

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestRequestOptionsApplyToGeneratedRequest(t *testing.T) {
	headers := http.Header{
		"X-Custom-Header": []string{"one", "two"},
		"Accept":          []string{"application/vnd.intercom.test+json"},
	}
	query := url.Values{
		"custom":   []string{"first", "second"},
		"per_page": []string{"25"},
	}
	ctx, err := WithRequestOptions(context.Background(), RequestOptions{
		Headers: headers,
		Query:   query,
	})
	if err != nil {
		t.Fatalf("WithRequestOptions returned error: %v", err)
	}
	headers.Set("X-Custom-Header", "changed")
	query.Set("custom", "changed")

	client, err := NewClient(
		"token",
		WithBaseURL("https://example.test"),
		WithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if got := req.Header.Values("X-Custom-Header"); len(got) != 2 || got[0] != "one" || got[1] != "two" {
				t.Fatalf("X-Custom-Header = %#v", got)
			}
			if got := req.Header.Get("Accept"); got != "application/vnd.intercom.test+json" {
				t.Fatalf("Accept = %q", got)
			}
			if got := req.URL.Query()["custom"]; len(got) != 2 || got[0] != "first" || got[1] != "second" {
				t.Fatalf("custom query = %#v", got)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"type":"admin","id":"1"}`)),
				Request:    req,
			}, nil
		})}),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	if _, err := client.Admins.Me(ctx); err != nil {
		t.Fatalf("Admins.Me returned error: %v", err)
	}
}

func TestRequestOptionsApplyToClientDo(t *testing.T) {
	ctx, err := WithRequestOptions(context.Background(), RequestOptions{
		Headers: http.Header{"Authorization": []string{"Bearer request-token"}},
		Query:   url.Values{"existing": nil},
	})
	if err != nil {
		t.Fatalf("WithRequestOptions returned error: %v", err)
	}

	client, err := NewClient(
		"token",
		WithBaseURL("https://example.test"),
		WithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if got := req.Header.Get("Authorization"); got != "Bearer request-token" {
				t.Fatalf("Authorization = %q", got)
			}
			if got := req.URL.Query().Get("existing"); got != "" {
				t.Fatalf("existing query = %q", got)
			}
			return &http.Response{
				StatusCode: http.StatusNoContent,
				Header:     make(http.Header),
				Body:       http.NoBody,
				Request:    req,
			}, nil
		})}),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	req, err := client.NewRequest(ctx, http.MethodGet, "/test?existing=value", nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}

	res, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	res.Body.Close()
}

func TestRequestOptionsRetryOverride(t *testing.T) {
	tests := []struct {
		name         string
		clientRetry  *RetryConfig
		requestRetry RetryConfig
		wantAttempts int
	}{
		{
			name:         "enable retries for one request",
			requestRetry: RetryConfig{MaxAttempts: 2, InitialBackoff: time.Nanosecond, MaxBackoff: time.Nanosecond},
			wantAttempts: 2,
		},
		{
			name:         "disable client retries for one request",
			clientRetry:  &RetryConfig{MaxAttempts: 2, InitialBackoff: time.Nanosecond, MaxBackoff: time.Nanosecond},
			requestRetry: RetryConfig{MaxAttempts: 1},
			wantAttempts: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attempts := 0
			options := []Option{
				WithBaseURL("https://example.test"),
				WithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					attempts++
					status := http.StatusServiceUnavailable
					if attempts > 1 {
						status = http.StatusNoContent
					}
					return &http.Response{StatusCode: status, Header: make(http.Header), Body: http.NoBody, Request: req}, nil
				})}),
			}
			if tt.clientRetry != nil {
				options = append(options, WithRetry(*tt.clientRetry))
			}
			client, err := NewClient("token", options...)
			if err != nil {
				t.Fatalf("NewClient returned error: %v", err)
			}
			ctx, err := WithRequestOptions(context.Background(), RequestOptions{Retry: &tt.requestRetry})
			if err != nil {
				t.Fatalf("WithRequestOptions returned error: %v", err)
			}
			req, err := client.NewRequest(ctx, http.MethodGet, "/test", nil)
			if err != nil {
				t.Fatalf("NewRequest returned error: %v", err)
			}

			_, _ = client.Do(req)
			if attempts != tt.wantAttempts {
				t.Fatalf("attempts = %d, want %d", attempts, tt.wantAttempts)
			}
		})
	}
}

func TestWithRequestOptionsValidationAndCopies(t *testing.T) {
	var nilContext context.Context
	if _, err := WithRequestOptions(nilContext, RequestOptions{}); err == nil {
		t.Fatal("nil context returned nil error")
	}
	if _, err := WithRequestOptions(context.Background(), RequestOptions{
		Retry: &RetryConfig{MaxAttempts: -1},
	}); err == nil {
		t.Fatal("invalid retry returned nil error")
	}

	retry := &RetryConfig{MaxAttempts: 2, StatusCodes: []int{http.StatusTeapot}}
	ctx, err := WithRequestOptions(context.Background(), RequestOptions{Retry: retry})
	if err != nil {
		t.Fatalf("WithRequestOptions returned error: %v", err)
	}
	retry.StatusCodes[0] = http.StatusOK
	options, ok := requestOptionsFromContext(ctx)
	if !ok || options.Retry == nil || options.Retry.StatusCodes[0] != http.StatusTeapot {
		t.Fatalf("request options = %#v", options)
	}

	empty, err := WithRequestOptions(context.Background(), RequestOptions{})
	if err != nil {
		t.Fatalf("WithRequestOptions returned error: %v", err)
	}
	if _, ok := requestOptionsFromContext(empty); !ok {
		t.Fatal("empty request options missing from context")
	}
	if _, ok := requestOptionsFromContext(nilContext); ok {
		t.Fatal("nil context returned request options")
	}
}

func TestApplyRequestOptionsNilRequest(t *testing.T) {
	applyRequestOptions(context.Background(), nil)
}

func TestRequestOptionsRetryNetworkError(t *testing.T) {
	attempts := 0
	client, err := NewClient(
		"token",
		WithBaseURL("https://example.test"),
		WithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			attempts++
			if attempts == 1 {
				return nil, timeoutNetError{}
			}
			return &http.Response{StatusCode: http.StatusNoContent, Header: make(http.Header), Body: http.NoBody, Request: req}, nil
		})}),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	ctx, err := WithRequestOptions(context.Background(), RequestOptions{
		Retry: &RetryConfig{MaxAttempts: 2, InitialBackoff: time.Nanosecond, MaxBackoff: time.Nanosecond},
	})
	if err != nil {
		t.Fatalf("WithRequestOptions returned error: %v", err)
	}
	req, err := client.NewRequest(ctx, http.MethodGet, "/test", nil)
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
