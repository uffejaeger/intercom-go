package intercom

import (
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestLiveAPIResponseCompatibility(t *testing.T) {
	tests := []struct {
		name     string
		response string
		call     func(context.Context, *Client) error
	}{
		{
			name:     "WhatsApp pagination accepts string per_page",
			response: `{"type":"list","ruleset_id":"ruleset-1","pages":{"type":"pages","per_page":"20","total_pages":1,"next":null},"total_count":0,"events":[]}`,
			call: func(ctx context.Context, client *Client) error {
				statuses, err := client.WhatsApp.GetMessageStatus(ctx, &WhatsAppMessageStatusParams{RulesetId: "ruleset-1"})
				if err != nil {
					return err
				}
				if statuses.Pages.PerPage != 20 {
					t.Fatalf("Pages.PerPage = %d, want 20", statuses.Pages.PerPage)
				}
				return nil
			},
		},
		{
			name:     "admin activity pagination accepts string next cursor",
			response: `{"type":"activity_log.list","activity_logs":[],"pages":{"type":"pages","next":"cursor-1","per_page":50}}`,
			call: func(ctx context.Context, client *Client) error {
				logs, err := client.Admins.ListActivityLogs(ctx, "1700000000")
				if err != nil {
					return err
				}
				if logs.Pages == nil || logs.Pages.Next == nil || logs.Pages.Next.StartingAfter == nil || *logs.Pages.Next.StartingAfter != "cursor-1" {
					t.Fatalf("Pages.Next = %#v", logs.Pages)
				}
				return nil
			},
		},
		{
			name:     "calls accept numeric IDs",
			response: `{"type":"list","data":[{"type":"call","id":123,"conversation_id":456,"admin_id":789,"contact_id":101112}],"pages":{"type":"pages"}}`,
			call: func(ctx context.Context, client *Client) error {
				calls, err := client.Calls.List(ctx)
				if err != nil {
					return err
				}
				if calls.Data == nil || len(*calls.Data) != 1 || (*calls.Data)[0].Id == nil || *(*calls.Data)[0].Id != "123" {
					t.Fatalf("calls.Data = %#v", calls.Data)
				}
				call := (*calls.Data)[0]
				if call.ConversationId == nil || *call.ConversationId != "456" || call.AdminId == nil || *call.AdminId != "789" || call.ContactId == nil || *call.ContactId != "101112" {
					t.Fatalf("call identifiers = %#v", call)
				}
				return nil
			},
		},
		{
			name:     "conversation statistics accept decimal integer values",
			response: `{"type":"conversation.list","conversations":[{"type":"conversation","id":"conversation-1","source":{"attachments":[{"type":"upload","width":"640","height":"480","filesize":"1234"}]},"statistics":{"median_time_to_reply":11.5,"last_closed_by_id":999,"assigned_team_first_response_time":[{"response_time":8.0,"team_id":42}]}}],"total_count":1}`,
			call: func(ctx context.Context, client *Client) error {
				conversations, err := client.Conversations.List(ctx)
				if err != nil {
					return err
				}
				if conversations.Conversations == nil || len(*conversations.Conversations) != 1 {
					t.Fatalf("Conversations = %#v", conversations.Conversations)
				}
				statistics := (*conversations.Conversations)[0].Statistics
				if statistics == nil || statistics.MedianTimeToReply == nil || *statistics.MedianTimeToReply != 11.5 {
					t.Fatalf("Statistics = %#v", statistics)
				}
				if statistics.AssignedTeamFirstResponseTime == nil || len(*statistics.AssignedTeamFirstResponseTime) != 1 || (*statistics.AssignedTeamFirstResponseTime)[0].ResponseTime == nil || *(*statistics.AssignedTeamFirstResponseTime)[0].ResponseTime != 8 {
					t.Fatalf("AssignedTeamFirstResponseTime = %#v", statistics.AssignedTeamFirstResponseTime)
				}
				if statistics.LastClosedById == nil || *statistics.LastClosedById != "999" {
					t.Fatalf("LastClosedById = %v", statistics.LastClosedById)
				}
				source := (*conversations.Conversations)[0].Source
				if source == nil || source.Attachments == nil || len(*source.Attachments) != 1 {
					t.Fatalf("Source = %#v", source)
				}
				attachment := (*source.Attachments)[0]
				if attachment.Width == nil || *attachment.Width != 640 || attachment.Height == nil || *attachment.Height != 480 || attachment.Filesize == nil || *attachment.Filesize != 1234 {
					t.Fatalf("attachment dimensions = %#v", attachment)
				}
				return nil
			},
		},
		{
			name:     "email settings accept RFC3339 timestamps",
			response: `{"type":"list","data":[{"type":"email","id":"email-1","created_at":"2025-07-21T14:44:35.000Z"}]}`,
			call: func(ctx context.Context, client *Client) error {
				emails, err := client.Emails.List(ctx)
				if err != nil {
					return err
				}
				if emails.Data == nil || len(*emails.Data) != 1 || (*emails.Data)[0].CreatedAt == nil || *(*emails.Data)[0].CreatedAt != 1753109075 {
					t.Fatalf("emails.Data = %#v", emails.Data)
				}
				return nil
			},
		},
		{
			name:     "admin detail accepts avatar objects",
			response: `{"type":"admin","id":"1","avatar":{"type":"avatar","image_url":"https://example.test/avatar.png"}}`,
			call: func(ctx context.Context, client *Client) error {
				admin, err := client.Admins.Retrieve(ctx, "1")
				if err != nil {
					return err
				}
				if admin.Avatar == nil || *admin.Avatar != "https://example.test/avatar.png" {
					t.Fatalf("Avatar = %v", admin.Avatar)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := liveRegressionClient(t, http.StatusOK, tt.response, nil)
			if err := tt.call(context.Background(), client); err != nil {
				t.Fatalf("call returned error: %v", err)
			}
		})
	}
}

func TestLiveAPINonUUIDRequestIDPreservesTypedError(t *testing.T) {
	client := liveRegressionClient(t, http.StatusNotFound, `{"type":"error.list","request_id":"0055o8c7almlkvraip50","errors":[{"code":"not_found","message":"Resource Not Found"}]}`, nil)
	checks := []struct {
		name string
		call func(context.Context) error
	}{
		{name: "brand", call: func(ctx context.Context) error { _, err := client.Brands.Retrieve(ctx, "missing"); return err }},
		{name: "macro", call: func(ctx context.Context) error { _, err := client.Macros.Get(ctx, "missing"); return err }},
		{name: "team", call: func(ctx context.Context) error { _, err := client.Teams.Retrieve(ctx, "missing"); return err }},
		{name: "news item", call: func(ctx context.Context) error { _, err := client.News.RetrieveItem(ctx, "1"); return err }},
		{name: "office-hours exception", call: func(ctx context.Context) error {
			_, err := client.OfficeHours.GetException(ctx, "schedule", "exception")
			return err
		}},
		{name: "data-connector execution", call: func(ctx context.Context) error {
			_, err := client.DataConnectors.GetExecutionResult(ctx, "connector", "execution")
			return err
		}},
		{name: "job", call: func(ctx context.Context) error { _, err := client.Workspace.JobStatus(ctx, "job"); return err }},
		{name: "visitor", call: func(ctx context.Context) error { _, err := client.Visitors.GetByUserID(ctx, "external-id"); return err }},
		{name: "article draft", call: func(ctx context.Context) error { _, err := client.Articles.GetDraft(ctx, "1"); return err }},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			err := check.call(context.Background())
			if !IsNotFound(err) {
				t.Fatalf("error = %T %v, want typed 404", err, err)
			}
			var apiErr *ErrorResponse
			if !errors.As(err, &apiErr) {
				t.Fatalf("error = %T, want *ErrorResponse", err)
			}
			if apiErr.RequestID != "0055o8c7almlkvraip50" {
				t.Fatalf("RequestID = %q", apiErr.RequestID)
			}
		})
	}
}

