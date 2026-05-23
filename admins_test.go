package intercom

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func newAdminsTestClient(t *testing.T, transport http.RoundTripper) *Client {
	t.Helper()
	client, err := NewClient(
		"token",
		WithBaseURL("https://example.test"),
		WithHTTPClient(&http.Client{Transport: transport}),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	return client
}

func TestAdminsMe(t *testing.T) {
	var method, path, authorization, version, userAgent, accept string

	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		method = req.Method
		path = req.URL.Path
		authorization = req.Header.Get("Authorization")
		version = req.Header.Get("Intercom-Version")
		userAgent = req.Header.Get("User-Agent")
		accept = req.Header.Get("Accept")

		return &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body:    io.NopCloser(strings.NewReader(`{"type":"admin","id":"991267459","email":"admin@example.com","name":"Admin User"}`)),
			Request: req,
		}, nil
	})

	client, err := NewClient(
		"token",
		WithBaseURL("https://example.test"),
		WithHTTPClient(&http.Client{Transport: transport}),
		WithUserAgent("test-agent"),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	admin, err := client.Admins.Me(context.Background())
	if err != nil {
		t.Fatalf("Me returned error: %v", err)
	}

	if method != http.MethodGet {
		t.Fatalf("method = %q", method)
	}
	if path != "/me" {
		t.Fatalf("path = %q", path)
	}
	if authorization != "Bearer token" {
		t.Fatalf("Authorization = %q", authorization)
	}
	if version != defaultAPIVersion {
		t.Fatalf("Intercom-Version = %q", version)
	}
	if userAgent != "test-agent" {
		t.Fatalf("User-Agent = %q", userAgent)
	}
	if accept != "application/json" {
		t.Fatalf("Accept = %q", accept)
	}
	if admin.Id == nil || *admin.Id != "991267459" {
		t.Fatalf("admin.Id = %v", admin.Id)
	}
	if admin.Email == nil || *admin.Email != "admin@example.com" {
		t.Fatalf("admin.Email = %v", admin.Email)
	}
}

