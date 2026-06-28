package intercom

import (
	"context"
	"fmt"
	"strconv"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// Collection is an Intercom help center collection.
type Collection = gen.CollectionSchema

// CollectionList is a list of Intercom collections.
type CollectionList = gen.CollectionListSchema

// CollectionCreate holds the fields for creating a collection.
type CollectionCreate = gen.CreateCollectionRequestSchema

// CollectionUpdate holds the fields for updating a collection.
type CollectionUpdate = gen.UpdateCollectionRequestSchema

// CollectionsService exposes collection-related Intercom API operations.
type CollectionsService struct {
	client *Client
}

// List returns all collections.
func (s *CollectionsService) List(ctx context.Context) (*CollectionList, error) {
	res, err := s.client.generated.ListAllCollectionsWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("list collections", res.StatusCode(), res.Body, res.JSON200)
}

// Create creates a collection.
func (s *CollectionsService) Create(ctx context.Context, collection CollectionCreate) (*Collection, error) {
	res, err := s.client.generated.CreateCollectionWithResponse(ctx, nil, gen.CreateCollectionJSONRequestBody(collection))
	if err != nil {
		return nil, err
	}
	return requireOK("create collection", res.StatusCode(), res.Body, res.JSON200)
}

// Retrieve returns a collection by ID.
func (s *CollectionsService) Retrieve(ctx context.Context, collectionID string) (*Collection, error) {
	id, err := requireIntID("collection", collectionID)
	if err != nil {
		return nil, err
	}
	res, err := s.client.generated.RetrieveCollectionWithResponse(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("retrieve collection", res.StatusCode(), res.Body, res.JSON200)
}

// Update updates a collection by ID.
func (s *CollectionsService) Update(ctx context.Context, collectionID string, collection CollectionUpdate) (*Collection, error) {
	id, err := requireIntID("collection", collectionID)
	if err != nil {
		return nil, err
	}
	res, err := s.client.generated.UpdateCollectionWithResponse(ctx, id, nil, gen.UpdateCollectionJSONRequestBody(collection))
	if err != nil {
		return nil, err
	}
	return requireOK("update collection", res.StatusCode(), res.Body, res.JSON200)
}

// Delete deletes a collection by ID.
func (s *CollectionsService) Delete(ctx context.Context, collectionID string) error {
	id, err := requireIntID("collection", collectionID)
	if err != nil {
		return err
	}
	res, err := s.client.generated.DeleteCollectionWithResponse(ctx, id, nil)
	if err != nil {
		return err
	}
	return requireEmpty(res.StatusCode(), res.Body)
}

func requireIntID(resource, id string) (int, error) {
	if id == "" {
		return 0, fmt.Errorf("intercom: %s ID is required", resource)
	}
	n, err := strconv.Atoi(id)
	if err != nil {
		return 0, fmt.Errorf("intercom: %s ID %q is not a valid integer: %w", resource, id, err)
	}
	return n, nil
}
