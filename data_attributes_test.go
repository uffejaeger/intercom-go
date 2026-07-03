package intercom

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

func TestDataAttributesServiceRequests(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		call       func(context.Context, *Client) error
		wantMethod string
		wantPath   string
		wantQuery  map[string]string
		wantBody   func(t *testing.T, body map[string]any)
	}{
		{
			name:     "list data attributes",
			response: `{"type":"list","data":[{"type":"data_attribute","id":77,"model":"contact","name":"plan"}]}`,
			call: func(ctx context.Context, client *Client) error {
				attributes, err := client.DataAttributes.List(ctx, DataAttributeListParams{
					Model:           DataAttributeModelContact,
					IncludeArchived: true,
				})
				if err != nil {
					return err
				}
				if attributes.Data == nil || len(*attributes.Data) != 1 {
					t.Fatalf("attributes.Data = %#v", attributes.Data)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/data_attributes",
			wantQuery: map[string]string{
				"model":            "contact",
				"include_archived": "true",
			},
		},
		{
			name:     "create data attribute",
			response: `{"type":"data_attribute","id":77,"model":"contact","name":"plan"}`,
			call: func(ctx context.Context, client *Client) error {
				description := "Customer plan"
				attribute, err := client.DataAttributes.Create(ctx, DataAttributeCreate{
					Name:        "plan",
					Model:       DataAttributeModelContact,
					DataType:    DataAttributeDataTypeList,
					Description: &description,
					Options:     []string{"free", "paid"},
				})
				if err != nil {
					return err
				}
				if attribute.Id == nil || *attribute.Id != 77 {
					t.Fatalf("attribute.Id = %v", attribute.Id)
				}
				return nil
			},
			wantMethod: http.MethodPost,
			wantPath:   "/data_attributes",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "name"); got != "plan" {
					t.Fatalf("name = %q", got)
				}
				if got := nestedString(body, "model"); got != "contact" {
					t.Fatalf("model = %q", got)
				}
				if got := nestedString(body, "data_type"); got != "options" {
					t.Fatalf("data_type = %q", got)
				}
				options, ok := body["options"].([]any)
				if !ok || len(options) != 2 {
					t.Fatalf("options = %#v", body["options"])
				}
			},
		},
		{
			name:     "update data attribute",
			response: `{"type":"data_attribute","id":77,"model":"contact","name":"plan","archived":true}`,
			call: func(ctx context.Context, client *Client) error {
				archived := true
				attribute, err := client.DataAttributes.Update(ctx, "77", DataAttributeUpdate{
					Archived: &archived,
					Options:  []string{"free", "paid"},
				})
				if err != nil {
					return err
				}
				if attribute.Archived == nil || !*attribute.Archived {
					t.Fatalf("attribute.Archived = %v", attribute.Archived)
				}
				return nil
			},
			wantMethod: http.MethodPut,
			wantPath:   "/data_attributes/77",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got, ok := body["archived"].(bool); !ok || !got {
					t.Fatalf("archived = %#v", body["archived"])
				}
				options, ok := body["options"].([]any)
				if !ok || len(options) != 2 {
					t.Fatalf("options = %#v", body["options"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod, gotPath string
			var gotQuery map[string]string
			var gotBody map[string]any

			client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
				gotMethod = req.Method
				gotPath = req.URL.Path
				gotQuery = map[string]string{}
				for key, values := range req.URL.Query() {
					if len(values) > 0 {
						gotQuery[key] = values[0]
					}
				}

				if req.Body != nil {
					defer req.Body.Close()
					if err := json.NewDecoder(req.Body).Decode(&gotBody); err != nil {
						return nil, err
					}
				}

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
			for key, want := range tt.wantQuery {
				if gotQuery[key] != want {
					t.Fatalf("Query[%q] = %q, want %q", key, gotQuery[key], want)
				}
			}
			if tt.wantBody != nil {
				tt.wantBody(t, gotBody)
			}
		})
	}
}

func TestDataAttributeModelHelpers(t *testing.T) {
	t.Run("list data attribute model", func(t *testing.T) {
		tests := []struct {
			name    string
			model   DataAttributeModel
			want    gen.LisDataAttributesParamsModel
			wantErr bool
		}{
			{name: "contact", model: DataAttributeModelContact, want: gen.LisDataAttributesParamsModelContact},
			{name: "company", model: DataAttributeModelCompany, want: gen.LisDataAttributesParamsModelCompany},
			{name: "conversation", model: DataAttributeModelConversation, want: gen.LisDataAttributesParamsModelConversation},
			{name: "empty", model: ""},
			{name: "unsupported", model: DataAttributeModel("workspace"), wantErr: true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := listDataAttributeModel(tt.model)
				if tt.wantErr {
					if err == nil {
						t.Fatal("expected error")
					}
					return
				}
				if err != nil {
					t.Fatalf("listDataAttributeModel returned error: %v", err)
				}
				if got != tt.want {
					t.Fatalf("model = %q, want %q", got, tt.want)
				}
			})
		}
	})

	t.Run("create data attribute model", func(t *testing.T) {
		tests := []struct {
			name    string
			model   DataAttributeModel
			wantErr bool
		}{
			{name: "contact", model: DataAttributeModelContact},
			{name: "company", model: DataAttributeModelCompany},
			{name: "conversation", model: DataAttributeModelConversation, wantErr: true},
			{name: "empty", model: "", wantErr: true},
			{name: "unsupported", model: DataAttributeModel("workspace"), wantErr: true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := createDataAttributeModel(tt.model)
				if tt.wantErr {
					if err == nil {
						t.Fatal("expected error")
					}
					return
				}
				if err != nil {
					t.Fatalf("createDataAttributeModel returned error: %v", err)
				}
				if got != tt.model {
					t.Fatalf("model = %q, want %q", got, tt.model)
				}
			})
		}
	})
}

