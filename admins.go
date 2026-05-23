package intercom

import (
	"context"
	"fmt"
	"strconv"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// Admin is the authenticated Intercom teammate account and workspace details.
type Admin = gen.AdminWithAppSchema

// AdminDetail is a single Intercom admin (without workspace details).
type AdminDetail = gen.AdminSchema

// AdminList is a list of Intercom admins.
type AdminList = gen.AdminListSchema

// AdminActivityLogs is a page of admin activity log entries.
type AdminActivityLogs = gen.ActivityLogListSchema

// AdminSetAway holds the fields for setting an admin's away status.
type AdminSetAway = gen.SetAwayAdminJSONRequestBody

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

// List returns all admins for the workspace.
func (s *AdminsService) List(ctx context.Context) (*AdminList, error) {
	res, err := s.client.generated.ListAdminsWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("list admins", res.StatusCode(), res.Body, res.JSON200)
}

// Retrieve retrieves an admin by ID.
func (s *AdminsService) Retrieve(ctx context.Context, adminID string) (*AdminDetail, error) {
	if adminID == "" {
		return nil, fmt.Errorf("intercom: admin ID is required")
	}
	id, err := strconv.Atoi(adminID)
	if err != nil {
		return nil, fmt.Errorf("intercom: admin ID %q is not a valid integer: %w", adminID, err)
	}
	res, err := s.client.generated.RetrieveAdminWithResponse(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("retrieve admin", res.StatusCode(), res.Body, res.JSON200)
}

// SetAway sets the away status of an admin.
func (s *AdminsService) SetAway(ctx context.Context, adminID string, req AdminSetAway) (*AdminDetail, error) {
	if adminID == "" {
		return nil, fmt.Errorf("intercom: admin ID is required")
	}
	id, err := strconv.Atoi(adminID)
	if err != nil {
		return nil, fmt.Errorf("intercom: admin ID %q is not a valid integer: %w", adminID, err)
	}
	res, err := s.client.generated.SetAwayAdminWithResponse(ctx, id, nil, gen.SetAwayAdminJSONRequestBody(req))
	if err != nil {
		return nil, err
	}
	return requireOK("set away admin", res.StatusCode(), res.Body, res.JSON200)
}

// ListActivityLogs returns activity logs for admins, starting from createdAtAfter (UNIX timestamp string).
func (s *AdminsService) ListActivityLogs(ctx context.Context, createdAtAfter string) (*AdminActivityLogs, error) {
	if createdAtAfter == "" {
		return nil, fmt.Errorf("intercom: created_at_after is required")
	}
	res, err := s.client.generated.ListActivityLogsWithResponse(ctx, &gen.ListActivityLogsParams{CreatedAtAfter: createdAtAfter})
	if err != nil {
		return nil, err
	}
	return requireOK("list activity logs", res.StatusCode(), res.Body, res.JSON200)
}
