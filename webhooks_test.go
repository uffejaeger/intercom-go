package intercom

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestParseWebhookPayload(t *testing.T) {
	payload := []byte(`{
		"type": "notification_event",
		"app_id": "app-1",
		"data": {
			"type": "notification_event_data",
			"item": {
				"type": "ticket",
				"id": "5",
				"ticket_id": "1",
				"unknown_future_field": {"kept": true}
			}
		},
		"links": {},
		"id": "notif-1",
		"topic": "ticket.created",
		"delivery_status": "pending",
		"delivery_attempts": 1,
		"delivered_at": 0,
		"first_sent_at": 1715937891,
		"created_at": 1715937890,
		"self": null
	}`)

	event, err := ParseWebhookPayload(payload)
	if err != nil {
		t.Fatalf("ParseWebhookPayload returned error: %v", err)
	}
	if event.Type != "notification_event" {
		t.Fatalf("Type = %q, want notification_event", event.Type)
	}
	if event.AppID != "app-1" {
		t.Fatalf("AppID = %q, want app-1", event.AppID)
	}
	if event.ID != "notif-1" {
		t.Fatalf("ID = %q, want notif-1", event.ID)
	}
	if event.Topic != "ticket.created" {
		t.Fatalf("Topic = %q, want ticket.created", event.Topic)
	}
	if event.DeliveryAttempts != 1 {
		t.Fatalf("DeliveryAttempts = %d, want 1", event.DeliveryAttempts)
	}
	if event.CreatedAt != 1715937890 {
		t.Fatalf("CreatedAt = %d, want 1715937890", event.CreatedAt)
	}
	if len(event.Raw) == 0 {
		t.Fatal("Raw is empty")
	}
	if event.Data.Type != "notification_event_data" {
		t.Fatalf("Data.Type = %q, want notification_event_data", event.Data.Type)
	}
	if len(event.Data.Raw) == 0 {
		t.Fatal("Data.Raw is empty")
	}
	if !strings.Contains(string(event.Data.Item), `"unknown_future_field"`) {
		t.Fatalf("Data.Item = %s, want raw unknown field preserved", string(event.Data.Item))
	}

	var item struct {
		Type     string `json:"type"`
		ID       string `json:"id"`
		TicketID string `json:"ticket_id"`
	}
	if err := event.DecodeItem(&item); err != nil {
		t.Fatalf("DecodeItem returned error: %v", err)
	}
	if item.Type != "ticket" || item.ID != "5" || item.TicketID != "1" {
		t.Fatalf("decoded item = %#v", item)
	}
}

func TestParseWebhook(t *testing.T) {
	event, err := ParseWebhook(strings.NewReader(`{"type":"notification_event","topic":"contact.created"}`))
	if err != nil {
		t.Fatalf("ParseWebhook returned error: %v", err)
	}
	if event.Topic != "contact.created" {
		t.Fatalf("Topic = %q, want contact.created", event.Topic)
	}
}

func TestParseWebhookErrors(t *testing.T) {
	tests := []struct {
		name    string
		reader  *strings.Reader
		payload []byte
		parse   func() (*WebhookEvent, error)
		want    string
	}{
		{
			name: "nil body",
			parse: func() (*WebhookEvent, error) {
				return ParseWebhook(nil)
			},
			want: "intercom: webhook body is nil",
		},
		{
			name: "invalid payload",
			parse: func() (*WebhookEvent, error) {
				return ParseWebhookPayload([]byte(`not json`))
			},
			want: "intercom: parse webhook payload:",
		},
		{
			name: "read error",
			parse: func() (*WebhookEvent, error) {
				return ParseWebhook(errReader{})
			},
			want: "intercom: read webhook body:",
		},
		{
			name: "invalid data",
			parse: func() (*WebhookEvent, error) {
				return ParseWebhookPayload([]byte(`{"data":"not object"}`))
			},
			want: "intercom: parse webhook data:",
		},
		{
			name: "empty item decode",
			parse: func() (*WebhookEvent, error) {
				event, err := ParseWebhookPayload([]byte(`{"data":{"type":"notification_event_data"}}`))
				if err != nil {
					return nil, err
				}
				return event, event.DecodeItem(&struct{}{})
			},
			want: "intercom: webhook event data item is empty",
		},
		{
			name: "nil event decode",
			parse: func() (*WebhookEvent, error) {
				var event *WebhookEvent
				return nil, event.DecodeItem(&struct{}{})
			},
			want: "intercom: webhook event is nil",
		},
		{
			name: "invalid item decode",
			parse: func() (*WebhookEvent, error) {
				event, err := ParseWebhookPayload([]byte(`{"data":{"type":"notification_event_data","item":"not object"}}`))
				if err != nil {
					return nil, err
				}
				var item struct {
					ID int `json:"id"`
				}
				return event, event.DecodeItem(&item)
			},
			want: "intercom: decode webhook event item:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.parse()
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %q, want to contain %q", err.Error(), tt.want)
			}
		})
	}
}

func TestVerifyWebhookSignature(t *testing.T) {
	payload := []byte(`{"type":"notification_event","topic":"ticket.created"}`)
	secret := "client-secret"
	signature := WebhookSignature(secret, payload)
	header := make(http.Header)
	header.Set(WebhookSignatureHeader, signature)

	if err := VerifyWebhookSignature(secret, payload, header); err != nil {
		t.Fatalf("VerifyWebhookSignature returned error: %v", err)
	}
}

func TestVerifyFinAgentWebhookSignature(t *testing.T) {
	payload := []byte(`{"type":"notification_event","topic":"fin_replied"}`)
	secret := "webhook-secret"
	signature := FinAgentWebhookSignature(secret, payload)

	if err := VerifyFinAgentWebhookSignature(secret, payload, signature); err != nil {
		t.Fatalf("VerifyFinAgentWebhookSignature returned error: %v", err)
	}
}

func TestVerifyWebhookSignatureErrors(t *testing.T) {
	payload := []byte(`{"type":"notification_event"}`)
	secret := "webhook-secret"

	tests := []struct {
		name      string
		secret    string
		payload   []byte
		signature string
		header    http.Header
		want      error
	}{
		{
			name:      "missing secret",
			signature: WebhookSignature(secret, payload),
			want:      ErrWebhookSignatureSecretRequired,
		},
		{
			name:   "missing signature",
			secret: secret,
			want:   ErrWebhookSignatureMissing,
		},
		{
			name:      "unsupported scheme",
			secret:    secret,
			signature: "sha256=abc123",
			want:      ErrWebhookSignatureUnsupported,
		},
		{
			name:      "invalid hex",
			secret:    secret,
			signature: "sha1=not-hex",
			want:      ErrWebhookSignatureInvalid,
		},
		{
			name:      "mismatch",
			secret:    secret,
			payload:   []byte(`{"type":"tampered"}`),
			signature: WebhookSignature(secret, payload),
			want:      ErrWebhookSignatureInvalid,
		},
		{
			name:   "nil header through generic verifier",
			secret: secret,
			header: nil,
			want:   ErrWebhookSignatureMissing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := payload
			if tt.payload != nil {
				payload = tt.payload
			}

			var err error
			if tt.header != nil || tt.name == "nil header through generic verifier" {
				err = VerifyWebhookSignature(tt.secret, payload, tt.header)
			} else {
				err = verifyWebhookSignature(tt.secret, payload, tt.signature)
			}

			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}
