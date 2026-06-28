package intercom

import (
	"context"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// FinReply is a request to continue a Fin conversation.
type FinReply = gen.ReplyToFinJSONBody

// FinStartConversation is a request to start a Fin conversation.
type FinStartConversation = gen.StartFinConversationJSONBody

// FinConversationResponse is the response from Fin conversation APIs.
type FinConversationResponse struct {
	ConversationID     *string                            `json:"conversation_id,omitempty"`
	CreatedAtMs        *string                            `json:"created_at_ms,omitempty"`
	Errors             *gen.FinAgentAttributeErrorsSchema `json:"errors,omitempty"`
	SSESubscriptionURL *string                            `json:"sse_subscription_url,omitempty"`
	Status             *string                            `json:"status,omitempty"`
	UserID             *string                            `json:"user_id,omitempty"`
}

// FinService exposes Fin-related Intercom API operations.
type FinService struct {
	client *Client
}

// Reply continues a Fin conversation.
func (s *FinService) Reply(ctx context.Context, req FinReply) (*FinConversationResponse, error) {
	res, err := s.client.generated.ReplyToFinWithResponse(ctx, nil, gen.ReplyToFinJSONRequestBody(req))
	if err != nil {
		return nil, err
	}
	return requireJSON[FinConversationResponse]("reply to fin", res.StatusCode(), res.Body)
}

// StartConversation starts a Fin conversation.
func (s *FinService) StartConversation(ctx context.Context, req FinStartConversation) (*FinConversationResponse, error) {
	res, err := s.client.generated.StartFinConversationWithResponse(ctx, nil, gen.StartFinConversationJSONRequestBody(req))
	if err != nil {
		return nil, err
	}
	return requireJSON[FinConversationResponse]("start fin conversation", res.StatusCode(), res.Body)
}

// GetVoiceCallByID retrieves a Fin voice call by ID.
func (s *FinService) GetVoiceCallByID(ctx context.Context, id string) (*FinVoiceCall, error) {
	n, err := requireIntID("voice call", id)
	if err != nil {
		return nil, err
	}
	return s.client.Calls.CollectFinVoiceCallByID(ctx, n)
}

// ListVoiceCallsByConversation returns Fin voice calls for a conversation.
func (s *FinService) ListVoiceCallsByConversation(ctx context.Context, conversationID string) ([]FinVoiceCall, error) {
	if conversationID == "" {
		return nil, errRequiredID("conversation")
	}
	return s.client.Calls.CollectFinVoiceCallsByConversationID(ctx, conversationID)
}

// GetVoiceCallByExternalID retrieves a Fin voice call by external ID.
func (s *FinService) GetVoiceCallByExternalID(ctx context.Context, externalID string) (*FinVoiceCall, error) {
	if externalID == "" {
		return nil, errRequiredID("external")
	}
	return s.client.Calls.CollectFinVoiceCallByExternalID(ctx, externalID)
}

// GetVoiceCallByPhoneNumber retrieves a Fin voice call lookup by phone number.
func (s *FinService) GetVoiceCallByPhoneNumber(ctx context.Context, phoneNumber string) (*FinVoiceCall, error) {
	if phoneNumber == "" {
		return nil, errRequiredID("phone number")
	}
	return s.client.Calls.CollectFinVoiceCallByPhoneNumber(ctx, phoneNumber)
}

// RegisterVoiceCall registers a Fin voice call.
func (s *FinService) RegisterVoiceCall(ctx context.Context, req RegisterFinVoiceCallRequest) (*FinVoiceCall, error) {
	return s.client.Calls.RegisterFinVoiceCall(ctx, req)
}
