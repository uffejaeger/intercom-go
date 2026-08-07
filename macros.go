package intercom

import (
	"context"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// Macro is an Intercom saved reply.
type Macro = gen.MacroSchema

// MacroList is a paginated list of saved replies.
type MacroList = gen.MacroListSchema

// MacroListParams configures macro listing.
type MacroListParams = gen.ListMacrosParams

// MacrosService exposes saved-reply operations.
type MacrosService struct{ client *Client }

// List returns saved replies visible to the authenticated admin.
func (s *MacrosService) List(ctx context.Context, params *MacroListParams) (*MacroList, error) {
	res, err := s.client.generated.ListMacrosWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	return requireOK("list macros", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

// Get returns a saved reply by ID.
func (s *MacrosService) Get(ctx context.Context, id string) (*Macro, error) {
	res, err := s.client.generated.GetMacroWithResponse(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("get macro", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}
