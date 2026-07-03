package intercom

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// offlineHTTPIntegrationTestClient exercises SDK calls through a real local HTTP server.
// Tests script responses up front, then assert the captured requests after SDK calls return.
type offlineHTTPIntegrationTestClient struct {
	Client *Client

	mu        sync.Mutex
	responses []offlineHTTPResponse
	requests  []offlineHTTPRequest
}

type offlineHTTPResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

type offlineHTTPRequest struct {
	Method   string
	Path     string
	RawQuery string
	Header   http.Header
	Body     []byte
}

func newOfflineHTTPIntegrationTestClient(t *testing.T, responses ...offlineHTTPResponse) *offlineHTTPIntegrationTestClient {
	t.Helper()

	fixture := &offlineHTTPIntegrationTestClient{
		responses: append([]offlineHTTPResponse(nil), responses...),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fixture.mu.Lock()
		fixture.requests = append(fixture.requests, offlineHTTPRequest{
			Method:   req.Method,
			Path:     req.URL.Path,
			RawQuery: req.URL.RawQuery,
			Header:   req.Header.Clone(),
			Body:     body,
		})
		if len(fixture.responses) == 0 {
			fixture.mu.Unlock()
			t.Errorf("unexpected request: %s %s", req.Method, req.URL.RequestURI())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		response := fixture.responses[0]
		fixture.responses = fixture.responses[1:]
		fixture.mu.Unlock()

		for key, values := range response.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(response.StatusCode)
		if _, err := w.Write(response.Body); err != nil {
			t.Errorf("write response body: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(
		"token",
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	fixture.Client = client

	return fixture
}

func offlineJSONResponse(statusCode int, body string) offlineHTTPResponse {
	return offlineHTTPResponse{
		StatusCode: statusCode,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: []byte(body),
	}
}

func offlineBinaryResponse(statusCode int, contentType string, body []byte) offlineHTTPResponse {
	return offlineHTTPResponse{
		StatusCode: statusCode,
		Header: http.Header{
			"Content-Type": []string{contentType},
		},
		Body: append([]byte(nil), body...),
	}
}

func offlineEmptyResponse(statusCode int) offlineHTTPResponse {
	return offlineHTTPResponse{StatusCode: statusCode}
}

func (fixture *offlineHTTPIntegrationTestClient) Request(t *testing.T, index int) offlineHTTPRequest {
	t.Helper()

	fixture.mu.Lock()
	defer fixture.mu.Unlock()

	if index < 0 || index >= len(fixture.requests) {
		t.Fatalf("request index %d out of range; captured %d request(s)", index, len(fixture.requests))
	}
	return fixture.requests[index]
}

func (fixture *offlineHTTPIntegrationTestClient) RequestCount() int {
	fixture.mu.Lock()
	defer fixture.mu.Unlock()
	return len(fixture.requests)
}

func (req offlineHTTPRequest) JSONBody(t *testing.T) map[string]any {
	t.Helper()

	var body map[string]any
	if err := json.Unmarshal(req.Body, &body); err != nil {
		t.Fatalf("decode request JSON body %q: %v", string(req.Body), err)
	}
	return body
}

func TestOfflineHTTPIntegrationHarnessRoutesThroughServer(t *testing.T) {
	fixture := newOfflineHTTPIntegrationTestClient(t, offlineJSONResponse(http.StatusOK, `{"type":"admin","id":"1","email":"admin@example.com"}`))

	admin, err := fixture.Client.Admins.Me(context.Background())
	if err != nil {
		t.Fatalf("Admins.Me returned error: %v", err)
	}
	if admin.Email == nil || *admin.Email != "admin@example.com" {
		t.Fatalf("admin.Email = %v", admin.Email)
	}

	if got := fixture.RequestCount(); got != 1 {
		t.Fatalf("captured requests = %d, want 1", got)
	}
	req := fixture.Request(t, 0)
	if req.Method != http.MethodGet {
		t.Fatalf("method = %q, want %q", req.Method, http.MethodGet)
	}
	if req.Path != "/me" {
		t.Fatalf("path = %q, want /me", req.Path)
	}
	if req.RawQuery != "" {
		t.Fatalf("query = %q, want empty", req.RawQuery)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer token" {
		t.Fatalf("Authorization = %q", got)
	}
	if got := req.Header.Get("Intercom-Version"); got != defaultAPIVersion {
		t.Fatalf("Intercom-Version = %q, want %q", got, defaultAPIVersion)
	}
	if got := req.Header.Get("User-Agent"); !strings.Contains(got, defaultUserAgent) {
		t.Fatalf("User-Agent = %q, want to contain %q", got, defaultUserAgent)
	}
	if got := req.Header.Get("Accept"); got != "application/json" {
		t.Fatalf("Accept = %q, want application/json", got)
	}
}

func TestOfflineHTTPIntegrationHarnessScriptedResponses(t *testing.T) {
	tests := []struct {
		name          string
		response      offlineHTTPResponse
		method        string
		path          string
		body          string
		accept        string
		wantStatus    int
		wantBody      string
		wantErrStatus int
		checkRequest  func(*testing.T, offlineHTTPRequest)
	}{
		{
			name:       "json",
			response:   offlineJSONResponse(http.StatusOK, `{"ok":true}`),
			method:     http.MethodPost,
			path:       "/echo?starting_after=cursor-1",
			body:       `{"message":"hello"}`,
			wantStatus: http.StatusOK,
			wantBody:   `{"ok":true}`,
			checkRequest: func(t *testing.T, req offlineHTTPRequest) {
				t.Helper()
				if req.RawQuery != "starting_after=cursor-1" {
					t.Fatalf("RawQuery = %q", req.RawQuery)
				}
				body := req.JSONBody(t)
				if body["message"] != "hello" {
					t.Fatalf("message = %#v", body["message"])
				}
			},
		},
		{
			name:       "binary",
			response:   offlineBinaryResponse(http.StatusOK, "application/octet-stream", []byte("export-bytes")),
			method:     http.MethodGet,
			path:       "/download/content/data/job-1",
			accept:     "application/octet-stream",
			wantStatus: http.StatusOK,
			wantBody:   "export-bytes",
			checkRequest: func(t *testing.T, req offlineHTTPRequest) {
				t.Helper()
				if got := req.Header.Get("Accept"); got != "application/octet-stream" {
					t.Fatalf("Accept = %q", got)
				}
			},
		},
		{
			name:       "empty",
			response:   offlineEmptyResponse(http.StatusNoContent),
			method:     http.MethodDelete,
			path:       "/tags/121",
			wantStatus: http.StatusNoContent,
			wantBody:   "",
		},
		{
			name:          "api error",
			response:      offlineJSONResponse(http.StatusNotFound, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"missing"}]}`),
			method:        http.MethodGet,
			path:          "/contacts/missing",
			wantErrStatus: http.StatusNotFound,
			checkRequest: func(t *testing.T, req offlineHTTPRequest) {
				t.Helper()
				if req.Path != "/contacts/missing" {
					t.Fatalf("Path = %q", req.Path)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newOfflineHTTPIntegrationTestClient(t, tt.response)

			var body io.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			}
			req, err := fixture.Client.NewRequest(context.Background(), tt.method, tt.path, body)
			if err != nil {
				t.Fatalf("NewRequest returned error: %v", err)
			}
			if tt.accept != "" {
				req.Header.Set("Accept", tt.accept)
			}

			res, err := fixture.Client.Do(req)
			if tt.wantErrStatus != 0 {
				if err == nil {
					t.Fatal("expected error")
				}
				var apiErr *ErrorResponse
				if !errors.As(err, &apiErr) {
					t.Fatalf("error = %T, want *ErrorResponse", err)
				}
				if apiErr.StatusCode != tt.wantErrStatus {
					t.Fatalf("StatusCode = %d, want %d", apiErr.StatusCode, tt.wantErrStatus)
				}
			} else {
				if err != nil {
					t.Fatalf("Do returned error: %v", err)
				}
				defer res.Body.Close()
				if res.StatusCode != tt.wantStatus {
					t.Fatalf("status = %d, want %d", res.StatusCode, tt.wantStatus)
				}
				data, err := io.ReadAll(res.Body)
				if err != nil {
					t.Fatalf("read response body: %v", err)
				}
				if string(data) != tt.wantBody {
					t.Fatalf("body = %q, want %q", string(data), tt.wantBody)
				}
			}

			if got := fixture.RequestCount(); got != 1 {
				t.Fatalf("captured requests = %d, want 1", got)
			}
			captured := fixture.Request(t, 0)
			if captured.Method != tt.method {
				t.Fatalf("method = %q, want %q", captured.Method, tt.method)
			}
			wantPath, _, _ := strings.Cut(tt.path, "?")
			if captured.Path != wantPath {
				t.Fatalf("path = %q, want %q", captured.Path, wantPath)
			}
			if tt.checkRequest != nil {
				tt.checkRequest(t, captured)
			}
		})
	}
}
