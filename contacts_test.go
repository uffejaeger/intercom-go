package intercom

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestContactsServiceRequests(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		call       func(context.Context, *Client) error
		wantMethod string
		wantPath   string
		wantBody   func(t *testing.T, body map[string]any)
	}{
		{
			name:     "get contact",
			response: `{"type":"contact","id":"contact-1","email":"contact@example.com"}`,
			call: func(ctx context.Context, client *Client) error {
				contact, err := client.Contacts.Get(ctx, "contact-1")
				if err != nil {
					return err
				}
				if contact.Id == nil || *contact.Id != "contact-1" {
					t.Fatalf("contact.Id = %v", contact.Id)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/contacts/contact-1",
		},
		{
			name:     "list contacts",
			response: `{"type":"list","data":[{"type":"contact","id":"contact-1"}],"total_count":1}`,
			call: func(ctx context.Context, client *Client) error {
				contacts, err := client.Contacts.List(ctx)
				if err != nil {
					return err
				}
				if contacts.Data == nil || len(*contacts.Data) != 1 {
					t.Fatalf("contacts.Data = %#v", contacts.Data)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/contacts",
		},
		{
			name:     "search contacts by scalar value",
			response: `{"type":"list","data":[{"type":"contact","id":"contact-1","email":"contact@example.com"}],"total_count":1}`,
			call: func(ctx context.Context, client *Client) error {
				contacts, err := client.Contacts.Search(ctx, ContactSearch{
					Field:         "email",
					Operator:      ContactSearchEquals,
					Value:         "contact@example.com",
					PerPage:       25,
					StartingAfter: "cursor-1",
				})
				if err != nil {
					return err
				}
				if contacts.TotalCount == nil || *contacts.TotalCount != 1 {
					t.Fatalf("contacts.TotalCount = %v", contacts.TotalCount)
				}
				return nil
			},
			wantMethod: http.MethodPost,
			wantPath:   "/contacts/search",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "query", "field"); got != "email" {
					t.Fatalf("query.field = %q", got)
				}
				if got := nestedString(body, "query", "operator"); got != string(ContactSearchEquals) {
					t.Fatalf("query.operator = %q", got)
				}
				if got := nestedString(body, "query", "value"); got != "contact@example.com" {
					t.Fatalf("query.value = %q", got)
				}
				if got := nestedFloat(body, "pagination", "per_page"); got != 25 {
					t.Fatalf("pagination.per_page = %v", got)
				}
				if got := nestedString(body, "pagination", "starting_after"); got != "cursor-1" {
					t.Fatalf("pagination.starting_after = %q", got)
				}
			},
		},
		{
			name:     "search contacts by string array value",
			response: `{"type":"list","data":[],"total_count":0}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Search(ctx, ContactSearch{
					Field:    "email",
					Operator: ContactSearchOperator("IN"),
					Value:    []string{"first@example.com", "second@example.com"},
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/contacts/search",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "query", "field"); got != "email" {
					t.Fatalf("query.field = %q", got)
				}
				if got := nestedString(body, "query", "operator"); got != "IN" {
					t.Fatalf("query.operator = %q", got)
				}
				if got := nested(body, "query", "value"); !reflect.DeepEqual(got, []any{"first@example.com", "second@example.com"}) {
					t.Fatalf("query.value = %#v", got)
				}
			},
		},
		{
			name:     "search contacts by int array value",
			response: `{"type":"list","data":[],"total_count":0}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Search(ctx, ContactSearch{
					Field:    "created_at",
					Operator: ContactSearchOperator("IN"),
					Value:    []int{1, 2},
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/contacts/search",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "query", "field"); got != "created_at" {
					t.Fatalf("query.field = %q", got)
				}
				if got := nestedString(body, "query", "operator"); got != "IN" {
					t.Fatalf("query.operator = %q", got)
				}
				if got := nested(body, "query", "value"); !reflect.DeepEqual(got, []any{float64(1), float64(2)}) {
					t.Fatalf("query.value = %#v", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var method, path, authorization, contentType string
			var body map[string]any

			transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				method = req.Method
				path = req.URL.Path
				authorization = req.Header.Get("Authorization")
				contentType = req.Header.Get("Content-Type")

				if req.Body != nil {
					if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
						t.Fatalf("decode request body: %v", err)
					}
				}

				return jsonResponse(req, http.StatusOK, tt.response), nil
			})

			client := newContactsTestClient(t, transport)
			if err := tt.call(context.Background(), client); err != nil {
				t.Fatalf("call returned error: %v", err)
			}

			if method != tt.wantMethod {
				t.Fatalf("method = %q, want %q", method, tt.wantMethod)
			}
			if path != tt.wantPath {
				t.Fatalf("path = %q, want %q", path, tt.wantPath)
			}
			if authorization != "Bearer token" {
				t.Fatalf("Authorization = %q", authorization)
			}
			if tt.wantBody != nil {
				if contentType != "application/json" {
					t.Fatalf("Content-Type = %q", contentType)
				}
				tt.wantBody(t, body)
			}
		})
	}
}

func TestContactsServiceErrors(t *testing.T) {
	tests := []struct {
		name string
		call func(context.Context, *Client) error
	}{
		{
			name: "get contact",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Get(ctx, "contact-1")
				return err
			},
		},
		{
			name: "list contacts",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.List(ctx)
				return err
			},
		},
		{
			name: "search contacts",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Search(ctx, ContactSearch{
					Field:    "email",
					Operator: ContactSearchEquals,
					Value:    "contact@example.com",
				})
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return jsonResponse(req, http.StatusUnauthorized, `{"type":"error.list","request_id":"12a938a3-314e-4939-b773-5cd45738bd21","errors":[{"code":"unauthorized","message":"Access Token Invalid"}]}`), nil
			})

			client := newContactsTestClient(t, transport)
			err := tt.call(context.Background(), client)
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
			if apiErr.RequestID != "12a938a3-314e-4939-b773-5cd45738bd21" {
				t.Fatalf("RequestID = %q", apiErr.RequestID)
			}
		})
	}
}

func TestContactSearchValidation(t *testing.T) {
	tests := []struct {
		name   string
		search ContactSearch
	}{
		{
			name: "missing field",
			search: ContactSearch{
				Operator: ContactSearchEquals,
				Value:    "contact@example.com",
			},
		},
		{
			name: "missing operator",
			search: ContactSearch{
				Field: "email",
				Value: "contact@example.com",
			},
		},
		{
			name: "unsupported value",
			search: ContactSearch{
				Field:    "email",
				Operator: ContactSearchEquals,
				Value:    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newContactsTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
				t.Fatal("unexpected HTTP request")
				return nil, nil
			}))

			_, err := client.Contacts.Search(context.Background(), tt.search)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func newContactsTestClient(t *testing.T, transport http.RoundTripper) *Client {
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

func jsonResponse(req *http.Request, statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body:    io.NopCloser(strings.NewReader(body)),
		Request: req,
	}
}

func nestedString(body map[string]any, keys ...string) string {
	value := nested(body, keys...)
	if value == nil {
		return ""
	}
	return value.(string)
}

func nestedFloat(body map[string]any, keys ...string) float64 {
	value := nested(body, keys...)
	if value == nil {
		return 0
	}
	return value.(float64)
}

func nested(body map[string]any, keys ...string) any {
	var current any = body
	for _, key := range keys {
		object, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = object[key]
	}
	return current
}
