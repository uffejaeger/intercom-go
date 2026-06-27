package intercom

import (
	"bytes"
	"context"
	"encoding/json"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// TicketList is a list of Intercom tickets.
type TicketList = gen.TicketListSchema

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

// Reply replies to a ticket. Pass either a generated admin or contact reply payload.
func (s *TicketsService) Reply(ctx context.Context, ticketID string, req any, skipNotifications *bool) (*TicketReply, error) {
	if ticketID == "" {
		return nil, errRequiredID("ticket")
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	if skipNotifications != nil {
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, err
		}
		payload["skip_notifications"] = *skipNotifications
		body, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
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
func (s *TicketsService) AttachTag(ctx context.Context, ticketID string, req gen.AttachTagToTicketJSONBody) (*Tag, error) {
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
func (s *TicketsService) DetachTag(ctx context.Context, ticketID, tagID string, req gen.DetachTagFromTicketJSONBody) (*Tag, error) {
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
