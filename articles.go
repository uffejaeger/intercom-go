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
type ArticleVersion = gen.ArticleVersionSchema
type ArticleVersionList = gen.ArticleVersionListSchema
type ArticleTag = gen.AttachTagToArticleJSONRequestBody
type ArticleDraftUpdate = gen.UpdateArticleRequestSchema
type ArticleDraftPublish = gen.PublishArticleDraftRequestSchema

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
	return requireOK("list articles", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

// Create creates a new article.
func (s *ArticlesService) Create(ctx context.Context, article ArticleCreate) (*Article, error) {
	res, err := s.client.generated.CreateArticleWithResponse(ctx, nil, gen.CreateArticleJSONRequestBody(article))
	if err != nil {
		return nil, err
	}
	return requireOK("create article", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
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
	return requireOK("retrieve article", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
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
	return requireOK("update article", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
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
	return requireOK("delete article", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
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
	return requireOK("search articles", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

// AttachTag attaches a tag to an article.
func (s *ArticlesService) AttachTag(ctx context.Context, articleID string, tag ArticleTag) (*Tag, error) {
	id, err := requireIntID("article", articleID)
	if err != nil {
		return nil, err
	}
	res, err := s.client.generated.AttachTagToArticleWithResponse(ctx, id, nil, tag)
	if err != nil {
		return nil, err
	}
	return requireOK("attach tag to article", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

// DetachTag removes a tag from an article.
func (s *ArticlesService) DetachTag(ctx context.Context, articleID, tagID string) (*Tag, error) {
	id, err := requireIntID("article", articleID)
	if err != nil {
		return nil, err
	}
	res, err := s.client.generated.DetachTagFromArticleWithResponse(ctx, id, tagID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("detach tag from article", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

// ListVersions returns versions of an article.
func (s *ArticlesService) ListVersions(ctx context.Context, articleID string, params *gen.ListArticleVersionsParams) (*ArticleVersionList, error) {
	id, err := requireIntID("article", articleID)
	if err != nil {
		return nil, err
	}
	res, err := s.client.generated.ListArticleVersionsWithResponse(ctx, id, params)
	if err != nil {
		return nil, err
	}
	return requireOK("list article versions", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

// GetVersion returns an article version.
func (s *ArticlesService) GetVersion(ctx context.Context, articleID, versionID string) (*ArticleVersion, error) {
	id, err := requireIntID("article", articleID)
	if err != nil {
		return nil, err
	}
	res, err := s.client.generated.RetrieveArticleVersionWithResponse(ctx, id, versionID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("get article version", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

// GetDraft returns an article draft.
func (s *ArticlesService) GetDraft(ctx context.Context, articleID string) (*Article, error) {
	id, err := requireIntID("article", articleID)
	if err != nil {
		return nil, err
	}
	res, err := s.client.generated.RetrieveArticleDraftWithResponse(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("get article draft", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

// StageDraft updates an article draft without publishing it.
func (s *ArticlesService) StageDraft(ctx context.Context, articleID string, draft ArticleDraftUpdate) (*Article, error) {
	id, err := requireIntID("article", articleID)
	if err != nil {
		return nil, err
	}
	res, err := s.client.generated.StageArticleDraftWithResponse(ctx, id, nil, draft)
	if err != nil {
		return nil, err
	}
	return requireOK("stage article draft", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

// PublishDraft publishes an article draft.
func (s *ArticlesService) PublishDraft(ctx context.Context, articleID string, draft ArticleDraftPublish) (*Article, error) {
	id, err := requireIntID("article", articleID)
	if err != nil {
		return nil, err
	}
	res, err := s.client.generated.PublishArticleDraftWithResponse(ctx, id, nil, draft)
	if err != nil {
		return nil, err
	}
	return requireOK("publish article draft", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}
