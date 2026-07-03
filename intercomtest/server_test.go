package intercomtest

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	intercom "github.com/uffejaeger/intercom-go"
)

func TestServerRoutesSDKCallsAndCapturesRequests(t *testing.T) {
	srv := NewServer(t,
		Route(http.MethodGet, "/me", JSON(http.StatusOK, `{"type":"admin","id":"admin-1","email":"admin@example.com"}`)),
	)

	client, err := srv.Client("token", intercom.WithAPIVersion("2.14"), intercom.WithUserAgent("test-agent"))
	if err != nil {
		t.Fatalf("Client returned error: %v", err)
	}

	admin, err := client.Admins.Me(context.Background())
	if err != nil {
		t.Fatalf("Admins.Me returned error: %v", err)
	}
	if admin.Email == nil || *admin.Email != "admin@example.com" {
		t.Fatalf("admin.Email = %v, want admin@example.com", admin.Email)
	}

	if got := srv.RequestCount(); got != 1 {
		t.Fatalf("RequestCount() = %d, want 1", got)
	}
	req := srv.Request(t, 0)
	if req.Method != http.MethodGet {
		t.Fatalf("Method = %q, want %q", req.Method, http.MethodGet)
	}
	if req.Path != "/me" {
		t.Fatalf("Path = %q, want /me", req.Path)
	}
	if req.RawQuery != "" {
		t.Fatalf("RawQuery = %q, want empty", req.RawQuery)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer token" {
		t.Fatalf("Authorization = %q, want Bearer token", got)
	}
	if got := req.Header.Get("Intercom-Version"); got != "2.14" {
		t.Fatalf("Intercom-Version = %q, want 2.14", got)
	}
	if got := req.Header.Get("User-Agent"); got != "test-agent" {
		t.Fatalf("User-Agent = %q, want test-agent", got)
	}
}

func TestServerScriptedResponses(t *testing.T) {
	tests := []struct {
		name       string
		response   Response
		method     string
		path       string
		body       string
		wantStatus int
		wantBody   string
		check      func(*testing.T, Request)
	}{
		{
			name:       "json",
			response:   JSON(http.StatusOK, map[string]any{"ok": true}),
			method:     http.MethodPost,
			path:       "/echo?starting_after=cursor-1",
			body:       `{"message":"hello"}`,
			wantStatus: http.StatusOK,
			wantBody:   `{"ok":true}`,
			check: func(t *testing.T, req Request) {
				t.Helper()
				if req.RawQuery != "starting_after=cursor-1" {
					t.Fatalf("RawQuery = %q, want starting_after=cursor-1", req.RawQuery)
				}
				body := req.JSONBody(t)
				if body["message"] != "hello" {
					t.Fatalf("message = %#v, want hello", body["message"])
				}
			},
		},
		{
			name:       "bytes",
			response:   Bytes(http.StatusOK, "application/octet-stream", []byte("export-bytes")),
			method:     http.MethodGet,
			path:       "/download/content/data/job-1",
			wantStatus: http.StatusOK,
			wantBody:   "export-bytes",
		},
		{
			name:       "no content",
			response:   NoContent(http.StatusNoContent),
			method:     http.MethodDelete,
			path:       "/tags/tag-1",
			wantStatus: http.StatusNoContent,
			wantBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedURL, err := url.Parse(tt.path)
			if err != nil {
				t.Fatalf("parse path: %v", err)
			}

			srv := NewServer(t, Route(tt.method, parsedURL.Path, tt.response))
			client, err := srv.Client("token")
			if err != nil {
				t.Fatalf("Client returned error: %v", err)
			}

			var body io.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			}
			req, err := client.NewRequest(context.Background(), tt.method, tt.path, body)
			if err != nil {
				t.Fatalf("NewRequest returned error: %v", err)
			}
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			res, err := client.Do(req)
			if err != nil {
				t.Fatalf("Do returned error: %v", err)
			}
			defer res.Body.Close()
			if res.StatusCode != tt.wantStatus {
				t.Fatalf("StatusCode = %d, want %d", res.StatusCode, tt.wantStatus)
			}
			responseBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("read response body: %v", err)
			}
			if string(responseBody) != tt.wantBody {
				t.Fatalf("body = %q, want %q", string(responseBody), tt.wantBody)
			}

			captured := srv.Request(t, 0)
			if tt.check != nil {
				tt.check(t, captured)
			}
		})
	}
}

func TestServerIntercomErrorResponse(t *testing.T) {
	srv := NewServer(t,
		Route(http.MethodGet, "/contacts/missing", Error(http.StatusNotFound, "req-1", "not_found", "missing")),
	)
	client, err := srv.Client("token")
	if err != nil {
		t.Fatalf("Client returned error: %v", err)
	}

	_, err = client.Contacts.Get(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}

	var apiErr *intercom.ErrorResponse
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %T, want *intercom.ErrorResponse", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Fatalf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusNotFound)
	}
	if apiErr.RequestID != "req-1" {
		t.Fatalf("RequestID = %q, want req-1", apiErr.RequestID)
	}
	if len(apiErr.Errors) != 1 || apiErr.Errors[0].Code != "not_found" {
		t.Fatalf("Errors = %#v, want not_found", apiErr.Errors)
	}
}

func TestServerRequestsReturnsCopies(t *testing.T) {
	srv := NewServer(t,
		Route(http.MethodGet, "/me", JSON(http.StatusOK, `{"type":"admin","id":"admin-1"}`)),
	)
	client, err := srv.Client("token")
	if err != nil {
		t.Fatalf("Client returned error: %v", err)
	}
	if _, err := client.Admins.Me(context.Background()); err != nil {
		t.Fatalf("Admins.Me returned error: %v", err)
	}

	requests := srv.Requests()
	requests[0].Header.Set("Authorization", "mutated")

	req := srv.Request(t, 0)
	if got := req.Header.Get("Authorization"); got != "Bearer token" {
		t.Fatalf("Authorization = %q, want Bearer token", got)
	}
}