func TestDataAttributeOptions(t *testing.T) {
	tests := []struct {
		name     string
		dataType DataAttributeDataType
		values   []string
		wantLen  int
		wantErr  bool
	}{
		{name: "string without options", dataType: DataAttributeDataTypeString},
		{name: "string with options", dataType: DataAttributeDataTypeString, values: []string{"free", "paid"}, wantErr: true},
		{name: "list without options", dataType: DataAttributeDataTypeList, wantErr: true},
		{name: "list one option", dataType: DataAttributeDataTypeList, values: []string{"free"}, wantErr: true},
		{name: "list empty option", dataType: DataAttributeDataTypeList, values: []string{"free", ""}, wantErr: true},
		{name: "list valid", dataType: DataAttributeDataTypeList, values: []string{"free", "paid"}, wantLen: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options, err := dataAttributeOptions(tt.dataType, tt.values)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("dataAttributeOptions returned error: %v", err)
			}
			if len(options) != tt.wantLen {
				t.Fatalf("len(options) = %d, want %d", len(options), tt.wantLen)
			}
		})
	}
}

func TestDataAttributesServiceTransportErrors(t *testing.T) {
	transportErr := errors.New("connection refused")
	client := newSupportingServicesTestClient(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, transportErr
	}))

	ctx := context.Background()

	tests := []struct {
		name string
		call func() error
	}{
		{
			name: "list",
			call: func() error {
				_, err := client.DataAttributes.List(ctx, DataAttributeListParams{})
				return err
			},
		},
		{
			name: "create",
			call: func() error {
				_, err := client.DataAttributes.Create(ctx, DataAttributeCreate{
					Name:     "plan",
					Model:    DataAttributeModelContact,
					DataType: DataAttributeDataTypeString,
				})
				return err
			},
		},
		{
			name: "update",
			call: func() error {
				archived := true
				_, err := client.DataAttributes.Update(ctx, "77", DataAttributeUpdate{Archived: &archived})
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err == nil {
				t.Fatal("expected transport error")
			}
		})
	}
}

func TestDataAttributesServiceValidation(t *testing.T) {
	client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected request: %s %s", req.Method, req.URL.Path)
		return nil, nil
	}))

	ctx := context.Background()
	if _, err := client.DataAttributes.Create(ctx, DataAttributeCreate{
		Name:  "plan",
		Model: DataAttributeModelContact,
	}); err == nil {
		t.Fatal("expected validation error")
	}
}
