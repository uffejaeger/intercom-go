package intercom

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	openapi_types "github.com/oapi-codegen/runtime/types"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

func newConversationsTestClient(t *testing.T, transport http.RoundTripper) *Client {
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

func TestConversationsServiceRequests(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		call       func(context.Context, *Client) error
		wantMethod string
		wantPath   string
		wantBody   func(t *testing.T, body map[string]any)
	}{
		{
			name:     "list conversations",
			response: `{"type":"conversation.list","conversations":[],"total_count":0}`,
			call: func(ctx context.Context, client *Client) error {
				list, err := client.Conversations.List(ctx)
				if err != nil {
					return err
				}
				if list.TotalCount == nil || *list.TotalCount != 0 {
					t.Fatalf("TotalCount = %v", list.TotalCount)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/conversations",
		},
		{
			name:     "create conversation",
			response: `{"type":"user_message","id":"msg-1","body":"hello","created_at":1700000000,"message_type":"inapp"}`,
			call: func(ctx context.Context, client *Client) error {
				msg, err := client.Conversations.Create(ctx, ConversationCreate{
					Body: "hello",
					From: struct {
						Id   openapi_types.UUID                    `json:"id"`
						Type gen.CreateConversationRequestFromType `json:"type"`
					}{Type: "user"},
				})
				if err != nil {
					return err
				}
				if msg.Id != "msg-1" {
					t.Fatalf("msg.Id = %q", msg.Id)
				}
				return nil
			},
			wantMethod: http.MethodPost,
			wantPath:   "/conversations",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "body"); got != "hello" {
					t.Fatalf("body = %q", got)
				}
			},
		},
		{
			name:     "get conversation",
			response: `{"type":"conversation","id":"123"}`,
			call: func(ctx context.Context, client *Client) error {
				conv, err := client.Conversations.Get(ctx, "123")
				if err != nil {
					return err
				}
				if conv.Id == nil || *conv.Id != "123" {
					t.Fatalf("conv.Id = %v", conv.Id)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/conversations/123",
		},
		{
			name:     "update conversation",
			response: `{"type":"conversation","id":"123"}`,
			call: func(ctx context.Context, client *Client) error {
				read := true
				_, err := client.Conversations.Update(ctx, "123", ConversationUpdate{Read: &read})
				return err
			},
			wantMethod: http.MethodPut,
			wantPath:   "/conversations/123",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got, _ := body["read"].(bool); !got {
					t.Fatalf("read = %v", body["read"])
				}
			},
		},
		{
			name:     "delete conversation",
			response: `{"id":"123","object":"conversation","deleted":true}`,
			call: func(ctx context.Context, client *Client) error {
				del, err := client.Conversations.Delete(ctx, "123")
				if err != nil {
					return err
				}
				if del.Id == nil || *del.Id != "123" {
					t.Fatalf("del.Id = %v", del.Id)
				}
				return nil
			},
			wantMethod: http.MethodDelete,
			wantPath:   "/conversations/123",
		},
		{
			name:     "search conversations",
			response: `{"type":"conversation.list","conversations":[],"total_count":0}`,
			call: func(ctx context.Context, client *Client) error {
				var query ConversationSearchQuery
				filter := gen.SingleFilterSearchRequestSchema{}
				field := "state"
				operator := gen.SingleFilterSearchRequestOperator("=")
				var value gen.SingleFilterSearchRequest_Value
				_ = value.FromSingleFilterSearchRequestValue0("open")
				filter.Field = &field
				filter.Operator = &operator
				filter.Value = &value
				_ = query.Query.FromSingleFilterSearchRequestSchema(filter)
				_, err := client.Conversations.Search(ctx, query)
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/conversations/search",
		},
		{
			name:     "reply to conversation",
			response: `{"type":"conversation","id":"conv-1"}`,
			call: func(ctx context.Context, client *Client) error {
				body := "admin reply"
				_, err := client.Conversations.Reply(ctx, "conv-1", ConversationAdminReply{
					AdminId:     "admin-1",
					Body:        &body,
					MessageType: "comment",
					Type:        "admin",
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/conversations/conv-1/reply",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "type"); got != "admin" {
					t.Fatalf("type = %q", got)
				}
				if got := nestedString(body, "admin_id"); got != "admin-1" {
					t.Fatalf("admin_id = %q", got)
				}
			},
		},
		{
			name:     "reply as contact",
			response: `{"type":"conversation","id":"conv-1"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.ReplyAsContact(ctx, "conv-1", ConversationContactReply{
					Body:        "contact reply",
					MessageType: "comment",
					Type:        "user",
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/conversations/conv-1/reply",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "type"); got != "user" {
					t.Fatalf("type = %q", got)
				}
			},
		},
		{
			name:     "assign conversation",
			response: `{"type":"conversation","id":"conv-1"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Assign(ctx, "conv-1", ConversationAssign{
					AdminId:     "admin-1",
					AssigneeId:  "admin-2",
					MessageType: "assignment",
					Type:        "admin",
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/conversations/conv-1/parts",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "message_type"); got != "assignment" {
					t.Fatalf("message_type = %q", got)
				}
			},
		},
		{
			name:     "close conversation",
			response: `{"type":"conversation","id":"conv-1"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Close(ctx, "conv-1", ConversationClose{
					AdminId:     "admin-1",
					MessageType: "close",
					Type:        "admin",
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/conversations/conv-1/parts",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "message_type"); got != "close" {
					t.Fatalf("message_type = %q", got)
				}
			},
		},
		{
			name:     "open conversation",
			response: `{"type":"conversation","id":"conv-1"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Open(ctx, "conv-1", ConversationOpen{
					AdminId:     "admin-1",
					MessageType: "open",
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/conversations/conv-1/parts",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "message_type"); got != "open" {
					t.Fatalf("message_type = %q", got)
				}
			},
		},
		{
			name:     "snooze conversation",
			response: `{"type":"conversation","id":"conv-1"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Snooze(ctx, "conv-1", ConversationSnooze{
					AdminId:      "admin-1",
					MessageType:  "snoozed",
					SnoozedUntil: 1700000000,
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/conversations/conv-1/parts",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "message_type"); got != "snoozed" {
					t.Fatalf("message_type = %q", got)
				}
			},
		},
		{
			name:     "attach contact to conversation",
			response: `{"type":"conversation","id":"conv-1"}`,
			call: func(ctx context.Context, client *Client) error {
				adminID := "admin-1"
				_, err := client.Conversations.AttachContact(ctx, "conv-1", ConversationAttachContact{
					AdminId: &adminID,
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/conversations/conv-1/customers",
		},
		{
			name:     "detach contact from conversation",
			response: `{"type":"conversation","id":"conv-1"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.DetachContact(ctx, "conv-1", "contact-1", ConversationDetachContact{
					AdminId: "admin-1",
				})
				return err
			},
			wantMethod: http.MethodDelete,
			wantPath:   "/conversations/conv-1/customers/contact-1",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "admin_id"); got != "admin-1" {
					t.Fatalf("admin_id = %q", got)
				}
			},
		},
		{
			name:     "redact conversation part",
			response: `{"type":"conversation","id":"conv-1"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.RedactPart(ctx, ConversationRedactPart{
					ConversationId:     "conv-1",
					ConversationPartId: "part-1",
					Type:               "conversation_part",
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/conversations/redact",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "type"); got != "conversation_part" {
					t.Fatalf("type = %q", got)
				}
			},
		},
		{
			name:     "redact conversation source",
			response: `{"type":"conversation","id":"conv-1"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.RedactSource(ctx, ConversationRedactSource{
					ConversationId: "conv-1",
					SourceId:       "src-1",
					Type:           "source",
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/conversations/redact",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "source_id"); got != "src-1" {
					t.Fatalf("source_id = %q", got)
				}
			},
		},
		{
			name:     "convert to ticket",
			response: `{"type":"ticket","id":"ticket-1"}`,
			call: func(ctx context.Context, client *Client) error {
				ticket, err := client.Conversations.ConvertToTicket(ctx, "123", ConversationToTicket{
					TicketTypeId: "type-1",
				})
				if err != nil {
					return err
				}
				if ticket.Id == nil || *ticket.Id != "ticket-1" {
					t.Fatalf("ticket.Id = %v", ticket.Id)
				}
				return nil
			},
			wantMethod: http.MethodPost,
			wantPath:   "/conversations/123/convert",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "ticket_type_id"); got != "type-1" {
					t.Fatalf("ticket_type_id = %q", got)
				}
			},
		},
		{
			name:     "attach tag to conversation",
			response: `{"type":"tag","id":"tag-1","name":"vip"}`,
			call: func(ctx context.Context, client *Client) error {
				tag, err := client.Conversations.AttachTag(ctx, "conv-1", "tag-1", "admin-1")
				if err != nil {
					return err
				}
				if tag.Id == nil || *tag.Id != "tag-1" {
					t.Fatalf("tag.Id = %v", tag.Id)
				}
				return nil
			},
			wantMethod: http.MethodPost,
			wantPath:   "/conversations/conv-1/tags",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "id"); got != "tag-1" {
					t.Fatalf("id = %q", got)
				}
				if got := nestedString(body, "admin_id"); got != "admin-1" {
					t.Fatalf("admin_id = %q", got)
				}
			},
		},
		{
			name:     "detach tag from conversation",
			response: `{"type":"tag","id":"tag-1","name":"vip"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.DetachTag(ctx, "conv-1", "tag-1", "admin-1")
				return err
			},
			wantMethod: http.MethodDelete,
			wantPath:   "/conversations/conv-1/tags/tag-1",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "admin_id"); got != "admin-1" {
					t.Fatalf("admin_id = %q", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var method, path, contentType string
			var body map[string]any

			transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				method = req.Method
				path = req.URL.Path
				contentType = req.Header.Get("Content-Type")

				if req.Body != nil {
					if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
						t.Fatalf("decode request body: %v", err)
					}
				}

				return jsonResponse(req, http.StatusOK, tt.response), nil
			})

			client := newConversationsTestClient(t, transport)
			if err := tt.call(context.Background(), client); err != nil {
				t.Fatalf("call returned error: %v", err)
			}

			if method != tt.wantMethod {
				t.Fatalf("method = %q, want %q", method, tt.wantMethod)
			}
			if path != tt.wantPath {
				t.Fatalf("path = %q, want %q", path, tt.wantPath)
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

func TestConversationsServiceErrors(t *testing.T) {
	tests := []struct {
		name string
		call func(context.Context, *Client) error
	}{
		{
			name: "list conversations",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.List(ctx)
				return err
			},
		},
		{
			name: "create conversation",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Create(ctx, ConversationCreate{Body: "hi"})
				return err
			},
		},
		{
			name: "get conversation",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Get(ctx, "123")
				return err
			},
		},
		{
			name: "search conversations",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Search(ctx, ConversationSearchQuery{})
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return jsonResponse(req, http.StatusUnauthorized, `{"type":"error.list","request_id":"12a938a3-314e-4939-b773-5cd45738bd21","errors":[{"code":"unauthorized","message":"Access Token Invalid"}]}`), nil
			})

			client := newConversationsTestClient(t, transport)
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
		})
	}
}

func TestConversationsTransportErrors(t *testing.T) {
	tests := []struct {
		name string
		call func(context.Context, *Client) error
	}{
		{
			name: "list",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.List(ctx)
				return err
			},
		},
		{
			name: "create",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Create(ctx, ConversationCreate{})
				return err
			},
		},
		{
			name: "get",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Get(ctx, "123")
				return err
			},
		},
		{
			name: "update",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Update(ctx, "123", ConversationUpdate{})
				return err
			},
		},
		{
			name: "delete",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Delete(ctx, "123")
				return err
			},
		},
		{
			name: "search",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Search(ctx, ConversationSearchQuery{})
				return err
			},
		},
		{
			name: "reply",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Reply(ctx, "conv-1", ConversationAdminReply{})
				return err
			},
		},
		{
			name: "reply as contact",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.ReplyAsContact(ctx, "conv-1", ConversationContactReply{})
				return err
			},
		},
		{
			name: "assign",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Assign(ctx, "conv-1", ConversationAssign{})
				return err
			},
		},
		{
			name: "close",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Close(ctx, "conv-1", ConversationClose{})
				return err
			},
		},
		{
			name: "open",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Open(ctx, "conv-1", ConversationOpen{})
				return err
			},
		},
		{
			name: "snooze",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Snooze(ctx, "conv-1", ConversationSnooze{})
				return err
			},
		},
		{
			name: "attach contact",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.AttachContact(ctx, "conv-1", ConversationAttachContact{})
				return err
			},
		},
		{
			name: "detach contact",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.DetachContact(ctx, "conv-1", "contact-1", ConversationDetachContact{})
				return err
			},
		},
		{
			name: "redact part",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.RedactPart(ctx, ConversationRedactPart{})
				return err
			},
		},
		{
			name: "redact source",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.RedactSource(ctx, ConversationRedactSource{})
				return err
			},
		},
		{
			name: "convert to ticket",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.ConvertToTicket(ctx, "123", ConversationToTicket{})
				return err
			},
		},
		{
			name: "attach tag",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.AttachTag(ctx, "conv-1", "tag-1", "admin-1")
				return err
			},
		},
		{
			name: "detach tag",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.DetachTag(ctx, "conv-1", "tag-1", "admin-1")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := roundTripFunc(func(*http.Request) (*http.Response, error) {
				return nil, errors.New("network down")
			})
			client := newConversationsTestClient(t, transport)
			if err := tt.call(context.Background(), client); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestConversationsValidation(t *testing.T) {
	tests := []struct {
		name string
		call func(context.Context, *Client) error
	}{
		{
			name: "get: empty ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Get(ctx, "")
				return err
			},
		},
		{
			name: "get: non-numeric ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Get(ctx, "not-a-number")
				return err
			},
		},
		{
			name: "get: partial-numeric ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Get(ctx, "123abc")
				return err
			},
		},
		{
			name: "update: empty ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Update(ctx, "", ConversationUpdate{})
				return err
			},
		},
		{
			name: "update: non-numeric ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Update(ctx, "not-a-number", ConversationUpdate{})
				return err
			},
		},
		{
			name: "delete: empty ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Delete(ctx, "")
				return err
			},
		},
		{
			name: "delete: non-numeric ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Delete(ctx, "not-a-number")
				return err
			},
		},
		{
			name: "reply: empty ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Reply(ctx, "", ConversationAdminReply{})
				return err
			},
		},
		{
			name: "reply as contact: empty ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.ReplyAsContact(ctx, "", ConversationContactReply{})
				return err
			},
		},
		{
			name: "assign: empty ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Assign(ctx, "", ConversationAssign{})
				return err
			},
		},
		{
			name: "close: empty ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Close(ctx, "", ConversationClose{})
				return err
			},
		},
		{
			name: "open: empty ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Open(ctx, "", ConversationOpen{})
				return err
			},
		},
		{
			name: "snooze: empty ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.Snooze(ctx, "", ConversationSnooze{})
				return err
			},
		},
		{
			name: "attach contact: empty conversation ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.AttachContact(ctx, "", ConversationAttachContact{})
				return err
			},
		},
		{
			name: "detach contact: empty conversation ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.DetachContact(ctx, "", "contact-1", ConversationDetachContact{})
				return err
			},
		},
		{
			name: "detach contact: empty contact ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.DetachContact(ctx, "conv-1", "", ConversationDetachContact{})
				return err
			},
		},
		{
			name: "convert to ticket: empty ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.ConvertToTicket(ctx, "", ConversationToTicket{})
				return err
			},
		},
		{
			name: "convert to ticket: non-numeric ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.ConvertToTicket(ctx, "not-a-number", ConversationToTicket{})
				return err
			},
		},
		{
			name: "attach tag: empty conversation ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.AttachTag(ctx, "", "tag-1", "admin-1")
				return err
			},
		},
		{
			name: "attach tag: empty tag ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.AttachTag(ctx, "conv-1", "", "admin-1")
				return err
			},
		},
		{
			name: "attach tag: empty admin ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.AttachTag(ctx, "conv-1", "tag-1", "")
				return err
			},
		},
		{
			name: "detach tag: empty conversation ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.DetachTag(ctx, "", "tag-1", "admin-1")
				return err
			},
		},
		{
			name: "detach tag: empty tag ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.DetachTag(ctx, "conv-1", "", "admin-1")
				return err
			},
		},
		{
			name: "detach tag: empty admin ID",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Conversations.DetachTag(ctx, "conv-1", "tag-1", "")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newConversationsTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
				t.Fatal("unexpected HTTP request")
				return nil, nil
			}))
			if err := tt.call(context.Background(), client); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}
