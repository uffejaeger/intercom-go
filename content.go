package intercom

import (
	"context"
	"net/http"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// ContentSnippet is reusable knowledge content for an AI agent or Copilot.
type ContentSnippet = gen.ContentSnippetSchema
type ContentSnippetList = gen.ContentSnippetListSchema
type ContentSearchResult = gen.ContentSearchResponseSchema
type ContentBulkAction = gen.ContentBulkActionResponseSchema
type ContentSnippetCreate = gen.ContentSnippetCreateRequestSchema
type ContentSnippetUpdate = gen.ContentSnippetUpdateRequestSchema
type ContentBulkActionRequest = gen.ContentBulkActionRequestSchema
type ContentSnippetListParams = gen.ListContentSnippetsParams
type ContentSearchParams = gen.SearchContentParams
type ContentSnippetTag = gen.AttachTagToContentSnippetJSONRequestBody

// ContentBulkActionContentType identifies the kind of content selected by a bulk action.
type ContentBulkActionContentType = gen.ContentBulkActionRequestContentIdsType

// ContentBulkActionContentID identifies one item selected by a bulk action.
type ContentBulkActionContentID = struct {
	Id   string                       `json:"id"`
	Type ContentBulkActionContentType `json:"type"`
}

// ContentBulkActionAudience configures segments for a set-audience bulk action.
type ContentBulkActionAudience = struct {
	AddSegmentIds    *[]int `json:"add_segment_ids,omitempty"`
	RemoveAll        *bool  `json:"remove_all,omitempty"`
	RemoveSegmentIds *[]int `json:"remove_segment_ids,omitempty"`
}

// ContentBulkActionAvailability configures availability for a set-availability bulk action.
type ContentBulkActionAvailability = struct {
	AiAgent    *bool `json:"ai_agent,omitempty"`
	Copilot    *bool `json:"copilot,omitempty"`
	SalesAgent *bool `json:"sales_agent,omitempty"`
}

// ContentBulkActionTags configures tags for an update-tags bulk action.
type ContentBulkActionTags = struct {
	AddTagIds    *[]int `json:"add_tag_ids,omitempty"`
	RemoveTagIds *[]int `json:"remove_tag_ids,omitempty"`
}

// ContentService exposes Knowledge Hub content and content-snippet operations.
type ContentService struct{ client *Client }

func (s *ContentService) Search(ctx context.Context, params *ContentSearchParams) (*ContentSearchResult, error) {
	res, err := s.client.generated.SearchContentWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	return requireOK("search content", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

func (s *ContentService) BulkAction(ctx context.Context, action ContentBulkActionRequest) (*ContentBulkAction, error) {
	res, err := s.client.generated.BulkContentActionsWithResponse(ctx, nil, action)
	if err != nil {
		return nil, err
	}
	return requireStatus("bulk content action", res.StatusCode(), http.StatusAccepted, res.Body, res.JSON202, responseHeaders(res.HTTPResponse))
}

func (s *ContentService) ListSnippets(ctx context.Context, params *ContentSnippetListParams) (*ContentSnippetList, error) {
	res, err := s.client.generated.ListContentSnippetsWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	return requireOK("list content snippets", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

func (s *ContentService) CreateSnippet(ctx context.Context, snippet ContentSnippetCreate) (*ContentSnippet, error) {
	res, err := s.client.generated.CreateContentSnippetWithResponse(ctx, nil, snippet)
	if err != nil {
		return nil, err
	}
	return requireCreated("create content snippet", res.StatusCode(), res.Body, res.JSON201, responseHeaders(res.HTTPResponse))
}

func (s *ContentService) GetSnippet(ctx context.Context, id string) (*ContentSnippet, error) {
	res, err := s.client.generated.GetContentSnippetWithResponse(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("get content snippet", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

func (s *ContentService) UpdateSnippet(ctx context.Context, id string, snippet ContentSnippetUpdate) (*ContentSnippet, error) {
	res, err := s.client.generated.UpdateContentSnippetWithResponse(ctx, id, nil, snippet)
	if err != nil {
		return nil, err
	}
	return requireOK("update content snippet", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

func (s *ContentService) DeleteSnippet(ctx context.Context, id string) error {
	res, err := s.client.generated.DeleteContentSnippetWithResponse(ctx, id, nil)
	if err != nil {
		return err
	}
	return requireEmpty(res.StatusCode(), res.Body, responseHeaders(res.HTTPResponse))
}

func (s *ContentService) AttachSnippetTag(ctx context.Context, id string, tag ContentSnippetTag) (*Tag, error) {
	res, err := s.client.generated.AttachTagToContentSnippetWithResponse(ctx, id, nil, tag)
	if err != nil {
		return nil, err
	}
	return requireOK("attach tag to content snippet", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

func (s *ContentService) DetachSnippetTag(ctx context.Context, id, tagID string) error {
	res, err := s.client.generated.DetachTagFromContentSnippetWithResponse(ctx, id, tagID, nil)
	if err != nil {
		return err
	}
	return requireEmpty(res.StatusCode(), res.Body, responseHeaders(res.HTTPResponse))
}
