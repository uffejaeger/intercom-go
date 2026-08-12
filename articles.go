package intercom

import (
	"context"
	"fmt"
	"strconv"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// Article is an Intercom article.
//
// ParentID and ParentType remain available for source compatibility with
// earlier SDK releases. API 2.16 provides ParentIds instead; ParentID is the
// first parent ID when one is present and ParentType is unavailable from the
// new response representation.
type Article struct {
	AiChatbotAvailability         *bool                                  `json:"ai_chatbot_availability,omitempty"`
	AiCopilotAvailability         *bool                                  `json:"ai_copilot_availability,omitempty"`
	AiSalesAgentAvailability      *bool                                  `json:"ai_sales_agent_availability,omitempty"`
	AuthorId                      *int                                   `json:"author_id,omitempty"`
	Body                          *string                                `json:"body,omitempty"`
	BodyMarkdown                  *string                                `json:"body_markdown,omitempty"`
	CreatedAt                     *int                                   `json:"created_at,omitempty"`
	CreatedById                   *int                                   `json:"created_by_id,omitempty"`
	DefaultLocale                 *string                                `json:"default_locale,omitempty"`
	Description                   *string                                `json:"description,omitempty"`
	DraftUpdatedAt                *int                                   `json:"draft_updated_at,omitempty"`
	ExcludeFromArticleSuggestions *bool                                  `json:"exclude_from_article_suggestions,omitempty"`
	HasUnpublishedChanges         *bool                                  `json:"has_unpublished_changes,omitempty"`
	HelpCenterAudience            *gen.ArticleListItemHelpCenterAudience `json:"help_center_audience,omitempty"`
	Id                            *string                                `json:"id,omitempty"`
	ParentId                      *int                                   `json:"parent_id,omitempty"`
	ParentIds                     *[]int                                 `json:"parent_ids,omitempty"`
	ParentType                    *string                                `json:"parent_type,omitempty"`
	ScheduledPublishAt            *int                                   `json:"scheduled_publish_at,omitempty"`
	ScheduledUnpublishAt          *int                                   `json:"scheduled_unpublish_at,omitempty"`
	State                         *gen.ArticleListItemState              `json:"state,omitempty"`
	Tags                          *gen.TagsSchema                        `json:"tags,omitempty"`
	Title                         *string                                `json:"title,omitempty"`
	TranslatedContent             *gen.ArticleTranslatedContentSchema    `json:"translated_content,omitempty"`
	Type                          *gen.ArticleListItemType               `json:"type,omitempty"`
	UpdatedAt                     *int                                   `json:"updated_at,omitempty"`
	UpdatedById                   *int                                   `json:"updated_by_id,omitempty"`
	Url                           *string                                `json:"url,omitempty"`
	WorkspaceId                   *string                                `json:"workspace_id,omitempty"`
}

// ArticleList is a page of Intercom articles.
type ArticleList struct {
	Data       *[]Article             `json:"data,omitempty"`
	Pages      *gen.CursorPagesSchema `json:"pages,omitempty"`
	TotalCount *int                   `json:"total_count,omitempty"`
	Type       *gen.ArticleListType   `json:"type,omitempty"`
}

// ArticleSearchResult is the result of an article search.
type ArticleSearchResult struct {
	Data       *ArticleSearchData             `json:"data,omitempty"`
	Pages      *gen.CursorPagesSchema         `json:"pages,omitempty"`
	TotalCount *int                           `json:"total_count,omitempty"`
	Type       *gen.ArticleSearchResponseType `json:"type,omitempty"`
}

// ArticleSearchData contains articles and highlights returned by an article search.
type ArticleSearchData struct {
	Articles   *[]Article                           `json:"articles,omitempty"`
	Highlights *[]gen.ArticleSearchHighlightsSchema `json:"highlights,omitempty"`
}

func articleFromGenerated(article *gen.ArticleSchema) *Article {
	if article == nil {
		return nil
	}
	result := &Article{
		AiChatbotAvailability:         article.AiChatbotAvailability,
		AiCopilotAvailability:         article.AiCopilotAvailability,
		AiSalesAgentAvailability:      article.AiSalesAgentAvailability,
		AuthorId:                      article.AuthorId,
		Body:                          article.Body,
		BodyMarkdown:                  article.BodyMarkdown,
		CreatedAt:                     article.CreatedAt,
		CreatedById:                   article.CreatedById,
		DefaultLocale:                 article.DefaultLocale,
		Description:                   article.Description,
		DraftUpdatedAt:                article.DraftUpdatedAt,
		ExcludeFromArticleSuggestions: article.ExcludeFromArticleSuggestions,
		HasUnpublishedChanges:         article.HasUnpublishedChanges,
		HelpCenterAudience:            article.HelpCenterAudience,
		Id:                            article.Id,
		ParentIds:                     article.ParentIds,
		ScheduledPublishAt:            article.ScheduledPublishAt,
		ScheduledUnpublishAt:          article.ScheduledUnpublishAt,
		State:                         article.State,
		Tags:                          article.Tags,
		Title:                         article.Title,
		TranslatedContent:             article.TranslatedContent,
		Type:                          article.Type,
		UpdatedAt:                     article.UpdatedAt,
		UpdatedById:                   article.UpdatedById,
		Url:                           article.Url,
		WorkspaceId:                   article.WorkspaceId,
	}
	if article.ParentIds != nil && len(*article.ParentIds) > 0 {
		parentID := (*article.ParentIds)[0]
		result.ParentId = &parentID
	}
	return result
}

func articleListFromGenerated(list *gen.ArticleListSchema) *ArticleList {
	if list == nil {
		return nil
	}
	result := &ArticleList{Pages: list.Pages, TotalCount: list.TotalCount, Type: list.Type}
	if list.Data == nil {
		return result
	}
	articles := make([]Article, 0, len(*list.Data))
	for i := range *list.Data {
		article := articleFromGenerated(&(*list.Data)[i])
		if article != nil {
			articles = append(articles, *article)
		}
	}
	result.Data = &articles
	return result
}

func articleSearchResultFromGenerated(result *gen.ArticleSearchResponseSchema) *ArticleSearchResult {
	if result == nil {
		return nil
	}
	converted := &ArticleSearchResult{Pages: result.Pages, TotalCount: result.TotalCount, Type: result.Type}
	if result.Data == nil {
		return converted
	}
	data := &ArticleSearchData{Highlights: result.Data.Highlights}
	if result.Data.Articles != nil {
		articles := make([]Article, 0, len(*result.Data.Articles))
		for i := range *result.Data.Articles {
			article := articleFromGenerated(&(*result.Data.Articles)[i])
			if article != nil {
				articles = append(articles, *article)
			}
		}
		data.Articles = &articles
	}
	converted.Data = data
	return converted
}

// ArticleDeleted is the result of deleting an article.
type ArticleDeleted = gen.DeletedArticleObjectSchema
type ArticleVersion = gen.ArticleVersionSchema
type ArticleVersionList = gen.ArticleVersionListSchema
type ArticleTag = gen.AttachTagToArticleJSONRequestBody
type ArticleDraftUpdate = gen.UpdateArticleRequestSchema
type ArticleDraftPublish = gen.PublishArticleDraftRequestSchema
type ArticleVersionListParams = gen.ListArticleVersionsParams

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
	articles, err := requireOK("list articles", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
	return articleListFromGenerated(articles), err
}

// Create creates a new article.
func (s *ArticlesService) Create(ctx context.Context, article ArticleCreate) (*Article, error) {
	res, err := s.client.generated.CreateArticleWithResponse(ctx, nil, gen.CreateArticleJSONRequestBody(article))
	if err != nil {
		return nil, err
	}
	created, err := requireOK("create article", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
	return articleFromGenerated(created), err
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
	article, err := requireOK("retrieve article", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
	return articleFromGenerated(article), err
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
	updated, err := requireOK("update article", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
	return articleFromGenerated(updated), err
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
	results, err := requireOK("search articles", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
	return articleSearchResultFromGenerated(results), err
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
func (s *ArticlesService) ListVersions(ctx context.Context, articleID string, params *ArticleVersionListParams) (*ArticleVersionList, error) {
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
	res, err := s.client.generated.RetrieveArticleDraft(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	article, err := requireHTTPJSON[gen.ArticleSchema]("get article draft", res)
	return articleFromGenerated(article), err
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
	article, err := requireOK("stage article draft", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
	return articleFromGenerated(article), err
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
	article, err := requireOK("publish article draft", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
	return articleFromGenerated(article), err
}
