package intercom

import (
	"context"
	"net/http"
	"testing"
	"time"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

func TestRemainingServicesRequests(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name       string
		statusCode int
		response   string
		call       func(context.Context, *Client) error
		wantMethod string
		wantPath   string
		wantQuery  string
	}{
		{
			name:       "create tag",
			statusCode: http.StatusOK,
			response:   `{"id":"105","name":"test"}`,
			call: func(ctx context.Context, client *Client) error {
				tag, err := client.Tags.Create(ctx, map[string]any{"name": "test"})
				if err != nil {
					return err
				}
				if tag.Id == nil || *tag.Id != "105" {
					t.Fatalf("tag.Id = %v", tag.Id)
				}
				return nil
			},
			wantMethod: http.MethodPost,
			wantPath:   "/tags",
		},
		{
			name:       "create message",
			statusCode: http.StatusOK,
			response:   `{"id":"19","type":"admin_message","body":"hello","created_at":1734537786,"message_type":"inapp"}`,
			call: func(ctx context.Context, client *Client) error {
				body := "hello"
				_, err := client.Messages.Create(ctx, MessageCreate{Body: &body})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/messages",
		},
		{
			name:       "list calls",
			statusCode: http.StatusOK,
			response:   `{"type":"list","data":[],"total_count":0}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Calls.List(ctx)
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/calls",
		},
		{
			name:       "list calls with transcripts",
			statusCode: http.StatusOK,
			response:   `{"type":"list","data":[]}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Calls.ListWithTranscripts(ctx, []string{"c1"})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/calls/search",
		},
		{
			name:       "call recording",
			statusCode: http.StatusOK,
			response:   `recording-bytes`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Calls.Recording(ctx, "123")
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/calls/123/recording",
		},
		{
			name:       "list internal articles",
			statusCode: http.StatusOK,
			response:   `{"type":"list","data":[],"total_count":0}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.InternalArticles.List(ctx)
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/internal_articles",
		},
		{
			name:       "search internal articles",
			statusCode: http.StatusOK,
			response:   `{"type":"list","data":{"internal_articles":[]},"total_count":0}`,
			call: func(ctx context.Context, client *Client) error {
				folderID := "folder-1"
				_, err := client.InternalArticles.Search(ctx, &folderID)
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/internal_articles/search",
			wantQuery:  "folder_id=folder-1",
		},
		{
			name:       "get ip allowlist",
			statusCode: http.StatusOK,
			response:   `{"type":"ip_allowlist","enabled":true,"ip_allowlist":["192.168.0.1"]}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Workspace.GetIPAllowlist(ctx)
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/ip_allowlist",
		},
		{
			name:       "job status",
			statusCode: http.StatusOK,
			response:   `{"type":"job","id":"2","status":"success"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Workspace.JobStatus(ctx, "2")
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/jobs/status/2",
		},
		{
			name:       "create data export",
			statusCode: http.StatusOK,
			response:   `{"job_identifier":"job-1","status":"pending"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Workspace.CreateDataExport(ctx, DataExportCreate{CreatedAtAfter: 1, CreatedAtBefore: 2})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/export/content/data",
		},
		{
			name:       "get reporting export job",
			statusCode: http.StatusOK,
			response:   `{"job_identifier":"job-1","status":"complete"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Workspace.GetReportingExportJob(ctx, "job-1", "app-1", "client-1")
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/export/reporting_data/job-1",
			wantQuery:  "app_id=app-1&client_id=client-1",
		},
		{
			name:       "download reporting export",
			statusCode: http.StatusOK,
			response:   `csv-bytes`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Workspace.DownloadReportingExport(ctx, "job-1", "app-1")
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/download/reporting_data/job-1",
			wantQuery:  "app_id=app-1",
		},
		{
			name:       "create phone switch",
			statusCode: http.StatusOK,
			response:   `{"type":"phone_switch","phone":"+4512345678"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.PhoneSwitches.Create(ctx, PhoneSwitchCreate{Phone: "+4512345678"})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/phone_call_redirects",
		},
		{
			name:       "retrieve visitor by user id",
			statusCode: http.StatusOK,
			response:   `{"type":"visitor","id":"v1","user_id":"u1"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Visitors.GetByUserID(ctx, "u1")
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/visitors",
			wantQuery:  "user_id=u1",
		},
		{
			name:       "convert visitor",
			statusCode: http.StatusOK,
			response:   `{"type":"contact","id":"c1"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Visitors.Convert(ctx, VisitorConvert{Type: "user"})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/visitors/convert",
		},
		{
			name:       "reply to fin",
			statusCode: http.StatusOK,
			response:   `{"conversation_id":"c1","status":"awaiting_user_reply","user_id":"u1"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Fin.Reply(ctx, FinReply{
					ConversationId: "c1",
					FinAgentMessageSchema: gen.FinAgentMessageSchema{
						Author:    "user",
						Body:      "hi",
						Timestamp: now,
					},
					FinAgentUserSchema: gen.FinAgentUserSchema{Id: "u1"},
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/fin/reply",
		},
		{
			name:       "register fin voice call",
			statusCode: http.StatusOK,
			response:   `{"id":1,"type":"ai_call"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Fin.RegisterVoiceCall(ctx, RegisterFinVoiceCallRequest{
					CallId:      "call-1",
					PhoneNumber: "+4512345678",
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/fin_voice/register",
		},
		{
			name:       "enqueue create ticket",
			statusCode: http.StatusOK,
			response:   `{"type":"job","id":"2","status":"pending"}`,
			call: func(ctx context.Context, client *Client) error {
				var contact gen.CreateTicketRequest_Contacts_Item
				_ = contact.FromCreateTicketRequestContacts0(gen.CreateTicketRequestContacts0{Id: "1234"})
				_, err := client.Tickets.EnqueueCreate(ctx, TicketCreate{
					TicketTypeId: "1234",
					Contacts:     []gen.CreateTicketRequest_Contacts_Item{contact},
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/tickets/enqueue",
		},
		{
			name:       "reply ticket",
			statusCode: http.StatusOK,
			response:   `{"type":"ticket_part","id":"156"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Tickets.Reply(ctx, "20", map[string]any{
					"type":         "admin",
					"admin_id":     "991267943",
					"message_type": "note",
					"body":         "hi",
				}, nil)
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/tickets/20/reply",
		},
		{
			name:       "attach tag to ticket",
			statusCode: http.StatusOK,
			response:   `{"type":"tag","id":"121","name":"Manual tag"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Tickets.AttachTag(ctx, "20", gen.AttachTagToTicketJSONBody{Id: "121", AdminId: "991267958"})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/tickets/20/tags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod, gotPath, gotQuery string
			transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				gotMethod = req.Method
				gotPath = req.URL.Path
				gotQuery = req.URL.RawQuery
				return &http.Response{
					StatusCode: tt.statusCode,
					Body:       http.NoBody,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Request:    req,
				}, nil
			})

			if tt.response != "" {
				transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
					gotMethod = req.Method
					gotPath = req.URL.Path
					gotQuery = req.URL.RawQuery
					return jsonResponse(req, tt.statusCode, tt.response), nil
				})
			}

			client := newSupportingServicesTestClient(t, transport)
			if err := tt.call(context.Background(), client); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotMethod != tt.wantMethod {
				t.Fatalf("method = %q, want %q", gotMethod, tt.wantMethod)
			}
			if gotPath != tt.wantPath {
				t.Fatalf("path = %q, want %q", gotPath, tt.wantPath)
			}
			if tt.wantQuery != "" && gotQuery != tt.wantQuery {
				t.Fatalf("query = %q, want %q", gotQuery, tt.wantQuery)
			}
		})
	}
}
