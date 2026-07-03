package intercom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// Conversation is an Intercom conversation.
type Conversation = gen.ConversationSchema

// ConversationList is a page of Intercom conversations.
type ConversationList = gen.ConversationListSchema

// ConversationDeleted is the result of deleting a conversation.
type ConversationDeleted = gen.ConversationDeletedSchema

// ConversationMessage is the result of creating a conversation.
type ConversationMessage = gen.MessageSchema

// ConversationHandlingEvent is a pause or resume event recorded for a conversation.
type ConversationHandlingEvent = gen.HandlingEventSchema

// ConversationHandlingEventList is a list of handling events for a conversation.
type ConversationHandlingEventList = gen.HandlingEventListSchema

// Ticket is an Intercom ticket.
type Ticket = gen.TicketSchema

// ConversationCreate holds the fields for creating a conversation.
type ConversationCreate = gen.CreateConversationRequestSchema

// ConversationUpdate holds the fields for updating a conversation.
type ConversationUpdate = gen.UpdateConversationRequestSchema

// ConversationAssign holds the fields for assigning a conversation.
type ConversationAssign = gen.AssignConversationRequestSchema

// ConversationClose holds the fields for closing a conversation.
type ConversationClose = gen.CloseConversationRequestSchema

// ConversationOpen holds the fields for re-opening a conversation.
type ConversationOpen = gen.OpenConversationRequestSchema

// ConversationSnooze holds the fields for snoozing a conversation.
type ConversationSnooze = gen.SnoozeConversationRequestSchema

// ConversationAdminReply holds the fields for replying to a conversation as an admin.
type ConversationAdminReply = gen.AdminReplyConversationRequestSchema

// ConversationContactReply holds the fields for replying to a conversation as a contact.
type ConversationContactReply = gen.ContactReplyBaseRequestSchema

// ConversationAttachContact holds the fields for attaching a contact to a conversation.
type ConversationAttachContact = gen.AttachContactToConversationRequestSchema

// ConversationDetachContact holds the fields for detaching a contact from a conversation.
type ConversationDetachContact = gen.DetachContactFromConversationRequest

// ConversationRedactPart holds the fields for redacting a conversation part.
type ConversationRedactPart = gen.RedactConversationRequest0

// ConversationRedactSource holds the fields for redacting a conversation source.
type ConversationRedactSource = gen.RedactConversationRequest1

// ConversationToTicket holds the fields for converting a conversation to a ticket.
type ConversationToTicket = gen.ConvertConversationToTicketRequestSchema

// ConversationSearchQuery holds the query for searching conversations.
type ConversationSearchQuery = gen.SearchRequestSchema

// ConversationsService exposes conversation-related Intercom API operations.
type ConversationsService struct {
	client *Client
}

// List returns all conversations.
func (s *ConversationsService) List(ctx context.Context) (*ConversationList, error) {
	res, err := s.client.generated.ListConversationsWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("list conversations", res.StatusCode(), res.Body, res.JSON200)
}

// Create creates a new conversation initiated from a contact.
func (s *ConversationsService) Create(ctx context.Context, conversation ConversationCreate) (*ConversationMessage, error) {
	res, err := s.client.generated.CreateConversationWithResponse(ctx, nil, gen.CreateConversationJSONRequestBody(conversation))
	if err != nil {
		return nil, err
	}
	return requireOK("create conversation", res.StatusCode(), res.Body, res.JSON200)
}