func TestLiveAPIContactNoteUsesStringContactID(t *testing.T) {
	const contactID = "6762f0ad1bb69f9f2193bb62"
	var gotPath string
	client := liveRegressionClient(t, http.StatusOK, `{"type":"note","id":"31","body":"<p>Hello</p>"}`, func(req *http.Request) {
		gotPath = req.URL.EscapedPath()
	})

	if _, err := client.Contacts.CreateNote(context.Background(), contactID, "Hello", "991267583"); err != nil {
		t.Fatalf("CreateNote returned error: %v", err)
	}
	if gotPath != "/contacts/"+contactID+"/notes" {
		t.Fatalf("path = %q", gotPath)
	}
}

func TestLiveAPIMergeHistoryFallsBackToContactField(t *testing.T) {
	const contactID = "6762f0ad1bb69f9f2193bb62"
	requestCount := 0
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requestCount++
		status := http.StatusNotFound
		body := `{"type":"error.list","request_id":"request-1","errors":[{"code":"not_found","message":"missing"}]}`
		switch requestCount {
		case 1:
			if req.URL.EscapedPath() != "/contacts/"+contactID+"/merge_history" {
				t.Fatalf("first path = %q", req.URL.EscapedPath())
			}
		case 2:
			if req.URL.EscapedPath() != "/contacts/"+contactID || req.URL.Query().Get("include_merge_history") != "true" {
				t.Fatalf("fallback URL = %s", req.URL.String())
			}
			status = http.StatusOK
			body = `{"type":"contact","id":"6762f0ad1bb69f9f2193bb62","role":"user","merge_history":[{"type":"merge_history","source_contact_id":"6762f0ad1bb69f9f2193bb61","source_contact_role":"lead","merged_at":1700000000}]}`
		default:
			t.Fatalf("unexpected request %d", requestCount)
		}
		return &http.Response{
			StatusCode: status,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    req,
		}, nil
	})
	client, err := NewClient("token", WithBaseURL("https://example.test"), WithHTTPClient(&http.Client{Transport: transport}))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	history, err := client.Contacts.ListMergeHistory(context.Background(), contactID)
	if err != nil {
		t.Fatalf("ListMergeHistory returned error: %v", err)
	}
	if requestCount != 2 || history.Data == nil || len(*history.Data) != 1 || (*history.Data)[0].SourceContactId == nil || *(*history.Data)[0].SourceContactId != "6762f0ad1bb69f9f2193bb61" {
		t.Fatalf("history = %#v; requests = %d", history, requestCount)
	}
}

