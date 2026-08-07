package intercom

import (
	"context"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// Audience is a group of contacts that can be targeted by Fin.
type Audience = gen.AudienceSchema

// AudienceList is a paginated list of audiences.
type AudienceList = gen.AudienceListSchema

// AudienceListParams configures audience listing.
type AudienceListParams = gen.ListAudiencesParams

// AudiencePredicate is a condition used to select contacts for an audience.
type AudiencePredicate = gen.PredicateSchema

// AudienceCreate configures a new audience.
type AudienceCreate = gen.CreateAudienceRequestSchema

// AudienceUpdate configures an audience update.
type AudienceUpdate = gen.UpdateAudienceRequestSchema

// AudiencesService exposes audience operations.
type AudiencesService struct{ client *Client }

// List returns audiences for the workspace.
func (s *AudiencesService) List(ctx context.Context, params *AudienceListParams) (*AudienceList, error) {
	res, err := s.client.generated.ListAudiencesWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	return requireOK("list audiences", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

// Create creates an audience.
func (s *AudiencesService) Create(ctx context.Context, audience AudienceCreate) (*Audience, error) {
	res, err := s.client.generated.CreateAudienceWithResponse(ctx, nil, audience)
	if err != nil {
		return nil, err
	}
	return requireCreated("create audience", res.StatusCode(), res.Body, res.JSON201, responseHeaders(res.HTTPResponse))
}

// Get returns an audience by ID.
func (s *AudiencesService) Get(ctx context.Context, id string) (*Audience, error) {
	res, err := s.client.generated.RetrieveAudienceWithResponse(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("get audience", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

// Update updates an audience.
func (s *AudiencesService) Update(ctx context.Context, id string, audience AudienceUpdate) (*Audience, error) {
	res, err := s.client.generated.UpdateAudienceWithResponse(ctx, id, nil, audience)
	if err != nil {
		return nil, err
	}
	return requireOK("update audience", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

// Delete removes an audience.
func (s *AudiencesService) Delete(ctx context.Context, id string) error {
	res, err := s.client.generated.DeleteAudienceWithResponse(ctx, id, nil)
	if err != nil {
		return err
	}
	return requireEmpty(res.StatusCode(), res.Body, responseHeaders(res.HTTPResponse))
}
