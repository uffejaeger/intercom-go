package intercom

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
)

func TestDataEventsServiceRequests(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		call       func(context.Context, *Client) error
		wantMethod string
		wantPath   string
		wantQuery  map[string]string
		wantBody   func(t *testing.T, body map[string]any)
	}{
		{
			name:       "list data events",
			statusCode: http.StatusOK,
			response:   `{"type":"event.summary","events":[{"name":"updated-plan","count":2}],"user_id":"user-1"}`,
			call: func(ctx context.Context, client *Client) error {
				summary, err := client.DataEvents.List(ctx, DataEventListFilter{UserID: "user-1", Summary: true})
				if err != nil {
					return err
				}
				if summary.UserId == nil || *summary.UserId != "user-1" {
					t.Fatalf("summary.UserId = %v", summary.UserId)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/events",
			wantQuery: map[string]string{
				"type":    "user",
				"user_id": "user-1",
				"summary": "true",
			},
		},
		{
			name:       "create data event",
			statusCode: http.StatusAccepted,
			response:   ``,
			call: func(ctx context.Context, client *Client) error {
				userID := "user-1"
				createdAt := 1710498600
				return client.DataEvents.Create(ctx, DataEventCreate{
					EventName: "updated-plan",
					CreatedAt: &createdAt,
					UserID:    &userID,
					Metadata:  map[string]any{"plan": "pro"},
				})
			},
			wantMethod: http.MethodPost,
			wantPath:   "/events",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "event_name"); got != "updated-plan" {
					t.Fatalf("event_name = %q", got)
				}
				if got := nestedString(body, "user_id"); got != "user-1" {
					t.Fatalf("user_id = %q", got)
				}
			},
		},
		{
			name:       "create data event summaries",
			statusCode: http.StatusOK,
			response:   ``,
			call: func(ctx context.Context, client *Client) error {
				first := 1710498600
				last := 1710499600
				return client.DataEvents.CreateSummaries(ctx, DataEventSummariesCreate{
					UserID: "user-1",
					EventSummaries: []DataEventSummaryCreate{
						{EventName: "updated-plan", Count: 2, First: &first, Last: &last},
					},
				})
			},
			wantMethod: http.MethodPost,
			wantPath:   "/events/summaries",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if got := nestedString(body, "user_id"); got != "user-1" {
					t.Fatalf("user_id = %q", got)
				}
				summaries, ok := body["event_summaries"].([]any)
				if !ok || len(summaries) != 1 {
					t.Fatalf("event_summaries = %#v", body["event_summaries"])
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
					if err := json.NewDecoder(req.Body).Decode(&gotBody); err != nil && tt.response != "" {
						return nil, err
					}
				}
				if tt.response == "" {
					return &http.Response{
						StatusCode: tt.statusCode,
						Header:     http.Header{"Content-Type": []string{"application/json"}},
						Body:       http.NoBody,
						Request:    req,
					}, nil
				}
				return jsonResponse(req, tt.statusCode, tt.response), nil
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

func TestDataEventsHelpers(t *testing.T) {
	t.Run("data event list query", func(t *testing.T) {
		values, err := dataEventListQuery(DataEventListFilter{IntercomUserID: "ic-1"})
		if err != nil {
			t.Fatalf("dataEventListQuery returned error: %v", err)
		}
		if got := values.Get("intercom_user_id"); got != "ic-1" {
			t.Fatalf("intercom_user_id = %q", got)
		}
	})

	t.Run("data event list query email", func(t *testing.T) {
		values, err := dataEventListQuery(DataEventListFilter{Email: "user@example.com"})
		if err != nil {
			t.Fatalf("dataEventListQuery returned error: %v", err)
		}
		if got := values.Get("email"); got != "user@example.com" {
			t.Fatalf("email = %q", got)
		}
	})

	t.Run("validate data event identifiers", func(t *testing.T) {
		userID := "user-1"
		if err := validateDataEventIdentifiers(&userID, nil, nil); err != nil {
			t.Fatalf("validateDataEventIdentifiers returned error: %v", err)
		}
		email := "user@example.com"
		if err := validateDataEventIdentifiers(nil, &email, nil); err != nil {
			t.Fatalf("validateDataEventIdentifiers returned error: %v", err)
		}
		id := "lead-1"
		if err := validateDataEventIdentifiers(nil, nil, &id); err != nil {
			t.Fatalf("validateDataEventIdentifiers returned error: %v", err)
		}
		if err := validateDataEventIdentifiers(nil, nil, nil); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestDataEventsTransportAndValidation(t *testing.T) {
	transportErr := errors.New("connection refused")
	client := newSupportingServicesTestClient(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, transportErr
	}))

	ctx := context.Background()
	userID := "user-1"

	tests := []struct {
		name string
		call func() error
	}{
		{"list transport", func() error {
			_, err := client.DataEvents.List(ctx, DataEventListFilter{UserID: "user-1"})
			return err
		}},
		{"create transport", func() error {
			return client.DataEvents.Create(ctx, DataEventCreate{EventName: "updated-plan", UserID: &userID})
		}},
		{"summaries transport", func() error {
			return client.DataEvents.CreateSummaries(ctx, DataEventSummariesCreate{
				UserID:         "user-1",
				EventSummaries: []DataEventSummaryCreate{{EventName: "updated-plan", Count: 1}},
			})
		}},
		{"list invalid", func() error {
			_, err := client.DataEvents.List(ctx, DataEventListFilter{})
			return err
		}},
		{"create invalid identifiers", func() error {
			email := "user@example.com"
			return client.DataEvents.Create(ctx, DataEventCreate{EventName: "updated-plan", UserID: &userID, Email: &email})
		}},
		{"create invalid name", func() error {
			return client.DataEvents.Create(ctx, DataEventCreate{UserID: &userID})
		}},
		{"summaries invalid", func() error {
			return client.DataEvents.CreateSummaries(ctx, DataEventSummariesCreate{})
		}},
		{"summaries invalid name", func() error {
			return client.DataEvents.CreateSummaries(ctx, DataEventSummariesCreate{
				UserID:         "user-1",
				EventSummaries: []DataEventSummaryCreate{{Count: 1}},
			})
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestDataEventsListReadError(t *testing.T) {
	client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       failingReadCloser{},
			Request:    req,
		}, nil
	}))

	if _, err := client.DataEvents.List(context.Background(), DataEventListFilter{UserID: "user-1"}); err == nil {
		t.Fatal("expected read error")
	}
}

type failingReadCloser struct{}

func (failingReadCloser) Read([]byte) (int, error) {
	return 0, errors.New("read failed")
}

func (failingReadCloser) Close() error {
	return nil
}

var _ io.ReadCloser = failingReadCloser{}
