package intercom

import (
	"context"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// HelpCenter is an Intercom help center.
type HelpCenter = gen.HelpCenterSchema

// HelpCenterList is a list of Intercom help centers.
type HelpCenterList = gen.HelpCenterListSchema

// HelpCentersService exposes help-center-related Intercom API operations.
type HelpCentersService struct {
	client *Client
}

// List returns all help centers.
func (s *HelpCentersService) List(ctx context.Context) (*HelpCenterList, error) {
	res, err := s.client.generated.ListHelpCentersWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("list help centers", res.StatusCode(), res.Body, res.JSON200)
}

// Retrieve returns a help center by ID.
func (s *HelpCentersService) Retrieve(ctx context.Context, helpCenterID string) (*HelpCenter, error) {
	id, err := requireIntID("help center", helpCenterID)
	if err != nil {
		return nil, err
	}
	res, err := s.client.generated.RetrieveHelpCenterWithResponse(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("retrieve help center", res.StatusCode(), res.Body, res.JSON200)
}
