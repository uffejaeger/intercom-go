package intercom

import (
	"context"
	"fmt"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// Brand is an Intercom brand.
type Brand = gen.BrandSchema

// BrandList is a list of Intercom brands.
type BrandList = gen.BrandListSchema

// BrandsService exposes brand-related Intercom API operations.
type BrandsService struct {
	client *Client
}

// List returns all brands for the workspace.
func (s *BrandsService) List(ctx context.Context) (*BrandList, error) {
	res, err := s.client.generated.ListBrandsWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("list brands", res.StatusCode(), res.Body, res.JSON200)
}

// Retrieve returns a brand by ID.
func (s *BrandsService) Retrieve(ctx context.Context, brandID string) (*Brand, error) {
	if brandID == "" {
		return nil, fmt.Errorf("intercom: brand ID is required")
	}
	res, err := s.client.generated.RetrieveBrandWithResponse(ctx, brandID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("retrieve brand", res.StatusCode(), res.Body, res.JSON200)
}
