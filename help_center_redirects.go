package intercom

import (
	"context"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

type HelpCenterRedirect = gen.HelpCenterRedirectSchema
type HelpCenterRedirectList = gen.HelpCenterRedirectListSchema
type HelpCenterRedirectDeleted = gen.DeletedHelpCenterRedirectObjectSchema
type HelpCenterRedirectCreate = gen.CreateHelpCenterRedirectRequestSchema
type HelpCenterRedirectListParams = gen.ListHelpCenterRedirectsParams

// HelpCenterRedirectsService exposes redirects within an Intercom help center.
type HelpCenterRedirectsService struct{ client *Client }

func (s *HelpCenterRedirectsService) List(ctx context.Context, helpCenterID string, params *HelpCenterRedirectListParams) (*HelpCenterRedirectList, error) {
	res, err := s.client.generated.ListHelpCenterRedirectsWithResponse(ctx, helpCenterID, params)
	if err != nil {
		return nil, err
	}
	return requireOK("list help center redirects", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

func (s *HelpCenterRedirectsService) Create(ctx context.Context, helpCenterID string, redirect HelpCenterRedirectCreate) (*HelpCenterRedirect, error) {
	res, err := s.client.generated.CreateHelpCenterRedirectWithResponse(ctx, helpCenterID, nil, redirect)
	if err != nil {
		return nil, err
	}
	return requireOK("create help center redirect", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

func (s *HelpCenterRedirectsService) Get(ctx context.Context, helpCenterID, id string) (*HelpCenterRedirect, error) {
	res, err := s.client.generated.RetrieveHelpCenterRedirectWithResponse(ctx, helpCenterID, id, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("get help center redirect", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

func (s *HelpCenterRedirectsService) Delete(ctx context.Context, helpCenterID, id string) (*HelpCenterRedirectDeleted, error) {
	res, err := s.client.generated.DeleteHelpCenterRedirectWithResponse(ctx, helpCenterID, id, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("delete help center redirect", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}
