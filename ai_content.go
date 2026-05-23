package intercom

import (
	"context"
	"fmt"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// ContentImportSource is an Intercom AI content import source.
type ContentImportSource = gen.ContentImportSourceSchema

// ContentImportSourceList is a list of content import sources.
type ContentImportSourceList = gen.ContentImportSourcesListSchema

// ContentImportSourceCreate holds the fields for creating a content import source.
type ContentImportSourceCreate = gen.CreateContentImportSourceRequestSchema

// ContentImportSourceUpdate holds the fields for updating a content import source.
type ContentImportSourceUpdate = gen.UpdateContentImportSourceRequestSchema

// ExternalPage is an Intercom AI external page.
type ExternalPage = gen.ExternalPageSchema

// ExternalPageList is a list of external pages.
type ExternalPageList = gen.ExternalPagesListSchema

// ExternalPageCreate holds the fields for creating an external page.
type ExternalPageCreate = gen.CreateExternalPageRequestSchema

// ExternalPageUpdate holds the fields for updating an external page.
type ExternalPageUpdate = gen.UpdateExternalPageRequestSchema

// AIContentService exposes AI content import source and external page Intercom API operations.
type AIContentService struct {
	client *Client
}

// ListContentImportSources returns all content import sources.
func (s *AIContentService) ListContentImportSources(ctx context.Context) (*ContentImportSourceList, error) {
	res, err := s.client.generated.ListContentImportSourcesWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("list content import sources", res.StatusCode(), res.Body, res.JSON200)
}

// CreateContentImportSource creates a new content import source.
func (s *AIContentService) CreateContentImportSource(ctx context.Context, req ContentImportSourceCreate) (*ContentImportSource, error) {
	res, err := s.client.generated.CreateContentImportSourceWithResponse(ctx, nil, gen.CreateContentImportSourceJSONRequestBody(req))
	if err != nil {
		return nil, err
	}
	return requireOK("create content import source", res.StatusCode(), res.Body, res.JSON200)
}

// GetContentImportSource retrieves a content import source by ID.
func (s *AIContentService) GetContentImportSource(ctx context.Context, sourceID string) (*ContentImportSource, error) {
	if sourceID == "" {
		return nil, fmt.Errorf("intercom: source ID is required")
	}
	res, err := s.client.generated.GetContentImportSourceWithResponse(ctx, sourceID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("get content import source", res.StatusCode(), res.Body, res.JSON200)
}

// UpdateContentImportSource updates a content import source.
func (s *AIContentService) UpdateContentImportSource(ctx context.Context, sourceID string, req ContentImportSourceUpdate) (*ContentImportSource, error) {
	if sourceID == "" {
		return nil, fmt.Errorf("intercom: source ID is required")
	}
	res, err := s.client.generated.UpdateContentImportSourceWithResponse(ctx, sourceID, nil, gen.UpdateContentImportSourceJSONRequestBody(req))
	if err != nil {
		return nil, err
	}
	return requireOK("update content import source", res.StatusCode(), res.Body, res.JSON200)
}

// DeleteContentImportSource deletes a content import source.
func (s *AIContentService) DeleteContentImportSource(ctx context.Context, sourceID string) error {
	if sourceID == "" {
		return fmt.Errorf("intercom: source ID is required")
	}
	res, err := s.client.generated.DeleteContentImportSourceWithResponse(ctx, sourceID, nil)
	if err != nil {
		return err
	}
	return requireEmpty(res.StatusCode(), res.Body)
}

// ListExternalPages returns all external pages.
func (s *AIContentService) ListExternalPages(ctx context.Context) (*ExternalPageList, error) {
	res, err := s.client.generated.ListExternalPagesWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("list external pages", res.StatusCode(), res.Body, res.JSON200)
}

// CreateExternalPage creates a new external page.
func (s *AIContentService) CreateExternalPage(ctx context.Context, req ExternalPageCreate) (*ExternalPage, error) {
	res, err := s.client.generated.CreateExternalPageWithResponse(ctx, nil, gen.CreateExternalPageJSONRequestBody(req))
	if err != nil {
		return nil, err
	}
	return requireOK("create external page", res.StatusCode(), res.Body, res.JSON200)
}

// GetExternalPage retrieves an external page by ID.
func (s *AIContentService) GetExternalPage(ctx context.Context, pageID string) (*ExternalPage, error) {
	if pageID == "" {
		return nil, fmt.Errorf("intercom: page ID is required")
	}
	res, err := s.client.generated.GetExternalPageWithResponse(ctx, pageID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("get external page", res.StatusCode(), res.Body, res.JSON200)
}

// UpdateExternalPage updates an external page.
func (s *AIContentService) UpdateExternalPage(ctx context.Context, pageID string, req ExternalPageUpdate) (*ExternalPage, error) {
	if pageID == "" {
		return nil, fmt.Errorf("intercom: page ID is required")
	}
	res, err := s.client.generated.UpdateExternalPageWithResponse(ctx, pageID, nil, gen.UpdateExternalPageJSONRequestBody(req))
	if err != nil {
		return nil, err
	}
	return requireOK("update external page", res.StatusCode(), res.Body, res.JSON200)
}

// DeleteExternalPage deletes an external page.
func (s *AIContentService) DeleteExternalPage(ctx context.Context, pageID string) (*ExternalPage, error) {
	if pageID == "" {
		return nil, fmt.Errorf("intercom: page ID is required")
	}
	res, err := s.client.generated.DeleteExternalPageWithResponse(ctx, pageID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("delete external page", res.StatusCode(), res.Body, res.JSON200)
}
