package intercom

import (
	"context"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// InternalArticle is an Intercom internal article.
type InternalArticle = gen.InternalArticleSchema

// InternalArticleList is a list of internal articles.
type InternalArticleList = gen.InternalArticleListSchema

// InternalArticleSearchResult is the result of an internal article search.
type InternalArticleSearchResult = gen.InternalArticleSearchResponseSchema

// InternalArticleDeleted is the result of deleting an internal article.
type InternalArticleDeleted = gen.DeletedInternalArticleObjectSchema

// InternalArticleCreate holds the fields for creating an internal article.
type InternalArticleCreate = gen.CreateInternalArticleRequestSchema

// InternalArticleUpdate holds the fields for updating an internal article.
type InternalArticleUpdate = gen.UpdateInternalArticleRequestSchema

// InternalArticlesService exposes internal-article-related Intercom API operations.
type InternalArticlesService struct {
	client *Client
}

// List returns all internal articles.
func (s *InternalArticlesService) List(ctx context.Context) (*InternalArticleList, error) {
	res, err := s.client.generated.ListInternalArticlesWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("list internal articles", res.StatusCode(), res.Body, res.JSON200)
}

// Create creates an internal article.
func (s *InternalArticlesService) Create(ctx context.Context, article InternalArticleCreate) (*InternalArticle, error) {
	res, err := s.client.generated.CreateInternalArticleWithResponse(ctx, nil, gen.CreateInternalArticleJSONRequestBody(article))
	if err != nil {
		return nil, err
	}
	return requireOK("create internal article", res.StatusCode(), res.Body, res.JSON200)
}

// Search searches internal articles.
func (s *InternalArticlesService) Search(ctx context.Context, folderID *string) (*InternalArticleSearchResult, error) {
	res, err := s.client.generated.SearchInternalArticlesWithResponse(ctx, &gen.SearchInternalArticlesParams{FolderId: folderID})
	if err != nil {
		return nil, err
	}
	return requireOK("search internal articles", res.StatusCode(), res.Body, res.JSON200)
}

// Retrieve returns an internal article by ID.
func (s *InternalArticlesService) Retrieve(ctx context.Context, articleID string) (*InternalArticle, error) {
	id, err := requireIntID("internal article", articleID)
	if err != nil {
		return nil, err
	}
	res, err := s.client.generated.RetrieveInternalArticleWithResponse(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("retrieve internal article", res.StatusCode(), res.Body, res.JSON200)
}

// Update updates an internal article by ID.
func (s *InternalArticlesService) Update(ctx context.Context, articleID string, article InternalArticleUpdate) (*InternalArticle, error) {
	id, err := requireIntID("internal article", articleID)
	if err != nil {
		return nil, err
	}
	res, err := s.client.generated.UpdateInternalArticleWithResponse(ctx, id, nil, gen.UpdateInternalArticleJSONRequestBody(article))
	if err != nil {
		return nil, err
	}
	return requireOK("update internal article", res.StatusCode(), res.Body, res.JSON200)
}

// Delete deletes an internal article by ID.
func (s *InternalArticlesService) Delete(ctx context.Context, articleID string) (*InternalArticleDeleted, error) {
	id, err := requireIntID("internal article", articleID)
	if err != nil {
		return nil, err
	}
	res, err := s.client.generated.DeleteInternalArticleWithResponse(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("delete internal article", res.StatusCode(), res.Body, res.JSON200)
}
