package intercom

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

func TestCoverageCompletionRequests(t *testing.T) {
	t.Run("calls transcript aliases", func(t *testing.T) {
		client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"text/plain"}},
				Body:       io.NopCloser(strings.NewReader("transcript-data")),
				Request:    req,
			}, nil
		}))

		for _, call := range []func(context.Context) ([]byte, error){
			func(ctx context.Context) ([]byte, error) { return client.Calls.GetTranscript(ctx, "call-1") },
			func(ctx context.Context) ([]byte, error) { return client.Calls.Transcript(ctx, "call-1") },
		} {
			data, err := call(context.Background())
			if err != nil {
				t.Fatalf("call returned error: %v", err)
			}
			if string(data) != "transcript-data" {
				t.Fatalf("data = %q", string(data))
			}
		}
	})

	t.Run("fin voice call wrappers", func(t *testing.T) {
		client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/fin/start":
				return jsonResponse(req, http.StatusOK, `{"conversation_id":"c1","status":"open","user_id":"u1"}`), nil
			case "/fin_voice/collect/1":
				return jsonResponse(req, http.StatusOK, `{"id":1,"status":"registered"}`), nil
			case "/fin_voice/conversation/conversation-1":
				return jsonResponse(req, http.StatusOK, `[{"id":1,"status":"registered"}]`), nil
			case "/fin_voice/external_id/external-1":
				return jsonResponse(req, http.StatusOK, `{"id":1,"external_call_id":"external-1","status":"registered"}`), nil
			case "/fin_voice/phone_number/+4512345678":
				return jsonResponse(req, http.StatusOK, `{"id":1,"user_phone_number":"+4512345678","status":"registered"}`), nil
			default:
				t.Fatalf("unexpected path: %s", req.URL.Path)
				return nil, nil
			}
		}))

		if _, err := client.Fin.StartConversation(context.Background(), FinStartConversation{}); err != nil {
			t.Fatalf("StartConversation returned error: %v", err)
		}
		if _, err := client.Fin.GetVoiceCallByID(context.Background(), "1"); err != nil {
			t.Fatalf("GetVoiceCallByID returned error: %v", err)
		}
		if _, err := client.Fin.ListVoiceCallsByConversation(context.Background(), "conversation-1"); err != nil {
			t.Fatalf("ListVoiceCallsByConversation returned error: %v", err)
		}
		if _, err := client.Fin.GetVoiceCallByExternalID(context.Background(), "external-1"); err != nil {
			t.Fatalf("GetVoiceCallByExternalID returned error: %v", err)
		}
		if _, err := client.Fin.GetVoiceCallByPhoneNumber(context.Background(), "+4512345678"); err != nil {
			t.Fatalf("GetVoiceCallByPhoneNumber returned error: %v", err)
		}
	})

	t.Run("internal articles requests", func(t *testing.T) {
		client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/internal_articles":
				if req.Method == http.MethodPost {
					return jsonResponse(req, http.StatusOK, `{"id":"10","title":"Internal article"}`), nil
				}
			case "/internal_articles/10":
				switch req.Method {
				case http.MethodGet, http.MethodPut:
					return jsonResponse(req, http.StatusOK, `{"id":"10","title":"Internal article"}`), nil
				case http.MethodDelete:
					return jsonResponse(req, http.StatusOK, `{"id":"10","object":"internal_article","deleted":true}`), nil
				}
			}
			t.Fatalf("unexpected request: %s %s", req.Method, req.URL.Path)
			return nil, nil
		}))

		if _, err := client.InternalArticles.Create(context.Background(), InternalArticleCreate{}); err != nil {
			t.Fatalf("Create returned error: %v", err)
		}
		if _, err := client.InternalArticles.Retrieve(context.Background(), "10"); err != nil {
			t.Fatalf("Retrieve returned error: %v", err)
		}
		if _, err := client.InternalArticles.Update(context.Background(), "10", InternalArticleUpdate{}); err != nil {
			t.Fatalf("Update returned error: %v", err)
		}
		if _, err := client.InternalArticles.Delete(context.Background(), "10"); err != nil {
			t.Fatalf("Delete returned error: %v", err)
		}
	})

	t.Run("visitors update", func(t *testing.T) {
		client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return jsonResponse(req, http.StatusOK, `{"type":"visitor","id":"v1","user_id":"u1"}`), nil
		}))
		if _, err := client.Visitors.Update(context.Background(), VisitorUpdate{}); err != nil {
			t.Fatalf("Update returned error: %v", err)
		}
	})

	t.Run("workspace requests", func(t *testing.T) {
		client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/ip_allowlist":
				return jsonResponse(req, http.StatusOK, `{"type":"ip_allowlist","enabled":true,"ip_allowlist":["1.1.1.1"]}`), nil
			case "/download/content/data/job-1":
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/octet-stream"}},
					Body:       io.NopCloser(strings.NewReader("export-bytes")),
					Request:    req,
				}, nil
			case "/export/content/data/job-1", "/export/cancel/job-1":
				return jsonResponse(req, http.StatusOK, `{"job_identifier":"job-1","status":"complete"}`), nil
			case "/export/reporting_data/get_datasets":
				return jsonResponse(req, http.StatusOK, `{"type":"list","data":[{"id":"dataset-1","name":"Conversations"}]}`), nil
			case "/export/reporting_data/enqueue":
				return jsonResponse(req, http.StatusOK, `{"job_identifier":"job-2","status":"pending"}`), nil
			case "/export/workflows/workflow-1":
				return jsonResponse(req, http.StatusOK, `{"app_id":1,"workflow":{"id":"workflow-1","title":"Workflow"}}`), nil
			default:
				t.Fatalf("unexpected request: %s %s", req.Method, req.URL.Path)
				return nil, nil
			}
		}))

		if _, err := client.Workspace.UpdateIPAllowlist(context.Background(), IPAllowlist{}); err != nil {
			t.Fatalf("UpdateIPAllowlist returned error: %v", err)
		}
		if data, err := client.Workspace.DownloadDataExport(context.Background(), "job-1"); err != nil || string(data) != "export-bytes" {
			t.Fatalf("DownloadDataExport = %q, %v", string(data), err)
		}
		if _, err := client.Workspace.GetDataExport(context.Background(), "job-1"); err != nil {
			t.Fatalf("GetDataExport returned error: %v", err)
		}
		if _, err := client.Workspace.CancelDataExport(context.Background(), "job-1"); err != nil {
			t.Fatalf("CancelDataExport returned error: %v", err)
		}
		if _, err := client.Workspace.ListReportingDatasets(context.Background()); err != nil {
			t.Fatalf("ListReportingDatasets returned error: %v", err)
		}
		if _, err := client.Workspace.CreateReportingExport(context.Background(), ReportingExportCreate{}); err != nil {
			t.Fatalf("CreateReportingExport returned error: %v", err)
		}
		if _, err := client.Workspace.ExportWorkflow(context.Background(), "workflow-1"); err != nil {
			t.Fatalf("ExportWorkflow returned error: %v", err)
		}
	})

	t.Run("tickets detach tag", func(t *testing.T) {
		client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return jsonResponse(req, http.StatusOK, `{"id":"121","name":"Manual tag"}`), nil
		}))
		_, err := client.Tickets.DetachTag(context.Background(), "20", "121", gen.DetachTagFromTicketJSONBody{AdminId: "991267958"})
		if err != nil {
			t.Fatalf("DetachTag returned error: %v", err)
		}
	})
}

