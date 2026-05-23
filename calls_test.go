package intercom

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func newCallsTestClient(t *testing.T, transport http.RoundTripper) *Client {
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

func TestCallsServiceRequests(t *testing.T) {
	errBody := `{"type":"error.list","errors":[{"code":"unauthorized"}],"request_id":"req-1"}`

	tests := []struct {
		name       string
		response   string
		call       func(context.Context, *Client) error
		wantMethod string
		wantPath   string
		wantBody   func(*testing.T, map[string]any)
	}{
		{
			name:     "list calls",
			response: `{"data":[{"id":"c1","type":"call"}],"pages":{}}`,
			call: func(ctx context.Context, c *Client) error {
				list, err := c.Calls.List(ctx)
				if err != nil {
					return err
				}
				if list.Data == nil || len(*list.Data) != 1 {
					t.Fatal("expected one call")
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/calls",
		},
		{
			name:     "get call",
			response: `{"id":"c1","type":"call"}`,
			call: func(ctx context.Context, c *Client) error {
				call, err := c.Calls.Get(ctx, "c1")
				if err != nil {
					return err
				}
				if call.Id == nil || *call.Id != "c1" {
					t.Fatalf("Id = %v", call.Id)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/calls/c1",
		},
		{
			name:     "list with transcripts",
			response: `{"data":[{"id":"c1","conversation_id":"conv1"}],"type":"list"}`,
			call: func(ctx context.Context, c *Client) error {
				list, err := c.Calls.ListWithTranscripts(ctx, []string{"conv1"})
				if err != nil {
					return err
				}
				if len(list.Data) != 1 {
					t.Fatalf("len(data) = %d, want 1", len(list.Data))
				}
				if list.Data[0].ConversationId == nil || *list.Data[0].ConversationId != "conv1" {
					t.Fatalf("ConversationId = %v", list.Data[0].ConversationId)
				}
				return nil
			},
			wantMethod: http.MethodPost,
			wantPath:   "/calls/search",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				ids, ok := body["conversation_ids"].([]any)
				if !ok || len(ids) != 1 || ids[0] != "conv1" {
					t.Fatalf("conversation_ids = %v", body["conversation_ids"])
				}
			},
		},
		{
			name:     "register fin voice call",
			response: `{"external_call_id":"ext-1","status":"registered"}`,
			call: func(ctx context.Context, c *Client) error {
				call, err := c.Calls.RegisterFinVoiceCall(ctx, RegisterFinVoiceCallRequest{
					CallId:      "ext-1",
					PhoneNumber: "+15550001234",
				})
				if err != nil {
					return err
				}
				if call.ExternalCallId == nil || *call.ExternalCallId != "ext-1" {
					t.Fatalf("ExternalCallId = %v", call.ExternalCallId)
				}
				return nil
			},
			wantMethod: http.MethodPost,
			wantPath:   "/fin_voice/register",
			wantBody: func(t *testing.T, body map[string]any) {
				t.Helper()
				if body["call_id"] != "ext-1" {
					t.Fatalf("call_id = %v", body["call_id"])
				}
			},
		},
		{
			name:     "collect fin voice call by ID",
			response: `{"external_call_id":"ext-1","status":"registered"}`,
			call: func(ctx context.Context, c *Client) error {
				call, err := c.Calls.CollectFinVoiceCallByID(ctx, 99)
				if err != nil {
					return err
				}
				_ = call
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/fin_voice/collect/99",
		},
		{
			name:     "collect fin voice call by external ID",
			response: `{"external_call_id":"ext-1","status":"registered"}`,
			call: func(ctx context.Context, c *Client) error {
				call, err := c.Calls.CollectFinVoiceCallByExternalID(ctx, "ext-1")
				if err != nil {
					return err
				}
				_ = call
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/fin_voice/external_id/ext-1",
		},
		{
			name:     "collect fin voice call by phone number",
			response: `{"external_call_id":"ext-1","status":"registered"}`,
			call: func(ctx context.Context, c *Client) error {
				call, err := c.Calls.CollectFinVoiceCallByPhoneNumber(ctx, "+15550001234")
				if err != nil {
					return err
				}
				_ = call
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/fin_voice/phone_number/+15550001234",
		},
		{
			name:     "collect fin voice calls by conversation ID",
			response: `[{"external_call_id":"ext-1","status":"registered"}]`,
			call: func(ctx context.Context, c *Client) error {
				calls, err := c.Calls.CollectFinVoiceCallsByConversationID(ctx, "conv-1")
				if err != nil {
					return err
				}
				if len(calls) != 1 {
					t.Fatalf("len(calls) = %d, want 1", len(calls))
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/fin_voice/conversation/conv-1",
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
			client := newCallsTestClient(t, transport)
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

	t.Run("get recording", func(t *testing.T) {
		var gotMethod, gotPath string
		transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
			gotMethod = req.Method
			gotPath = req.URL.Path
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"audio/mpeg"}},
				Body:       io.NopCloser(strings.NewReader("audio-data")),
				Request:    req,
			}, nil
		})
		client := newCallsTestClient(t, transport)
		data, err := client.Calls.GetRecording(context.Background(), "c1")
		if err != nil {
			t.Fatalf("GetRecording returned error: %v", err)
		}
		if string(data) != "audio-data" {
			t.Fatalf("data = %q", string(data))
		}
		if gotMethod != http.MethodGet {
			t.Fatalf("method = %q, want GET", gotMethod)
		}
		if gotPath != "/calls/c1/recording" {
			t.Fatalf("path = %q, want /calls/c1/recording", gotPath)
		}
	})

	t.Run("get transcript", func(t *testing.T) {
		var gotMethod, gotPath string
		transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
			gotMethod = req.Method
			gotPath = req.URL.Path
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"text/plain"}},
				Body:       io.NopCloser(strings.NewReader("transcript-data")),
				Request:    req,
			}, nil
		})
		client := newCallsTestClient(t, transport)
		data, err := client.Calls.GetTranscript(context.Background(), "c1")
		if err != nil {
			t.Fatalf("GetTranscript returned error: %v", err)
		}
		if string(data) != "transcript-data" {
			t.Fatalf("data = %q", string(data))
		}
		if gotMethod != http.MethodGet {
			t.Fatalf("method = %q, want GET", gotMethod)
		}
		if gotPath != "/calls/c1/transcript" {
			t.Fatalf("path = %q, want /calls/c1/transcript", gotPath)
		}
	})

	t.Run("api errors", func(t *testing.T) {
		calls := []struct {
			name string
			call func(context.Context, *Client) error
		}{
			{"list", func(ctx context.Context, c *Client) error { _, err := c.Calls.List(ctx); return err }},
			{"get", func(ctx context.Context, c *Client) error { _, err := c.Calls.Get(ctx, "c1"); return err }},
			{"get recording", func(ctx context.Context, c *Client) error {
				_, err := c.Calls.GetRecording(ctx, "c1")
				return err
			}},
			{"get transcript", func(ctx context.Context, c *Client) error {
				_, err := c.Calls.GetTranscript(ctx, "c1")
				return err
			}},
			{"list with transcripts", func(ctx context.Context, c *Client) error {
				_, err := c.Calls.ListWithTranscripts(ctx, []string{"conv1"})
				return err
			}},
			{"register fin voice call", func(ctx context.Context, c *Client) error {
				_, err := c.Calls.RegisterFinVoiceCall(ctx, RegisterFinVoiceCallRequest{CallId: "ext-1", PhoneNumber: "+1"})
				return err
			}},
			{"collect fin voice call by ID", func(ctx context.Context, c *Client) error {
				_, err := c.Calls.CollectFinVoiceCallByID(ctx, 1)
				return err
			}},
			{"collect fin voice call by external ID", func(ctx context.Context, c *Client) error {
				_, err := c.Calls.CollectFinVoiceCallByExternalID(ctx, "ext-1")
				return err
			}},
			{"collect fin voice call by phone number", func(ctx context.Context, c *Client) error {
				_, err := c.Calls.CollectFinVoiceCallByPhoneNumber(ctx, "+1")
				return err
			}},
			{"collect fin voice calls by conversation ID", func(ctx context.Context, c *Client) error {
				_, err := c.Calls.CollectFinVoiceCallsByConversationID(ctx, "conv-1")
				return err
			}},
		}
		for _, tt := range calls {
			t.Run(tt.name, func(t *testing.T) {
				client := newCallsTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
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
			{"list", func(ctx context.Context, c *Client) error { _, err := c.Calls.List(ctx); return err }},
			{"get", func(ctx context.Context, c *Client) error { _, err := c.Calls.Get(ctx, "c1"); return err }},
			{"get recording", func(ctx context.Context, c *Client) error {
				_, err := c.Calls.GetRecording(ctx, "c1")
				return err
			}},
			{"get transcript", func(ctx context.Context, c *Client) error {
				_, err := c.Calls.GetTranscript(ctx, "c1")
				return err
			}},
			{"list with transcripts", func(ctx context.Context, c *Client) error {
				_, err := c.Calls.ListWithTranscripts(ctx, []string{"conv1"})
				return err
			}},
			{"collect by phone number", func(ctx context.Context, c *Client) error {
				_, err := c.Calls.CollectFinVoiceCallByPhoneNumber(ctx, "+1")
				return err
			}},
			{"collect by conversation ID", func(ctx context.Context, c *Client) error {
				_, err := c.Calls.CollectFinVoiceCallsByConversationID(ctx, "conv-1")
				return err
			}},
		}
		for _, tt := range calls {
			t.Run(tt.name, func(t *testing.T) {
				client := newCallsTestClient(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
					return nil, transportErr
				}))
				if err := tt.call(context.Background(), client); err == nil {
					t.Fatal("expected transport error")
				}
			})
		}
	})

	// The generated client only parses JSON content types. Non-JSON responses bypass
	// schema unmarshaling and hit our status-check and body-parse branches directly.
	t.Run("non-json error status", func(t *testing.T) {
		// status != 200, non-JSON content: exercises the parseErrorResponse branch
		calls := []struct {
			name string
			call func(context.Context, *Client) error
		}{
			{"get recording", func(ctx context.Context, c *Client) error {
				_, err := c.Calls.GetRecording(ctx, "c1")
				return err
			}},
			{"list with transcripts", func(ctx context.Context, c *Client) error {
				_, err := c.Calls.ListWithTranscripts(ctx, []string{"conv1"})
				return err
			}},
			{"collect by phone number", func(ctx context.Context, c *Client) error {
				_, err := c.Calls.CollectFinVoiceCallByPhoneNumber(ctx, "+1")
				return err
			}},
			{"collect by conversation ID", func(ctx context.Context, c *Client) error {
				_, err := c.Calls.CollectFinVoiceCallsByConversationID(ctx, "conv-1")
				return err
			}},
		}
		for _, tt := range calls {
			t.Run(tt.name, func(t *testing.T) {
				client := newCallsTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusInternalServerError,
						Header:     http.Header{"Content-Type": []string{"text/plain"}},
						Body:       io.NopCloser(strings.NewReader("internal error")),
						Request:    req,
					}, nil
				}))
				if err := tt.call(context.Background(), client); err == nil {
					t.Fatal("expected error")
				}
			})
		}
	})

	t.Run("non-json body unmarshal error", func(t *testing.T) {
		// status 200, non-JSON content: exercises the json.Unmarshal error branch
		calls := []struct {
			name string
			call func(context.Context, *Client) error
		}{
			{"list with transcripts", func(ctx context.Context, c *Client) error {
				_, err := c.Calls.ListWithTranscripts(ctx, []string{"conv1"})
				return err
			}},
			{"collect by phone number", func(ctx context.Context, c *Client) error {
				_, err := c.Calls.CollectFinVoiceCallByPhoneNumber(ctx, "+1")
				return err
			}},
		}
		for _, tt := range calls {
			t.Run(tt.name, func(t *testing.T) {
				client := newCallsTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{"Content-Type": []string{"text/plain"}},
						Body:       io.NopCloser(strings.NewReader("not-json{")),
						Request:    req,
					}, nil
				}))
				if err := tt.call(context.Background(), client); err == nil {
					t.Fatal("expected unmarshal error")
				}
			})
		}
	})
}

