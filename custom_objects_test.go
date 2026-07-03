package intercom

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestCustomObjectsServiceRequests(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		call       func(context.Context, *Client) error
		wantMethod string
		wantPath   string
		wantQuery  string
	}{
		{
			name:     "create or update",
			response: `{"id":"22","type":"Order","external_id":"external-1","custom_attributes":{"order_number":"ORDER-12345"}}`,
			call: func(ctx context.Context, client *Client) error {
				externalID := "external-1"
				instance, err := client.CustomObjects.CreateOrUpdate(ctx, "Order", CustomObjectInstanceCreateOrUpdate{
					ExternalId:       &externalID,
					CustomAttributes: &map[string]string{"order_number": "ORDER-12345"},
				})
				if err != nil {
					return err
				}
				if instance.Id == nil || *instance.Id != "22" {
					t.Fatalf("instance.Id = %v", instance.Id)
				}
				return nil
			},
			wantMethod: http.MethodPost,
			wantPath:   "/custom_object_instances/Order",
		},
		{
			name:     "get by id",
			response: `{"id":"22","type":"Order","external_id":"external-1"}`,
			call: func(ctx context.Context, client *Client) error {
				instance, err := client.CustomObjects.Get(ctx, "Order", "22")
				if err != nil {
					return err
				}
				if instance.Id == nil || *instance.Id != "22" {
					t.Fatalf("instance.Id = %v", instance.Id)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/custom_object_instances/Order/22",
		},
		{
			name:     "get by external id",
			response: `{"id":"22","type":"Order","external_id":"external-1"}`,
			call: func(ctx context.Context, client *Client) error {
				instance, err := client.CustomObjects.GetByExternalID(ctx, "Order", "external-1")
				if err != nil {
					return err
				}
				if instance.ExternalId == nil || *instance.ExternalId != "external-1" {
					t.Fatalf("instance.ExternalId = %v", instance.ExternalId)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/custom_object_instances/Order",
			wantQuery:  "external_id=external-1",
		},
		{
			name:     "delete by id",
			response: `{"id":"22","object":"Order","deleted":true}`,
			call: func(ctx context.Context, client *Client) error {
				deleted, err := client.CustomObjects.Delete(ctx, "Order", "22")
				if err != nil {
					return err
				}
				if deleted.Deleted == nil || !*deleted.Deleted {
					t.Fatalf("deleted.Deleted = %v", deleted.Deleted)
				}
				return nil
			},
			wantMethod: http.MethodDelete,
			wantPath:   "/custom_object_instances/Order/22",
		},
		{
			name:     "delete by external id",
			response: `{"id":"22","object":"Order","deleted":true}`,
			call: func(ctx context.Context, client *Client) error {
				deleted, err := client.CustomObjects.DeleteByExternalID(ctx, "Order", "external-1")
				if err != nil {
					return err
				}
				if deleted.Deleted == nil || !*deleted.Deleted {
					t.Fatalf("deleted.Deleted = %v", deleted.Deleted)
				}
				return nil
			},
			wantMethod: http.MethodDelete,
			wantPath:   "/custom_object_instances/Order",
			wantQuery:  "external_id=external-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod, gotPath, gotQuery string
			client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
				gotMethod = req.Method
				gotPath = req.URL.Path
				gotQuery = req.URL.RawQuery
				return jsonResponse(req, http.StatusOK, tt.response), nil
			}))

			if err := tt.call(context.Background(), client); err != nil {
				t.Fatalf("call returned error: %v", err)
			}
			if gotMethod != tt.wantMethod {
				t.Fatalf("method = %q, want %q", gotMethod, tt.wantMethod)
			}
			if gotPath != tt.wantPath {
				t.Fatalf("path = %q, want %q", gotPath, tt.wantPath)
			}
			if gotQuery != tt.wantQuery {
				t.Fatalf("query = %q, want %q", gotQuery, tt.wantQuery)
			}
		})
	}
}

