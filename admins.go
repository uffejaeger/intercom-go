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

// AdminActivityLogEventTypes is the list of supported admin activity-log event types.
type AdminActivityLogEventTypes = gen.ActivityLogEventTypeListSchema

// AdminActivityLogEventTypesParams configures an activity-log event-type request.
type AdminActivityLogEventTypesParams = gen.ListActivityLogEventTypesParams

// AdminActivityLogSearchParams configures an activity-log search request.
type AdminActivityLogSearchParams = gen.SearchActivityLogsParams

// AdminActivityLogSearch is the body for an activity-log search request.
type AdminActivityLogSearch = gen.SearchActivityLogsJSONRequestBody

// AdminSetAway holds the fields for setting an admin's away status.
type AdminSetAway = gen.SetAwayAdminJSONRequestBody

// AdminsService exposes admin-related Intercom API operations.
type AdminsService struct {
	client *Client
}

// ListActivityLogEventTypes returns available admin activity-log event types.
func (s *AdminsService) ListActivityLogEventTypes(ctx context.Context, params *AdminActivityLogEventTypesParams) (*AdminActivityLogEventTypes, error) {
	res, err := s.client.generated.ListActivityLogEventTypesWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	return requireOK("list activity-log event types", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

// SearchActivityLogs searches admin activity logs.
func (s *AdminsService) SearchActivityLogs(ctx context.Context, params *AdminActivityLogSearchParams, request AdminActivityLogSearch) (*AdminActivityLogs, error) {
	res, err := s.client.generated.SearchActivityLogsWithResponse(ctx, params, request)
	if err != nil {
		return nil, err
	}
	return requireOK("search activity logs", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

// Me identifies the currently authenticated admin.
func (s *AdminsService) Me(ctx context.Context) (*Admin, error) {
	res, err := s.client.generated.IdentifyAdminWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("identify admin", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

// List returns all admins for the workspace.
func (s *AdminsService) List(ctx context.Context) (*AdminList, error) {
	res, err := s.client.generated.ListAdminsWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("list admins", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
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
	return requireOK("retrieve admin", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
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
	return requireOK("set away admin", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
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
	return requireOK("list activity logs", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}
