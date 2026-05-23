package intercom

import (
	"context"
	"fmt"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// Contact is an Intercom contact.
type Contact = gen.ContactSchema

// ContactList is a page of Intercom contacts.
type ContactList = gen.ContactListSchema

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
	if err := query.FromSingleFilterSearchRequestSchema(filter); err != nil {
		return gen.SearchContactsJSONRequestBody{}, err
	}

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
			if err := generatedItem.FromSingleFilterSearchRequestValue20(item); err != nil {
				return generated, err
			}
			items = append(items, generatedItem)
		}
		return generated, generated.FromSingleFilterSearchRequestValue2(items)
	case []int:
		items := make([]gen.SingleFilterSearchRequest_Value_2_Item, 0, len(typed))
		for _, item := range typed {
			var generatedItem gen.SingleFilterSearchRequest_Value_2_Item
			if err := generatedItem.FromSingleFilterSearchRequestValue21(item); err != nil {
				return generated, err
			}
			items = append(items, generatedItem)
		}
		return generated, generated.FromSingleFilterSearchRequestValue2(items)
	default:
		return generated, fmt.Errorf("intercom: unsupported contact search value type %T", value)
	}
}
