package intercom

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func newBrandsTestClient(t *testing.T, transport http.RoundTripper) *Client {
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

func TestBrandsServiceRequests(t *testing.T) {
	errBody := `{"type":"error.list","errors":[{"code":"unauthorized"}],"request_id":"req-1"}`

	tests := []struct {
		name       string
		response   string
		call       func(context.Context, *Client) error
		wantMethod string
		wantPath   string
	}{
		{
			name:     "list brands",
			response: `{"data":[{"id":"1","name":"My Brand"}]}`,
			call: func(ctx context.Context, c *Client) error {
				list, err := c.Brands.List(ctx)
				if err != nil {
					return err
				}
				if list.Data == nil || len(*list.Data) != 1 {
					t.Fatal("expected one brand")
				}
				if (*list.Data)[0].Name == nil || *(*list.Data)[0].Name != "My Brand" {
					t.Fatalf("Name = %v", (*list.Data)[0].Name)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/brands",
		},
		{
			name:     "retrieve brand",
			response: `{"id":"42","name":"My Brand"}`,
			call: func(ctx context.Context, c *Client) error {
				b, err := c.Brands.Retrieve(ctx, "42")
				if err != nil {
					return err
				}
				if b.Id == nil || *b.Id != "42" {
					t.Fatalf("Id = %v", b.Id)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/brands/42",
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
			client := newBrandsTestClient(t, transport)
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

	t.Run("api errors", func(t *testing.T) {
		calls := []struct {
			name string
			call func(context.Context, *Client) error
		}{
			{"list", func(ctx context.Context, c *Client) error { _, err := c.Brands.List(ctx); return err }},
			{"retrieve", func(ctx context.Context, c *Client) error { _, err := c.Brands.Retrieve(ctx, "42"); return err }},
		}
		for _, tt := range calls {
			t.Run(tt.name, func(t *testing.T) {
				client := newBrandsTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
					return jsonResponse(req, http.StatusUnauthorized, errBody), nil
				}))
				if err := tt.call(context.Background(), client); err == nil {
					t.Fatal("expected error")
				}
			})
		}
	})

	t.Run("transport errors", func(t *testing.T) {
		transportErr := errors.New("connection refused")
		calls := []struct {
			name string
			call func(context.Context, *Client) error
		}{
			{"list", func(ctx context.Context, c *Client) error { _, err := c.Brands.List(ctx); return err }},
			{"retrieve", func(ctx context.Context, c *Client) error { _, err := c.Brands.Retrieve(ctx, "42"); return err }},
		}
		for _, tt := range calls {
			t.Run(tt.name, func(t *testing.T) {
				client := newBrandsTestClient(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
					return nil, transportErr
				}))
				if err := tt.call(context.Background(), client); err == nil {
					t.Fatal("expected transport error")
				}
			})
		}
	})
}

func TestBrandsValidation(t *testing.T) {
	client := newBrandsTestClient(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("unexpected HTTP request")
		return nil, nil
	}))
	ctx := context.Background()

	if _, err := client.Brands.Retrieve(ctx, ""); err == nil {
		t.Fatal("expected validation error for empty brand ID")
	}
}
