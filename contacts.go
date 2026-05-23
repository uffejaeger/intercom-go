package intercom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// Contact is an Intercom contact.
type Contact = gen.ContactSchema

// ContactList is a page of Intercom contacts.
type ContactList = gen.ContactListSchema

// ContactDeleted is the result of deleting a contact.
type ContactDeleted = gen.ContactDeleted

// ContactArchived is the result of archiving a contact.
type ContactArchived = gen.ContactArchived

// ContactUnarchived is the result of unarchiving a contact.
type ContactUnarchived = gen.ContactUnarchived

// ContactBlocked is the result of blocking a contact.
type ContactBlocked = gen.ContactBlockedSchema

// ContactCreate holds the fields for creating a contact.
type ContactCreate = gen.CreateContactRequestSchema

// ContactUpdate holds the fields for updating a contact.
type ContactUpdate = gen.UpdateContactRequestSchema

// Note is an Intercom note on a contact.
type Note = gen.NoteSchema

// NoteList is a list of notes.
type NoteList = gen.NoteListSchema

// ContactSegments is the list of segments a contact belongs to.
type ContactSegments = gen.ContactSegmentsSchema

// SubscriptionType is an Intercom subscription type.
type SubscriptionType = gen.SubscriptionTypeSchema

// SubscriptionTypeList is a list of subscription types.
type SubscriptionTypeList = gen.SubscriptionTypeListSchema

// Tag is an Intercom tag.
type Tag = gen.TagSchema

// TagList is a list of tags.
type TagList = gen.TagListSchema

// ContactsService exposes contact-related Intercom API operations.
type ContactsService struct {
	client *Client
}

// ContactSearchOperator is an Intercom contact search operator.
type ContactSearchOperator string

const (
	ContactSearchEquals      ContactSearchOperator = "="
	ContactSearchNotEquals   ContactSearchOperator = "!="
	ContactSearchGreaterThan ContactSearchOperator = ">"
	ContactSearchLessThan    ContactSearchOperator = "<"
	ContactSearchContains    ContactSearchOperator = "~"
)

// ContactSearch describes a single-filter contact search.
type ContactSearch struct {
	Field         string
	Operator      ContactSearchOperator
	Value         any
	PerPage       int
	StartingAfter string
}

// Get retrieves a contact by Intercom contact ID.
func (s *ContactsService) Get(ctx context.Context, contactID string) (*Contact, error) {
	if contactID == "" {
		return nil, fmt.Errorf("intercom: contact ID is required")
	}

	res, err := s.client.generated.ShowContactWithResponse(ctx, contactID, nil)
	if err != nil {
		return nil, err
	}

	return requireOK("get contact", res.StatusCode(), res.Body, res.JSON200)
}

// GetByExternalID retrieves a contact by external ID.
func (s *ContactsService) GetByExternalID(ctx context.Context, externalID string) (*Contact, error) {
	if externalID == "" {
		return nil, fmt.Errorf("intercom: external ID is required")
	}

	res, err := s.client.generated.ShowContactByExternalIdWithResponse(ctx, externalID, nil)
	if err != nil {
		return nil, err
	}

	return requireOK("get contact by external ID", res.StatusCode(), res.Body, res.JSON200)
}

// List returns contacts.
func (s *ContactsService) List(ctx context.Context) (*ContactList, error) {
	res, err := s.client.generated.ListContactsWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}

	return requireOK("list contacts", res.StatusCode(), res.Body, res.JSON200)
}

// Search searches contacts using one Intercom search filter.
func (s *ContactsService) Search(ctx context.Context, search ContactSearch) (*ContactList, error) {
	body, err := search.toGenerated()
	if err != nil {
		return nil, err
	}

	res, err := s.client.generated.SearchContactsWithResponse(ctx, nil, body)
	if err != nil {
		return nil, err
	}

	return requireOK("search contacts", res.StatusCode(), res.Body, res.JSON200)
}