// Get retrieves a conversation by Intercom conversation ID.
func (s *ConversationsService) Get(ctx context.Context, conversationID string) (*Conversation, error) {
	if conversationID == "" {
		return nil, fmt.Errorf("intercom: conversation ID is required")
	}
	id, err := conversationIDToInt(conversationID)
	if err != nil {
		return nil, err
	}
	res, err := s.client.generated.RetrieveConversationWithResponse(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("get conversation", res.StatusCode(), res.Body, res.JSON200)
}

// ListHandlingEvents returns pause and resume events for a conversation.
func (s *ConversationsService) ListHandlingEvents(ctx context.Context, conversationID string) (*ConversationHandlingEventList, error) {
	if conversationID == "" {
		return nil, fmt.Errorf("intercom: conversation ID is required")
	}
	res, err := s.client.generated.ListHandlingEventsWithResponse(ctx, conversationID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("list conversation handling events", res.StatusCode(), res.Body, res.JSON200)
}

// Update updates an existing conversation.
func (s *ConversationsService) Update(ctx context.Context, conversationID string, conversation ConversationUpdate) (*Conversation, error) {
	if conversationID == "" {
		return nil, fmt.Errorf("intercom: conversation ID is required")
	}
	id, err := conversationIDToInt(conversationID)
	if err != nil {
		return nil, err
	}
	res, err := s.client.generated.UpdateConversationWithResponse(ctx, id, nil, gen.UpdateConversationJSONRequestBody(conversation))
	if err != nil {
		return nil, err
	}
	return requireOK("update conversation", res.StatusCode(), res.Body, res.JSON200)
}

// Delete deletes a conversation.
func (s *ConversationsService) Delete(ctx context.Context, conversationID string) (*ConversationDeleted, error) {
	if conversationID == "" {
		return nil, fmt.Errorf("intercom: conversation ID is required")
	}
	id, err := conversationIDToInt(conversationID)
	if err != nil {
		return nil, err
	}
	res, err := s.client.generated.DeleteConversationWithResponse(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("delete conversation", res.StatusCode(), res.Body, res.JSON200)
}

// Search searches conversations using an Intercom search query.
func (s *ConversationsService) Search(ctx context.Context, query ConversationSearchQuery) (*ConversationList, error) {
	res, err := s.client.generated.SearchConversationsWithResponse(ctx, nil, gen.SearchConversationsJSONRequestBody(query))
	if err != nil {
		return nil, err
	}
	return requireOK("search conversations", res.StatusCode(), res.Body, res.JSON200)
}

// Reply sends an admin reply to a conversation.
func (s *ConversationsService) Reply(ctx context.Context, conversationID string, reply ConversationAdminReply) (*Conversation, error) {
	if conversationID == "" {
		return nil, fmt.Errorf("intercom: conversation ID is required")
	}
	b, _ := json.Marshal(reply)
	res, err := s.client.generated.ReplyConversationWithBodyWithResponse(ctx, conversationID, nil, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	return requireOK("reply to conversation", res.StatusCode(), res.Body, res.JSON200)
}

// ReplyAsContact sends a contact reply to a conversation.
func (s *ConversationsService) ReplyAsContact(ctx context.Context, conversationID string, reply ConversationContactReply) (*Conversation, error) {
	if conversationID == "" {
		return nil, fmt.Errorf("intercom: conversation ID is required")
	}
	b, _ := json.Marshal(reply)
	res, err := s.client.generated.ReplyConversationWithBodyWithResponse(ctx, conversationID, nil, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	return requireOK("reply to conversation as contact", res.StatusCode(), res.Body, res.JSON200)
}

// Assign assigns a conversation to an admin or team.
func (s *ConversationsService) Assign(ctx context.Context, conversationID string, assign ConversationAssign) (*Conversation, error) {
	if conversationID == "" {
		return nil, fmt.Errorf("intercom: conversation ID is required")
	}
	b, _ := json.Marshal(assign)
	res, err := s.client.generated.ManageConversationWithBodyWithResponse(ctx, conversationID, nil, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	return requireOK("assign conversation", res.StatusCode(), res.Body, res.JSON200)
}

// Close closes a conversation.
func (s *ConversationsService) Close(ctx context.Context, conversationID string, req ConversationClose) (*Conversation, error) {
	if conversationID == "" {
		return nil, fmt.Errorf("intercom: conversation ID is required")
	}
	b, _ := json.Marshal(req)
	res, err := s.client.generated.ManageConversationWithBodyWithResponse(ctx, conversationID, nil, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	return requireOK("close conversation", res.StatusCode(), res.Body, res.JSON200)
}

// Open re-opens a snoozed or closed conversation.
func (s *ConversationsService) Open(ctx context.Context, conversationID string, req ConversationOpen) (*Conversation, error) {
	if conversationID == "" {
		return nil, fmt.Errorf("intercom: conversation ID is required")
	}
	b, _ := json.Marshal(req)
	res, err := s.client.generated.ManageConversationWithBodyWithResponse(ctx, conversationID, nil, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	return requireOK("open conversation", res.StatusCode(), res.Body, res.JSON200)
}

// Snooze snoozes a conversation until the given time.
func (s *ConversationsService) Snooze(ctx context.Context, conversationID string, req ConversationSnooze) (*Conversation, error) {
	if conversationID == "" {
		return nil, fmt.Errorf("intercom: conversation ID is required")
	}
	b, _ := json.Marshal(req)
	res, err := s.client.generated.ManageConversationWithBodyWithResponse(ctx, conversationID, nil, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	return requireOK("snooze conversation", res.StatusCode(), res.Body, res.JSON200)
}

// AttachContact attaches a contact to a conversation.
func (s *ConversationsService) AttachContact(ctx context.Context, conversationID string, req ConversationAttachContact) (*Conversation, error) {
	if conversationID == "" {
		return nil, fmt.Errorf("intercom: conversation ID is required")
	}
	res, err := s.client.generated.AttachContactToConversationWithResponse(ctx, conversationID, nil, gen.AttachContactToConversationJSONRequestBody(req))
	if err != nil {
		return nil, err
	}
	return requireOK("attach contact to conversation", res.StatusCode(), res.Body, res.JSON200)
}

// DetachContact detaches a contact from a conversation.
func (s *ConversationsService) DetachContact(ctx context.Context, conversationID, contactID string, req ConversationDetachContact) (*Conversation, error) {
	if conversationID == "" {
		return nil, fmt.Errorf("intercom: conversation ID is required")
	}
	if contactID == "" {
		return nil, fmt.Errorf("intercom: contact ID is required")
	}
	res, err := s.client.generated.DetachContactFromConversationWithResponse(ctx, conversationID, contactID, nil, gen.DetachContactFromConversationJSONRequestBody(req))
	if err != nil {
		return nil, err
	}
	return requireOK("detach contact from conversation", res.StatusCode(), res.Body, res.JSON200)
}

// RedactPart redacts a message part from a conversation.
func (s *ConversationsService) RedactPart(ctx context.Context, req ConversationRedactPart) (*Conversation, error) {
	b, _ := json.Marshal(req)
	res, err := s.client.generated.RedactConversationWithBodyWithResponse(ctx, nil, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	return requireOK("redact conversation part", res.StatusCode(), res.Body, res.JSON200)
}

// RedactSource redacts a source message from a conversation.
func (s *ConversationsService) RedactSource(ctx context.Context, req ConversationRedactSource) (*Conversation, error) {
	b, _ := json.Marshal(req)
	res, err := s.client.generated.RedactConversationWithBodyWithResponse(ctx, nil, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	return requireOK("redact conversation source", res.StatusCode(), res.Body, res.JSON200)
}

// ConvertToTicket converts a conversation to a ticket.
func (s *ConversationsService) ConvertToTicket(ctx context.Context, conversationID string, req ConversationToTicket) (*Ticket, error) {
	if conversationID == "" {
		return nil, fmt.Errorf("intercom: conversation ID is required")
	}
	id, err := conversationIDToInt(conversationID)
	if err != nil {
		return nil, err
	}
	res, err := s.client.generated.ConvertConversationToTicketWithResponse(ctx, id, nil, gen.ConvertConversationToTicketJSONRequestBody(req))
	if err != nil {
		return nil, err
	}
	return requireOK("convert conversation to ticket", res.StatusCode(), res.Body, res.JSON200)
}

// AttachTag attaches a tag to a conversation.
func (s *ConversationsService) AttachTag(ctx context.Context, conversationID, tagID, adminID string) (*Tag, error) {
	if conversationID == "" {
		return nil, fmt.Errorf("intercom: conversation ID is required")
	}
	if tagID == "" {
		return nil, fmt.Errorf("intercom: tag ID is required")
	}
	if adminID == "" {
		return nil, fmt.Errorf("intercom: admin ID is required")
	}
	res, err := s.client.generated.AttachTagToConversationWithResponse(ctx, conversationID, nil, gen.AttachTagToConversationJSONRequestBody{
		Id:      tagID,
		AdminId: adminID,
	})
	if err != nil {
		return nil, err
	}
	return requireOK("attach tag to conversation", res.StatusCode(), res.Body, res.JSON200)
}

// DetachTag detaches a tag from a conversation.
func (s *ConversationsService) DetachTag(ctx context.Context, conversationID, tagID, adminID string) (*Tag, error) {
	if conversationID == "" {
		return nil, fmt.Errorf("intercom: conversation ID is required")
	}
	if tagID == "" {
		return nil, fmt.Errorf("intercom: tag ID is required")
	}
	if adminID == "" {
		return nil, fmt.Errorf("intercom: admin ID is required")
	}
	res, err := s.client.generated.DetachTagFromConversationWithResponse(ctx, conversationID, tagID, nil, gen.DetachTagFromConversationJSONRequestBody{
		AdminId: adminID,
	})
	if err != nil {
		return nil, err
	}
	return requireOK("detach tag from conversation", res.StatusCode(), res.Body, res.JSON200)
}

func conversationIDToInt(conversationID string) (int, error) {
	id, err := strconv.Atoi(conversationID)
	if err != nil {
		return 0, fmt.Errorf("intercom: conversation ID %q is not a valid integer: %w", conversationID, err)
	}
	return id, nil
}
