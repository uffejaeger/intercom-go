package intercom

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func newAIContentTestClient(t *testing.T, transport http.RoundTripper) *Client {
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

const (
	contentImportSourceJSON     = `{"id":1,"url":"https://example.com","created_at":0,"last_synced_at":0,"status":"active","sync_behavior":"api","type":"external_page","updated_at":0}`
	contentImportSourceListJSON = `{"type":"list","data":[]}`
	externalPageJSON            = `{"id":"page-1","url":"https://example.com/page","created_at":0,"last_ingested_at":0,"ai_agent_availability":false,"ai_copilot_availability":false,"external_id":"ext-1","html":"","locale":"en","source_id":1,"title":"Test","type":"external_page"}`
	externalPageListJSON        = `{"type":"list","data":[]}`
)

func TestAIContentServiceRequests(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		call       func(context.Context, *Client) error
		wantMethod string
		wantPath   string
	}{
		{
			name:     "list content import sources",
			response: contentImportSourceListJSON,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.AIContent.ListContentImportSources(ctx)
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/ai/content_import_sources",
		},
		{
			name:     "create content import source",
			response: contentImportSourceJSON,
			call: func(ctx context.Context, client *Client) error {
				src, err := client.AIContent.CreateContentImportSource(ctx, ContentImportSourceCreate{Url: "https://example.com"})
				if err != nil {
					return err
				}
				if src.Id != 1 {
					t.Fatalf("Id = %v", src.Id)
				}
				return nil
			},
			wantMethod: http.MethodPost,
			wantPath:   "/ai/content_import_sources",
		},
		{
			name:     "get content import source",
			response: contentImportSourceJSON,
			call: func(ctx context.Context, client *Client) error {
				src, err := client.AIContent.GetContentImportSource(ctx, "src-1")
				if err != nil {
					return err
				}
				if src.Id != 1 {
					t.Fatalf("Id = %v", src.Id)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/ai/content_import_sources/src-1",
		},
		{
			name:     "update content import source",
			response: contentImportSourceJSON,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.AIContent.UpdateContentImportSource(ctx, "src-1", ContentImportSourceUpdate{})
				return err
			},
			wantMethod: http.MethodPut,
			wantPath:   "/ai/content_import_sources/src-1",
		},
		{
			name:     "delete content import source",
			response: ``,
			call: func(ctx context.Context, client *Client) error {
				return client.AIContent.DeleteContentImportSource(ctx, "src-1")
			},
			wantMethod: http.MethodDelete,
			wantPath:   "/ai/content_import_sources/src-1",
		},
		{
			name:     "list external pages",
			response: externalPageListJSON,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.AIContent.ListExternalPages(ctx)
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/ai/external_pages",
		},
		{
			name:     "create external page",
			response: externalPageJSON,
			call: func(ctx context.Context, client *Client) error {
				page, err := client.AIContent.CreateExternalPage(ctx, ExternalPageCreate{})
				if err != nil {
					return err
				}
				if page.Id != "page-1" {
					t.Fatalf("Id = %v", page.Id)
				}
				return nil
			},
			wantMethod: http.MethodPost,
			wantPath:   "/ai/external_pages",
		},
		{
			name:     "get external page",
			response: externalPageJSON,
			call: func(ctx context.Context, client *Client) error {
				page, err := client.AIContent.GetExternalPage(ctx, "page-1")
				if err != nil {
					return err
				}
				if page.Id != "page-1" {
					t.Fatalf("Id = %v", page.Id)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/ai/external_pages/page-1",
		},
		{
			name:     "update external page",
			response: externalPageJSON,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.AIContent.UpdateExternalPage(ctx, "page-1", ExternalPageUpdate{})
				return err
			},
			wantMethod: http.MethodPut,
			wantPath:   "/ai/external_pages/page-1",
		},
		{
			name:     "delete external page",
			response: externalPageJSON,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.AIContent.DeleteExternalPage(ctx, "page-1")
				return err
			},
			wantMethod: http.MethodDelete,
			wantPath:   "/ai/external_pages/page-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod, gotPath string
			transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				gotMethod = req.Method
				gotPath = req.URL.Path
				return jsonResponse(req, http.StatusOK, tt.response), nil
			})
			client := newAIContentTestClient(t, transport)
			if err := tt.call(context.Background(), client); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotMethod != tt.wantMethod {
				t.Fatalf("method = %q, want %q", gotMethod, tt.wantMethod)
			}
			if gotPath != tt.wantPath {
				t.Fatalf("path = %q, want %q", gotPath, tt.wantPath)
			}
		})
	}
}

