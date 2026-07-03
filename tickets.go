package intercom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// TicketList is a list of Intercom tickets.
type TicketList = gen.TicketListSchema

// TicketContact is a contact selector included in a ticket create request.
type TicketContact = gen.CreateTicketRequest_Contacts_Item

// TicketContactID selects a ticket contact by Intercom contact ID.
type TicketContactID = gen.CreateTicketRequestContacts0

// TicketContactExternalID selects a ticket contact by external ID.
type TicketContactExternalID = gen.CreateTicketRequestContacts1

// TicketContactEmail selects a ticket contact by email address.
type TicketContactEmail = gen.CreateTicketRequestContacts2

// NewTicketContactByID constructs a ticket contact selector from an Intercom contact ID.
func NewTicketContactByID(id string) TicketContact {
	var contact TicketContact
	_ = contact.FromCreateTicketRequestContacts0(TicketContactID{Id: id})
	return contact
}

// NewTicketContactByExternalID constructs a ticket contact selector from an external contact ID.
func NewTicketContactByExternalID(externalID string) TicketContact {
	var contact TicketContact
	_ = contact.FromCreateTicketRequestContacts1(TicketContactExternalID{ExternalId: externalID})
	return contact
}

// NewTicketContactByEmail constructs a ticket contact selector from a contact email.
func NewTicketContactByEmail(email string) TicketContact {
	var contact TicketContact
	_ = contact.FromCreateTicketRequestContacts2(TicketContactEmail{Email: email})
	return contact
}

// TicketCreate holds the fields for creating a ticket.
type TicketCreate = gen.CreateTicketJSONBody

// TicketEnqueued is an async ticket creation job.
type TicketEnqueued = gen.JobsSchema

// TicketUpdate holds the fields for updating a ticket.
type TicketUpdate = gen.UpdateTicketJSONBody

// TicketSearchQuery holds the query for searching tickets.
type TicketSearchQuery = gen.SearchRequestSchema

// TicketState is an Intercom ticket state.
type TicketState = gen.TicketStateDetailedSchema

// TicketStateList is a list of ticket states.
type TicketStateList = gen.TicketStateListSchema

// TicketType is an Intercom ticket type.
type TicketType = gen.TicketTypeSchema

// TicketTypeList is a list of ticket types.
type TicketTypeList = gen.TicketTypeListSchema

// TicketTypeAttribute is an Intercom ticket type attribute.
type TicketTypeAttribute = gen.TicketTypeAttributeSchema

// TicketTypeCreate holds the fields for creating a ticket type.
type TicketTypeCreate = gen.CreateTicketTypeRequestSchema

// TicketTypeUpdate holds the fields for updating a ticket type.
type TicketTypeUpdate = gen.UpdateTicketTypeRequestSchema

// TicketTypeAttributeCreate holds the fields for creating a ticket type attribute.
type TicketTypeAttributeCreate = gen.CreateTicketTypeAttributeRequestSchema

// TicketTypeAttributeUpdate holds the fields for updating a ticket type attribute.
type TicketTypeAttributeUpdate = gen.UpdateTicketTypeAttributeRequestSchema

// TicketReply is the result of replying to a ticket.
type TicketReply = gen.TicketReplySchema

// TicketReplyMessageType identifies the kind of reply sent to a ticket.
type TicketReplyMessageType string

const (
	// TicketReplyMessageTypeComment posts a visible comment reply to the ticket.
	TicketReplyMessageTypeComment TicketReplyMessageType = "comment"
	// TicketReplyMessageTypeNote posts an internal note on the ticket.
	TicketReplyMessageTypeNote TicketReplyMessageType = "note"
	// TicketReplyMessageTypeQuickReply posts a reply that includes quick-reply options.
	TicketReplyMessageTypeQuickReply TicketReplyMessageType = "quick_reply"
)

// TicketReplyOption is a quick-reply option included in a ticket reply.
type TicketReplyOption struct {
	Text string `json:"text"`
	UUID string `json:"uuid"`
}

// TicketReplyContact identifies the contact replying to a ticket.
type TicketReplyContact struct {
	Email          *string `json:"email,omitempty"`
	IntercomUserID *string `json:"intercom_user_id,omitempty"`
	UserID         *string `json:"user_id,omitempty"`
}

