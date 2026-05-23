package intercom

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

func newCompaniesTestClient(t *testing.T, transport http.RoundTripper) *Client {
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
	companyJSON          = `{"id":"comp-1","name":"Acme"}`
	companyListJSON      = `{"type":"list","data":[],"total_count":0}`
	companyScrollJSON    = `{"type":"list","data":[],"total_count":0,"scroll_param":"token-1"}`
	companyDeletedJSON   = `{"id":"comp-1","object":"company","deleted":true}`
	companyContactsJSON  = `{"type":"list","data":[],"total_count":0}`
	companySegmentsJSON  = `{"type":"list","data":[]}`
	noteListJSON         = `{"type":"list","notes":[]}`
	contactCompaniesJSON = `{"type":"company.list","companies":[],"total_count":0}`
)

func TestCompaniesServiceRequests(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		call       func(context.Context, *Client) error
		wantMethod string
		wantPath   string
		wantBody   func(t *testing.T, body map[string]any)
	}{
		{
			name:     "create or update company",
			response: companyJSON,
			call: func(ctx context.Context, client *Client) error {
				name := "Acme"
				c, err := client.Companies.CreateOrUpdate(ctx, CompanyCreate{Name: &name})
				if err != nil {
					return err
				}
				if c.Id == nil || *c.Id != "comp-1" {
					t.Fatalf("Id = %v", c.Id)
				}
				return nil
			},
			wantMethod: http.MethodPost,
			wantPath:   "/companies",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "name"); got != "Acme" {
					t.Fatalf("name = %q", got)
				}
			},
		},
		{
			name:     "retrieve company",
			response: companyJSON,
			call: func(ctx context.Context, client *Client) error {
				c, err := client.Companies.Retrieve(ctx, "comp-1")
				if err != nil {
					return err
				}
				if c.Id == nil || *c.Id != "comp-1" {
					t.Fatalf("Id = %v", c.Id)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/companies/comp-1",
		},
		{
			name:     "retrieve company by ID",
			response: companyListJSON,
			call: func(ctx context.Context, client *Client) error {
				list, err := client.Companies.RetrieveByID(ctx, "ext-1")
				if err != nil {
					return err
				}
				if list.TotalCount == nil || *list.TotalCount != 0 {
					t.Fatalf("TotalCount = %v", list.TotalCount)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/companies",
		},
		{
			name:     "update company",
			response: companyJSON,
			call: func(ctx context.Context, client *Client) error {
				name := "Acme Updated"
				c, err := client.Companies.Update(ctx, "comp-1", CompanyUpdate{Name: &name})
				if err != nil {
					return err
				}
				if c.Id == nil || *c.Id != "comp-1" {
					t.Fatalf("Id = %v", c.Id)
				}
				return nil
			},
			wantMethod: http.MethodPut,
			wantPath:   "/companies/comp-1",
		},
		{
			name:     "delete company",
			response: companyDeletedJSON,
			call: func(ctx context.Context, client *Client) error {
				d, err := client.Companies.Delete(ctx, "comp-1")
				if err != nil {
					return err
				}
				if d.Id == nil || *d.Id != "comp-1" {
					t.Fatalf("Id = %v", d.Id)
				}
				return nil
			},
			wantMethod: http.MethodDelete,
			wantPath:   "/companies/comp-1",
		},
		{
			name:     "list all companies",
			response: companyListJSON,
			call: func(ctx context.Context, client *Client) error {
				list, err := client.Companies.ListAll(ctx)
				if err != nil {
					return err
				}
				if list.TotalCount == nil || *list.TotalCount != 0 {
					t.Fatalf("TotalCount = %v", list.TotalCount)
				}
				return nil
			},
			wantMethod: http.MethodPost,
			wantPath:   "/companies/list",
		},
		{
			name:     "scroll companies",
			response: companyScrollJSON,
			call: func(ctx context.Context, client *Client) error {
				scroll, err := client.Companies.Scroll(ctx)
				if err != nil {
					return err
				}
				if scroll.ScrollParam == nil || *scroll.ScrollParam != "token-1" {
					t.Fatalf("ScrollParam = %v", scroll.ScrollParam)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/companies/scroll",
		},
		{
			name:     "list company notes",
			response: noteListJSON,
			call: func(ctx context.Context, client *Client) error {
				list, err := client.Companies.ListNotes(ctx, "comp-1")
				if err != nil {
					return err
				}
				_ = list
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/companies/comp-1/notes",
		},
		{
			name:     "list company contacts",
			response: companyContactsJSON,
			call: func(ctx context.Context, client *Client) error {
				list, err := client.Companies.ListContacts(ctx, "comp-1")
				if err != nil {
					return err
				}
				if list.TotalCount == nil || *list.TotalCount != 0 {
					t.Fatalf("TotalCount = %v", list.TotalCount)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/companies/comp-1/contacts",
		},
		{
			name:     "list company segments",
			response: companySegmentsJSON,
			call: func(ctx context.Context, client *Client) error {
				list, err := client.Companies.ListSegments(ctx, "comp-1")
				if err != nil {
					return err
				}
				_ = list
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/companies/comp-1/segments",
		},
		{
			name:     "attach contact to company",
			response: companyJSON,
			call: func(ctx context.Context, client *Client) error {
				c, err := client.Companies.AttachContact(ctx, "contact-1", "comp-1")
				if err != nil {
					return err
				}
				if c.Id == nil || *c.Id != "comp-1" {
					t.Fatalf("Id = %v", c.Id)
				}
				return nil
			},
			wantMethod: http.MethodPost,
			wantPath:   "/contacts/contact-1/companies",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "id"); got != "comp-1" {
					t.Fatalf("id = %q", got)
				}
			},
		},
		{
			name:     "detach contact from company",
			response: companyJSON,
			call: func(ctx context.Context, client *Client) error {
				c, err := client.Companies.DetachContact(ctx, "contact-1", "comp-1")
				if err != nil {
					return err
				}
				if c.Id == nil || *c.Id != "comp-1" {
					t.Fatalf("Id = %v", c.Id)
				}
				return nil
			},
			wantMethod: http.MethodDelete,
			wantPath:   "/contacts/contact-1/companies/comp-1",
		},
		{
			name:     "list companies for contact",
			response: contactCompaniesJSON,
			call: func(ctx context.Context, client *Client) error {
				list, err := client.Companies.ListForContact(ctx, "contact-1")
				if err != nil {
					return err
				}
				if list.TotalCount == nil || *list.TotalCount != 0 {
					t.Fatalf("TotalCount = %v", list.TotalCount)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/contacts/contact-1/companies",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod, gotPath string
			transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				gotMethod = req.Method
				gotPath = req.URL.Path
				if tt.wantBody != nil {
					var body map[string]any
					if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
						t.Fatalf("decode body: %v", err)
					}
					tt.wantBody(t, body)
				}
				return jsonResponse(req, http.StatusOK, tt.response), nil
			})
			client := newCompaniesTestClient(t, transport)
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

func TestCompaniesServiceErrors(t *testing.T) {
	errBody := `{"type":"error.list","errors":[{"code":"unauthorized","message":"Access Token Invalid"}],"request_id":"12a938a3-314e-4939-b773-5cd45738bd21"}`

	tests := []struct {
		name string
		call func(context.Context, *Client) error
	}{
		{
			name: "create or update company",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Companies.CreateOrUpdate(ctx, CompanyCreate{})
				return err
			},
		},
		{
			name: "retrieve company",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Companies.Retrieve(ctx, "comp-1")
				return err
			},
		},
		{
			name: "list all companies",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Companies.ListAll(ctx)
				return err
			},
		},
		{
			name: "attach contact to company",
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Companies.AttachContact(ctx, "contact-1", "comp-1")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return jsonResponse(req, http.StatusUnauthorized, errBody), nil
			})
			client := newCompaniesTestClient(t, transport)
			err := tt.call(context.Background(), client)
			if err == nil {
				t.Fatal("expected error for non-200 response")
			}
		})
	}
}

func TestCompaniesTransportErrors(t *testing.T) {
	transportErr := errors.New("dial tcp: connection refused")

	calls := []struct {
		name string
		call func(context.Context, *Client) error
	}{
		{"create or update", func(ctx context.Context, c *Client) error {
			_, err := c.Companies.CreateOrUpdate(ctx, CompanyCreate{})
			return err
		}},
		{"retrieve", func(ctx context.Context, c *Client) error { _, err := c.Companies.Retrieve(ctx, "comp-1"); return err }},
		{"retrieve by ID", func(ctx context.Context, c *Client) error {
			_, err := c.Companies.RetrieveByID(ctx, "ext-1")
			return err
		}},
		{"update", func(ctx context.Context, c *Client) error {
			_, err := c.Companies.Update(ctx, "comp-1", CompanyUpdate{})
			return err
		}},
		{"delete", func(ctx context.Context, c *Client) error { _, err := c.Companies.Delete(ctx, "comp-1"); return err }},
		{"list all", func(ctx context.Context, c *Client) error { _, err := c.Companies.ListAll(ctx); return err }},
		{"scroll", func(ctx context.Context, c *Client) error { _, err := c.Companies.Scroll(ctx); return err }},
		{"list notes", func(ctx context.Context, c *Client) error { _, err := c.Companies.ListNotes(ctx, "comp-1"); return err }},
		{"list contacts", func(ctx context.Context, c *Client) error {
			_, err := c.Companies.ListContacts(ctx, "comp-1")
			return err
		}},
		{"list segments", func(ctx context.Context, c *Client) error {
			_, err := c.Companies.ListSegments(ctx, "comp-1")
			return err
		}},
		{"attach contact", func(ctx context.Context, c *Client) error {
			_, err := c.Companies.AttachContact(ctx, "c-1", "comp-1")
			return err
		}},
		{"detach contact", func(ctx context.Context, c *Client) error {
			_, err := c.Companies.DetachContact(ctx, "c-1", "comp-1")
			return err
		}},
		{"list for contact", func(ctx context.Context, c *Client) error {
			_, err := c.Companies.ListForContact(ctx, "c-1")
			return err
		}},
	}

	for _, tt := range calls {
		t.Run(tt.name, func(t *testing.T) {
			transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return nil, transportErr
			})
			client := newCompaniesTestClient(t, transport)
			if err := tt.call(context.Background(), client); err == nil {
				t.Fatal("expected transport error")
			}
		})
	}
}

