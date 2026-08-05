package intercom

import (
	"context"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

type ConversationAttribute = gen.ConversationAttribute
type ConversationAttributeList = gen.ConversationAttributeListSchema
type ConversationAttributeCreate = gen.CreateConversationAttributeRequest
type ConversationAttributeUpdate = gen.UpdateConversationAttributeRequestSchema
type ConversationAttributeOptionCreate = gen.CreateConversationAttributeOptionRequestSchema
type ConversationAttributeOptionUpdate = gen.UpdateConversationAttributeOptionRequestSchema
type ConversationAttributeListParams = gen.ListConversationAttributesParams

// ConversationAttributesService exposes custom attributes for conversations.
type ConversationAttributesService struct{ client *Client }

func (s *ConversationAttributesService) List(ctx context.Context, params *ConversationAttributeListParams) (*ConversationAttributeList, error) {
	res, err := s.client.generated.ListConversationAttributesWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	return requireOK("list conversation attributes", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

func (s *ConversationAttributesService) Create(ctx context.Context, attribute ConversationAttributeCreate) (*ConversationAttribute, error) {
	res, err := s.client.generated.CreateConversationAttributeWithResponse(ctx, nil, attribute)
	if err != nil {
		return nil, err
	}
	return requireOK("create conversation attribute", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

func (s *ConversationAttributesService) Get(ctx context.Context, id int) (*ConversationAttribute, error) {
	res, err := s.client.generated.GetConversationAttributeWithResponse(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("get conversation attribute", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

func (s *ConversationAttributesService) Update(ctx context.Context, id int, attribute ConversationAttributeUpdate) (*ConversationAttribute, error) {
	res, err := s.client.generated.UpdateConversationAttributeWithResponse(ctx, id, nil, attribute)
	if err != nil {
		return nil, err
	}
	return requireOK("update conversation attribute", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

func (s *ConversationAttributesService) Delete(ctx context.Context, id int) (*ConversationAttribute, error) {
	res, err := s.client.generated.DeleteConversationAttributeWithResponse(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("delete conversation attribute", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

func (s *ConversationAttributesService) CreateOption(ctx context.Context, attributeID int, option ConversationAttributeOptionCreate) (*ConversationAttribute, error) {
	res, err := s.client.generated.CreateConversationAttributeOptionWithResponse(ctx, attributeID, nil, option)
	if err != nil {
		return nil, err
	}
	return requireOK("create conversation attribute option", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

func (s *ConversationAttributesService) UpdateOption(ctx context.Context, attributeID int, optionID string, option ConversationAttributeOptionUpdate) (*ConversationAttribute, error) {
	res, err := s.client.generated.UpdateConversationAttributeOptionWithResponse(ctx, attributeID, optionID, nil, option)
	if err != nil {
		return nil, err
	}
	return requireOK("update conversation attribute option", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

func (s *ConversationAttributesService) DeleteOption(ctx context.Context, attributeID int, optionID string) (*ConversationAttribute, error) {
	res, err := s.client.generated.DeleteConversationAttributeOptionWithResponse(ctx, attributeID, optionID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("delete conversation attribute option", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}
