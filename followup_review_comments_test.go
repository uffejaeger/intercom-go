package intercom

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestDoPreservesExplicitAccept(t *testing.T) {
	var gotAccept string

	client, err := NewClient(
		"token",
		WithBaseURL("https://example.test"),
		WithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			gotAccept = req.Header.Get("Accept")
			return &http.Response{
				StatusCode: http.StatusNoContent,
				Body:       http.NoBody,
				Header:     make(http.Header),
				Request:    req,
			}, nil
		})}),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	req, err := client.NewRequest(context.Background(), http.MethodGet, "/exports", nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}
	req.Header.Set("Accept", "application/octet-stream")

	if _, err := client.Do(req); err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	if gotAccept != "application/octet-stream" {
		t.Fatalf("Accept = %q", gotAccept)
	}
}

func TestWorkspaceDownloadRequestsUseOctetStreamAccept(t *testing.T) {
	tests := []struct {
		name string
		call func(context.Context, *Client) error
	}{
		{
			name: "content export download",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Workspace.DownloadDataExport(ctx, "job-1")
				return err
			},
		},
		{
			name: "reporting export download",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Workspace.DownloadReportingExport(ctx, "job-1", "app-1")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotAccept string
			client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
				gotAccept = req.Header.Get("Accept")
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/octet-stream"}},
					Body:       io.NopCloser(strings.NewReader("bytes")),
					Request:    req,
				}, nil
			}))
			if err := tt.call(context.Background(), client); err != nil {
				t.Fatalf("call returned error: %v", err)
			}
			if gotAccept != "application/octet-stream" {
				t.Fatalf("Accept = %q", gotAccept)
			}
		})
	}
}

func TestFinRequestsUsePublicTypes(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name string
		call func(context.Context, *Client) error
	}{
		{
			name: "reply",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Fin.Reply(ctx, FinReply{
					ConversationId: "c1",
					Message:        FinMessage{Author: FinMessageAuthorUser, Body: "hi", Timestamp: now},
					User:           FinUser{Id: "u1"},
				})
				return err
			},
		},
		{
			name: "start conversation",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Fin.StartConversation(ctx, FinStartConversation{
					ConversationId: "c1",
					Message:        FinMessage{Author: FinMessageAuthorUser, Body: "hi", Timestamp: now},
					User:           FinUser{Id: "u1"},
				})
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
				var payload map[string]any
				if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if payload["message"] == nil {
					t.Fatalf("payload = %#v", payload)
				}
				if payload["user"] == nil {
					t.Fatalf("payload = %#v", payload)
				}
				return jsonResponse(req, http.StatusOK, `{"conversation_id":"c1","status":"awaiting_user_reply","user_id":"u1"}`), nil
			}))
			if err := tt.call(context.Background(), client); err != nil {
				t.Fatalf("call returned error: %v", err)
			}
		})
	}
}