// Create creates a new contact.
func (s *ContactsService) Create(ctx context.Context, contact ContactCreate) (*Contact, error) {
	body, err := marshalBody(contact)
	if err != nil {
		return nil, err
	}
	res, err := s.client.generated.CreateContactWithBodyWithResponse(ctx, nil, "application/json", body)
	if err != nil {
		return nil, err
	}
	return requireOK("create contact", res.StatusCode(), res.Body, res.JSON200)
}

// Update updates an existing contact.
func (s *ContactsService) Update(ctx context.Context, contactID string, contact ContactUpdate) (*Contact, error) {
	if contactID == "" {
		return nil, fmt.Errorf("intercom: contact ID is required")
	}
	body, err := marshalBody(contact)
	if err != nil {
		return nil, err
	}
	res, err := s.client.generated.UpdateContactWithBodyWithResponse(ctx, contactID, nil, "application/json", body)
	if err != nil {
		return nil, err
	}
	return requireOK("update contact", res.StatusCode(), res.Body, res.JSON200)
}

// Merge merges a lead (from) into a user (into).
func (s *ContactsService) Merge(ctx context.Context, from, into string) (*Contact, error) {
	if from == "" {
		return nil, fmt.Errorf("intercom: from contact ID is required")
	}
	if into == "" {
		return nil, fmt.Errorf("intercom: into contact ID is required")
	}

	res, err := s.client.generated.MergeContactWithResponse(ctx, nil, gen.MergeContactJSONRequestBody{
		From: &from,
		Into: &into,
	})
	if err != nil {
		return nil, err
	}

	return requireOK("merge contact", res.StatusCode(), res.Body, res.JSON200)
}

// Archive archives a contact.
func (s *ContactsService) Archive(ctx context.Context, contactID string) (*ContactArchived, error) {
	if contactID == "" {
		return nil, fmt.Errorf("intercom: contact ID is required")
	}

	res, err := s.client.generated.ArchiveContactWithResponse(ctx, contactID, nil)
	if err != nil {
		return nil, err
	}

	return requireOK("archive contact", res.StatusCode(), res.Body, res.JSON200)
}

// Unarchive unarchives a contact.
func (s *ContactsService) Unarchive(ctx context.Context, contactID string) (*ContactUnarchived, error) {
	if contactID == "" {
		return nil, fmt.Errorf("intercom: contact ID is required")
	}

	res, err := s.client.generated.UnarchiveContactWithResponse(ctx, contactID, nil)
	if err != nil {
		return nil, err
	}

	return requireOK("unarchive contact", res.StatusCode(), res.Body, res.JSON200)
}

// Block blocks a contact.
func (s *ContactsService) Block(ctx context.Context, contactID string) (*ContactBlocked, error) {
	if contactID == "" {
		return nil, fmt.Errorf("intercom: contact ID is required")
	}

	res, err := s.client.generated.BlockContactWithResponse(ctx, contactID, nil)
	if err != nil {
		return nil, err
	}

	return requireOK("block contact", res.StatusCode(), res.Body, res.JSON200)
}

// Delete deletes a contact.
func (s *ContactsService) Delete(ctx context.Context, contactID string) (*ContactDeleted, error) {
	if contactID == "" {
		return nil, fmt.Errorf("intercom: contact ID is required")
	}

	res, err := s.client.generated.DeleteContactWithResponse(ctx, contactID, nil)
	if err != nil {
		return nil, err
	}

	return requireOK("delete contact", res.StatusCode(), res.Body, res.JSON200)
}

// ListNotes returns notes for a contact.
func (s *ContactsService) ListNotes(ctx context.Context, contactID string) (*NoteList, error) {
	if contactID == "" {
		return nil, fmt.Errorf("intercom: contact ID is required")
	}

	res, err := s.client.generated.ListNotesWithResponse(ctx, contactID, nil)
	if err != nil {
		return nil, err
	}

	return requireOK("list notes", res.StatusCode(), res.Body, res.JSON200)
}

