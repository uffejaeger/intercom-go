package intercom

import (
	"context"
	"fmt"
	"net/http"

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

	if res.StatusCode() != http.StatusOK {
		return nil, parseErrorResponse(res.StatusCode(), res.Body)
	}
	if res.JSON200 == nil {
		return nil, fmt.Errorf("intercom: identify admin returned status %d without a response body", res.StatusCode())
	}

	return res.JSON200, nil
}
