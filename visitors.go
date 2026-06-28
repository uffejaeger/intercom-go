package intercom

import (
	"context"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// Visitor is an Intercom visitor.
type Visitor = gen.VisitorSchema

// VisitorConverted is the contact returned after converting a visitor.
type VisitorConverted = gen.ContactSchema

// VisitorUpdate holds the fields for updating a visitor.
type VisitorUpdate struct {
	CustomAttributes *map[string]string `json:"custom_attributes,omitempty"`
	ID               *string            `json:"id,omitempty"`
	Name             *string            `json:"name,omitempty"`
	UserID           *string            `json:"user_id,omitempty"`
}

func (u VisitorUpdate) toGenerated() gen.UpdateVisitorRequestSchema {
	return gen.UpdateVisitorRequestSchema{
		CustomAttributes: u.CustomAttributes,
		Id:               u.ID,
		Name:             u.Name,
		UserId:           u.UserID,
	}
}

// VisitorConvertContact identifies the contact retained after converting a visitor.
type VisitorConvertContact struct {
	Email  *string `json:"email,omitempty"`
	ID     *string `json:"id,omitempty"`
	UserID *string `json:"user_id,omitempty"`
}

func (c VisitorConvertContact) toGenerated() gen.ConvertVisitorRequest_User {
	return gen.ConvertVisitorRequest_User{
		Email:  c.Email,
		Id:     c.ID,
		UserId: c.UserID,
	}
}

// VisitorConvertSource identifies the visitor to convert.
type VisitorConvertSource struct {
	Email  *string `json:"email,omitempty"`
	ID     *string `json:"id,omitempty"`
	UserID *string `json:"user_id,omitempty"`
}

func (v VisitorConvertSource) toGenerated() gen.ConvertVisitorRequest_Visitor {
	return gen.ConvertVisitorRequest_Visitor{
		Email:  v.Email,
		Id:     v.ID,
		UserId: v.UserID,
	}
}

// VisitorConvert holds the fields for converting a visitor.
type VisitorConvert struct {
	Type    string                `json:"type"`
	User    VisitorConvertContact `json:"user"`
	Visitor VisitorConvertSource  `json:"visitor"`
}

func (r VisitorConvert) toGenerated() gen.ConvertVisitorRequestSchema {
	return gen.ConvertVisitorRequestSchema{
		Type:    r.Type,
		User:    r.User.toGenerated(),
		Visitor: r.Visitor.toGenerated(),
	}
}

// VisitorsService exposes visitor-related Intercom API operations.
type VisitorsService struct {
	client *Client
}

// GetByUserID retrieves a visitor by user ID.
func (s *VisitorsService) GetByUserID(ctx context.Context, userID string) (*Visitor, error) {
	if userID == "" {
		return nil, errRequiredID("user")
	}
	res, err := s.client.generated.RetrieveVisitorWithUserIdWithResponse(ctx, &gen.RetrieveVisitorWithUserIdParams{UserId: userID})
	if err != nil {
		return nil, err
	}
	return requireOK("retrieve visitor", res.StatusCode(), res.Body, res.JSON200)
}

// Update updates a visitor.
func (s *VisitorsService) Update(ctx context.Context, update VisitorUpdate) (*Visitor, error) {
	res, err := s.client.generated.UpdateVisitorWithResponse(ctx, nil, gen.UpdateVisitorJSONRequestBody(update.toGenerated()))
	if err != nil {
		return nil, err
	}
	return requireOK("update visitor", res.StatusCode(), res.Body, res.JSON200)
}

// Convert converts a visitor into a contact.
func (s *VisitorsService) Convert(ctx context.Context, req VisitorConvert) (*VisitorConverted, error) {
	res, err := s.client.generated.ConvertVisitorWithResponse(ctx, nil, gen.ConvertVisitorJSONRequestBody(req.toGenerated()))
	if err != nil {
		return nil, err
	}
	return requireOK("convert visitor", res.StatusCode(), res.Body, res.JSON200)
}
