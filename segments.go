package intercom

import (
	"context"
	"fmt"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// Segment is an Intercom segment.
type Segment = gen.SegmentSchema

// SegmentList is a list of Intercom segments.
type SegmentList = gen.SegmentListSchema

// SegmentsService exposes segment-related Intercom API operations.
type SegmentsService struct {
	client *Client
}

// List returns all segments for the workspace.
func (s *SegmentsService) List(ctx context.Context) (*SegmentList, error) {
	res, err := s.client.generated.ListSegmentsWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("list segments", res.StatusCode(), res.Body, res.JSON200)
}

// Retrieve returns a segment by ID.
func (s *SegmentsService) Retrieve(ctx context.Context, segmentID string) (*Segment, error) {
	if segmentID == "" {
		return nil, fmt.Errorf("intercom: segment ID is required")
	}
	res, err := s.client.generated.RetrieveSegmentWithResponse(ctx, segmentID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("retrieve segment", res.StatusCode(), res.Body, res.JSON200)
}
