package intercom

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
	"strings"
)

const (
	// WebhookSignatureHeader is the documented Intercom webhook signature header.
	WebhookSignatureHeader = "X-Hub-Signature"
	webhookSignaturePrefix = "sha1="

	// FinAgentWebhookSignatureHeader is the documented Fin Agent API webhook signature header.
	FinAgentWebhookSignatureHeader = "X-Fin-Agent-API-Webhook-Signature"
	finAgentWebhookSignaturePrefix = "sha256="
)

var (
	// ErrWebhookSignatureSecretRequired is returned when signature verification is attempted without a secret.
	ErrWebhookSignatureSecretRequired = errors.New("intercom: webhook signature secret is required")
	// ErrWebhookSignatureMissing is returned when the expected webhook signature header is absent.
	ErrWebhookSignatureMissing = errors.New("intercom: webhook signature is missing")
	// ErrWebhookSignatureInvalid is returned when the webhook signature does not match the payload.
	ErrWebhookSignatureInvalid = errors.New("intercom: webhook signature is invalid")
	// ErrWebhookSignatureUnsupported is returned when the signature uses an unsupported scheme.
	ErrWebhookSignatureUnsupported = errors.New("intercom: webhook signature scheme is unsupported")
)

// WebhookEvent is the common Intercom webhook notification envelope.
//
// Topic identifies the event, for example "ticket.created". Data.Item contains
// the raw topic-specific Intercom object so callers can decode only the topics
// they handle while remaining forward-compatible with new topics.
type WebhookEvent struct {
	Type             string           `json:"type"`
	AppID            string           `json:"app_id"`
	Data             WebhookEventData `json:"data"`
	Links            map[string]any   `json:"links,omitempty"`
	ID               string           `json:"id"`
	Topic            string           `json:"topic"`
	DeliveryStatus   string           `json:"delivery_status"`
	DeliveryAttempts int              `json:"delivery_attempts"`
	DeliveredAt      int64            `json:"delivered_at"`
	FirstSentAt      int64            `json:"first_sent_at"`
	CreatedAt        int64            `json:"created_at"`
	Self             any              `json:"self,omitempty"`
	Raw              json.RawMessage  `json:"-"`
}

// WebhookEventData is the common data envelope inside an Intercom webhook notification.
type WebhookEventData struct {
	Type string          `json:"type"`
	Item json.RawMessage `json:"item"`
	Raw  json.RawMessage `json:"-"`
}

// DecodeItem decodes the topic-specific webhook item into v.
func (e *WebhookEvent) DecodeItem(v any) error {
	if e == nil {
		return errors.New("intercom: webhook event is nil")
	}
	return e.Data.DecodeItem(v)
}

// DecodeItem decodes the topic-specific webhook item into v.
func (d WebhookEventData) DecodeItem(v any) error {
	if len(d.Item) == 0 {
		return errors.New("intercom: webhook event data item is empty")
	}
	if err := json.Unmarshal(d.Item, v); err != nil {
		return fmt.Errorf("intercom: decode webhook event item: %w", err)
	}
	return nil
}

// ParseWebhook reads and parses an Intercom webhook notification.
//
// If you need to verify a webhook signature, read the request body once, call
// VerifyWebhookSignature with the raw bytes, and then call ParseWebhookPayload
// with the same bytes. Do not verify a re-encoded JSON body.
func ParseWebhook(r io.Reader) (*WebhookEvent, error) {
	if r == nil {
		return nil, errors.New("intercom: webhook body is nil")
	}
	payload, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("intercom: read webhook body: %w", err)
	}
	return ParseWebhookPayload(payload)
}