func TestCustomObjectsServiceErrors(t *testing.T) {
	ctx := context.Background()
	apiErr := `{"type":"error.list","errors":[{"code":"not_found","message":"not found"}],"request_id":"req-1"}`

	t.Run("api errors", func(t *testing.T) {
		client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return jsonResponse(req, http.StatusNotFound, apiErr), nil
		}))
		calls := []func(context.Context, *Client) error{
			func(ctx context.Context, client *Client) error {
				_, err := client.CustomObjects.CreateOrUpdate(ctx, "Order", CustomObjectInstanceCreateOrUpdate{})
				return err
			},
			func(ctx context.Context, client *Client) error {
				_, err := client.CustomObjects.Get(ctx, "Order", "22")
				return err
			},
			func(ctx context.Context, client *Client) error {
				_, err := client.CustomObjects.GetByExternalID(ctx, "Order", "external-1")
				return err
			},
			func(ctx context.Context, client *Client) error {
				_, err := client.CustomObjects.Delete(ctx, "Order", "22")
				return err
			},
			func(ctx context.Context, client *Client) error {
				_, err := client.CustomObjects.DeleteByExternalID(ctx, "Order", "external-1")
				return err
			},
		}
		for _, call := range calls {
			if err := call(ctx, client); err == nil {
				t.Fatal("expected api error")
			}
		}
	})

	t.Run("transport errors", func(t *testing.T) {
		transportErr := errors.New("connection refused")
		client := newSupportingServicesTestClient(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, transportErr
		}))
		calls := []func(context.Context, *Client) error{
			func(ctx context.Context, client *Client) error {
				_, err := client.CustomObjects.CreateOrUpdate(ctx, "Order", CustomObjectInstanceCreateOrUpdate{})
				return err
			},
			func(ctx context.Context, client *Client) error {
				_, err := client.CustomObjects.Get(ctx, "Order", "22")
				return err
			},
			func(ctx context.Context, client *Client) error {
				_, err := client.CustomObjects.GetByExternalID(ctx, "Order", "external-1")
				return err
			},
			func(ctx context.Context, client *Client) error {
				_, err := client.CustomObjects.Delete(ctx, "Order", "22")
				return err
			},
			func(ctx context.Context, client *Client) error {
				_, err := client.CustomObjects.DeleteByExternalID(ctx, "Order", "external-1")
				return err
			},
		}
		for _, call := range calls {
			if err := call(ctx, client); err == nil {
				t.Fatal("expected transport error")
			}
		}
	})
}

func TestCustomObjectsValidation(t *testing.T) {
	client := newSupportingServicesTestClient(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("unexpected HTTP request")
		return nil, nil
	}))
	ctx := context.Background()

	if _, err := client.CustomObjects.CreateOrUpdate(ctx, "", CustomObjectInstanceCreateOrUpdate{}); err == nil {
		t.Fatal("expected validation error for empty custom object type")
	}
	if _, err := client.CustomObjects.Get(ctx, "", "22"); err == nil {
		t.Fatal("expected validation error for empty custom object type")
	}
	if _, err := client.CustomObjects.Get(ctx, "Order", ""); err == nil {
		t.Fatal("expected validation error for empty instance ID")
	}
	if _, err := client.CustomObjects.GetByExternalID(ctx, "", "external-1"); err == nil {
		t.Fatal("expected validation error for empty custom object type")
	}
	if _, err := client.CustomObjects.GetByExternalID(ctx, "Order", ""); err == nil {
		t.Fatal("expected validation error for empty external ID")
	}
	if _, err := client.CustomObjects.Delete(ctx, "", "22"); err == nil {
		t.Fatal("expected validation error for empty custom object type")
	}
	if _, err := client.CustomObjects.Delete(ctx, "Order", ""); err == nil {
		t.Fatal("expected validation error for empty instance ID")
	}
	if _, err := client.CustomObjects.DeleteByExternalID(ctx, "", "external-1"); err == nil {
		t.Fatal("expected validation error for empty custom object type")
	}
	if _, err := client.CustomObjects.DeleteByExternalID(ctx, "Order", ""); err == nil {
		t.Fatal("expected validation error for empty external ID")
	}
}