// NewTicketReplyContactByEmail constructs a ticket reply contact selector from an email address.
func NewTicketReplyContactByEmail(email string) TicketReplyContact {
	return TicketReplyContact{Email: &email}
}

// NewTicketReplyContactByIntercomUserID constructs a ticket reply contact selector from an Intercom contact ID.
func NewTicketReplyContactByIntercomUserID(intercomUserID string) TicketReplyContact {
	return TicketReplyContact{IntercomUserID: &intercomUserID}
}

// NewTicketReplyContactByUserID constructs a ticket reply contact selector from an external user ID.
func NewTicketReplyContactByUserID(userID string) TicketReplyContact {
	return TicketReplyContact{UserID: &userID}
}

func (c TicketReplyContact) payload() (map[string]any, error) {
	count := 0
	payload := map[string]any{}
	if c.Email != nil {
		count++
		payload["email"] = *c.Email
	}
	if c.IntercomUserID != nil {
		count++
		payload["intercom_user_id"] = *c.IntercomUserID
	}
	if c.UserID != nil {
		count++
		payload["user_id"] = *c.UserID
	}
	if count != 1 {
		return nil, fmt.Errorf("intercom: exactly one ticket reply contact identifier is required")
	}
	return payload, nil
}

// TicketAdminReply replies to a ticket on behalf of an admin.
type TicketAdminReply struct {
	AdminID        string                 `json:"admin_id"`
	AttachmentURLs *[]string              `json:"attachment_urls,omitempty"`
	Body           *string                `json:"body,omitempty"`
	CreatedAt      *int                   `json:"created_at,omitempty"`
	CrossPost      *bool                  `json:"cross_post,omitempty"`
	MessageType    TicketReplyMessageType `json:"message_type"`
	ReplyOptions   *[]TicketReplyOption   `json:"reply_options,omitempty"`
}

func (TicketAdminReply) isTicketReplyRequest() {}

func (r TicketAdminReply) payload(skipNotifications *bool) (map[string]any, error) {
	payload := map[string]any{
		"admin_id":     r.AdminID,
		"message_type": r.MessageType,
		"type":         "admin",
	}
	if r.AttachmentURLs != nil {
		payload["attachment_urls"] = *r.AttachmentURLs
	}
	if r.Body != nil {
		payload["body"] = *r.Body
	}
	if r.CreatedAt != nil {
		payload["created_at"] = *r.CreatedAt
	}
	if r.CrossPost != nil {
		payload["cross_post"] = *r.CrossPost
	}
	if r.ReplyOptions != nil {
		payload["reply_options"] = *r.ReplyOptions
	}
	if skipNotifications != nil {
		payload["skip_notifications"] = *skipNotifications
	}
	return payload, nil
}

// TicketContactReply replies to a ticket on behalf of a contact.
type TicketContactReply struct {
	AttachmentURLs *[]string            `json:"attachment_urls,omitempty"`
	Body           string               `json:"body"`
	Contact        TicketReplyContact   `json:"-"`
	CreatedAt      *int                 `json:"created_at,omitempty"`
	ReplyOptions   *[]TicketReplyOption `json:"reply_options,omitempty"`
}

func (TicketContactReply) isTicketReplyRequest() {}

func (r TicketContactReply) payload(skipNotifications *bool) (map[string]any, error) {
	payload, err := r.Contact.payload()
	if err != nil {
		return nil, err
	}
	payload["body"] = r.Body
	payload["message_type"] = TicketReplyMessageTypeComment
	payload["type"] = "user"
	if r.AttachmentURLs != nil {
		payload["attachment_urls"] = *r.AttachmentURLs
	}
	if r.CreatedAt != nil {
		payload["created_at"] = *r.CreatedAt
	}
	if r.ReplyOptions != nil {
		payload["reply_options"] = *r.ReplyOptions
	}
	if skipNotifications != nil {
		payload["skip_notifications"] = *skipNotifications
	}
	return payload, nil
}

// TicketReplyRequest is a supported request body for the ticket reply endpoint.
type TicketReplyRequest interface {
	isTicketReplyRequest()
	payload(skipNotifications *bool) (map[string]any, error)
}

// TicketTagAttachRequest holds the fields for attaching a tag to a ticket.
type TicketTagAttachRequest = gen.AttachTagToTicketJSONBody

// TicketTagDetachRequest holds the fields for detaching a tag from a ticket.
type TicketTagDetachRequest = gen.DetachTagFromTicketJSONBody

