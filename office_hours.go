package intercom

import (
	"context"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// OfficeHoursSchedule defines recurring workspace opening hours.
type OfficeHoursSchedule = gen.OfficeHoursScheduleSchema

// OfficeHoursScheduleList is a list of office-hours schedules.
type OfficeHoursScheduleList = gen.OfficeHoursScheduleListSchema

// OfficeHoursException overrides a schedule for one date.
type OfficeHoursException = gen.OfficeHoursExceptionSchema

// OfficeHoursExceptionList is a list of office-hours exceptions.
type OfficeHoursExceptionList = gen.OfficeHoursExceptionListSchema

// OfficeHoursTimeInterval is an open interval in an office-hours schedule or exception.
type OfficeHoursTimeInterval = gen.OfficeHoursTimeIntervalSchema

// OfficeHoursScheduleCreate configures a schedule.
type OfficeHoursScheduleCreate = gen.CreateOfficeHoursScheduleRequestSchema

// OfficeHoursScheduleUpdate configures a schedule update.
type OfficeHoursScheduleUpdate = gen.UpdateOfficeHoursScheduleRequestSchema

// OfficeHoursExceptionCreate configures an exception.
type OfficeHoursExceptionCreate = gen.CreateOfficeHoursExceptionRequestSchema

// OfficeHoursExceptionUpdate configures an exception update.
type OfficeHoursExceptionUpdate = gen.UpdateOfficeHoursExceptionRequestSchema

// OfficeHoursExceptionUpdateType identifies the exception behavior for an update.
type OfficeHoursExceptionUpdateType = gen.UpdateOfficeHoursExceptionRequestExceptionType

// OfficeHoursScheduleListParams configures an office-hours schedule list request.
type OfficeHoursScheduleListParams = gen.ListOfficeHoursSchedulesParams

// OfficeHoursExceptionListParams configures an office-hours exception list request.
type OfficeHoursExceptionListParams = gen.ListOfficeHoursExceptionsParams

// OfficeHoursService exposes workspace office-hours schedules and exceptions.
type OfficeHoursService struct{ client *Client }

// ListSchedules returns workspace office-hours schedules.
func (s *OfficeHoursService) ListSchedules(ctx context.Context, params *OfficeHoursScheduleListParams) (*OfficeHoursScheduleList, error) {
	res, err := s.client.generated.ListOfficeHoursSchedulesWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	return requireOK("list office hours schedules", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

// CreateSchedule creates an office-hours schedule.
func (s *OfficeHoursService) CreateSchedule(ctx context.Context, schedule OfficeHoursScheduleCreate) (*OfficeHoursSchedule, error) {
	res, err := s.client.generated.CreateOfficeHoursScheduleWithResponse(ctx, nil, schedule)
	if err != nil {
		return nil, err
	}
	return requireCreated("create office hours schedule", res.StatusCode(), res.Body, res.JSON201, responseHeaders(res.HTTPResponse))
}

// GetSchedule returns an office-hours schedule by ID.
func (s *OfficeHoursService) GetSchedule(ctx context.Context, id string) (*OfficeHoursSchedule, error) {
	res, err := s.client.generated.GetOfficeHoursScheduleWithResponse(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("get office hours schedule", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

// UpdateSchedule updates an office-hours schedule.
func (s *OfficeHoursService) UpdateSchedule(ctx context.Context, id string, schedule OfficeHoursScheduleUpdate) (*OfficeHoursSchedule, error) {
	res, err := s.client.generated.UpdateOfficeHoursScheduleWithResponse(ctx, id, nil, schedule)
	if err != nil {
		return nil, err
	}
	return requireOK("update office hours schedule", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

// DeleteSchedule removes an office-hours schedule.
func (s *OfficeHoursService) DeleteSchedule(ctx context.Context, id string) error {
	res, err := s.client.generated.DeleteOfficeHoursScheduleWithResponse(ctx, id, nil)
	if err != nil {
		return err
	}
	return requireEmpty(res.StatusCode(), res.Body, responseHeaders(res.HTTPResponse))
}

// ListExceptions returns exceptions for a schedule.
func (s *OfficeHoursService) ListExceptions(ctx context.Context, scheduleID string, params *OfficeHoursExceptionListParams) (*OfficeHoursExceptionList, error) {
	res, err := s.client.generated.ListOfficeHoursExceptionsWithResponse(ctx, scheduleID, params)
	if err != nil {
		return nil, err
	}
	return requireOK("list office hours exceptions", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

// CreateException adds an exception to a schedule.
func (s *OfficeHoursService) CreateException(ctx context.Context, scheduleID string, exception OfficeHoursExceptionCreate) (*OfficeHoursException, error) {
	res, err := s.client.generated.CreateOfficeHoursExceptionWithResponse(ctx, scheduleID, nil, exception)
	if err != nil {
		return nil, err
	}
	return requireCreated("create office hours exception", res.StatusCode(), res.Body, res.JSON201, responseHeaders(res.HTTPResponse))
}

// GetException returns a schedule exception by ID.
func (s *OfficeHoursService) GetException(ctx context.Context, scheduleID, id string) (*OfficeHoursException, error) {
	res, err := s.client.generated.GetOfficeHoursExceptionWithResponse(ctx, scheduleID, id, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("get office hours exception", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

// UpdateException updates a schedule exception.
func (s *OfficeHoursService) UpdateException(ctx context.Context, scheduleID, id string, exception OfficeHoursExceptionUpdate) (*OfficeHoursException, error) {
	res, err := s.client.generated.UpdateOfficeHoursExceptionWithResponse(ctx, scheduleID, id, nil, exception)
	if err != nil {
		return nil, err
	}
	return requireOK("update office hours exception", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

// DeleteException removes a schedule exception.
func (s *OfficeHoursService) DeleteException(ctx context.Context, scheduleID, id string) error {
	res, err := s.client.generated.DeleteOfficeHoursExceptionWithResponse(ctx, scheduleID, id, nil)
	if err != nil {
		return err
	}
	return requireEmpty(res.StatusCode(), res.Body, responseHeaders(res.HTTPResponse))
}
