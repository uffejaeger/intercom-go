package intercom

import (
	"context"
	"net/http"
	"testing"
)

func newSupportingServicesTestClient(t *testing.T, transport http.RoundTripper) *Client {
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

func TestSupportingServicesRequests(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		call       func(context.Context, *Client) error
		wantMethod string
		wantPath   string
	}{
		{
			name:       "list brands",
			statusCode: http.StatusOK,
			response:   `{"type":"list","data":[{"type":"brand","id":"1","name":"Default"}]}`,
			call: func(ctx context.Context, client *Client) error {
				brands, err := client.Brands.List(ctx)
				if err != nil {
					return err
				}
				if brands.Data == nil || len(*brands.Data) != 1 {
					t.Fatalf("brands.Data = %#v", brands.Data)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/brands",
		},
		{
			name:       "retrieve brand",
			statusCode: http.StatusOK,
			response:   `{"type":"brand","id":"1","name":"Default"}`,
			call: func(ctx context.Context, client *Client) error {
				brand, err := client.Brands.Retrieve(ctx, "1")
				if err != nil {
					return err
				}
				if brand.Id == nil || *brand.Id != "1" {
					t.Fatalf("brand.Id = %v", brand.Id)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/brands/1",
		},
		{
			name:       "retrieve note",
			statusCode: http.StatusOK,
			response:   `{"type":"note","id":"7","body":"Followed up"}`,
			call: func(ctx context.Context, client *Client) error {
				note, err := client.Notes.Retrieve(ctx, "7")
				if err != nil {
					return err
				}
				if note.Id == nil || *note.Id != "7" {
					t.Fatalf("note.Id = %v", note.Id)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/notes/7",
		},
		{
			name:       "list segments",
			statusCode: http.StatusOK,
			response:   `{"type":"segment.list","segments":[{"type":"segment","id":"segment-1","name":"VIP"}]}`,
			call: func(ctx context.Context, client *Client) error {
				segments, err := client.Segments.List(ctx)
				if err != nil {
					return err
				}
				if segments.Segments == nil || len(*segments.Segments) != 1 {
					t.Fatalf("segments.Segments = %#v", segments.Segments)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/segments",
		},
		{
			name:       "retrieve segment",
			statusCode: http.StatusOK,
			response:   `{"type":"segment","id":"segment-1","name":"VIP"}`,
			call: func(ctx context.Context, client *Client) error {
				segment, err := client.Segments.Retrieve(ctx, "segment-1")
				if err != nil {
					return err
				}
				if segment.Id == nil || *segment.Id != "segment-1" {
					t.Fatalf("segment.Id = %v", segment.Id)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/segments/segment-1",
		},
		{
			name:       "list subscription types",
			statusCode: http.StatusOK,
			response:   `{"type":"list","data":[{"type":"subscription","id":"sub-1","name":"Product updates"}]}`,
			call: func(ctx context.Context, client *Client) error {
				subscriptions, err := client.SubscriptionTypes.List(ctx)
				if err != nil {
					return err
				}
				if subscriptions.Data == nil || len(*subscriptions.Data) != 1 {
					t.Fatalf("subscriptions.Data = %#v", subscriptions.Data)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/subscription_types",
		},
		{
			name:       "list tags",
			statusCode: http.StatusOK,
			response:   `{"type":"list","data":[{"type":"tag","id":"102","name":"Manual tag 1"}]}`,
			call: func(ctx context.Context, client *Client) error {
				tags, err := client.Tags.List(ctx)
				if err != nil {
					return err
				}
				if tags.Data == nil || len(*tags.Data) != 1 {
					t.Fatalf("tags.Data = %#v", tags.Data)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/tags",
		},
		{
			name:       "retrieve tag",
			statusCode: http.StatusOK,
			response:   `{"type":"tag","id":"102","name":"Manual tag 1"}`,
			call: func(ctx context.Context, client *Client) error {
				tag, err := client.Tags.Retrieve(ctx, "102")
				if err != nil {
					return err
				}
				if tag.Id == nil || *tag.Id != "102" {
					t.Fatalf("tag.Id = %v", tag.Id)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/tags/102",
		},
		{
			name:       "delete tag",
			statusCode: http.StatusNoContent,
			response:   ``,
			call: func(ctx context.Context, client *Client) error {
				return client.Tags.Delete(ctx, "102")
			},
			wantMethod: http.MethodDelete,
			wantPath:   "/tags/102",
		},
		{
			name:       "list teams",
			statusCode: http.StatusOK,
			response:   `{"type":"team.list","teams":[{"type":"team","id":"14","name":"Support"}]}`,
			call: func(ctx context.Context, client *Client) error {
				teams, err := client.Teams.List(ctx)
				if err != nil {
					return err
				}
				if teams.Teams == nil || len(*teams.Teams) != 1 {
					t.Fatalf("teams.Teams = %#v", teams.Teams)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/teams",
		},
		{
			name:       "retrieve team",
			statusCode: http.StatusOK,
			response:   `{"type":"team","id":"14","name":"Support"}`,
			call: func(ctx context.Context, client *Client) error {
				team, err := client.Teams.Retrieve(ctx, "14")
				if err != nil {
					return err
				}
				if team.Id == nil || *team.Id != "14" {
					t.Fatalf("team.Id = %v", team.Id)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/teams/14",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod, gotPath string
			transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				gotMethod = req.Method
				gotPath = req.URL.Path
				return jsonResponse(req, tt.statusCode, tt.response), nil
			})

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
		})
	}
}

func TestSupportingServicesValidation(t *testing.T) {
	client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected request: %s %s", req.Method, req.URL.Path)
		return nil, nil
	}))

	tests := []struct {
		name string
		call func() error
	}{
		{name: "brand ID required", call: func() error { _, err := client.Brands.Retrieve(context.Background(), ""); return err }},
		{name: "note ID required", call: func() error { _, err := client.Notes.Retrieve(context.Background(), ""); return err }},
		{name: "note ID integer", call: func() error { _, err := client.Notes.Retrieve(context.Background(), "abc"); return err }},
		{name: "segment ID required", call: func() error { _, err := client.Segments.Retrieve(context.Background(), ""); return err }},
		{name: "tag ID required for retrieve", call: func() error { _, err := client.Tags.Retrieve(context.Background(), ""); return err }},
		{name: "tag ID required for delete", call: func() error { return client.Tags.Delete(context.Background(), "") }},
		{name: "team ID required", call: func() error { _, err := client.Teams.Retrieve(context.Background(), ""); return err }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