// TicketsService exposes ticket-related Intercom API operations.
type TicketsService struct {
	client *Client
}

// Create creates a ticket.
func (s *TicketsService) Create(ctx context.Context, ticket TicketCreate) (*Ticket, error) {
	res, err := s.client.generated.CreateTicketWithResponse(ctx, nil, gen.CreateTicketJSONRequestBody(ticket))
	if err != nil {
		return nil, err
	}
	return requireOK("create ticket", res.StatusCode(), res.Body, res.JSON200)
}

// EnqueueCreate enqueues asynchronous ticket creation.
func (s *TicketsService) EnqueueCreate(ctx context.Context, ticket TicketCreate) (*TicketEnqueued, error) {
	res, err := s.client.generated.EnqueueCreateTicketWithResponse(ctx, nil, gen.EnqueueCreateTicketJSONRequestBody(ticket))
	if err != nil {
		return nil, err
	}
	return requireOK("enqueue create ticket", res.StatusCode(), res.Body, res.JSON200)
}

// Search searches tickets using an Intercom search query.
func (s *TicketsService) Search(ctx context.Context, query TicketSearchQuery) (*TicketList, error) {
	return s.SearchWithOptions(ctx, query, CursorPageOptions{})
}

// SearchWithOptions searches tickets using cursor pagination options.
func (s *TicketsService) SearchWithOptions(ctx context.Context, query TicketSearchQuery, options CursorPageOptions) (*TicketList, error) {
	if pagination := NewSearchPagination(options); pagination != nil {
		query.Pagination = pagination
	}
	res, err := s.client.generated.SearchTicketsWithResponse(ctx, nil, gen.SearchTicketsJSONRequestBody(query))
	if err != nil {
		return nil, err
	}
	return requireOK("search tickets", res.StatusCode(), res.Body, res.JSON200)
}

