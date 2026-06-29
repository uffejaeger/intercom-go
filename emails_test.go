package intercom

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestEmailsServiceRequests(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		call       func(context.Context, *Client) error
		wantMethod string
		wantPath   string
	}{
		{
			name:     "list emails",
			response: `{"type":"list","data":[{"type":"email_setting","id":"1","email":"support@company.com"}]}`,
			call: func(ctx context.Context, client *Client) error {
				emails, err := client.Emails.List(ctx)
				if err != nil {
					return err
				}
				if emails.Data == nil || len(*emails.Data) != 1 {
					t.Fatalf("emails.Data = %#v", emails.Data)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/emails",
		},
		{
			name:     "retrieve email",
			response: `{"type":"email_setting","id":"10","email":"support@company.com"}`,
			call: func(ctx context.Context, client *Client) error {
				email, err := client.Emails.Retrieve(ctx, "10")
				if err != nil {
					return err
				}
				if email.Id == nil || *email.Id != "10" {
					t.Fatalf("email.Id = %v", email.Id)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/emails/10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod, gotPath string
			client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
				gotMethod = req.Method
				gotPath = req.URL.Path
				return jsonResponse(req, http.StatusOK, tt.response), nil
			}))

			if err := tt.call(context.Background(), client); err != nil {
				t.Fatalf("call returned error: %v", err)
			}
			if gotMethod != tt.wantMethod {
				t.Fatalf("Method = %q, want %q", gotMethod, tt.wantMethod)
			}
			if gotPath != tt.wantPath {
				t.Fatalf("Path = %q, want %q", gotPath, tt.wantPath)
			}
		})
	}
}

func TestEmailsServiceValidationAndTransport(t *testing.T) {
	client := newSupportingServicesTestClient(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("connection refused")
	}))

	if _, err := client.Emails.Retrieve(context.Background(), ""); err == nil {
		t.Fatal("expected validation error")
	}
	if _, err := client.Emails.List(context.Background()); err == nil {
		t.Fatal("expected transport error")
	}
	if _, err := client.Emails.Retrieve(context.Background(), "10"); err == nil {
		t.Fatal("expected transport error")
	}
}
