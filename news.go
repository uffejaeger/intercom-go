package intercom

import (
	"context"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// NewsItem is an Intercom news item.
type NewsItem = gen.NewsItemSchema

// NewsListType identifies the type of a paginated news response.
type NewsListType = gen.PaginatedResponseType

// NewsCursorPages holds pagination details for a paginated news response.
type NewsCursorPages = gen.CursorPagesSchema

// NewsItemList is a paginated list of Intercom news items.
type NewsItemList struct {
	Data       *[]NewsItem      `json:"data,omitempty"`
	Pages      *NewsCursorPages `json:"pages,omitempty"`
	TotalCount *int             `json:"total_count,omitempty"`
	Type       *NewsListType    `json:"type,omitempty"`
}

// NewsItemCreate holds the fields for creating a news item.
type NewsItemCreate = gen.NewsItemRequestSchema

// NewsItemUpdate holds the fields for updating a news item.
type NewsItemUpdate = gen.NewsItemRequestSchema

// Newsfeed is an Intercom newsfeed.
type Newsfeed = gen.NewsfeedSchema

// NewsfeedList is a paginated list of Intercom newsfeeds.
type NewsfeedList struct {
	Data       *[]Newsfeed      `json:"data,omitempty"`
	Pages      *NewsCursorPages `json:"pages,omitempty"`
	TotalCount *int             `json:"total_count,omitempty"`
	Type       *NewsListType    `json:"type,omitempty"`
}

// NewsService exposes news-related Intercom API operations.
type NewsService struct {
	client *Client
}

// ListItems returns all news items.
func (s *NewsService) ListItems(ctx context.Context) (*NewsItemList, error) {
	res, err := s.client.generated.ListNewsItemsWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return requireJSON[NewsItemList]("list news items", res.StatusCode(), res.Body)
}

// CreateItem creates a news item.
func (s *NewsService) CreateItem(ctx context.Context, item NewsItemCreate) (*NewsItem, error) {
	res, err := s.client.generated.CreateNewsItemWithResponse(ctx, nil, gen.CreateNewsItemJSONRequestBody(item))
	if err != nil {
		return nil, err
	}
	return requireOK("create news item", res.StatusCode(), res.Body, res.JSON200)
}

// RetrieveItem returns a news item by ID.
func (s *NewsService) RetrieveItem(ctx context.Context, newsItemID string) (*NewsItem, error) {
	id, err := requireIntID("news item", newsItemID)
	if err != nil {
		return nil, err
	}
	res, err := s.client.generated.RetrieveNewsItemWithResponse(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("retrieve news item", res.StatusCode(), res.Body, res.JSON200)
}

// UpdateItem updates a news item by ID.
func (s *NewsService) UpdateItem(ctx context.Context, newsItemID string, item NewsItemUpdate) (*NewsItem, error) {
	id, err := requireIntID("news item", newsItemID)
	if err != nil {
		return nil, err
	}
	res, err := s.client.generated.UpdateNewsItemWithResponse(ctx, id, nil, gen.UpdateNewsItemJSONRequestBody(item))
	if err != nil {
		return nil, err
	}
	return requireOK("update news item", res.StatusCode(), res.Body, res.JSON200)
}

// DeleteItem deletes a news item by ID.
func (s *NewsService) DeleteItem(ctx context.Context, newsItemID string) error {
	id, err := requireIntID("news item", newsItemID)
	if err != nil {
		return err
	}
	res, err := s.client.generated.DeleteNewsItemWithResponse(ctx, id, nil)
	if err != nil {
		return err
	}
	return requireEmpty(res.StatusCode(), res.Body)
}

// ListFeeds returns all newsfeeds.
func (s *NewsService) ListFeeds(ctx context.Context) (*NewsfeedList, error) {
	res, err := s.client.generated.ListNewsfeedsWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return requireJSON[NewsfeedList]("list newsfeeds", res.StatusCode(), res.Body)
}

// RetrieveFeed returns a newsfeed by ID.
func (s *NewsService) RetrieveFeed(ctx context.Context, newsfeedID string) (*Newsfeed, error) {
	if newsfeedID == "" {
		return nil, errRequiredID("newsfeed")
	}
	res, err := s.client.generated.RetrieveNewsfeedWithResponse(ctx, newsfeedID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("retrieve newsfeed", res.StatusCode(), res.Body, res.JSON200)
}

// ListFeedItems returns live items for a newsfeed.
func (s *NewsService) ListFeedItems(ctx context.Context, newsfeedID string) (*NewsItemList, error) {
	if newsfeedID == "" {
		return nil, errRequiredID("newsfeed")
	}
	res, err := s.client.generated.ListLiveNewsfeedItemsWithResponse(ctx, newsfeedID, nil)
	if err != nil {
		return nil, err
	}
	return requireJSON[NewsItemList]("list newsfeed items", res.StatusCode(), res.Body)
}