// ParseWebhookPayload parses an Intercom webhook notification from raw bytes.
func ParseWebhookPayload(payload []byte) (*WebhookEvent, error) {
	var raw webhookEventRaw
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("intercom: parse webhook payload: %w", err)
	}

	event := &WebhookEvent{
		Type:             raw.Type,
		AppID:            raw.AppID,
		Links:            raw.Links,
		ID:               raw.ID,
		Topic:            raw.Topic,
		DeliveryStatus:   raw.DeliveryStatus,
		DeliveryAttempts: raw.DeliveryAttempts,
		DeliveredAt:      raw.DeliveredAt,
		FirstSentAt:      raw.FirstSentAt,
		CreatedAt:        raw.CreatedAt,
		Self:             raw.Self,
		Raw:              append(json.RawMessage(nil), payload...),
	}

	if len(raw.Data) > 0 {
		if err := json.Unmarshal(raw.Data, &event.Data); err != nil {
			return nil, fmt.Errorf("intercom: parse webhook data: %w", err)
		}
		event.Data.Raw = append(json.RawMessage(nil), raw.Data...)
	}

	return event, nil
}

// VerifyWebhookSignature verifies the documented Intercom webhook signature.
//
// Intercom signs webhook notifications with X-Hub-Signature using a
// "sha1=<hex>" HMAC-SHA1 over the raw request body and the app client secret.
func VerifyWebhookSignature(clientSecret string, payload []byte, header http.Header) error {
	if header == nil {
		return verifyWebhookSignature(clientSecret, payload, "")
	}
	return verifyWebhookSignature(clientSecret, payload, header.Get(WebhookSignatureHeader))
}

// VerifyFinAgentWebhookSignature verifies a Fin Agent API webhook signature header value.
func VerifyFinAgentWebhookSignature(secret string, payload []byte, signature string) error {
	return verifyHMACSignature(secret, payload, signature, finAgentWebhookSignaturePrefix, sha256.New)
}

// WebhookSignature returns the Intercom webhook signature header value for payload.
//
// This is primarily useful in tests and local fixtures.
func WebhookSignature(clientSecret string, payload []byte) string {
	mac := hmac.New(sha1.New, []byte(clientSecret))
	_, _ = mac.Write(payload)
	return webhookSignaturePrefix + hex.EncodeToString(mac.Sum(nil))
}

// FinAgentWebhookSignature returns the Fin Agent API signature header value for payload.
//
// This is primarily useful in tests and local fixtures.
func FinAgentWebhookSignature(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	return finAgentWebhookSignaturePrefix + hex.EncodeToString(mac.Sum(nil))
}

func verifyWebhookSignature(clientSecret string, payload []byte, signature string) error {
	return verifyHMACSignature(clientSecret, payload, signature, webhookSignaturePrefix, sha1.New)
}

func verifyHMACSignature(secret string, payload []byte, signature string, prefix string, newHash func() hash.Hash) error {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return ErrWebhookSignatureSecretRequired
	}

	signature = strings.TrimSpace(signature)
	if signature == "" {
		return ErrWebhookSignatureMissing
	}
	if !strings.HasPrefix(signature, prefix) {
		return ErrWebhookSignatureUnsupported
	}

	got, err := hex.DecodeString(strings.TrimPrefix(signature, prefix))
	if err != nil {
		return ErrWebhookSignatureInvalid
	}

	mac := hmac.New(newHash, []byte(secret))
	_, _ = mac.Write(payload)
	if !hmac.Equal(got, mac.Sum(nil)) {
		return ErrWebhookSignatureInvalid
	}

	return nil
}

type webhookEventRaw struct {
	Type             string          `json:"type"`
	AppID            string          `json:"app_id"`
	Data             json.RawMessage `json:"data"`
	Links            map[string]any  `json:"links"`
	ID               string          `json:"id"`
	Topic            string          `json:"topic"`
	DeliveryStatus   string          `json:"delivery_status"`
	DeliveryAttempts int             `json:"delivery_attempts"`
	DeliveredAt      int64           `json:"delivered_at"`
	FirstSentAt      int64           `json:"first_sent_at"`
	CreatedAt        int64           `json:"created_at"`
	Self             any             `json:"self"`
}
