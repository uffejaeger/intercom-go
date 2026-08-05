package intercom

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

func TestAPI216ServicesCoverSuccessAndTransportFailures(t *testing.T) {
	ctx := context.Background()
	calls := api216Calls(ctx)

	t.Run("success", func(t *testing.T) {
		client := newAPI216TestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: api216Status(req),
				Status:     http.StatusText(api216Status(req)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
				Request:    req,
			}, nil
		}))

		for name, call := range calls {
			t.Run(name, func(t *testing.T) {
				if err := call(client); err != nil {
					t.Fatalf("call returned error: %v", err)
				}
			})
		}
	})

	t.Run("transport failure", func(t *testing.T) {
		client := newAPI216TestClient(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("transport unavailable")
		}))

		for name, call := range calls {
			t.Run(name, func(t *testing.T) {
				if err := call(client); err == nil {
					t.Fatal("expected transport error")
				}
			})
		}
	})
}

func TestAPI216ValidationAndResponseFailures(t *testing.T) {
	client := newAPI216TestClient(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("validation failure should not send a request")
		return nil, nil
	}))

	for name, call := range map[string]func() error{
		"article attach tag": func() error {
			_, err := client.Articles.AttachTag(context.Background(), "invalid", ArticleTag{})
			return err
		},
		"article detach tag": func() error { _, err := client.Articles.DetachTag(context.Background(), "invalid", "tag"); return err },
		"article versions":   func() error { _, err := client.Articles.ListVersions(context.Background(), "invalid", nil); return err },
		"article version": func() error {
			_, err := client.Articles.GetVersion(context.Background(), "invalid", "version")
			return err
		},
		"article draft": func() error { _, err := client.Articles.GetDraft(context.Background(), "invalid"); return err },
		"article stage": func() error {
			_, err := client.Articles.StageDraft(context.Background(), "invalid", ArticleDraftUpdate{})
			return err
		},
		"article publish": func() error {
			_, err := client.Articles.PublishDraft(context.Background(), "invalid", ArticleDraftPublish{})
			return err
		},
		"internal article attach tag": func() error {
			_, err := client.InternalArticles.AttachTag(context.Background(), "invalid", InternalArticleTag{})
			return err
		},
		"internal article detach tag": func() error {
			_, err := client.InternalArticles.DetachTag(context.Background(), "invalid", "tag")
			return err
		},
		"company note": func() error {
			_, err := client.Companies.CreateNote(context.Background(), "", CompanyNoteCreate{})
			return err
		},
		"custom object list": func() error { _, err := client.CustomObjects.List(context.Background(), "", nil); return err },
	} {
		t.Run(name, func(t *testing.T) {
			if err := call(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}

	if _, err := requireStatus("test", http.StatusInternalServerError, http.StatusOK, nil, new(struct{})); err == nil {
		t.Fatal("expected status error")
	}
	if _, err := requireStatus("test", http.StatusOK, http.StatusOK, nil, (*struct{})(nil)); err == nil {
		t.Fatal("expected missing response-body error")
	}
}

func newAPI216TestClient(t *testing.T, transport http.RoundTripper) *Client {
	t.Helper()
	client, err := NewClient("token", WithBaseURL("https://example.test"), WithHTTPClient(&http.Client{Transport: transport}))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	return client
}

func api216Status(req *http.Request) int {
	if req.Method == http.MethodPost {
		switch {
		case strings.Contains(req.URL.Path, "bulk"):
			return http.StatusAccepted
		case req.URL.Path == "/audiences", req.URL.Path == "/content_snippets", req.URL.Path == "/data_connectors", req.URL.Path == "/office_hours_schedules", strings.Contains(req.URL.Path, "/office_hours_exceptions") && strings.HasSuffix(req.URL.Path, "/office_hours_exceptions"):
			return http.StatusCreated
		}
	}
	return http.StatusOK
}

func api216Calls(ctx context.Context) map[string]func(*Client) error {
	return map[string]func(*Client) error{
		"audiences list":        func(c *Client) error { _, err := c.Audiences.List(ctx, nil); return err },
		"audiences create":      func(c *Client) error { _, err := c.Audiences.Create(ctx, AudienceCreate{}); return err },
		"audiences get":         func(c *Client) error { _, err := c.Audiences.Get(ctx, "audience"); return err },
		"audiences update":      func(c *Client) error { _, err := c.Audiences.Update(ctx, "audience", AudienceUpdate{}); return err },
		"audiences delete":      func(c *Client) error { return c.Audiences.Delete(ctx, "audience") },
		"macros list":           func(c *Client) error { _, err := c.Macros.List(ctx, nil); return err },
		"macros get":            func(c *Client) error { _, err := c.Macros.Get(ctx, "macro"); return err },
		"office schedules list": func(c *Client) error { _, err := c.OfficeHours.ListSchedules(ctx, nil); return err },
		"office schedules create": func(c *Client) error {
			_, err := c.OfficeHours.CreateSchedule(ctx, OfficeHoursScheduleCreate{})
			return err
		},
		"office schedules get": func(c *Client) error { _, err := c.OfficeHours.GetSchedule(ctx, "schedule"); return err },
		"office schedules update": func(c *Client) error {
			_, err := c.OfficeHours.UpdateSchedule(ctx, "schedule", OfficeHoursScheduleUpdate{})
			return err
		},
		"office schedules delete": func(c *Client) error { return c.OfficeHours.DeleteSchedule(ctx, "schedule") },
		"office exceptions list":  func(c *Client) error { _, err := c.OfficeHours.ListExceptions(ctx, "schedule", nil); return err },
		"office exceptions create": func(c *Client) error {
			_, err := c.OfficeHours.CreateException(ctx, "schedule", OfficeHoursExceptionCreate{})
			return err
		},
		"office exceptions get": func(c *Client) error { _, err := c.OfficeHours.GetException(ctx, "schedule", "exception"); return err },
		"office exceptions update": func(c *Client) error {
			_, err := c.OfficeHours.UpdateException(ctx, "schedule", "exception", OfficeHoursExceptionUpdate{})
			return err
		},
		"office exceptions delete": func(c *Client) error { return c.OfficeHours.DeleteException(ctx, "schedule", "exception") },
		"content search":           func(c *Client) error { _, err := c.Content.Search(ctx, nil); return err },
		"content bulk":             func(c *Client) error { _, err := c.Content.BulkAction(ctx, ContentBulkActionRequest{}); return err },
		"content snippets list":    func(c *Client) error { _, err := c.Content.ListSnippets(ctx, nil); return err },
		"content snippets create":  func(c *Client) error { _, err := c.Content.CreateSnippet(ctx, ContentSnippetCreate{}); return err },
		"content snippets get":     func(c *Client) error { _, err := c.Content.GetSnippet(ctx, "snippet"); return err },
		"content snippets update": func(c *Client) error {
			_, err := c.Content.UpdateSnippet(ctx, "snippet", ContentSnippetUpdate{})
			return err
		},
		"content snippets delete": func(c *Client) error { return c.Content.DeleteSnippet(ctx, "snippet") },
		"content snippets attach tag": func(c *Client) error {
			_, err := c.Content.AttachSnippetTag(ctx, "snippet", ContentSnippetTag{})
			return err
		},
		"content snippets detach tag": func(c *Client) error { return c.Content.DetachSnippetTag(ctx, "snippet", "tag") },
		"data connectors list":        func(c *Client) error { _, err := c.DataConnectors.List(ctx, nil); return err },
		"data connectors create":      func(c *Client) error { _, err := c.DataConnectors.Create(ctx, DataConnectorCreate{}); return err },
		"data connectors get":         func(c *Client) error { _, err := c.DataConnectors.Get(ctx, "connector"); return err },
		"data connectors update": func(c *Client) error {
			_, err := c.DataConnectors.Update(ctx, "connector", DataConnectorUpdate{})
			return err
		},
		"data connectors delete": func(c *Client) error { _, err := c.DataConnectors.Delete(ctx, "connector"); return err },
		"data connectors executions": func(c *Client) error {
			_, err := c.DataConnectors.ListExecutionResults(ctx, "connector", nil)
			return err
		},
		"data connectors execution": func(c *Client) error {
			_, err := c.DataConnectors.GetExecutionResult(ctx, "connector", "result")
			return err
		},
		"conversation attributes list": func(c *Client) error { _, err := c.ConversationAttributes.List(ctx, nil); return err },
		"conversation attributes create": func(c *Client) error {
			_, err := c.ConversationAttributes.Create(ctx, ConversationAttributeCreate{})
			return err
		},
		"conversation attributes get": func(c *Client) error { _, err := c.ConversationAttributes.Get(ctx, 1); return err },
		"conversation attributes update": func(c *Client) error {
			_, err := c.ConversationAttributes.Update(ctx, 1, ConversationAttributeUpdate{})
			return err
		},
		"conversation attributes delete": func(c *Client) error { _, err := c.ConversationAttributes.Delete(ctx, 1); return err },
		"conversation attribute options create": func(c *Client) error {
			_, err := c.ConversationAttributes.CreateOption(ctx, 1, ConversationAttributeOptionCreate{})
			return err
		},
		"conversation attribute options update": func(c *Client) error {
			_, err := c.ConversationAttributes.UpdateOption(ctx, 1, "option", ConversationAttributeOptionUpdate{})
			return err
		},
		"conversation attribute options delete": func(c *Client) error { _, err := c.ConversationAttributes.DeleteOption(ctx, 1, "option"); return err },
		"help center redirects list":            func(c *Client) error { _, err := c.HelpCenterRedirects.List(ctx, "center", nil); return err },
		"help center redirects create": func(c *Client) error {
			_, err := c.HelpCenterRedirects.Create(ctx, "center", HelpCenterRedirectCreate{})
			return err
		},
		"help center redirects get":    func(c *Client) error { _, err := c.HelpCenterRedirects.Get(ctx, "center", "redirect"); return err },
		"help center redirects delete": func(c *Client) error { _, err := c.HelpCenterRedirects.Delete(ctx, "center", "redirect"); return err },
		"articles attach tag":          func(c *Client) error { _, err := c.Articles.AttachTag(ctx, "1", ArticleTag{}); return err },
		"articles detach tag":          func(c *Client) error { _, err := c.Articles.DetachTag(ctx, "1", "tag"); return err },
		"articles versions":            func(c *Client) error { _, err := c.Articles.ListVersions(ctx, "1", nil); return err },
		"articles version":             func(c *Client) error { _, err := c.Articles.GetVersion(ctx, "1", "version"); return err },
		"articles draft":               func(c *Client) error { _, err := c.Articles.GetDraft(ctx, "1"); return err },
		"articles stage":               func(c *Client) error { _, err := c.Articles.StageDraft(ctx, "1", ArticleDraftUpdate{}); return err },
		"articles publish":             func(c *Client) error { _, err := c.Articles.PublishDraft(ctx, "1", ArticleDraftPublish{}); return err },
		"internal articles attach tag": func(c *Client) error {
			_, err := c.InternalArticles.AttachTag(ctx, "1", InternalArticleTag{})
			return err
		},
		"internal articles detach tag": func(c *Client) error { _, err := c.InternalArticles.DetachTag(ctx, "1", "tag"); return err },
		"companies create note": func(c *Client) error {
			_, err := c.Companies.CreateNote(ctx, "company", CompanyNoteCreate{})
			return err
		},
		"contacts banners":        func(c *Client) error { _, err := c.Contacts.ListBanners(ctx, "contact"); return err },
		"contacts dismiss banner": func(c *Client) error { _, err := c.Contacts.DismissBanner(ctx, "contact", "view"); return err },
		"contacts merge history":  func(c *Client) error { _, err := c.Contacts.ListMergeHistory(ctx, "contact"); return err },
		"admins event types":      func(c *Client) error { _, err := c.Admins.ListActivityLogEventTypes(ctx, nil); return err },
		"admins search logs": func(c *Client) error {
			_, err := c.Admins.SearchActivityLogs(ctx, nil, gen.SearchActivityLogsJSONRequestBody{})
			return err
		},
		"conversations deleted": func(c *Client) error { _, err := c.Conversations.ListDeletedIDs(ctx, nil); return err },
		"conversations merge": func(c *Client) error {
			_, err := c.Conversations.Merge(ctx, "conversation", gen.MergeConversationJSONRequestBody{})
			return err
		},
		"conversations side": func(c *Client) error {
			_, err := c.Conversations.ListSideConversations(ctx, "conversation", nil)
			return err
		},
		"custom objects list": func(c *Client) error { _, err := c.CustomObjects.List(ctx, "Order", nil); return err },
		"teams metrics":       func(c *Client) error { _, err := c.Teams.Metrics(ctx, "team", nil); return err },
		"tickets change type": func(c *Client) error {
			_, err := c.Tickets.ChangeType(ctx, "ticket", gen.ChangeTicketTypeJSONRequestBody{})
			return err
		},
		"tickets link conversation": func(c *Client) error {
			_, err := c.Tickets.LinkConversation(ctx, "ticket", gen.LinkConversationToTicketJSONRequestBody{})
			return err
		},
		"tickets unlink conversation": func(c *Client) error {
			_, err := c.Tickets.UnlinkConversation(ctx, "ticket", "conversation")
			return err
		},
		"fin submit csat":          func(c *Client) error { _, err := c.Fin.SubmitCSAT(ctx, gen.SubmitFinCsatJSONRequestBody{}); return err },
		"whatsapp get status":      func(c *Client) error { _, err := c.WhatsApp.GetMessageStatus(ctx, nil); return err },
		"whatsapp retrieve status": func(c *Client) error { _, err := c.WhatsApp.RetrieveMessageStatus(ctx, nil); return err },
	}
}
