package intercom

import (
	"context"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// FinAttachment is an attachment included with a Fin request.
type FinAttachment = gen.FinAgentAttachmentSchema

// FinAttachmentType identifies the type of attachment included with a Fin request.
type FinAttachmentType = gen.FinAgentAttachmentType

const (
	FinAttachmentTypeFile FinAttachmentType = gen.File
	FinAttachmentTypeURL  FinAttachmentType = gen.Url
)

// FinConversationMetadata provides extra context for a Fin conversation.
type FinConversationMetadata = gen.FinAgentConversationMetadataSchema

// FinMessage is a message exchanged within a Fin conversation.
type FinMessage = gen.FinAgentMessageSchema

// FinMessageAuthor identifies who sent a Fin conversation message.
type FinMessageAuthor = gen.FinAgentMessageAuthor

const (
	FinMessageAuthorAgent FinMessageAuthor = gen.FinAgentMessageAuthorAgent
	FinMessageAuthorFin   FinMessageAuthor = gen.FinAgentMessageAuthorFin
	FinMessageAuthorUser  FinMessageAuthor = gen.FinAgentMessageAuthorUser
)

// FinUser identifies the user participating in a Fin conversation.
type FinUser = gen.FinAgentUserSchema

// FinReply is a request to continue a Fin conversation.
type FinReply struct {
	Attachments    *[]FinAttachment `json:"attachments,omitempty"`
	ConversationId string           `json:"conversation_id"`
	Message        FinMessage       `json:"message"`
	User           FinUser          `json:"user"`
}

func (r FinReply) toGenerated() gen.ReplyToFinJSONBody {
	return gen.ReplyToFinJSONBody{
		Attachments:           r.Attachments,
		ConversationId:        r.ConversationId,
		FinAgentMessageSchema: gen.FinAgentMessageSchema(r.Message),
		FinAgentUserSchema:    gen.FinAgentUserSchema(r.User),
	}
}

// FinStartConversation is a request to start a Fin conversation.
type FinStartConversation struct {
	Attachments          *[]FinAttachment         `json:"attachments,omitempty"`
	ConversationId       string                   `json:"conversation_id"`
	ConversationMetadata *FinConversationMetadata `json:"conversation_metadata,omitempty"`
	Message              FinMessage               `json:"message"`
	User                 FinUser                  `json:"user"`
}

func (r FinStartConversation) toGenerated() gen.StartFinConversationJSONBody {
	return gen.StartFinConversationJSONBody{
		Attachments:                        r.Attachments,
		ConversationId:                     r.ConversationId,
		FinAgentConversationMetadataSchema: r.ConversationMetadata,
		FinAgentMessageSchema:              gen.FinAgentMessageSchema(r.Message),
		FinAgentUserSchema:                 gen.FinAgentUserSchema(r.User),
	}
}

// FinConversationResponse is the response from Fin conversation APIs.
type FinConversationResponse struct {
	ConversationID     *string                `json:"conversation_id,omitempty"`
	CreatedAtMs        *string                `json:"created_at_ms,omitempty"`
	Errors             *FinConversationErrors `json:"errors,omitempty"`
	SSESubscriptionURL *string                `json:"sse_subscription_url,omitempty"`
	Status             *string                `json:"status,omitempty"`
	UserID             *string                `json:"user_id,omitempty"`
}

// FinConversationAttributeErrors groups validation errors for a Fin object.
type FinConversationAttributeErrors struct {
	Attributes *map[string]string `json:"attributes,omitempty"`
}

// FinConversationErrors groups validation errors returned by Fin conversation endpoints.
type FinConversationErrors struct {
	Conversation *FinConversationAttributeErrors `json:"conversation,omitempty"`
	User         *FinConversationAttributeErrors `json:"user,omitempty"`
}

// FinService exposes Fin-related Intercom API operations.
type FinService struct {
	client *Client
}

// Reply continues a Fin conversation.
func (s *FinService) Reply(ctx context.Context, req FinReply) (*FinConversationResponse, error) {
	res, err := s.client.generated.ReplyToFinWithResponse(ctx, nil, gen.ReplyToFinJSONRequestBody(req.toGenerated()))
	if err != nil {
		return nil, err
	}
	return requireJSON[FinConversationResponse]("reply to fin", res.StatusCode(), res.Body)
}

// StartConversation starts a Fin conversation.
func (s *FinService) StartConversation(ctx context.Context, req FinStartConversation) (*FinConversationResponse, error) {
	res, err := s.client.generated.StartFinConversationWithResponse(ctx, nil, gen.StartFinConversationJSONRequestBody(req.toGenerated()))
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
