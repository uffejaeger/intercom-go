package intercom

import (
	"context"
	"fmt"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// WhatsAppMessageStatusList is a list of WhatsApp message statuses.
type WhatsAppMessageStatusList = gen.WhatsappMessageStatusListSchema

// WhatsAppMessageStatus is a WhatsApp message status.
type WhatsAppMessageStatus = gen.WhatsappMessageStatusSchema

// WhatsAppMessageStatusParams configures a WhatsApp message-status list request.
type WhatsAppMessageStatusParams = gen.GetWhatsAppMessageStatusParams

// WhatsAppMessageStatusRetrieveParams configures a WhatsApp message-status retrieval request.
type WhatsAppMessageStatusRetrieveParams = gen.RetrieveWhatsAppMessageStatusParams

// WhatsAppService exposes WhatsApp message delivery-status lookups.
type WhatsAppService struct{ client *Client }

// GetMessageStatus gets a WhatsApp message status.
func (s *WhatsAppService) GetMessageStatus(ctx context.Context, params *WhatsAppMessageStatusParams) (*WhatsAppMessageStatusList, error) {
	if params == nil || params.RulesetId == "" {
		return nil, fmt.Errorf("intercom: WhatsApp ruleset ID is required")
	}
	res, err := s.client.generated.GetWhatsAppMessageStatusWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	return requireOK("get WhatsApp message status", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

// RetrieveMessageStatus retrieves a WhatsApp message status by reference.
func (s *WhatsAppService) RetrieveMessageStatus(ctx context.Context, params *WhatsAppMessageStatusRetrieveParams) (*WhatsAppMessageStatus, error) {
	if params == nil || params.MessageId == "" {
		return nil, fmt.Errorf("intercom: WhatsApp message ID is required")
	}
	res, err := s.client.generated.RetrieveWhatsAppMessageStatusWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	return requireOK("retrieve WhatsApp message status", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}
