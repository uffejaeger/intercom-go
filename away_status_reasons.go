package intercom

import (
	"context"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// AwayStatusReason is an Intercom away status reason.
type AwayStatusReason = gen.AwayStatusReasonSchema

// AwayStatusReasonsService exposes away status reason Intercom API operations.
type AwayStatusReasonsService struct {
	client *Client
}

// List returns all away status reasons for the workspace.
func (s *AwayStatusReasonsService) List(ctx context.Context) ([]AwayStatusReason, error) {
	res, err := s.client.generated.ListAwayStatusReasonsWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	list, err := requireOK("list away status reasons", res.StatusCode(), res.Body, res.JSON200)
	if err != nil {
		return nil, err
	}
	if list.Data == nil {
		return nil, nil
	}
	return *list.Data, nil
}