func TestAIContentServiceErrors(t *testing.T) {
	errBody := `{"type":"error.list","errors":[{"code":"unauthorized"}],"request_id":"12a938a3-314e-4939-b773-5cd45738bd21"}`

	tests := []struct {
		name string
		call func(context.Context, *Client) error
	}{
		{"list sources", func(ctx context.Context, c *Client) error {
			_, err := c.AIContent.ListContentImportSources(ctx)
			return err
		}},
		{"create source", func(ctx context.Context, c *Client) error {
			_, err := c.AIContent.CreateContentImportSource(ctx, ContentImportSourceCreate{})
			return err
		}},
		{"get source", func(ctx context.Context, c *Client) error {
			_, err := c.AIContent.GetContentImportSource(ctx, "src-1")
			return err
		}},
		{"update source", func(ctx context.Context, c *Client) error {
			_, err := c.AIContent.UpdateContentImportSource(ctx, "src-1", ContentImportSourceUpdate{})
			return err
		}},
		{"delete source", func(ctx context.Context, c *Client) error {
			return c.AIContent.DeleteContentImportSource(ctx, "src-1")
		}},
		{"list pages", func(ctx context.Context, c *Client) error {
			_, err := c.AIContent.ListExternalPages(ctx)
			return err
		}},
		{"create page", func(ctx context.Context, c *Client) error {
			_, err := c.AIContent.CreateExternalPage(ctx, ExternalPageCreate{})
			return err
		}},
		{"get page", func(ctx context.Context, c *Client) error {
			_, err := c.AIContent.GetExternalPage(ctx, "page-1")
			return err
		}},
		{"update page", func(ctx context.Context, c *Client) error {
			_, err := c.AIContent.UpdateExternalPage(ctx, "page-1", ExternalPageUpdate{})
			return err
		}},
		{"delete page", func(ctx context.Context, c *Client) error {
			_, err := c.AIContent.DeleteExternalPage(ctx, "page-1")
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newAIContentTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return jsonResponse(req, http.StatusUnauthorized, errBody), nil
			}))
			if err := tt.call(context.Background(), client); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestAIContentTransportErrors(t *testing.T) {
	transportErr := errors.New("connection refused")

	calls := []struct {
		name string
		call func(context.Context, *Client) error
	}{
		{"list sources", func(ctx context.Context, c *Client) error {
			_, err := c.AIContent.ListContentImportSources(ctx)
			return err
		}},
		{"create source", func(ctx context.Context, c *Client) error {
			_, err := c.AIContent.CreateContentImportSource(ctx, ContentImportSourceCreate{})
			return err
		}},
		{"get source", func(ctx context.Context, c *Client) error {
			_, err := c.AIContent.GetContentImportSource(ctx, "src-1")
			return err
		}},
		{"update source", func(ctx context.Context, c *Client) error {
			_, err := c.AIContent.UpdateContentImportSource(ctx, "src-1", ContentImportSourceUpdate{})
			return err
		}},
		{"delete source", func(ctx context.Context, c *Client) error {
			return c.AIContent.DeleteContentImportSource(ctx, "src-1")
		}},
		{"list pages", func(ctx context.Context, c *Client) error {
			_, err := c.AIContent.ListExternalPages(ctx)
			return err
		}},
		{"create page", func(ctx context.Context, c *Client) error {
			_, err := c.AIContent.CreateExternalPage(ctx, ExternalPageCreate{})
			return err
		}},
		{"get page", func(ctx context.Context, c *Client) error {
			_, err := c.AIContent.GetExternalPage(ctx, "page-1")
			return err
		}},
		{"update page", func(ctx context.Context, c *Client) error {
			_, err := c.AIContent.UpdateExternalPage(ctx, "page-1", ExternalPageUpdate{})
			return err
		}},
		{"delete page", func(ctx context.Context, c *Client) error {
			_, err := c.AIContent.DeleteExternalPage(ctx, "page-1")
			return err
		}},
	}

	for _, tt := range calls {
		t.Run(tt.name, func(t *testing.T) {
			client := newAIContentTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return nil, transportErr
			}))
			if err := tt.call(context.Background(), client); err == nil {
				t.Fatal("expected transport error")
			}
		})
	}
}

func TestAIContentValidation(t *testing.T) {
	client := newAIContentTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		t.Fatal("unexpected HTTP request")
		return nil, nil
	}))
	ctx := context.Background()

	tests := []struct {
		name string
		call func() error
	}{
		{"get source: empty ID", func() error { _, err := client.AIContent.GetContentImportSource(ctx, ""); return err }},
		{"update source: empty ID", func() error {
			_, err := client.AIContent.UpdateContentImportSource(ctx, "", ContentImportSourceUpdate{})
			return err
		}},
		{"delete source: empty ID", func() error { return client.AIContent.DeleteContentImportSource(ctx, "") }},
		{"get page: empty ID", func() error { _, err := client.AIContent.GetExternalPage(ctx, ""); return err }},
		{"update page: empty ID", func() error {
			_, err := client.AIContent.UpdateExternalPage(ctx, "", ExternalPageUpdate{})
			return err
		}},
		{"delete page: empty ID", func() error { _, err := client.AIContent.DeleteExternalPage(ctx, ""); return err }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