func TestCompaniesValidation(t *testing.T) {
	client := newCompaniesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		t.Fatal("unexpected HTTP request")
		return nil, nil
	}))
	ctx := context.Background()

	tests := []struct {
		name string
		call func() error
	}{
		{"retrieve: empty ID", func() error { _, err := client.Companies.Retrieve(ctx, ""); return err }},
		{"retrieve by ID: empty ID", func() error { _, err := client.Companies.RetrieveByID(ctx, ""); return err }},
		{"update: empty ID", func() error { _, err := client.Companies.Update(ctx, "", CompanyUpdate{}); return err }},
		{"delete: empty ID", func() error { _, err := client.Companies.Delete(ctx, ""); return err }},
		{"list notes: empty ID", func() error { _, err := client.Companies.ListNotes(ctx, ""); return err }},
		{"list contacts: empty ID", func() error { _, err := client.Companies.ListContacts(ctx, ""); return err }},
		{"list segments: empty ID", func() error { _, err := client.Companies.ListSegments(ctx, ""); return err }},
		{"attach contact: empty contact ID", func() error { _, err := client.Companies.AttachContact(ctx, "", "comp-1"); return err }},
		{"attach contact: empty company ID", func() error { _, err := client.Companies.AttachContact(ctx, "c-1", ""); return err }},
		{"detach contact: empty contact ID", func() error { _, err := client.Companies.DetachContact(ctx, "", "comp-1"); return err }},
		{"detach contact: empty company ID", func() error { _, err := client.Companies.DetachContact(ctx, "c-1", ""); return err }},
		{"list for contact: empty ID", func() error { _, err := client.Companies.ListForContact(ctx, ""); return err }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
