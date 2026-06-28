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
		client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			gotPath = req.URL.Path
			return jsonResponse(req, http.StatusOK, `{"id":"121","name":"Manual tag"}`), nil
		}))
		if _, err := client.Tickets.AttachTag(context.Background(), "20", TicketTagAttachRequest{Id: "121", AdminId: "991267958"}); err != nil {
			t.Fatalf("AttachTag returned error: %v", err)
		}
		if gotPath != "/tickets/20/tags" {
			t.Fatalf("path = %q", gotPath)
		}
	})
}
