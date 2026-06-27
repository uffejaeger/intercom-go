package intercom

import "context"

// SubscriptionTypesService exposes subscription-type-related Intercom API operations.
type SubscriptionTypesService struct {
	client *Client
}

// List returns all subscription types for the workspace.
func (s *SubscriptionTypesService) List(ctx context.Context) (*SubscriptionTypeList, error) {
	res, err := s.client.generated.ListSubscriptionTypesWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("list subscription types", res.StatusCode(), res.Body, res.JSON200)
}
