package intercom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// TagDetail is a single Intercom tag returned by the tag lookup endpoint.
type TagDetail = gen.TagBasicSchema

// TagsService exposes tag-related Intercom API operations.
type TagsService struct {
	client *Client
}

// List returns all tags for the workspace.
func (s *TagsService) List(ctx context.Context) (*TagList, error) {
	res, err := s.client.generated.ListTagsWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("list tags", res.StatusCode(), res.Body, res.JSON200)
}

// Create creates or updates a tag, or applies tag operations supported by the endpoint.
func (s *TagsService) Create(ctx context.Context, req any) (*TagDetail, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("intercom: marshal tag request: %w", err)
	}
	res, err := s.client.generated.CreateTagWithBodyWithResponse(ctx, nil, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	return requireOK("create tag", res.StatusCode(), res.Body, res.JSON200)
}

// Retrieve returns a tag by ID.
func (s *TagsService) Retrieve(ctx context.Context, tagID string) (*TagDetail, error) {
	if tagID == "" {
		return nil, fmt.Errorf("intercom: tag ID is required")
	}
	res, err := s.client.generated.FindTagWithResponse(ctx, tagID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("retrieve tag", res.StatusCode(), res.Body, res.JSON200)
}

// Delete deletes a tag by ID.
func (s *TagsService) Delete(ctx context.Context, tagID string) error {
	if tagID == "" {
		return fmt.Errorf("intercom: tag ID is required")
	}
	res, err := s.client.generated.DeleteTagWithResponse(ctx, tagID, nil)
	if err != nil {
		return err
	}
	return requireEmpty(res.StatusCode(), res.Body)
}
