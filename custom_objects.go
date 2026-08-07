package intercom

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// CustomObjectInstance is an Intercom custom object instance.
type CustomObjectInstance = gen.CustomObjectInstanceSchema

// CustomObjectInstanceDeleted is a deleted Intercom custom object instance response.
type CustomObjectInstanceDeleted = gen.CustomObjectInstanceDeletedSchema

// CustomObjectInstanceCreateOrUpdate holds the fields for creating or updating a custom object instance.
type CustomObjectInstanceCreateOrUpdate = gen.CreateOrUpdateCustomObjectInstanceRequestSchema

// CustomObjectInstanceList is a paginated list of custom object instances.
type CustomObjectInstanceList = gen.CustomObjectInstancesPaginatedListSchema

// CustomObjectInstanceListParams configures a custom object instance list request.
type CustomObjectInstanceListParams = gen.ListCustomObjectInstancesParams

// CustomObjectsService exposes custom-object instance Intercom API operations.
type CustomObjectsService struct {
	client *Client
}

// List returns instances for a custom object type.
func (s *CustomObjectsService) List(ctx context.Context, customObjectType string, params *CustomObjectInstanceListParams) (*CustomObjectInstanceList, error) {
	if err := requireCustomObjectType(customObjectType); err != nil {
		return nil, err
	}
	res, err := s.client.generated.ListCustomObjectInstancesWithResponse(ctx, customObjectType, params)
	if err != nil {
		return nil, err
	}
	return requireOK("list custom object instances", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
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
	return requireOK("create or update custom object instance", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
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
	return requireOK("get custom object instance", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

// GetByExternalID returns a custom object instance by external ID.
func (s *CustomObjectsService) GetByExternalID(ctx context.Context, customObjectType, externalID string) (*CustomObjectInstance, error) {
	if err := requireCustomObjectType(customObjectType); err != nil {
		return nil, err
	}
	if externalID == "" {
		return nil, fmt.Errorf("intercom: custom object instance external ID is required")
	}
	path := "/custom_object_instances/" + url.PathEscape(customObjectType) + "?external_id=" + url.QueryEscape(externalID)
	req, err := s.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("intercom: read custom object instance by external ID response: %w", err)
	}
	return requireJSON[CustomObjectInstance]("get custom object instance by external ID", res.StatusCode, body, res.Header)
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
	return requireOK("delete custom object instance", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
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
	return requireOK("delete custom object instance by external ID", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

func requireCustomObjectType(customObjectType string) error {
	if customObjectType == "" {
		return fmt.Errorf("intercom: custom object type is required")
	}
	return nil
}