func TestCoverageCompletionTransportErrors(t *testing.T) {
	transportErr := errors.New("connection refused")
	client := newSupportingServicesTestClient(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, transportErr
	}))
	ctx := context.Background()
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	body := "hello"
	name := "collection"
	contact := gen.CreateTicketRequest_Contacts_Item{}
	_ = contact.FromCreateTicketRequestContacts0(gen.CreateTicketRequestContacts0{Id: "1234"})

	tests := []struct {
		name string
		call func() error
	}{
		{"collections list", func() error { _, err := client.Collections.List(ctx); return err }},
		{"collections create", func() error { _, err := client.Collections.Create(ctx, CollectionCreate{Name: name}); return err }},
		{"collections retrieve", func() error { _, err := client.Collections.Retrieve(ctx, "10"); return err }},
		{"collections update", func() error { _, err := client.Collections.Update(ctx, "10", CollectionUpdate{}); return err }},
		{"collections delete", func() error { return client.Collections.Delete(ctx, "10") }},
		{"help centers list", func() error { _, err := client.HelpCenters.List(ctx); return err }},
		{"help centers retrieve", func() error { _, err := client.HelpCenters.Retrieve(ctx, "10"); return err }},
		{"messages create", func() error { _, err := client.Messages.Create(ctx, MessageCreate{Body: &body}); return err }},
		{"news list items", func() error { _, err := client.News.ListItems(ctx); return err }},
		{"news create item", func() error { _, err := client.News.CreateItem(ctx, NewsItemCreate{Title: "news"}); return err }},
		{"news retrieve item", func() error { _, err := client.News.RetrieveItem(ctx, "10"); return err }},
		{"news update item", func() error { _, err := client.News.UpdateItem(ctx, "10", NewsItemUpdate{Title: "news"}); return err }},
		{"news delete item", func() error { return client.News.DeleteItem(ctx, "10") }},
		{"news list feeds", func() error { _, err := client.News.ListFeeds(ctx); return err }},
		{"news retrieve feed", func() error { _, err := client.News.RetrieveFeed(ctx, "feed-1"); return err }},
		{"news list feed items", func() error { _, err := client.News.ListFeedItems(ctx, "feed-1"); return err }},
		{"phone switches create", func() error {
			_, err := client.PhoneSwitches.Create(ctx, PhoneSwitchCreate{Phone: "+4512345678"})
			return err
		}},
		{"segments list", func() error { _, err := client.Segments.List(ctx); return err }},
		{"segments retrieve", func() error { _, err := client.Segments.Retrieve(ctx, "segment-1"); return err }},
		{"subscription types list", func() error { _, err := client.SubscriptionTypes.List(ctx); return err }},
		{"tags list", func() error { _, err := client.Tags.List(ctx); return err }},
		{"tags create", func() error { _, err := client.Tags.Create(ctx, map[string]any{"name": "tag"}); return err }},
		{"tags retrieve", func() error { _, err := client.Tags.Retrieve(ctx, "10"); return err }},
		{"tags delete", func() error { return client.Tags.Delete(ctx, "10") }},
		{"teams list", func() error { _, err := client.Teams.List(ctx); return err }},
		{"teams retrieve", func() error { _, err := client.Teams.Retrieve(ctx, "10"); return err }},
		{"tickets create", func() error {
			_, err := client.Tickets.Create(ctx, TicketCreate{TicketTypeId: "1234", Contacts: []gen.CreateTicketRequest_Contacts_Item{contact}})
			return err
		}},
		{"tickets enqueue create", func() error {
			_, err := client.Tickets.EnqueueCreate(ctx, TicketCreate{TicketTypeId: "1234", Contacts: []gen.CreateTicketRequest_Contacts_Item{contact}})
			return err
		}},
		{"tickets search", func() error { _, err := client.Tickets.Search(ctx, TicketSearchQuery{}); return err }},
		{"tickets get", func() error { _, err := client.Tickets.Get(ctx, "20"); return err }},
		{"tickets update", func() error { _, err := client.Tickets.Update(ctx, "20", TicketUpdate{}); return err }},
		{"tickets delete", func() error { return client.Tickets.Delete(ctx, "20") }},
		{"tickets reply", func() error { _, err := client.Tickets.Reply(ctx, "20", map[string]any{"body": "hi"}, nil); return err }},
		{"tickets list states", func() error { _, err := client.Tickets.ListStates(ctx); return err }},
		{"tickets list types", func() error { _, err := client.Tickets.ListTypes(ctx); return err }},
		{"tickets get type", func() error { _, err := client.Tickets.GetType(ctx, "1"); return err }},
		{"tickets create type", func() error { _, err := client.Tickets.CreateType(ctx, TicketTypeCreate{Name: "Bug"}); return err }},
		{"tickets update type", func() error { _, err := client.Tickets.UpdateType(ctx, "1", TicketTypeUpdate{}); return err }},
		{"tickets create type attribute", func() error {
			_, err := client.Tickets.CreateTypeAttribute(ctx, "1", TicketTypeAttributeCreate{Name: "Priority"})
			return err
		}},
		{"tickets update type attribute", func() error {
			_, err := client.Tickets.UpdateTypeAttribute(ctx, "1", "2", TicketTypeAttributeUpdate{})
			return err
		}},
		{"tickets attach tag", func() error {
			_, err := client.Tickets.AttachTag(ctx, "20", gen.AttachTagToTicketJSONBody{Id: "121", AdminId: "1"})
			return err
		}},
		{"tickets detach tag", func() error {
			_, err := client.Tickets.DetachTag(ctx, "20", "121", gen.DetachTagFromTicketJSONBody{AdminId: "1"})
			return err
		}},
		{"visitors get by user id", func() error { _, err := client.Visitors.GetByUserID(ctx, "u1"); return err }},
		{"visitors update", func() error { _, err := client.Visitors.Update(ctx, VisitorUpdate{}); return err }},
		{"visitors convert", func() error { _, err := client.Visitors.Convert(ctx, VisitorConvert{Type: "user"}); return err }},
		{"workspace get ip allowlist", func() error { _, err := client.Workspace.GetIPAllowlist(ctx); return err }},
		{"workspace update ip allowlist", func() error { _, err := client.Workspace.UpdateIPAllowlist(ctx, IPAllowlist{}); return err }},
		{"workspace job status", func() error { _, err := client.Workspace.JobStatus(ctx, "1"); return err }},
		{"workspace create data export", func() error { _, err := client.Workspace.CreateDataExport(ctx, DataExportCreate{}); return err }},
		{"workspace download data export", func() error { _, err := client.Workspace.DownloadDataExport(ctx, "job-1"); return err }},
		{"workspace get data export", func() error { _, err := client.Workspace.GetDataExport(ctx, "job-1"); return err }},
		{"workspace cancel data export", func() error { _, err := client.Workspace.CancelDataExport(ctx, "job-1"); return err }},
		{"workspace list reporting datasets", func() error { _, err := client.Workspace.ListReportingDatasets(ctx); return err }},
		{"workspace create reporting export", func() error {
			_, err := client.Workspace.CreateReportingExport(ctx, ReportingExportCreate{})
			return err
		}},
		{"workspace get reporting export job", func() error {
			_, err := client.Workspace.GetReportingExportJob(ctx, "job-1", "app-1", "client-1")
			return err
		}},
		{"workspace download reporting export", func() error { _, err := client.Workspace.DownloadReportingExport(ctx, "job-1", "app-1"); return err }},
		{"workspace export workflow", func() error { _, err := client.Workspace.ExportWorkflow(ctx, "workflow-1"); return err }},
		{"internal articles list", func() error { _, err := client.InternalArticles.List(ctx); return err }},
		{"internal articles create", func() error { _, err := client.InternalArticles.Create(ctx, InternalArticleCreate{}); return err }},
		{"internal articles search", func() error { _, err := client.InternalArticles.Search(ctx, nil); return err }},
		{"internal articles retrieve", func() error { _, err := client.InternalArticles.Retrieve(ctx, "10"); return err }},
		{"internal articles update", func() error { _, err := client.InternalArticles.Update(ctx, "10", InternalArticleUpdate{}); return err }},
		{"internal articles delete", func() error { _, err := client.InternalArticles.Delete(ctx, "10"); return err }},
		{"calls get transcript", func() error { _, err := client.Calls.GetTranscript(ctx, "call-1"); return err }},
		{"calls transcript", func() error { _, err := client.Calls.Transcript(ctx, "call-1"); return err }},
		{"fin reply", func() error {
			_, err := client.Fin.Reply(ctx, FinReply{
				ConversationId:        "c1",
				FinAgentMessageSchema: gen.FinAgentMessageSchema{Author: "user", Body: "hi", Timestamp: now},
				FinAgentUserSchema:    gen.FinAgentUserSchema{Id: "u1"},
			})
			return err
		}},
		{"fin start conversation", func() error { _, err := client.Fin.StartConversation(ctx, FinStartConversation{}); return err }},
		{"fin get voice call by id", func() error { _, err := client.Fin.GetVoiceCallByID(ctx, "1"); return err }},
		{"fin list voice calls by conversation", func() error { _, err := client.Fin.ListVoiceCallsByConversation(ctx, "conversation-1"); return err }},
		{"fin get voice call by external id", func() error { _, err := client.Fin.GetVoiceCallByExternalID(ctx, "external-1"); return err }},
		{"fin get voice call by phone number", func() error { _, err := client.Fin.GetVoiceCallByPhoneNumber(ctx, "+4512345678"); return err }},
		{"fin register voice call", func() error { _, err := client.Fin.RegisterVoiceCall(ctx, RegisterFinVoiceCallRequest{}); return err }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err == nil {
				t.Fatal("expected transport error")
			}
		})
	}
}