// CreateNote creates a note on a contact.
func (s *ContactsService) CreateNote(ctx context.Context, contactID string, body string, adminID string) (*Note, error) {
	if contactID == "" {
		return nil, fmt.Errorf("intercom: contact ID is required")
	}
	if body == "" {
		return nil, fmt.Errorf("intercom: note body is required")
	}

	id, err := contactIDToInt(contactID)
	if err != nil {
		return nil, err
	}

	req := gen.CreateNoteJSONRequestBody{Body: body}
	if adminID != "" {
		req.AdminId = &adminID
	}

	res, err := s.client.generated.CreateNoteWithResponse(ctx, id, nil, req)
	if err != nil {
		return nil, err
	}

	return requireOK("create note", res.StatusCode(), res.Body, res.JSON200)
}

// ListSegments returns segments a contact belongs to.
func (s *ContactsService) ListSegments(ctx context.Context, contactID string) (*ContactSegments, error) {
	if contactID == "" {
		return nil, fmt.Errorf("intercom: contact ID is required")
	}

	res, err := s.client.generated.ListSegmentsForAContactWithResponse(ctx, contactID, nil)
	if err != nil {
		return nil, err
	}

	return requireOK("list segments for contact", res.StatusCode(), res.Body, res.JSON200)
}

// ListSubscriptions returns subscription types for a contact.
func (s *ContactsService) ListSubscriptions(ctx context.Context, contactID string) (*SubscriptionTypeList, error) {
	if contactID == "" {
		return nil, fmt.Errorf("intercom: contact ID is required")
	}

	res, err := s.client.generated.ListSubscriptionsForAContactWithResponse(ctx, contactID, nil)
	if err != nil {
		return nil, err
	}

	return requireOK("list subscriptions for contact", res.StatusCode(), res.Body, res.JSON200)
}

// AttachSubscription attaches a subscription type to a contact.
func (s *ContactsService) AttachSubscription(ctx context.Context, contactID, subscriptionID, consentType string) (*SubscriptionType, error) {
	if contactID == "" {
		return nil, fmt.Errorf("intercom: contact ID is required")
	}
	if subscriptionID == "" {
		return nil, fmt.Errorf("intercom: subscription ID is required")
	}
	if consentType == "" {
		return nil, fmt.Errorf("intercom: consent type is required")
	}

	res, err := s.client.generated.AttachSubscriptionTypeToContactWithResponse(ctx, contactID, nil, gen.AttachSubscriptionTypeToContactJSONRequestBody{
		Id:          subscriptionID,
		ConsentType: consentType,
	})
	if err != nil {
		return nil, err
	}

	return requireOK("attach subscription to contact", res.StatusCode(), res.Body, res.JSON200)
}

// DetachSubscription detaches a subscription type from a contact.
func (s *ContactsService) DetachSubscription(ctx context.Context, contactID, subscriptionID string) (*SubscriptionType, error) {
	if contactID == "" {
		return nil, fmt.Errorf("intercom: contact ID is required")
	}
	if subscriptionID == "" {
		return nil, fmt.Errorf("intercom: subscription ID is required")
	}

	res, err := s.client.generated.DetachSubscriptionTypeToContactWithResponse(ctx, contactID, subscriptionID, nil)
	if err != nil {
		return nil, err
	}

	return requireOK("detach subscription from contact", res.StatusCode(), res.Body, res.JSON200)
}

// ListTags returns tags attached to a contact.
func (s *ContactsService) ListTags(ctx context.Context, contactID string) (*TagList, error) {
	if contactID == "" {
		return nil, fmt.Errorf("intercom: contact ID is required")
	}

	res, err := s.client.generated.ListTagsForAContactWithResponse(ctx, contactID, nil)
	if err != nil {
		return nil, err
	}

	return requireOK("list tags for contact", res.StatusCode(), res.Body, res.JSON200)
}

// AttachTag attaches a tag to a contact.
func (s *ContactsService) AttachTag(ctx context.Context, contactID, tagID string) (*Tag, error) {
	if contactID == "" {
		return nil, fmt.Errorf("intercom: contact ID is required")
	}
	if tagID == "" {
		return nil, fmt.Errorf("intercom: tag ID is required")
	}

	res, err := s.client.generated.AttachTagToContactWithResponse(ctx, contactID, nil, gen.AttachTagToContactJSONRequestBody{
		Id: tagID,
	})
	if err != nil {
		return nil, err
	}

	return requireOK("attach tag to contact", res.StatusCode(), res.Body, res.JSON200)
}

