package intercom

import (
	"context"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// WhatsAppService exposes WhatsApp message delivery-status lookups.
type WhatsAppService struct{ client *Client }

// GetMessageStatus gets a WhatsApp message status.
func (s *WhatsAppService) GetMessageStatus(ctx context.Context, params *gen.GetWhatsAppMessageStatusParams) (*gen.GetWhatsAppMessageStatusResponse, error) {
	return s.client.generated.GetWhatsAppMessageStatusWithResponse(ctx, params)
}

// RetrieveMessageStatus retrieves a WhatsApp message status by reference.
func (s *WhatsAppService) RetrieveMessageStatus(ctx context.Context, params *gen.RetrieveWhatsAppMessageStatusParams) (*gen.RetrieveWhatsAppMessageStatusResponse, error) {
	return s.client.generated.RetrieveWhatsAppMessageStatusWithResponse(ctx, params)
}
