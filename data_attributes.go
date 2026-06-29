package intercom

import (
	"context"
	"fmt"
	"strconv"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// DataAttribute is an Intercom data attribute.
type DataAttribute = gen.DataAttributeSchema

// DataAttributeList is a list of Intercom data attributes.
type DataAttributeList = gen.DataAttributeListSchema

// DataAttributeModel identifies which Intercom model a data attribute belongs to.
type DataAttributeModel string

const (
	DataAttributeModelContact      DataAttributeModel = "contact"
	DataAttributeModelCompany      DataAttributeModel = "company"
	DataAttributeModelConversation DataAttributeModel = "conversation"
)

// DataAttributeDataType identifies the stored value type for a data attribute.
type DataAttributeDataType string

const (
	DataAttributeDataTypeString  DataAttributeDataType = "string"
	DataAttributeDataTypeInteger DataAttributeDataType = "integer"
	DataAttributeDataTypeFloat   DataAttributeDataType = "float"
	DataAttributeDataTypeBoolean DataAttributeDataType = "boolean"
	DataAttributeDataTypeDate    DataAttributeDataType = "date"
	DataAttributeDataTypeList    DataAttributeDataType = "list"
)

// DataAttributeListParams configures data attribute listing.
type DataAttributeListParams struct {
	Model           DataAttributeModel
	IncludeArchived bool
}

// DataAttributeCreate holds the fields for creating a data attribute.
type DataAttributeCreate struct {
	Name              string
	Model             DataAttributeModel
	DataType          DataAttributeDataType
	Description       *string
	MessengerWritable *bool
	Options           []string
}

// DataAttributeUpdate holds the fields for updating a data attribute.
type DataAttributeUpdate struct {
	Archived          *bool
	Description       *string
	MessengerWritable *bool
	Options           []string
}

// DataAttributesService exposes data-attribute Intercom API operations.
type DataAttributesService struct {
	client *Client
}

// List returns data attributes for the workspace.
func (s *DataAttributesService) List(ctx context.Context, params DataAttributeListParams) (*DataAttributeList, error) {
	genParams := &gen.LisDataAttributesParams{}
	if params.Model != "" {
		model, err := listDataAttributeModel(params.Model)
		if err != nil {
			return nil, err
		}
		genParams.Model = &model
	}
	if params.IncludeArchived {
		includeArchived := true
		genParams.IncludeArchived = &includeArchived
	}
	if genParams.Model == nil && genParams.IncludeArchived == nil {
		genParams = nil
	}

	res, err := s.client.generated.LisDataAttributesWithResponse(ctx, genParams)
	if err != nil {
		return nil, err
	}

	return requireOK("list data attributes", res.StatusCode(), res.Body, res.JSON200)
}

// Create creates a new contact or company data attribute.
func (s *DataAttributesService) Create(ctx context.Context, attribute DataAttributeCreate) (*DataAttribute, error) {
	if attribute.Name == "" {
		return nil, fmt.Errorf("intercom: data attribute name is required")
	}
	model, err := createDataAttributeModel(attribute.Model)
	if err != nil {
		return nil, err
	}
	if attribute.DataType == "" {
		return nil, fmt.Errorf("intercom: data attribute data type is required")
	}
	options, err := dataAttributeOptions(attribute.DataType, attribute.Options)
	if err != nil {
		return nil, err
	}

	body, _ := marshalBody(struct {
		Name              string                `json:"name"`
		Model             string                `json:"model"`
		DataType          string                `json:"data_type"`
		Description       *string               `json:"description,omitempty"`
		MessengerWritable *bool                 `json:"messenger_writable,omitempty"`
		Options           []dataAttributeOption `json:"options,omitempty"`
	}{
		Name:              attribute.Name,
		Model:             string(model),
		DataType:          string(attribute.DataType),
		Description:       attribute.Description,
		MessengerWritable: attribute.MessengerWritable,
		Options:           options,
	})

	res, err := s.client.generated.CreateDataAttributeWithBodyWithResponse(ctx, nil, "application/json", body)
	if err != nil {
		return nil, err
	}

	return requireOK("create data attribute", res.StatusCode(), res.Body, res.JSON200)
}

// Update updates or archives a data attribute.
func (s *DataAttributesService) Update(ctx context.Context, dataAttributeID string, attribute DataAttributeUpdate) (*DataAttribute, error) {
	if dataAttributeID == "" {
		return nil, fmt.Errorf("intercom: data attribute ID is required")
	}
	id, err := strconv.Atoi(dataAttributeID)
	if err != nil {
		return nil, fmt.Errorf("intercom: data attribute ID %q is not a valid integer: %w", dataAttributeID, err)
	}
	var options []dataAttributeOption
	if len(attribute.Options) > 0 {
		options, err = dataAttributeOptions(DataAttributeDataTypeList, attribute.Options)
		if err != nil {
			return nil, err
		}
	}

	body, _ := marshalBody(struct {
		Archived          *bool                 `json:"archived,omitempty"`
		Description       *string               `json:"description,omitempty"`
		MessengerWritable *bool                 `json:"messenger_writable,omitempty"`
		Options           []dataAttributeOption `json:"options,omitempty"`
	}{
		Archived:          attribute.Archived,
		Description:       attribute.Description,
		MessengerWritable: attribute.MessengerWritable,
		Options:           options,
	})

	res, err := s.client.generated.UpdateDataAttributeWithBodyWithResponse(ctx, id, nil, "application/json", body)
	if err != nil {
		return nil, err
	}

	return requireOK("update data attribute", res.StatusCode(), res.Body, res.JSON200)
}

type dataAttributeOption struct {
	Value string `json:"value"`
}

func listDataAttributeModel(model DataAttributeModel) (gen.LisDataAttributesParamsModel, error) {
	switch model {
	case DataAttributeModelContact:
		return gen.LisDataAttributesParamsModelContact, nil
	case DataAttributeModelCompany:
		return gen.LisDataAttributesParamsModelCompany, nil
	case DataAttributeModelConversation:
		return gen.LisDataAttributesParamsModelConversation, nil
	case "":
		return "", nil
	default:
		return "", fmt.Errorf("intercom: unsupported data attribute model %q", model)
	}
}

func createDataAttributeModel(model DataAttributeModel) (DataAttributeModel, error) {
	switch model {
	case DataAttributeModelContact, DataAttributeModelCompany:
		return model, nil
	case DataAttributeModelConversation:
		return "", fmt.Errorf("intercom: conversation data attributes cannot be created")
	case "":
		return "", fmt.Errorf("intercom: data attribute model is required")
	default:
		return "", fmt.Errorf("intercom: unsupported data attribute model %q", model)
	}
}

func dataAttributeOptions(dataType DataAttributeDataType, values []string) ([]dataAttributeOption, error) {
	if len(values) == 0 {
		if dataType == DataAttributeDataTypeList {
			return nil, fmt.Errorf("intercom: list data attributes require at least two options")
		}
		return nil, nil
	}
	if dataType != DataAttributeDataTypeList {
		return nil, fmt.Errorf("intercom: data attribute options are only supported for list data attributes")
	}
	if len(values) < 2 {
		return nil, fmt.Errorf("intercom: list data attributes require at least two options")
	}

	options := make([]dataAttributeOption, 0, len(values))
	for _, value := range values {
		if value == "" {
			return nil, fmt.Errorf("intercom: data attribute options cannot be empty")
		}
		options = append(options, dataAttributeOption{Value: value})
	}
	return options, nil
}