// DetachTag detaches a tag from a contact.
func (s *ContactsService) DetachTag(ctx context.Context, contactID, tagID string) (*Tag, error) {
	if contactID == "" {
		return nil, fmt.Errorf("intercom: contact ID is required")
	}
	if tagID == "" {
		return nil, fmt.Errorf("intercom: tag ID is required")
	}

	res, err := s.client.generated.DetachTagFromContactWithResponse(ctx, contactID, tagID, nil)
	if err != nil {
		return nil, err
	}

	return requireOK("detach tag from contact", res.StatusCode(), res.Body, res.JSON200)
}

func (s ContactSearch) toGenerated() (gen.SearchContactsJSONRequestBody, error) {
	if s.Field == "" {
		return gen.SearchContactsJSONRequestBody{}, fmt.Errorf("intercom: contact search field is required")
	}
	if s.Operator == "" {
		return gen.SearchContactsJSONRequestBody{}, fmt.Errorf("intercom: contact search operator is required")
	}

	value, err := contactSearchValue(s.Value)
	if err != nil {
		return gen.SearchContactsJSONRequestBody{}, err
	}

	operator := gen.SingleFilterSearchRequestOperator(s.Operator)
	filter := gen.SingleFilterSearchRequestSchema{
		Field:    &s.Field,
		Operator: &operator,
		Value:    &value,
	}

	var query gen.SearchRequest_Query
	_ = query.FromSingleFilterSearchRequestSchema(filter) // json.Marshal on a simple struct, never fails

	body := gen.SearchContactsJSONRequestBody{
		Query: query,
	}
	if s.PerPage > 0 || s.StartingAfter != "" {
		body.Pagination = &gen.StartingAfterPagingSchema{}
		if s.PerPage > 0 {
			body.Pagination.PerPage = &s.PerPage
		}
		if s.StartingAfter != "" {
			body.Pagination.StartingAfter = &s.StartingAfter
		}
	}

	return body, nil
}

func contactSearchValue(value any) (gen.SingleFilterSearchRequest_Value, error) {
	var generated gen.SingleFilterSearchRequest_Value

	switch typed := value.(type) {
	case string:
		return generated, generated.FromSingleFilterSearchRequestValue0(typed)
	case int:
		return generated, generated.FromSingleFilterSearchRequestValue1(typed)
	case []string:
		items := make([]gen.SingleFilterSearchRequest_Value_2_Item, 0, len(typed))
		for _, item := range typed {
			var generatedItem gen.SingleFilterSearchRequest_Value_2_Item
			_ = generatedItem.FromSingleFilterSearchRequestValue20(item) // json.Marshal(string) never fails
			items = append(items, generatedItem)
		}
		return generated, generated.FromSingleFilterSearchRequestValue2(items)
	case []int:
		items := make([]gen.SingleFilterSearchRequest_Value_2_Item, 0, len(typed))
		for _, item := range typed {
			var generatedItem gen.SingleFilterSearchRequest_Value_2_Item
			_ = generatedItem.FromSingleFilterSearchRequestValue21(item) // json.Marshal(int) never fails
			items = append(items, generatedItem)
		}
		return generated, generated.FromSingleFilterSearchRequestValue2(items)
	default:
		return generated, fmt.Errorf("intercom: unsupported contact search value type %T", value)
	}
}

// marshalBody marshals v to JSON and returns it as a *bytes.Reader.
func marshalBody(v any) (*bytes.Reader, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("intercom: marshal request body: %w", err)
	}
	return bytes.NewReader(b), nil
}

// contactIDToInt converts a string contact ID to int as required by some generated endpoints.
func contactIDToInt(contactID string) (int, error) {
	id, err := strconv.Atoi(contactID)
	if err != nil {
		return 0, fmt.Errorf("intercom: contact ID %q is not a valid integer: %w", contactID, err)
	}
	return id, nil
}