func TestCallsValidation(t *testing.T) {
	client := newCallsTestClient(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("unexpected HTTP request")
		return nil, nil
	}))
	ctx := context.Background()

	tests := []struct {
		name string
		call func() error
	}{
		{"get: empty ID", func() error { _, err := client.Calls.Get(ctx, ""); return err }},
		{"get recording: empty ID", func() error { _, err := client.Calls.GetRecording(ctx, ""); return err }},
		{"get transcript: empty ID", func() error { _, err := client.Calls.GetTranscript(ctx, ""); return err }},
		{"list with transcripts: empty IDs", func() error {
			_, err := client.Calls.ListWithTranscripts(ctx, nil)
			return err
		}},
		{"list with transcripts: too many IDs", func() error {
			ids := make([]string, 21)
			_, err := client.Calls.ListWithTranscripts(ctx, ids)
			return err
		}},
		{"collect by external ID: empty", func() error {
			_, err := client.Calls.CollectFinVoiceCallByExternalID(ctx, "")
			return err
		}},
		{"collect by phone: empty", func() error {
			_, err := client.Calls.CollectFinVoiceCallByPhoneNumber(ctx, "")
			return err
		}},
		{"collect by conversation ID: empty", func() error {
			_, err := client.Calls.CollectFinVoiceCallsByConversationID(ctx, "")
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
