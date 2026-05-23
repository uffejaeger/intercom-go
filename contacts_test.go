package intercom

import (
	"context"
	"encoding/json"
	"errors"
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
			name:     "get contact by external ID",
			response: `{"type":"contact","id":"contact-1","external_id":"ext-1"}`,
			call: func(ctx context.Context, client *Client) error {
				contact, err := client.Contacts.GetByExternalID(ctx, "ext-1")
				if err != nil {
					return err
				}
				if contact.Id == nil || *contact.Id != "contact-1" {
					t.Fatalf("contact.Id = %v", contact.Id)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/contacts/find_by_external_id/ext-1",
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
			name:     "search contacts by int scalar value",
			response: `{"type":"list","data":[],"total_count":0}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Search(ctx, ContactSearch{
					Field:    "created_at",
					Operator: ContactSearchGreaterThan,
					Value:    1700000000,
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/contacts/search",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedFloat(body, "query", "value"); got != float64(1700000000) {
					t.Fatalf("query.value = %v", got)
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
		{
			name:     "create contact",
			response: `{"type":"contact","id":"contact-1","email":"new@example.com"}`,
			call: func(ctx context.Context, client *Client) error {
				email := "new@example.com"
				contact, err := client.Contacts.Create(ctx, ContactCreate{Email: &email})
				if err != nil {
					return err
				}
				if contact.Id == nil || *contact.Id != "contact-1" {
					t.Fatalf("contact.Id = %v", contact.Id)
				}
				return nil
			},
			wantMethod: http.MethodPost,
			wantPath:   "/contacts",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "email"); got != "new@example.com" {
					t.Fatalf("email = %q", got)
				}
			},
		},
		{
			name:     "update contact",
			response: `{"type":"contact","id":"contact-1","name":"Updated Name"}`,
			call: func(ctx context.Context, client *Client) error {
				name := "Updated Name"
				contact, err := client.Contacts.Update(ctx, "contact-1", ContactUpdate{Name: &name})
				if err != nil {
					return err
				}
				if contact.Name == nil || *contact.Name != "Updated Name" {
					t.Fatalf("contact.Name = %v", contact.Name)
				}
				return nil
			},
			wantMethod: http.MethodPut,
			wantPath:   "/contacts/contact-1",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "name"); got != "Updated Name" {
					t.Fatalf("name = %q", got)
				}
			},
		},
		{
			name:     "merge contacts",
			response: `{"type":"contact","id":"contact-2"}`,
			call: func(ctx context.Context, client *Client) error {
				contact, err := client.Contacts.Merge(ctx, "lead-1", "contact-2")
				if err != nil {
					return err
				}
				if contact.Id == nil || *contact.Id != "contact-2" {
					t.Fatalf("contact.Id = %v", contact.Id)
				}
				return nil
			},
			wantMethod: http.MethodPost,
			wantPath:   "/contacts/merge",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "from"); got != "lead-1" {
					t.Fatalf("from = %q", got)
				}
				if got := nestedString(body, "into"); got != "contact-2" {
					t.Fatalf("into = %q", got)
				}
			},
		},
		{
			name:     "archive contact",
			response: `{"type":"contact","id":"contact-1","archived":true}`,
			call: func(ctx context.Context, client *Client) error {
				result, err := client.Contacts.Archive(ctx, "contact-1")
				if err != nil {
					return err
				}
				if result.Id == nil || *result.Id != "contact-1" {
					t.Fatalf("result.Id = %v", result.Id)
				}
				return nil
			},
			wantMethod: http.MethodPost,
			wantPath:   "/contacts/contact-1/archive",
		},
		{
			name:     "unarchive contact",
			response: `{"type":"contact","id":"contact-1","archived":false}`,
			call: func(ctx context.Context, client *Client) error {
				result, err := client.Contacts.Unarchive(ctx, "contact-1")
				if err != nil {
					return err
				}
				if result.Id == nil || *result.Id != "contact-1" {
					t.Fatalf("result.Id = %v", result.Id)
				}
				return nil
			},
			wantMethod: http.MethodPost,
			wantPath:   "/contacts/contact-1/unarchive",
		},
		{
			name:     "block contact",
			response: `{"type":"contact","id":"contact-1"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Block(ctx, "contact-1")
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/contacts/contact-1/block",
		},
		{
			name:     "delete contact",
			response: `{"id":"contact-1","object":"contact","deleted":true}`,
			call: func(ctx context.Context, client *Client) error {
				result, err := client.Contacts.Delete(ctx, "contact-1")
				if err != nil {
					return err
				}
				if result.Id == nil || *result.Id != "contact-1" {
					t.Fatalf("result.Id = %v", result.Id)
				}
				return nil
			},
			wantMethod: http.MethodDelete,
			wantPath:   "/contacts/contact-1",
		},
		{
			name:     "list notes",
			response: `{"type":"list","data":[{"type":"note","id":"1","body":"a note"}]}`,
			call: func(ctx context.Context, client *Client) error {
				notes, err := client.Contacts.ListNotes(ctx, "contact-1")
				if err != nil {
					return err
				}
				if notes.Data == nil || len(*notes.Data) != 1 {
					t.Fatalf("notes.Data = %v", notes.Data)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/contacts/contact-1/notes",
		},
		{
			name:     "create note",
			response: `{"type":"note","id":"1","body":"a note"}`,
			call: func(ctx context.Context, client *Client) error {
				note, err := client.Contacts.CreateNote(ctx, "123", "a note", "admin-1")
				if err != nil {
					return err
				}
				if note.Body == nil || *note.Body != "a note" {
					t.Fatalf("note.Body = %v", note.Body)
				}
				return nil
			},
			wantMethod: http.MethodPost,
			wantPath:   "/contacts/123/notes",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "body"); got != "a note" {
					t.Fatalf("body = %q", got)
				}
				if got := nestedString(body, "admin_id"); got != "admin-1" {
					t.Fatalf("admin_id = %q", got)
				}
			},
		},
		{
			name:     "list segments",
			response: `{"type":"list","data":[{"type":"segment","id":"seg-1"}]}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.ListSegments(ctx, "contact-1")
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/contacts/contact-1/segments",
		},
		{
			name:     "list subscriptions",
			response: `{"type":"list","data":[{"type":"subscription_type","id":"sub-1"}]}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.ListSubscriptions(ctx, "contact-1")
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/contacts/contact-1/subscriptions",
		},
		{
			name:     "attach subscription",
			response: `{"type":"subscription_type","id":"sub-1"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.AttachSubscription(ctx, "contact-1", "sub-1", "opt_in")
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/contacts/contact-1/subscriptions",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "id"); got != "sub-1" {
					t.Fatalf("id = %q", got)
				}
				if got := nestedString(body, "consent_type"); got != "opt_in" {
					t.Fatalf("consent_type = %q", got)
				}
			},
		},
		{
			name:     "detach subscription",
			response: `{"type":"subscription_type","id":"sub-1"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.DetachSubscription(ctx, "contact-1", "sub-1")
				return err
			},
			wantMethod: http.MethodDelete,
			wantPath:   "/contacts/contact-1/subscriptions/sub-1",
		},
		{
			name:     "list tags",
			response: `{"type":"list","data":[{"type":"tag","id":"tag-1","name":"vip"}]}`,
			call: func(ctx context.Context, client *Client) error {
				tags, err := client.Contacts.ListTags(ctx, "contact-1")
				if err != nil {
					return err
				}
				if tags.Data == nil || len(*tags.Data) != 1 {
					t.Fatalf("tags.Data = %v", tags.Data)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/contacts/contact-1/tags",
		},
		{
			name:     "attach tag",
			response: `{"type":"tag","id":"tag-1","name":"vip"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.AttachTag(ctx, "contact-1", "tag-1")
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/contacts/contact-1/tags",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "id"); got != "tag-1" {
					t.Fatalf("id = %q", got)
				}
			},
		},
		{
			name:     "detach tag",
			response: `{"type":"tag","id":"tag-1","name":"vip"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.DetachTag(ctx, "contact-1", "tag-1")
				return err
			},
			wantMethod: http.MethodDelete,
			wantPath:   "/contacts/contact-1/tags/tag-1",
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
		{
			name: "create contact",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Create(ctx, ContactCreate{})
				return err
			},
		},
		{
			name: "update contact",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Update(ctx, "contact-1", ContactUpdate{})
				return err
			},
		},
		{
			name: "delete contact",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Delete(ctx, "contact-1")
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

func TestContactsGetEmptyID(t *testing.T) {
	client := newContactsTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		t.Fatal("unexpected HTTP request")
		return nil, nil
	}))

	_, err := client.Contacts.Get(context.Background(), "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestContactsTransportErrors(t *testing.T) {
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
		{
			name: "get by external ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.GetByExternalID(ctx, "ext-1")
				return err
			},
		},
		{
			name: "create contact",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Create(ctx, ContactCreate{})
				return err
			},
		},
		{
			name: "update contact",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Update(ctx, "contact-1", ContactUpdate{})
				return err
			},
		},
		{
			name: "merge contact",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Merge(ctx, "lead-1", "contact-2")
				return err
			},
		},
		{
			name: "archive contact",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Archive(ctx, "contact-1")
				return err
			},
		},
		{
			name: "unarchive contact",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Unarchive(ctx, "contact-1")
				return err
			},
		},
		{
			name: "block contact",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Block(ctx, "contact-1")
				return err
			},
		},
		{
			name: "delete contact",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Delete(ctx, "contact-1")
				return err
			},
		},
		{
			name: "list notes",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.ListNotes(ctx, "contact-1")
				return err
			},
		},
		{
			name: "create note",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.CreateNote(ctx, "123", "body", "")
				return err
			},
		},
		{
			name: "list segments",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.ListSegments(ctx, "contact-1")
				return err
			},
		},
		{
			name: "list subscriptions",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.ListSubscriptions(ctx, "contact-1")
				return err
			},
		},
		{
			name: "attach subscription",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.AttachSubscription(ctx, "contact-1", "sub-1", "opt_in")
				return err
			},
		},
		{
			name: "detach subscription",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.DetachSubscription(ctx, "contact-1", "sub-1")
				return err
			},
		},
		{
			name: "list tags",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.ListTags(ctx, "contact-1")
				return err
			},
		},
		{
			name: "attach tag",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.AttachTag(ctx, "contact-1", "tag-1")
				return err
			},
		},
		{
			name: "detach tag",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.DetachTag(ctx, "contact-1", "tag-1")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := roundTripFunc(func(*http.Request) (*http.Response, error) {
				return nil, errors.New("network down")
			})

			client := newContactsTestClient(t, transport)
			if err := tt.call(context.Background(), client); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestMarshalBodyPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for non-serializable type")
		}
	}()
	marshalBody(make(chan int))
}

