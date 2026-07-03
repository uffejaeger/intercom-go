package intercom

import (
	"context"
	"fmt"
	"io"
	"net/url"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// DataEvent is an Intercom data event.
type DataEvent = gen.DataEventSchema

// DataEventSummary is an Intercom data event summary response.
type DataEventSummary = gen.DataEventSummarySchema

// DataEventSummaryItem is a summarized data event.
type DataEventSummaryItem = gen.DataEventSummaryItemSchema

// DataEventListFilter identifies which user's events to query.
type DataEventListFilter struct {
	UserID         string
	IntercomUserID string
	Email          string
	Summary        bool
}

// DataEventCreate holds the fields for submitting a data event.
type DataEventCreate struct {
	EventName string
	CreatedAt *int
	UserID    *string
	Email     *string
	ID        *string
	Metadata  map[string]any
}

// DataEventSummaryCreate holds the fields for a single summary item.
type DataEventSummaryCreate struct {
	EventName string
	Count     int
	First     *int
	Last      *int
}

// DataEventSummariesCreate holds the fields for creating event summaries.
type DataEventSummariesCreate struct {
	UserID         string
	EventSummaries []DataEventSummaryCreate
}

// DataEventsService exposes Intercom data-event API operations.
type DataEventsService struct {
	client *Client
}

// List returns data event summaries for a user or lead.
func (s *DataEventsService) List(ctx context.Context, filter DataEventListFilter) (*DataEventSummary, error) {
	query, err := dataEventListQuery(filter)
	if err != nil {
		return nil, err
	}

	req, _ := s.client.NewRequest(ctx, "GET", "/events?"+query.Encode(), nil)

	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return requireJSON[DataEventSummary]("list data events", res.StatusCode, body)
}

// Create submits a data event.
func (s *DataEventsService) Create(ctx context.Context, event DataEventCreate) error {
	if event.EventName == "" {
		return fmt.Errorf("intercom: data event name is required")
	}
	if event.CreatedAt == nil {
		return fmt.Errorf("intercom: data event created_at is required")
	}
	if err := validateDataEventIdentifiers(event.UserID, event.Email, event.ID); err != nil {
		return err
	}

	body, _ := marshalBody(struct {
		EventName string         `json:"event_name"`
		CreatedAt *int           `json:"created_at,omitempty"`
		UserID    *string        `json:"user_id,omitempty"`
		Email     *string        `json:"email,omitempty"`
		ID        *string        `json:"id,omitempty"`
		Metadata  map[string]any `json:"metadata,omitempty"`
	}{
		EventName: event.EventName,
		CreatedAt: event.CreatedAt,
		UserID:    event.UserID,
		Email:     event.Email,
		ID:        event.ID,
		Metadata:  event.Metadata,
	})

	res, err := s.client.generated.CreateDataEventWithBodyWithResponse(ctx, nil, "application/json", body)
	if err != nil {
		return err
	}

	return requireEmpty(res.StatusCode(), res.Body)
}

// CreateSummaries submits summarized data events for a user.
func (s *DataEventsService) CreateSummaries(ctx context.Context, summaries DataEventSummariesCreate) error {
	if summaries.UserID == "" {
		return fmt.Errorf("intercom: data event summary user ID is required")
	}
	if len(summaries.EventSummaries) == 0 {
		return fmt.Errorf("intercom: at least one data event summary is required")
	}

	items := make([]struct {
		EventName string `json:"event_name"`
		Count     int    `json:"count"`
		First     *int   `json:"first,omitempty"`
		Last      *int   `json:"last,omitempty"`
	}, 0, len(summaries.EventSummaries))
	for _, summary := range summaries.EventSummaries {
		if summary.EventName == "" {
			return fmt.Errorf("intercom: data event summary name is required")
		}
		items = append(items, struct {
			EventName string `json:"event_name"`
			Count     int    `json:"count"`
			First     *int   `json:"first,omitempty"`
			Last      *int   `json:"last,omitempty"`
		}{
			EventName: summary.EventName,
			Count:     summary.Count,
			First:     summary.First,
			Last:      summary.Last,
		})
	}

	body, _ := marshalBody(struct {
		UserID         string `json:"user_id"`
		EventSummaries any    `json:"event_summaries"`
	}{
		UserID:         summaries.UserID,
		EventSummaries: items,
	})

	res, err := s.client.generated.DataEventSummariesWithBodyWithResponse(ctx, nil, "application/json", body)
	if err != nil {
		return err
	}

	return requireEmpty(res.StatusCode(), res.Body)
}

func dataEventListQuery(filter DataEventListFilter) (url.Values, error) {
	values := url.Values{}
	values.Set("type", "user")
	if filter.Summary {
		values.Set("summary", "true")
	}

	identifiers := 0
	if filter.UserID != "" {
		values.Set("user_id", filter.UserID)
		identifiers++
	}
	if filter.IntercomUserID != "" {
		values.Set("intercom_user_id", filter.IntercomUserID)
		identifiers++
	}
	if filter.Email != "" {
		values.Set("email", filter.Email)
		identifiers++
	}
	if identifiers != 1 {
		return nil, fmt.Errorf("intercom: exactly one data event filter identifier is required")
	}

	return values, nil
}

func validateDataEventIdentifiers(userID, email, id *string) error {
	identifiers := 0
	if userID != nil && *userID != "" {
		identifiers++
	}
	if email != nil && *email != "" {
		identifiers++
	}
	if id != nil && *id != "" {
		identifiers++
	}
	if identifiers != 1 {
		return fmt.Errorf("intercom: exactly one data event identifier is required")
	}
	return nil
}
