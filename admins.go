package intercom

import (
	"context"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// Admin is the authenticated Intercom teammate account and workspace details.
type Admin = gen.AdminWithAppSchema

// AdminsService exposes admin-related Intercom API operations.
type AdminsService struct {
	client *Client
}

// Me identifies the currently authenticated admin.
func (s *AdminsService) Me(ctx context.Context) (*Admin, error) {
	res, err := s.client.generated.IdentifyAdminWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}

	return requireOK("identify admin", res.StatusCode(), res.Body, res.JSON200)
}