func TestContactsValidation(t *testing.T) {
	tests := []struct {
		name string
		call func(context.Context, *Client) error
	}{
		{
			name: "get contact: empty ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Get(ctx, "")
				return err
			},
		},
		{
			name: "get by external ID: empty ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.GetByExternalID(ctx, "")
				return err
			},
		},
		{
			name: "update contact: empty ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Update(ctx, "", ContactUpdate{})
				return err
			},
		},
		{
			name: "merge: empty from",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Merge(ctx, "", "contact-2")
				return err
			},
		},
		{
			name: "merge: empty into",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Merge(ctx, "lead-1", "")
				return err
			},
		},
		{
			name: "archive: empty ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Archive(ctx, "")
				return err
			},
		},
		{
			name: "unarchive: empty ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Unarchive(ctx, "")
				return err
			},
		},
		{
			name: "block: empty ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Block(ctx, "")
				return err
			},
		},
		{
			name: "delete: empty ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Delete(ctx, "")
				return err
			},
		},
		{
			name: "list notes: empty contact ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.ListNotes(ctx, "")
				return err
			},
		},
		{
			name: "create note: empty contact ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.CreateNote(ctx, "", "body", "")
				return err
			},
		},
		{
			name: "create note: empty body",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.CreateNote(ctx, "123", "", "")
				return err
			},
		},
		{
			name: "create note: non-numeric contact ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.CreateNote(ctx, "not-an-int", "body", "")
				return err
			},
		},
		{
			name: "list segments: empty contact ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.ListSegments(ctx, "")
				return err
			},
		},
		{
			name: "list subscriptions: empty contact ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.ListSubscriptions(ctx, "")
				return err
			},
		},
		{
			name: "attach subscription: empty contact ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.AttachSubscription(ctx, "", "sub-1", "opt_in")
				return err
			},
		},
		{
			name: "attach subscription: empty subscription ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.AttachSubscription(ctx, "contact-1", "", "opt_in")
				return err
			},
		},
		{
			name: "attach subscription: empty consent type",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.AttachSubscription(ctx, "contact-1", "sub-1", "")
				return err
			},
		},
		{
			name: "detach subscription: empty contact ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.DetachSubscription(ctx, "", "sub-1")
				return err
			},
		},
		{
			name: "detach subscription: empty subscription ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.DetachSubscription(ctx, "contact-1", "")
				return err
			},
		},
		{
			name: "list tags: empty contact ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.ListTags(ctx, "")
				return err
			},
		},
		{
			name: "attach tag: empty contact ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.AttachTag(ctx, "", "tag-1")
				return err
			},
		},
		{
			name: "attach tag: empty tag ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.AttachTag(ctx, "contact-1", "")
				return err
			},
		},
		{
			name: "detach tag: empty contact ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.DetachTag(ctx, "", "tag-1")
				return err
			},
		},
		{
			name: "detach tag: empty tag ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.DetachTag(ctx, "contact-1", "")
				return err
			},
		},
		{
			name: "search: missing field",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Search(ctx, ContactSearch{Operator: ContactSearchEquals, Value: "x"})
				return err
			},
		},
		{
			name: "search: missing operator",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Search(ctx, ContactSearch{Field: "email", Value: "x"})
				return err
			},
		},
		{
			name: "search: unsupported value type",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Contacts.Search(ctx, ContactSearch{Field: "email", Operator: ContactSearchEquals, Value: true})
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newContactsTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
				t.Fatal("unexpected HTTP request")
				return nil, nil
			}))
			if err := tt.call(context.Background(), client); err == nil {
				t.Fatal("expected error, got nil")
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