func TestAdminsMeError(t *testing.T) {
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body:    io.NopCloser(strings.NewReader(`{"type":"error.list","request_id":"req-1","errors":[{"code":"unauthorized","message":"Access Token Invalid"}]}`)),
			Request: req,
		}, nil
	})

	client, err := NewClient(
		"token",
		WithBaseURL("https://example.test"),
		WithHTTPClient(&http.Client{Transport: transport}),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.Admins.Me(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	apiErr, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("error type = %T, want *ErrorResponse", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("StatusCode = %d", apiErr.StatusCode)
	}
	if apiErr.RequestID != "req-1" {
		t.Fatalf("RequestID = %q", apiErr.RequestID)
	}
}

func TestAdminsMeTransportError(t *testing.T) {
	transport := roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("network down")
	})

	client, err := NewClient(
		"token",
		WithBaseURL("https://example.test"),
		WithHTTPClient(&http.Client{Transport: transport}),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.Admins.Me(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAdminsServiceRequests(t *testing.T) {
	errBody := `{"type":"error.list","errors":[{"code":"unauthorized"}],"request_id":"12a938a3-314e-4939-b773-5cd45738bd21"}`

	tests := []struct {
		name       string
		response   string
		call       func(context.Context, *Client) error
		wantMethod string
		wantPath   string
		wantBody   func(t *testing.T, body map[string]any)
	}{
		{
			name:     "list admins",
			response: `{"type":"admin.list","admins":[]}`,
			call: func(ctx context.Context, client *Client) error {
				list, err := client.Admins.List(ctx)
				if err != nil {
					return err
				}
				if list.Admins == nil {
					t.Fatal("Admins is nil")
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/admins",
		},
		{
			name:     "retrieve admin",
			response: `{"type":"admin","id":"42","name":"Alice"}`,
			call: func(ctx context.Context, client *Client) error {
				a, err := client.Admins.Retrieve(ctx, "42")
				if err != nil {
					return err
				}
				if a.Id == nil || *a.Id != "42" {
					t.Fatalf("Id = %v", a.Id)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/admins/42",
		},
		{
			name:     "set away admin",
			response: `{"type":"admin","id":"42","name":"Alice"}`,
			call: func(ctx context.Context, client *Client) error {
				a, err := client.Admins.SetAway(ctx, "42", AdminSetAway{AwayModeEnabled: true, AwayModeReassign: false})
				if err != nil {
					return err
				}
				if a.Id == nil || *a.Id != "42" {
					t.Fatalf("Id = %v", a.Id)
				}
				return nil
			},
			wantMethod: http.MethodPut,
			wantPath:   "/admins/42/away",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if v, ok := body["away_mode_enabled"].(bool); !ok || !v {
					t.Fatalf("away_mode_enabled = %v", body["away_mode_enabled"])
				}
			},
		},
		{
			name:     "list activity logs",
			response: `{"type":"activity_log.list","activity_logs":[]}`,
			call: func(ctx context.Context, client *Client) error {
				logs, err := client.Admins.ListActivityLogs(ctx, "1700000000")
				if err != nil {
					return err
				}
				_ = logs
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/admins/activity_logs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod, gotPath string
			transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				gotMethod = req.Method
				gotPath = req.URL.Path
				if tt.wantBody != nil {
					var body map[string]any
					if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
						t.Fatalf("decode body: %v", err)
					}
					tt.wantBody(t, body)
				}
				return jsonResponse(req, http.StatusOK, tt.response), nil
			})
			client := newAdminsTestClient(t, transport)
			if err := tt.call(context.Background(), client); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotMethod != tt.wantMethod {
				t.Fatalf("method = %q, want %q", gotMethod, tt.wantMethod)
			}
			if gotPath != tt.wantPath {
				t.Fatalf("path = %q, want %q", gotPath, tt.wantPath)
			}
		})
	}

	t.Run("api errors", func(t *testing.T) {
		calls := []struct {
			name string
			call func(context.Context, *Client) error
		}{
			{"list", func(ctx context.Context, c *Client) error { _, err := c.Admins.List(ctx); return err }},
			{"retrieve", func(ctx context.Context, c *Client) error { _, err := c.Admins.Retrieve(ctx, "42"); return err }},
			{"set away", func(ctx context.Context, c *Client) error {
				_, err := c.Admins.SetAway(ctx, "42", AdminSetAway{})
				return err
			}},
			{"list activity logs", func(ctx context.Context, c *Client) error {
				_, err := c.Admins.ListActivityLogs(ctx, "1700000000")
				return err
			}},
		}
		for _, tt := range calls {
			t.Run(tt.name, func(t *testing.T) {
				client := newAdminsTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
					return jsonResponse(req, http.StatusUnauthorized, errBody), nil
				}))
				if err := tt.call(context.Background(), client); err == nil {
					t.Fatal("expected error")
				}
			})
		}
	})

	t.Run("transport errors", func(t *testing.T) {
		transportErr := errors.New("connection refused")
		calls := []struct {
			name string
			call func(context.Context, *Client) error
		}{
			{"list", func(ctx context.Context, c *Client) error { _, err := c.Admins.List(ctx); return err }},
			{"retrieve", func(ctx context.Context, c *Client) error { _, err := c.Admins.Retrieve(ctx, "42"); return err }},
			{"set away", func(ctx context.Context, c *Client) error {
				_, err := c.Admins.SetAway(ctx, "42", AdminSetAway{})
				return err
			}},
			{"list activity logs", func(ctx context.Context, c *Client) error {
				_, err := c.Admins.ListActivityLogs(ctx, "1700000000")
				return err
			}},
		}
		for _, tt := range calls {
			t.Run(tt.name, func(t *testing.T) {
				client := newAdminsTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
					return nil, transportErr
				}))
				if err := tt.call(context.Background(), client); err == nil {
					t.Fatal("expected transport error")
				}
			})
		}
	})
}

func TestAdminsValidation(t *testing.T) {
	client := newAdminsTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		t.Fatal("unexpected HTTP request")
		return nil, nil
	}))
	ctx := context.Background()

	tests := []struct {
		name string
		call func() error
	}{
		{"retrieve: empty ID", func() error { _, err := client.Admins.Retrieve(ctx, ""); return err }},
		{"retrieve: partial-numeric ID", func() error { _, err := client.Admins.Retrieve(ctx, "42abc"); return err }},
		{"set away: empty ID", func() error { _, err := client.Admins.SetAway(ctx, "", AdminSetAway{}); return err }},
		{"set away: partial-numeric ID", func() error {
			_, err := client.Admins.SetAway(ctx, "42abc", AdminSetAway{})
			return err
		}},
		{"list activity logs: empty timestamp", func() error {
			_, err := client.Admins.ListActivityLogs(ctx, "")
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