func TestTicketHelpersExposeConstructiblePublicRequests(t *testing.T) {
	tests := []struct {
		name    string
		contact TicketContact
		check   func(*testing.T, map[string]any)
	}{
		{
			name:    "contact by id",
			contact: NewTicketContactByID("1234"),
			check: func(t *testing.T, contact map[string]any) {
				t.Helper()
				if contact["id"] != "1234" {
					t.Fatalf("contact = %#v", contact)
				}
			},
		},
		{
			name:    "contact by external id",
			contact: NewTicketContactByExternalID("external-1"),
			check: func(t *testing.T, contact map[string]any) {
				t.Helper()
				if contact["external_id"] != "external-1" {
					t.Fatalf("contact = %#v", contact)
				}
			},
		},
		{
			name:    "contact by email",
			contact: NewTicketContactByEmail("user@example.com"),
			check: func(t *testing.T, contact map[string]any) {
				t.Helper()
				if contact["email"] != "user@example.com" {
					t.Fatalf("contact = %#v", contact)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
				var payload map[string]any
				if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				contacts, ok := payload["contacts"].([]any)
				if !ok || len(contacts) != 1 {
					t.Fatalf("contacts = %#v", payload["contacts"])
				}
				contact, ok := contacts[0].(map[string]any)
				if !ok {
					t.Fatalf("contact = %#v", contacts[0])
				}
				tt.check(t, contact)
				return jsonResponse(req, http.StatusOK, `{"type":"job","id":"2","status":"pending"}`), nil
			}))
			_, err := client.Tickets.EnqueueCreate(context.Background(), TicketCreate{
				TicketTypeId: "ticket-type-1",
				Contacts:     []TicketContact{tt.contact},
			})
			if err != nil {
				t.Fatalf("EnqueueCreate returned error: %v", err)
			}
		})
	}

	t.Run("public tag request types", func(t *testing.T) {
		var gotPath string
		var gotBody map[string]any
		client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			gotPath = req.URL.Path
			if err := json.NewDecoder(req.Body).Decode(&gotBody); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			return jsonResponse(req, http.StatusOK, `{"id":"121","name":"Manual tag"}`), nil
		}))
		if _, err := client.Tickets.AttachTag(context.Background(), "20", TicketTagAttachRequest{Id: "121", AdminId: "991267958"}); err != nil {
			t.Fatalf("AttachTag returned error: %v", err)
		}
		if gotPath != "/tickets/20/tags" {
			t.Fatalf("path = %q", gotPath)
		}
		if gotBody["id"] != "121" || gotBody["admin_id"] != "991267958" {
			t.Fatalf("body = %#v", gotBody)
		}
	})

	t.Run("ticket reply request types", func(t *testing.T) {
		var gotBody map[string]any
		skip := true
		client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if err := json.NewDecoder(req.Body).Decode(&gotBody); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			return jsonResponse(req, http.StatusOK, `{"type":"ticket_part","id":"156"}`), nil
		}))
		if _, err := client.Tickets.Reply(context.Background(), "20", TicketContactReply{
			Body:        "hi",
			Contact:     NewTicketReplyContactByEmail("user@example.com"),
			MessageType: TicketReplyMessageTypeComment,
		}, &skip); err != nil {
			t.Fatalf("Reply returned error: %v", err)
		}
		if gotBody["email"] != "user@example.com" || gotBody["type"] != "user" || gotBody["skip_notifications"] != true {
			t.Fatalf("body = %#v", gotBody)
		}
	})

	t.Run("tag create request types", func(t *testing.T) {
		var gotBody map[string]any
		client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if err := json.NewDecoder(req.Body).Decode(&gotBody); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			return jsonResponse(req, http.StatusOK, `{"id":"105","name":"test"}`), nil
		}))
		if _, err := client.Tags.Create(context.Background(), TagCreateOrUpdateRequest{Name: "test"}); err != nil {
			t.Fatalf("Create returned error: %v", err)
		}
		if gotBody["name"] != "test" {
			t.Fatalf("body = %#v", gotBody)
		}
	})

	t.Run("tag union request variants", func(t *testing.T) {
		TagCreateOrUpdateRequest{Name: "direct"}.isTagCreateRequest()
		TagCompanyRequest{Name: "direct"}.isTagCreateRequest()
		TagCompanyUntagRequest{Name: "direct"}.isTagCreateRequest()
		TagUsersRequest{Name: "direct"}.isTagCreateRequest()

		tests := []struct {
			name string
			req  TagCreateRequest
			want func(*testing.T, map[string]any)
		}{
			{
				name: "company tagging",
				req: TagCompanyRequest{
					Name:      "enterprise",
					Companies: []TagCompanyReference{{ID: stringPtr("company-1")}},
				},
				want: func(t *testing.T, body map[string]any) {
					t.Helper()
					if body["name"] != "enterprise" {
						t.Fatalf("body = %#v", body)
					}
					companies, ok := body["companies"].([]any)
					if !ok || len(companies) != 1 {
						t.Fatalf("body = %#v", body)
					}
				},
			},
			{
				name: "company untagging",
				req: TagCompanyUntagRequest{
					Name:      "enterprise",
					Companies: []TagCompanyUntagReference{{ID: stringPtr("company-1"), Untag: true}},
				},
				want: func(t *testing.T, body map[string]any) {
					t.Helper()
					companies := body["companies"].([]any)
					company := companies[0].(map[string]any)
					if company["untag"] != true {
						t.Fatalf("body = %#v", body)
					}
				},
			},
			{
				name: "user tagging",
				req: TagUsersRequest{
					Name:  "enterprise",
					Users: []TagUserReference{{ID: "contact-1"}},
				},
				want: func(t *testing.T, body map[string]any) {
					t.Helper()
					users, ok := body["users"].([]any)
					if !ok || len(users) != 1 {
						t.Fatalf("body = %#v", body)
					}
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tt.req.isTagCreateRequest()
				var gotBody map[string]any
				client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
					if err := json.NewDecoder(req.Body).Decode(&gotBody); err != nil {
						t.Fatalf("decode body: %v", err)
					}
					return jsonResponse(req, http.StatusOK, `{"id":"105","name":"test"}`), nil
				}))
				if _, err := client.Tags.Create(context.Background(), tt.req); err != nil {
					t.Fatalf("Create returned error: %v", err)
				}
				tt.want(t, gotBody)
			})
		}
	})

	t.Run("tag create marshal error", func(t *testing.T) {
		client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			t.Fatal("unexpected request")
			return nil, nil
		}))
		if _, err := client.Tags.Create(context.Background(), invalidTagCreateRequest{}); err == nil {
			t.Fatal("expected marshal error")
		}
	})

	t.Run("tag create transport error", func(t *testing.T) {
		client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return nil, io.ErrUnexpectedEOF
		}))
		if _, err := client.Tags.Create(context.Background(), TagCreateOrUpdateRequest{Name: "test"}); err == nil {
			t.Fatal("expected transport error")
		}
	})

	t.Run("ticket admin reply request types", func(t *testing.T) {
		var gotBody map[string]any
		skip := false
		body := "internal note"
		createdAt := 123
		crossPost := true
		replyOptions := []TicketReplyOption{{Text: "Yes", UUID: "e1d7f8f2-1234-4c9f-9f4c-6f4f1a0d0001"}}
		attachments := []string{"https://example.test/a.png"}
		client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if err := json.NewDecoder(req.Body).Decode(&gotBody); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			return jsonResponse(req, http.StatusOK, `{"type":"ticket_part","id":"156"}`), nil
		}))
		req := TicketAdminReply{
			AdminID:        "991267943",
			AttachmentURLs: &attachments,
			Body:           &body,
			CreatedAt:      &createdAt,
			CrossPost:      &crossPost,
			MessageType:    TicketReplyMessageTypeQuickReply,
			ReplyOptions:   &replyOptions,
		}
		req.isTicketReplyRequest()
		if _, err := client.Tickets.Reply(context.Background(), "20", req, &skip); err != nil {
			t.Fatalf("Reply returned error: %v", err)
		}
		if gotBody["type"] != "admin" || gotBody["skip_notifications"] != false || gotBody["cross_post"] != true {
			t.Fatalf("body = %#v", gotBody)
		}
	})

	t.Run("ticket intercom user reply selector", func(t *testing.T) {
		var gotBody map[string]any
		attachments := []string{"https://example.test/a.png"}
		createdAt := 456
		replyOptions := []TicketReplyOption{{Text: "Done", UUID: "e1d7f8f2-1234-4c9f-9f4c-6f4f1a0d0002"}}
		client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if err := json.NewDecoder(req.Body).Decode(&gotBody); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			return jsonResponse(req, http.StatusOK, `{"type":"ticket_part","id":"156"}`), nil
		}))
		reply := TicketContactReply{
			AttachmentURLs: &attachments,
			Body:           "hi",
			Contact:        NewTicketReplyContactByIntercomUserID("contact-1"),
			CreatedAt:      &createdAt,
			MessageType:    TicketReplyMessageTypeComment,
			ReplyOptions:   &replyOptions,
		}
		reply.isTicketReplyRequest()
		if _, err := client.Tickets.Reply(context.Background(), "20", reply, nil); err != nil {
			t.Fatalf("Reply returned error: %v", err)
		}
		if gotBody["intercom_user_id"] != "contact-1" || gotBody["created_at"] != float64(456) {
			t.Fatalf("body = %#v", gotBody)
		}
	})

	t.Run("ticket reply marshal error", func(t *testing.T) {
		client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			t.Fatal("unexpected request")
			return nil, nil
		}))
		if _, err := client.Tickets.Reply(context.Background(), "20", invalidTicketReplyRequest{}, nil); err == nil {
			t.Fatal("expected marshal error")
		}
	})

	t.Run("ticket reply transport error", func(t *testing.T) {
		client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return nil, io.ErrUnexpectedEOF
		}))
		if _, err := client.Tickets.Reply(context.Background(), "20", TicketContactReply{
			Body:        "hi",
			Contact:     NewTicketReplyContactByEmail("user@example.com"),
			MessageType: TicketReplyMessageTypeComment,
		}, nil); err == nil {
			t.Fatal("expected transport error")
		}
	})
}

func stringPtr(s string) *string {
	return &s
}

type invalidTagCreateRequest struct{}

func (invalidTagCreateRequest) isTagCreateRequest() {}

func (invalidTagCreateRequest) MarshalJSON() ([]byte, error) {
	return nil, io.ErrUnexpectedEOF
}

type invalidTicketReplyRequest struct{}

func (invalidTicketReplyRequest) isTicketReplyRequest() {}

func (invalidTicketReplyRequest) payload(skipNotifications *bool) (map[string]any, error) {
	return map[string]any{"bad": make(chan int)}, nil
}