func TestLiveAPIMergeHistoryFallbackErrors(t *testing.T) {
	tests := []struct {
		name         string
		secondStatus int
		secondBody   string
		secondErr    error
	}{
		{name: "transport error", secondErr: errors.New("fallback failed")},
		{name: "API error", secondStatus: http.StatusInternalServerError, secondBody: `{"type":"error.list","request_id":"request-2","errors":[{"code":"server_error","message":"failed"}]}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestCount := 0
			transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				requestCount++
				if requestCount == 2 && tt.secondErr != nil {
					return nil, tt.secondErr
				}
				status := http.StatusNotFound
				body := `{"type":"error.list","request_id":"request-1","errors":[{"code":"not_found","message":"missing"}]}`
				if requestCount == 2 {
					status = tt.secondStatus
					body = tt.secondBody
				}
				return &http.Response{
					StatusCode: status,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(strings.NewReader(body)),
					Request:    req,
				}, nil
			})
			client, err := NewClient("token", WithBaseURL("https://example.test"), WithHTTPClient(&http.Client{Transport: transport}))
			if err != nil {
				t.Fatalf("NewClient returned error: %v", err)
			}

			if _, err := client.Contacts.ListMergeHistory(context.Background(), "6762f0ad1bb69f9f2193bb62"); err == nil {
				t.Fatal("expected fallback error")
			}
		})
	}
}

func TestLiveAPIConversationCreateUsesStringContactID(t *testing.T) {
	typeOfRequest := reflect.TypeFor[ConversationCreate]()
	from, ok := typeOfRequest.FieldByName("From")
	if !ok {
		t.Fatal("ConversationCreate.From is missing")
	}
	id, ok := from.Type.FieldByName("Id")
	if !ok {
		t.Fatal("ConversationCreate.From.Id is missing")
	}
	if id.Type.Kind() != reflect.String {
		t.Fatalf("ConversationCreate.From.Id kind = %s, want string", id.Type.Kind())
	}
}

func TestLiveAPIEmptyArticleDraftResponseIsSemanticError(t *testing.T) {
	client := liveRegressionClient(t, http.StatusOK, "", nil)

	_, err := client.Articles.GetDraft(context.Background(), "123")
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "unexpected end of JSON input") {
		t.Fatalf("error leaked generated JSON failure: %v", err)
	}
	if !strings.Contains(err.Error(), "without a response body") {
		t.Fatalf("error = %v, want missing response body error", err)
	}
}

func liveRegressionClient(t *testing.T, status int, response string, inspect func(*http.Request)) *Client {
	t.Helper()
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if inspect != nil {
			inspect(req)
		}
		return &http.Response{
			StatusCode: status,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(response)),
			Request:    req,
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
	return client
}
