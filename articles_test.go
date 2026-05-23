package intercom

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func newArticlesTestClient(t *testing.T, transport http.RoundTripper) *Client {
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
	articleJSON        = `{"id":"1","title":"Hello World","type":"article"}`
	articleListJSON    = `{"type":"list","data":[],"total_count":0}`
	articleSearchJSON  = `{"type":"list","data":{"articles":[],"highlights":[]},"total_count":0}`
	articleDeletedJSON = `{"id":"1","object":"article","deleted":true}`
)

func TestArticlesServiceRequests(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		call       func(context.Context, *Client) error
		wantMethod string
		wantPath   string
	}{
		{
			name:     "list articles",
			response: articleListJSON,
			call: func(ctx context.Context, client *Client) error {
				list, err := client.Articles.List(ctx)
				if err != nil {
					return err
				}
				if list.TotalCount == nil || *list.TotalCount != 0 {
					t.Fatalf("TotalCount = %v", list.TotalCount)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/articles",
		},
		{
			name:     "create article",
			response: articleJSON,
			call: func(ctx context.Context, client *Client) error {
				a, err := client.Articles.Create(ctx, ArticleCreate{Title: "Hello World", AuthorId: 1})
				if err != nil {
					return err
				}
				if a.Title == nil || *a.Title != "Hello World" {
					t.Fatalf("Title = %v", a.Title)
				}
				return nil
			},
			wantMethod: http.MethodPost,
			wantPath:   "/articles",
		},
		{
			name:     "retrieve article",
			response: articleJSON,
			call: func(ctx context.Context, client *Client) error {
				a, err := client.Articles.Retrieve(ctx, "1")
				if err != nil {
					return err
				}
				if a.Title == nil || *a.Title != "Hello World" {
					t.Fatalf("Title = %v", a.Title)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/articles/1",
		},
		{
			name:     "update article",
			response: articleJSON,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Articles.Update(ctx, "1", ArticleUpdate{})
				return err
			},
			wantMethod: http.MethodPut,
			wantPath:   "/articles/1",
		},
		{
			name:     "delete article",
			response: articleDeletedJSON,
			call: func(ctx context.Context, client *Client) error {
				d, err := client.Articles.Delete(ctx, "1")
				if err != nil {
					return err
				}
				if d.Deleted == nil || !*d.Deleted {
					t.Fatalf("Deleted = %v", d.Deleted)
				}
				return nil
			},
			wantMethod: http.MethodDelete,
			wantPath:   "/articles/1",
		},
		{
			name:     "search articles",
			response: articleSearchJSON,
			call: func(ctx context.Context, client *Client) error {
				res, err := client.Articles.Search(ctx, ArticleSearch{Phrase: "hello"})
				if err != nil {
					return err
				}
				if res.TotalCount == nil || *res.TotalCount != 0 {
					t.Fatalf("TotalCount = %v", res.TotalCount)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/articles/search",
		},
		{
			name:     "search articles with options",
			response: articleSearchJSON,
			call: func(ctx context.Context, client *Client) error {
				highlight := true
				_, err := client.Articles.Search(ctx, ArticleSearch{
					State:        "published",
					HelpCenterID: 42,
					Highlight:    &highlight,
				})
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/articles/search",
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
			client := newArticlesTestClient(t, transport)
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

func TestArticlesServiceErrors(t *testing.T) {
	errBody := `{"type":"error.list","errors":[{"code":"unauthorized"}],"request_id":"12a938a3-314e-4939-b773-5cd45738bd21"}`

	calls := []struct {
		name string
		call func(context.Context, *Client) error
	}{
		{"list", func(ctx context.Context, c *Client) error { _, err := c.Articles.List(ctx); return err }},
		{"create", func(ctx context.Context, c *Client) error {
			_, err := c.Articles.Create(ctx, ArticleCreate{Title: "x", AuthorId: 1})
			return err
		}},
		{"retrieve", func(ctx context.Context, c *Client) error { _, err := c.Articles.Retrieve(ctx, "1"); return err }},
		{"update", func(ctx context.Context, c *Client) error {
			_, err := c.Articles.Update(ctx, "1", ArticleUpdate{})
			return err
		}},
		{"delete", func(ctx context.Context, c *Client) error { _, err := c.Articles.Delete(ctx, "1"); return err }},
		{"search", func(ctx context.Context, c *Client) error {
			_, err := c.Articles.Search(ctx, ArticleSearch{})
			return err
		}},
	}

	for _, tt := range calls {
		t.Run(tt.name, func(t *testing.T) {
			client := newArticlesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return jsonResponse(req, http.StatusUnauthorized, errBody), nil
			}))
			if err := tt.call(context.Background(), client); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestArticlesTransportErrors(t *testing.T) {
	transportErr := errors.New("connection refused")

	calls := []struct {
		name string
		call func(context.Context, *Client) error
	}{
		{"list", func(ctx context.Context, c *Client) error { _, err := c.Articles.List(ctx); return err }},
		{"create", func(ctx context.Context, c *Client) error {
			_, err := c.Articles.Create(ctx, ArticleCreate{Title: "x", AuthorId: 1})
			return err
		}},
		{"retrieve", func(ctx context.Context, c *Client) error { _, err := c.Articles.Retrieve(ctx, "1"); return err }},
		{"update", func(ctx context.Context, c *Client) error {
			_, err := c.Articles.Update(ctx, "1", ArticleUpdate{})
			return err
		}},
		{"delete", func(ctx context.Context, c *Client) error { _, err := c.Articles.Delete(ctx, "1"); return err }},
		{"search", func(ctx context.Context, c *Client) error {
			_, err := c.Articles.Search(ctx, ArticleSearch{})
			return err
		}},
	}

	for _, tt := range calls {
		t.Run(tt.name, func(t *testing.T) {
			client := newArticlesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return nil, transportErr
			}))
			if err := tt.call(context.Background(), client); err == nil {
				t.Fatal("expected transport error")
			}
		})
	}
}

func TestArticlesValidation(t *testing.T) {
	client := newArticlesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		t.Fatal("unexpected HTTP request")
		return nil, nil
	}))
	ctx := context.Background()

	tests := []struct {
		name string
		call func() error
	}{
		{"retrieve: empty ID", func() error { _, err := client.Articles.Retrieve(ctx, ""); return err }},
		{"retrieve: partial-numeric ID", func() error { _, err := client.Articles.Retrieve(ctx, "1abc"); return err }},
		{"update: empty ID", func() error { _, err := client.Articles.Update(ctx, "", ArticleUpdate{}); return err }},
		{"update: partial-numeric ID", func() error { _, err := client.Articles.Update(ctx, "1abc", ArticleUpdate{}); return err }},
		{"delete: empty ID", func() error { _, err := client.Articles.Delete(ctx, ""); return err }},
		{"delete: partial-numeric ID", func() error { _, err := client.Articles.Delete(ctx, "1abc"); return err }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
