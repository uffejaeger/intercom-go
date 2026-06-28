package intercom

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestVisitorsServiceRequests(t *testing.T) {
	name := "Visitor Name"
	userID := "visitor-user-1"
	contactID := "contact-1"
	visitorEmail := "visitor@example.com"
	contactEmail := "contact@example.com"
	customAttributes := map[string]string{"plan": "pro"}

	tests := []struct {
		name       string
		response   string
		call       func(context.Context, *Client) error
		wantMethod string
		wantPath   string
		wantQuery  string
		wantBody   func(*testing.T, map[string]any)
	}{
		{
			name:     "get by user id",
			response: `{"type":"visitor","id":"v1","user_id":"u1"}`,
			call: func(ctx context.Context, client *Client) error {
				visitor, err := client.Visitors.GetByUserID(ctx, "u1")
				if err != nil {
					return err
				}
				if visitor.UserId == nil || *visitor.UserId != "u1" {
					t.Fatalf("visitor.UserId = %v", visitor.UserId)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/visitors",
			wantQuery:  "user_id=u1",
		},
		{
			name:     "update visitor",
			response: `{"type":"visitor","id":"v1","user_id":"visitor-user-1","name":"Visitor Name"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Visitors.Update(ctx, VisitorUpdate{
					Name:             &name,
					UserID:           &userID,
					CustomAttributes: &customAttributes,
				})
				return err
			},
			wantMethod: http.MethodPut,
			wantPath:   "/visitors",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if body["name"] != "Visitor Name" {
					t.Fatalf("body[name] = %#v", body["name"])
				}
				if body["user_id"] != "visitor-user-1" {
					t.Fatalf("body[user_id] = %#v", body["user_id"])
				}
				attrs, ok := body["custom_attributes"].(map[string]any)
				if !ok || attrs["plan"] != "pro" {
					t.Fatalf("body[custom_attributes] = %#v", body["custom_attributes"])
				}
			},
		},
		{
			name:     "convert visitor",
			response: `{"type":"contact","id":"c1","external_id":"contact-1"}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Visitors.Convert(ctx, VisitorConvert{
					Type: "user",
					User: VisitorConvertContact{
						ID:    &contactID,
						Email: &contactEmail,
					},
					Visitor: VisitorConvertSource{
						UserID: &userID,
						Email:  &visitorEmail,
					},
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/visitors/convert",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if body["type"] != "user" {
					t.Fatalf("body[type] = %#v", body["type"])
				}
				user, ok := body["user"].(map[string]any)
				if !ok || user["id"] != "contact-1" || user["email"] != "contact@example.com" {
					t.Fatalf("body[user] = %#v", body["user"])
				}
				visitor, ok := body["visitor"].(map[string]any)
				if !ok || visitor["user_id"] != "visitor-user-1" || visitor["email"] != "visitor@example.com" {
					t.Fatalf("body[visitor] = %#v", body["visitor"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod, gotPath, gotQuery string
			transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				gotMethod = req.Method
				gotPath = req.URL.Path
				gotQuery = req.URL.RawQuery
				if tt.wantBody != nil {
					var body map[string]any
					if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
						t.Fatalf("decode body: %v", err)
					}
					tt.wantBody(t, body)
				}
				return jsonResponse(req, http.StatusOK, tt.response), nil
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
			if tt.wantQuery != "" && gotQuery != tt.wantQuery {
				t.Fatalf("query = %q, want %q", gotQuery, tt.wantQuery)
			}
		})
	}
}