// Get retrieves a ticket by ID.
func (s *TicketsService) Get(ctx context.Context, ticketID string) (*Ticket, error) {
	if ticketID == "" {
		return nil, errRequiredID("ticket")
	}
	res, err := s.client.generated.GetTicketWithResponse(ctx, ticketID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("get ticket", res.StatusCode(), res.Body, res.JSON200)
}

// Update updates a ticket by ID.
func (s *TicketsService) Update(ctx context.Context, ticketID string, ticket TicketUpdate) (*Ticket, error) {
	if ticketID == "" {
		return nil, errRequiredID("ticket")
	}
	res, err := s.client.generated.UpdateTicketWithResponse(ctx, ticketID, nil, gen.UpdateTicketJSONRequestBody(ticket))
	if err != nil {
		return nil, err
	}
	return requireOK("update ticket", res.StatusCode(), res.Body, res.JSON200)
}

// Delete deletes a ticket by ID.
func (s *TicketsService) Delete(ctx context.Context, ticketID string) error {
	if ticketID == "" {
		return errRequiredID("ticket")
	}
	res, err := s.client.generated.DeleteTicketWithResponse(ctx, ticketID, nil)
	if err != nil {
		return err
	}
	return requireEmpty(res.StatusCode(), res.Body)
}

// Reply replies to a ticket.
func (s *TicketsService) Reply(ctx context.Context, ticketID string, req TicketReplyRequest, skipNotifications *bool) (*TicketReply, error) {
	if ticketID == "" {
		return nil, errRequiredID("ticket")
	}
	payload, err := req.payload(skipNotifications)
	if err != nil {
		return nil, err
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("intercom: marshal ticket reply request: %w", err)
	}
	res, err := s.client.generated.ReplyTicketWithBodyWithResponse(ctx, ticketID, nil, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	return requireOK("reply ticket", res.StatusCode(), res.Body, res.JSON200)
}

// ListStates returns all ticket states for the workspace.
func (s *TicketsService) ListStates(ctx context.Context) (*TicketStateList, error) {
	res, err := s.client.generated.ListTicketStatesWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("list ticket states", res.StatusCode(), res.Body, res.JSON200)
}

// ListTypes returns all ticket types for the workspace.
func (s *TicketsService) ListTypes(ctx context.Context) (*TicketTypeList, error) {
	res, err := s.client.generated.ListTicketTypesWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("list ticket types", res.StatusCode(), res.Body, res.JSON200)
}

// GetType retrieves a ticket type by ID.
func (s *TicketsService) GetType(ctx context.Context, ticketTypeID string) (*TicketType, error) {
	if ticketTypeID == "" {
		return nil, errRequiredID("ticket type")
	}
	res, err := s.client.generated.GetTicketTypeWithResponse(ctx, ticketTypeID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("get ticket type", res.StatusCode(), res.Body, res.JSON200)
}

// CreateType creates a ticket type.
func (s *TicketsService) CreateType(ctx context.Context, ticketType TicketTypeCreate) (*TicketType, error) {
	res, err := s.client.generated.CreateTicketTypeWithResponse(ctx, nil, gen.CreateTicketTypeJSONRequestBody(ticketType))
	if err != nil {
		return nil, err
	}
	return requireOK("create ticket type", res.StatusCode(), res.Body, res.JSON200)
}

// UpdateType updates a ticket type by ID.
func (s *TicketsService) UpdateType(ctx context.Context, ticketTypeID string, ticketType TicketTypeUpdate) (*TicketType, error) {
	if ticketTypeID == "" {
		return nil, errRequiredID("ticket type")
	}
	res, err := s.client.generated.UpdateTicketTypeWithResponse(ctx, ticketTypeID, nil, gen.UpdateTicketTypeJSONRequestBody(ticketType))
	if err != nil {
		return nil, err
	}
	return requireOK("update ticket type", res.StatusCode(), res.Body, res.JSON200)
}

// CreateTypeAttribute creates an attribute for a ticket type.
func (s *TicketsService) CreateTypeAttribute(ctx context.Context, ticketTypeID string, attribute TicketTypeAttributeCreate) (*TicketTypeAttribute, error) {
	if ticketTypeID == "" {
		return nil, errRequiredID("ticket type")
	}
	res, err := s.client.generated.CreateTicketTypeAttributeWithResponse(ctx, ticketTypeID, nil, gen.CreateTicketTypeAttributeJSONRequestBody(attribute))
	if err != nil {
		return nil, err
	}
	return requireOK("create ticket type attribute", res.StatusCode(), res.Body, res.JSON200)
}

// UpdateTypeAttribute updates an attribute for a ticket type.
func (s *TicketsService) UpdateTypeAttribute(ctx context.Context, ticketTypeID, attributeID string, attribute TicketTypeAttributeUpdate) (*TicketTypeAttribute, error) {
	if ticketTypeID == "" {
		return nil, errRequiredID("ticket type")
	}
	if attributeID == "" {
		return nil, errRequiredID("attribute")
	}
	res, err := s.client.generated.UpdateTicketTypeAttributeWithResponse(ctx, ticketTypeID, attributeID, nil, gen.UpdateTicketTypeAttributeJSONRequestBody(attribute))
	if err != nil {
		return nil, err
	}
	return requireOK("update ticket type attribute", res.StatusCode(), res.Body, res.JSON200)
}

// AttachTag attaches a tag to a ticket.
func (s *TicketsService) AttachTag(ctx context.Context, ticketID string, req TicketTagAttachRequest) (*Tag, error) {
	if ticketID == "" {
		return nil, errRequiredID("ticket")
	}
	res, err := s.client.generated.AttachTagToTicketWithResponse(ctx, ticketID, nil, gen.AttachTagToTicketJSONRequestBody(req))
	if err != nil {
		return nil, err
	}
	return requireOK("attach tag to ticket", res.StatusCode(), res.Body, res.JSON200)
}

// DetachTag detaches a tag from a ticket.
func (s *TicketsService) DetachTag(ctx context.Context, ticketID, tagID string, req TicketTagDetachRequest) (*Tag, error) {
	if ticketID == "" {
		return nil, errRequiredID("ticket")
	}
	if tagID == "" {
		return nil, errRequiredID("tag")
	}
	res, err := s.client.generated.DetachTagFromTicketWithResponse(ctx, ticketID, tagID, nil, gen.DetachTagFromTicketJSONRequestBody(req))
	if err != nil {
		return nil, err
	}
	return requireOK("detach tag from ticket", res.StatusCode(), res.Body, res.JSON200)
}

func errRequiredID(resource string) error {
	return genErrRequiredID(resource)
}

func genErrRequiredID(resource string) error {
	return &requiredIDError{resource: resource}
}

type requiredIDError struct {
	resource string
}

func (e *requiredIDError) Error() string {
	return "intercom: " + e.resource + " ID is required"
}
