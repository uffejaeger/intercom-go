package intercom

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestOfflineSDKIntegrationRepresentativeFlows(t *testing.T) {
	tests := []struct {
		name         string
		responses    []offlineHTTPResponse
		call         func(context.Context, *Client) error
		checkRequest func(*testing.T, offlineHTTPRequest)
	}{
		{
			name: "contact search sends json body and decodes list response",
			responses: []offlineHTTPResponse{
				offlineJSONResponse(http.StatusOK, `{"type":"list","data":[{"type":"contact","id":"contact-1","email":"customer@example.com"}],"total_count":1}`),
			},
			call: func(ctx context.Context, client *Client) error {
				contacts, err := client.Contacts.Search(ctx, ContactSearch{
					Field:         "email",
					Operator:      ContactSearchEquals,
					Value:         "customer@example.com",
					PerPage:       25,
					StartingAfter: "cursor-1",
				})
				if err != nil {
					return err
				}
				if contacts.TotalCount == nil || *contacts.TotalCount != 1 {
					t.Fatalf("contacts.TotalCount = %v, want 1", contacts.TotalCount)
				}
				if contacts.Data == nil || len(*contacts.Data) != 1 {
					t.Fatalf("contacts.Data = %#v, want 1 contact", contacts.Data)
				}
				contact := (*contacts.Data)[0]
				if contact.Id == nil || *contact.Id != "contact-1" {
					t.Fatalf("contact.Id = %v, want contact-1", contact.Id)
				}
				if contact.Email == nil || *contact.Email != "customer@example.com" {
					t.Fatalf("contact.Email = %v, want customer@example.com", contact.Email)
				}
				return nil
			},
			checkRequest: func(t *testing.T, req offlineHTTPRequest) {
				t.Helper()
				assertOfflineSDKIntegrationRequest(t, req, http.MethodPost, "/contacts/search")
				if got := req.Header.Get("Content-Type"); got != "application/json" {
					t.Fatalf("Content-Type = %q, want application/json", got)
				}

				body := req.JSONBody(t)
				if got := offlineSDKNestedString(body, "query", "field"); got != "email" {
					t.Fatalf("query.field = %q, want email", got)
				}
				if got := offlineSDKNestedString(body, "query", "operator"); got != string(ContactSearchEquals) {
					t.Fatalf("query.operator = %q, want %q", got, ContactSearchEquals)
				}
				if got := offlineSDKNestedString(body, "query", "value"); got != "customer@example.com" {
					t.Fatalf("query.value = %q, want customer@example.com", got)
				}
				if got := offlineSDKNestedFloat(body, "pagination", "per_page"); got != 25 {
					t.Fatalf("pagination.per_page = %v, want 25", got)
				}
				if got := offlineSDKNestedString(body, "pagination", "starting_after"); got != "cursor-1" {
					t.Fatalf("pagination.starting_after = %q, want cursor-1", got)
				}
			},
		},
		{
			name: "data export download requests octet stream and returns bytes",
			responses: []offlineHTTPResponse{
				offlineBinaryResponse(http.StatusOK, "application/octet-stream", []byte("export-bytes")),
			},
			call: func(ctx context.Context, client *Client) error {
				data, err := client.Workspace.DownloadDataExport(ctx, "job-1")
				if err != nil {
					return err
				}
				if string(data) != "export-bytes" {
					t.Fatalf("downloaded data = %q, want export-bytes", string(data))
				}
				return nil
			},
			checkRequest: func(t *testing.T, req offlineHTTPRequest) {
				t.Helper()
				assertOfflineSDKIntegrationRequest(t, req, http.MethodGet, "/download/content/data/job-1")
				if got := req.Header.Get("Accept"); got != "application/octet-stream" {
					t.Fatalf("Accept = %q, want application/octet-stream", got)
				}
			},
		},
		{
			name: "api error response maps to ErrorResponse",
			responses: []offlineHTTPResponse{
				offlineJSONResponse(http.StatusNotFound, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"missing"}]}`),
			},
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Get(ctx, "missing")
				if err == nil {
					return errors.New("expected error")
				}

				var apiErr *ErrorResponse
				if !errors.As(err, &apiErr) {
					t.Fatalf("error = %T, want *ErrorResponse", err)
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
				return nil
			},
			checkRequest: func(t *testing.T, req offlineHTTPRequest) {
				t.Helper()
				assertOfflineSDKIntegrationRequest(t, req, http.MethodGet, "/contacts/missing")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newOfflineHTTPIntegrationTestClient(t, tt.responses...)

			if err := tt.call(context.Background(), fixture.Client); err != nil {
				t.Fatalf("SDK call returned error: %v", err)
			}
			if got := fixture.RequestCount(); got != 1 {
				t.Fatalf("captured requests = %d, want 1", got)
			}
			tt.checkRequest(t, fixture.Request(t, 0))
		})
	}
}

func TestOfflineSDKIntegrationContextCancellation(t *testing.T) {
	fixture := newOfflineHTTPIntegrationTestClient(t, offlineHTTPResponse{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body:  []byte(`{"type":"contact","id":"slow-contact"}`),
		Delay: 100 * time.Millisecond,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := fixture.Client.Contacts.Get(ctx, "slow-contact")
	if err == nil {
		t.Fatal("expected context deadline error")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), context.DeadlineExceeded.Error()) {
		t.Fatalf("error = %v, want context deadline exceeded", err)
	}
	if got := fixture.RequestCount(); got != 1 {
		t.Fatalf("captured requests = %d, want 1", got)
	}
	assertOfflineSDKIntegrationRequest(t, fixture.Request(t, 0), http.MethodGet, "/contacts/slow-contact")
}

func assertOfflineSDKIntegrationRequest(t *testing.T, req offlineHTTPRequest, method, path string) {
	t.Helper()

	if req.Method != method {
		t.Fatalf("method = %q, want %q", req.Method, method)
	}
	if req.Path != path {
		t.Fatalf("path = %q, want %q", req.Path, path)
	}
	if req.RawQuery != "" {
		t.Fatalf("query = %q, want empty", req.RawQuery)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer token" {
		t.Fatalf("Authorization = %q, want Bearer token", got)
	}
	if got := req.Header.Get("Intercom-Version"); got != defaultAPIVersion {
		t.Fatalf("Intercom-Version = %q, want %q", got, defaultAPIVersion)
	}
	if got := req.Header.Get("User-Agent"); !strings.Contains(got, defaultUserAgent) {
		t.Fatalf("User-Agent = %q, want to contain %q", got, defaultUserAgent)
	}
}

func offlineSDKNestedString(body map[string]any, keys ...string) string {
	value := offlineSDKNestedValue(body, keys...)
	text, _ := value.(string)
	return text
}

func offlineSDKNestedFloat(body map[string]any, keys ...string) float64 {
	value := offlineSDKNestedValue(body, keys...)
	number, _ := value.(float64)
	return number
}

func offlineSDKNestedValue(body map[string]any, keys ...string) any {
	var value any = body
	for _, key := range keys {
		next, ok := value.(map[string]any)
		if !ok {
			return nil
		}
		value = next[key]
	}
	return value
}