func TestCoverageCompletionValidation(t *testing.T) {
	client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected request: %s %s", req.Method, req.URL.Path)
		return nil, nil
	}))
	ctx := context.Background()

	tests := []struct {
		name string
		call func() error
	}{
		{"help centers retrieve empty", func() error { _, err := client.HelpCenters.Retrieve(ctx, ""); return err }},
		{"help centers retrieve non-numeric", func() error { _, err := client.HelpCenters.Retrieve(ctx, "abc"); return err }},
		{"internal articles retrieve empty", func() error { _, err := client.InternalArticles.Retrieve(ctx, ""); return err }},
		{"internal articles retrieve non-numeric", func() error { _, err := client.InternalArticles.Retrieve(ctx, "abc"); return err }},
		{"internal articles update empty", func() error { _, err := client.InternalArticles.Update(ctx, "", InternalArticleUpdate{}); return err }},
		{"internal articles update non-numeric", func() error {
			_, err := client.InternalArticles.Update(ctx, "abc", InternalArticleUpdate{})
			return err
		}},
		{"internal articles delete empty", func() error { _, err := client.InternalArticles.Delete(ctx, ""); return err }},
		{"internal articles delete non-numeric", func() error { _, err := client.InternalArticles.Delete(ctx, "abc"); return err }},
		{"segments retrieve empty", func() error { _, err := client.Segments.Retrieve(ctx, ""); return err }},
		{"tags retrieve empty", func() error { _, err := client.Tags.Retrieve(ctx, ""); return err }},
		{"tags delete empty", func() error { return client.Tags.Delete(ctx, "") }},
		{"news retrieve item empty", func() error { _, err := client.News.RetrieveItem(ctx, ""); return err }},
		{"news retrieve item non-numeric", func() error { _, err := client.News.RetrieveItem(ctx, "abc"); return err }},
		{"news update item empty", func() error { _, err := client.News.UpdateItem(ctx, "", NewsItemUpdate{}); return err }},
		{"news update item non-numeric", func() error { _, err := client.News.UpdateItem(ctx, "abc", NewsItemUpdate{}); return err }},
		{"news delete item empty", func() error { return client.News.DeleteItem(ctx, "") }},
		{"news delete item non-numeric", func() error { return client.News.DeleteItem(ctx, "abc") }},
		{"news retrieve feed empty", func() error { _, err := client.News.RetrieveFeed(ctx, ""); return err }},
		{"news list feed items empty", func() error { _, err := client.News.ListFeedItems(ctx, ""); return err }},
		{"notes retrieve empty", func() error { _, err := client.Notes.Retrieve(ctx, ""); return err }},
		{"notes retrieve non-numeric", func() error { _, err := client.Notes.Retrieve(ctx, "abc"); return err }},
		{"collections delete empty", func() error { return client.Collections.Delete(ctx, "") }},
		{"collections delete non-numeric", func() error { return client.Collections.Delete(ctx, "abc") }},
		{"collections update empty", func() error { _, err := client.Collections.Update(ctx, "", CollectionUpdate{}); return err }},
		{"collections update non-numeric", func() error { _, err := client.Collections.Update(ctx, "abc", CollectionUpdate{}); return err }},
		{"tickets get empty", func() error { _, err := client.Tickets.Get(ctx, ""); return err }},
		{"tickets update empty", func() error { _, err := client.Tickets.Update(ctx, "", TicketUpdate{}); return err }},
		{"tickets delete empty", func() error { return client.Tickets.Delete(ctx, "") }},
		{"tickets reply empty", func() error { _, err := client.Tickets.Reply(ctx, "", map[string]any{}, nil); return err }},
		{"tickets get type empty", func() error { _, err := client.Tickets.GetType(ctx, ""); return err }},
		{"tickets update type empty", func() error { _, err := client.Tickets.UpdateType(ctx, "", TicketTypeUpdate{}); return err }},
		{"tickets create type attribute empty type", func() error {
			_, err := client.Tickets.CreateTypeAttribute(ctx, "", TicketTypeAttributeCreate{})
			return err
		}},
		{"tickets update type attribute empty type", func() error {
			_, err := client.Tickets.UpdateTypeAttribute(ctx, "", "1", TicketTypeAttributeUpdate{})
			return err
		}},
		{"tickets update type attribute empty attribute", func() error {
			_, err := client.Tickets.UpdateTypeAttribute(ctx, "1", "", TicketTypeAttributeUpdate{})
			return err
		}},
		{"tickets attach tag empty", func() error { _, err := client.Tickets.AttachTag(ctx, "", gen.AttachTagToTicketJSONBody{}); return err }},
		{"tickets detach tag empty ticket", func() error {
			_, err := client.Tickets.DetachTag(ctx, "", "1", gen.DetachTagFromTicketJSONBody{})
			return err
		}},
		{"tickets detach tag empty tag", func() error {
			_, err := client.Tickets.DetachTag(ctx, "1", "", gen.DetachTagFromTicketJSONBody{})
			return err
		}},
		{"visitors get empty", func() error { _, err := client.Visitors.GetByUserID(ctx, ""); return err }},
		{"workspace job status empty", func() error { _, err := client.Workspace.JobStatus(ctx, ""); return err }},
		{"workspace download data export empty", func() error { _, err := client.Workspace.DownloadDataExport(ctx, ""); return err }},
		{"workspace get data export empty", func() error { _, err := client.Workspace.GetDataExport(ctx, ""); return err }},
		{"workspace cancel data export empty", func() error { _, err := client.Workspace.CancelDataExport(ctx, ""); return err }},
		{"workspace get reporting export job empty id", func() error {
			_, err := client.Workspace.GetReportingExportJob(ctx, "", "app-1", "client-1")
			return err
		}},
		{"workspace get reporting export job empty app", func() error {
			_, err := client.Workspace.GetReportingExportJob(ctx, "job-1", "", "client-1")
			return err
		}},
		{"workspace get reporting export job empty client", func() error { _, err := client.Workspace.GetReportingExportJob(ctx, "job-1", "app-1", ""); return err }},
		{"workspace download reporting export empty id", func() error { _, err := client.Workspace.DownloadReportingExport(ctx, "", "app-1"); return err }},
		{"workspace download reporting export empty app", func() error { _, err := client.Workspace.DownloadReportingExport(ctx, "job-1", ""); return err }},
		{"workspace export workflow empty", func() error { _, err := client.Workspace.ExportWorkflow(ctx, ""); return err }},
		{"calls get transcript empty", func() error { _, err := client.Calls.GetTranscript(ctx, ""); return err }},
		{"calls transcript empty", func() error { _, err := client.Calls.Transcript(ctx, ""); return err }},
		{"fin get voice call by id empty", func() error { _, err := client.Fin.GetVoiceCallByID(ctx, ""); return err }},
		{"fin get voice call by id non-numeric", func() error { _, err := client.Fin.GetVoiceCallByID(ctx, "abc"); return err }},
		{"fin list voice calls empty", func() error { _, err := client.Fin.ListVoiceCallsByConversation(ctx, ""); return err }},
		{"fin get voice call by external empty", func() error { _, err := client.Fin.GetVoiceCallByExternalID(ctx, ""); return err }},
		{"fin get voice call by phone empty", func() error { _, err := client.Fin.GetVoiceCallByPhoneNumber(ctx, ""); return err }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestCoverageCompletionBehavior(t *testing.T) {
	t.Run("requireJSON decode error", func(t *testing.T) {
		if _, err := requireJSON[map[string]any]("reporting export", http.StatusOK, []byte("{")); err == nil {
			t.Fatal("expected decode error")
		}
	})

	t.Run("requireJSON status error", func(t *testing.T) {
		if _, err := requireJSON[map[string]any]("reporting export", http.StatusForbidden, []byte(`{"type":"error.list","errors":[{"code":"forbidden"}]}`)); err == nil {
			t.Fatal("expected status error")
		}
	})

	t.Run("notes retrieve transport error", func(t *testing.T) {
		client := newSupportingServicesTestClient(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("connection refused")
		}))
		if _, err := client.Notes.Retrieve(context.Background(), "10"); err == nil {
			t.Fatal("expected transport error")
		}
	})

	t.Run("tags create marshal error", func(t *testing.T) {
		client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			t.Fatal("unexpected request")
			return nil, nil
		}))
		if _, err := client.Tags.Create(context.Background(), make(chan int)); err == nil {
			t.Fatal("expected marshal error")
		}
	})

	t.Run("tickets reply marshal error", func(t *testing.T) {
		client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			t.Fatal("unexpected request")
			return nil, nil
		}))
		if _, err := client.Tickets.Reply(context.Background(), "20", make(chan int), nil); err == nil {
			t.Fatal("expected marshal error")
		}
	})

	t.Run("tickets reply skip notifications", func(t *testing.T) {
		skip := true
		client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			if !strings.Contains(string(body), `"skip_notifications":true`) {
				t.Fatalf("body = %s", string(body))
			}
			return jsonResponse(req, http.StatusOK, `{"type":"ticket_part","id":"156"}`), nil
		}))
		if _, err := client.Tickets.Reply(context.Background(), "20", map[string]any{"body": "hi"}, &skip); err != nil {
			t.Fatalf("Reply returned error: %v", err)
		}
	})

	t.Run("tickets reply skip notifications false", func(t *testing.T) {
		skip := false
		client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			if !strings.Contains(string(body), `"skip_notifications":false`) {
				t.Fatalf("body = %s", string(body))
			}
			return jsonResponse(req, http.StatusOK, `{"type":"ticket_part","id":"156"}`), nil
		}))
		if _, err := client.Tickets.Reply(context.Background(), "20", map[string]any{"body": "hi"}, &skip); err != nil {
			t.Fatalf("Reply returned error: %v", err)
		}
	})

	t.Run("tickets reply skip notifications non-object payload", func(t *testing.T) {
		skip := true
		client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			t.Fatal("unexpected request")
			return nil, nil
		}))
		if _, err := client.Tickets.Reply(context.Background(), "20", []string{"hi"}, &skip); err == nil {
			t.Fatal("expected unmarshal error")
		}
	})

	t.Run("workspace download data export api error", func(t *testing.T) {
		client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return jsonResponse(req, http.StatusForbidden, `{"type":"error.list","errors":[{"code":"forbidden"}]}`), nil
		}))
		if _, err := client.Workspace.DownloadDataExport(context.Background(), "job-1"); err == nil {
			t.Fatal("expected API error")
		}
	})

	t.Run("workspace download reporting export api error", func(t *testing.T) {
		client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return jsonResponse(req, http.StatusForbidden, `{"type":"error.list","errors":[{"code":"forbidden"}]}`), nil
		}))
		if _, err := client.Workspace.DownloadReportingExport(context.Background(), "job-1", "app-1"); err == nil {
			t.Fatal("expected API error")
		}
	})

	t.Run("required id error string", func(t *testing.T) {
		if got := (&requiredIDError{resource: "ticket"}).Error(); got != "intercom: ticket ID is required" {
			t.Fatalf("Error() = %q", got)
		}
	})
}
