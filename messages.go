package intercom

import (
	"context"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// Message is an Intercom outbound message.
type Message = gen.MessageSchema

// MessageCreate holds the fields for creating a message.
type MessageCreate = gen.CreateMessageJSONRequestBody

// MessagesService exposes message-related Intercom API operations.
type MessagesService struct {
	client *Client
}

// Create creates a message.
func (s *MessagesService) Create(ctx context.Context, message MessageCreate) (*Message, error) {
	res, err := s.client.generated.CreateMessageWithResponse(ctx, nil, message)
	if err != nil {
		return nil, err
	}
	return requireOK("create message", res.StatusCode(), res.Body, res.JSON200)
}
