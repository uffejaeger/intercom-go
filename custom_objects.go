package intercom

import (
	"context"
	"fmt"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// CustomObjectInstance is an Intercom custom object instance.
type CustomObjectInstance = gen.CustomObjectInstanceSchema

// CustomObjectInstanceDeleted is a deleted Intercom custom object instance response.
type CustomObjectInstanceDeleted = gen.CustomObjectInstanceDeletedSchema

// CustomObjectInstanceCreateOrUpdate holds the fields for creating or updating a custom object instance.
type CustomObjectInstanceCreateOrUpdate = gen.CreateOrUpdateCustomObjectInstanceRequestSchema

// CustomObjectsService exposes custom-object instance Intercom API operations.
type CustomObjectsService struct {
	client *Client
}

// CreateOrUpdate creates or updates a custom object instance for a custom object type.
func (s *CustomObjectsService) CreateOrUpdate(ctx context.Context, customObjectType string, instance CustomObjectInstanceCreateOrUpdate) (*CustomObjectInstance, error) {
	if err := requireCustomObjectType(customObjectType); err != nil {
		return nil, err
	}
	res, err := s.client.generated.CreateCustomObjectInstancesWithResponse(ctx, customObjectType, nil, gen.CreateCustomObjectInstancesJSONRequestBody(instance))
	if err != nil {
		return nil, err
	}
	return requireOK("create or update custom object instance", res.StatusCode(), res.Body, res.JSON200)
}

// Get returns a custom object instance by Intercom ID.
func (s *CustomObjectsService) Get(ctx context.Context, customObjectType, instanceID string) (*CustomObjectInstance, error) {
	if err := requireCustomObjectType(customObjectType); err != nil {
		return nil, err
	}
	if instanceID == "" {
		return nil, fmt.Errorf("intercom: custom object instance ID is required")
	}
	res, err := s.client.generated.GetCustomObjectInstancesByIdWithResponse(ctx, customObjectType, instanceID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("get custom object instance", res.StatusCode(), res.Body, res.JSON200)
}

// GetByExternalID returns a custom object instance by external ID.
func (s *CustomObjectsService) GetByExternalID(ctx context.Context, customObjectType, externalID string) (*CustomObjectInstance, error) {
	if err := requireCustomObjectType(customObjectType); err != nil {
		return nil, err
	}
	if externalID == "" {
		return nil, fmt.Errorf("intercom: custom object instance external ID is required")
	}
	params := &gen.GetCustomObjectInstancesByExternalIdParams{ExternalId: externalID}
	res, err := s.client.generated.GetCustomObjectInstancesByExternalIdWithResponse(ctx, customObjectType, params)
	if err != nil {
		return nil, err
	}
	return requireOK("get custom object instance by external ID", res.StatusCode(), res.Body, res.JSON200)
}

// Delete deletes a custom object instance by Intercom ID.
func (s *CustomObjectsService) Delete(ctx context.Context, customObjectType, instanceID string) (*CustomObjectInstanceDeleted, error) {
	if err := requireCustomObjectType(customObjectType); err != nil {
		return nil, err
	}
	if instanceID == "" {
		return nil, fmt.Errorf("intercom: custom object instance ID is required")
	}
	res, err := s.client.generated.DeleteCustomObjectInstancesByExternalIdWithResponse(ctx, customObjectType, instanceID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("delete custom object instance", res.StatusCode(), res.Body, res.JSON200)
}

// DeleteByExternalID deletes a custom object instance by external ID.
func (s *CustomObjectsService) DeleteByExternalID(ctx context.Context, customObjectType, externalID string) (*CustomObjectInstanceDeleted, error) {
	if err := requireCustomObjectType(customObjectType); err != nil {
		return nil, err
	}
	if externalID == "" {
		return nil, fmt.Errorf("intercom: custom object instance external ID is required")
	}
	params := &gen.DeleteCustomObjectInstancesByIdParams{ExternalId: externalID}
	res, err := s.client.generated.DeleteCustomObjectInstancesByIdWithResponse(ctx, customObjectType, params)
	if err != nil {
		return nil, err
	}
	return requireOK("delete custom object instance by external ID", res.StatusCode(), res.Body, res.JSON200)
}

func requireCustomObjectType(customObjectType string) error {
	if customObjectType == "" {
		return fmt.Errorf("intercom: custom object type is required")
	}
	return nil
}
