package intercom

import (
	"context"
	"fmt"
	"strconv"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// Article is an Intercom article.
type Article = gen.ArticleSchema

// ArticleList is a page of Intercom articles.
type ArticleList = gen.ArticleListSchema

// ArticleSearchResult is the result of an article search.
type ArticleSearchResult = gen.ArticleSearchResponseSchema

// ArticleDeleted is the result of deleting an article.
type ArticleDeleted = gen.DeletedArticleObjectSchema

// ArticleCreate holds the fields for creating an article.
type ArticleCreate = gen.CreateArticleRequestSchema

// ArticleUpdate holds the fields for updating an article.
type ArticleUpdate = gen.UpdateArticleRequestSchema

// ArticleSearch holds the parameters for searching articles.
type ArticleSearch struct {
	Phrase       string
	State        string
	HelpCenterID int
	Highlight    *bool
}

// ArticlesService exposes article-related Intercom API operations.
type ArticlesService struct {
	client *Client
}

// List returns all articles.
func (s *ArticlesService) List(ctx context.Context) (*ArticleList, error) {
	res, err := s.client.generated.ListArticlesWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("list articles", res.StatusCode(), res.Body, res.JSON200)
}

// Create creates a new article.
func (s *ArticlesService) Create(ctx context.Context, article ArticleCreate) (*Article, error) {
	res, err := s.client.generated.CreateArticleWithResponse(ctx, nil, gen.CreateArticleJSONRequestBody(article))
	if err != nil {
		return nil, err
	}
	return requireOK("create article", res.StatusCode(), res.Body, res.JSON200)
}

// Retrieve retrieves an article by ID.
func (s *ArticlesService) Retrieve(ctx context.Context, articleID string) (*Article, error) {
	if articleID == "" {
		return nil, fmt.Errorf("intercom: article ID is required")
	}
	id, err := strconv.Atoi(articleID)
	if err != nil {
		return nil, fmt.Errorf("intercom: article ID %q is not a valid integer: %w", articleID, err)
	}
	res, err := s.client.generated.RetrieveArticleWithResponse(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("retrieve article", res.StatusCode(), res.Body, res.JSON200)
}

// Update updates an article.
func (s *ArticlesService) Update(ctx context.Context, articleID string, article ArticleUpdate) (*Article, error) {
	if articleID == "" {
		return nil, fmt.Errorf("intercom: article ID is required")
	}
	id, err := strconv.Atoi(articleID)
	if err != nil {
		return nil, fmt.Errorf("intercom: article ID %q is not a valid integer: %w", articleID, err)
	}
	res, err := s.client.generated.UpdateArticleWithResponse(ctx, id, nil, gen.UpdateArticleJSONRequestBody(article))
	if err != nil {
		return nil, err
	}
	return requireOK("update article", res.StatusCode(), res.Body, res.JSON200)
}

// Delete deletes an article.
func (s *ArticlesService) Delete(ctx context.Context, articleID string) (*ArticleDeleted, error) {
	if articleID == "" {
		return nil, fmt.Errorf("intercom: article ID is required")
	}
	id, err := strconv.Atoi(articleID)
	if err != nil {
		return nil, fmt.Errorf("intercom: article ID %q is not a valid integer: %w", articleID, err)
	}
	res, err := s.client.generated.DeleteArticleWithResponse(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("delete article", res.StatusCode(), res.Body, res.JSON200)
}

// Search searches articles by phrase, state, help center, or highlight options.
func (s *ArticlesService) Search(ctx context.Context, search ArticleSearch) (*ArticleSearchResult, error) {
	params := &gen.SearchArticlesParams{}
	if search.Phrase != "" {
		params.Phrase = &search.Phrase
	}
	if search.State != "" {
		params.State = &search.State
	}
	if search.HelpCenterID != 0 {
		params.HelpCenterId = &search.HelpCenterID
	}
	if search.Highlight != nil {
		params.Highlight = search.Highlight
	}
	res, err := s.client.generated.SearchArticlesWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	return requireOK("search articles", res.StatusCode(), res.Body, res.JSON200)
}
