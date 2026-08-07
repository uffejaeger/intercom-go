package intercom

import (
	"context"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

type DataConnector = gen.DataConnectorSchema
type DataConnectorDetail = gen.DataConnectorDetailSchema
type DataConnectorList = gen.DataConnectorListSchema
type DataConnectorExecutionResult = gen.DataConnectorExecutionResultSchema
type DataConnectorExecutionResultList = gen.DataConnectorExecutionResultListSchema
type DataConnectorDeleted = gen.DeletedDataConnectorObjectSchema
type DataConnectorCreate = gen.CreateDataConnectorRequestSchema
type DataConnectorUpdate = gen.UpdateDataConnectorRequestSchema
type DataConnectorListParams = gen.ListDataConnectorsParams
type DataConnectorExecutionListParams = gen.ListDataConnectorExecutionResultsParams

// DataConnectorExecutionSuccess filters execution results by success status.
type DataConnectorExecutionSuccess = gen.ListDataConnectorExecutionResultsParamsSuccess

// DataConnectorExecutionErrorType filters execution results by error type.
type DataConnectorExecutionErrorType = gen.ListDataConnectorExecutionResultsParamsErrorType

// DataConnectorExecutionIncludeBodies controls whether execution bodies are returned.
type DataConnectorExecutionIncludeBodies = gen.ListDataConnectorExecutionResultsParamsIncludeBodies

// DataConnectorCreateAudience identifies a user type that can use a new connector.
type DataConnectorCreateAudience = gen.CreateDataConnectorRequestAudiences

// DataConnectorCreateDataInputType identifies the type of a new connector input.
type DataConnectorCreateDataInputType = gen.CreateDataConnectorRequestDataInputsType

// DataConnectorCreateDataInput configures one input accepted by a new connector.
type DataConnectorCreateDataInput = struct {
	DefaultValue *string                           `json:"default_value,omitempty"`
	Description  *string                           `json:"description,omitempty"`
	Name         *string                           `json:"name,omitempty"`
	Required     *bool                             `json:"required,omitempty"`
	Type         *DataConnectorCreateDataInputType `json:"type,omitempty"`
}

// DataConnectorHeader configures an HTTP header sent by a connector.
type DataConnectorHeader = struct {
	Name  *string `json:"name,omitempty"`
	Value *string `json:"value,omitempty"`
}

// DataConnectorCreateHTTPMethod identifies the HTTP method for a new connector.
type DataConnectorCreateHTTPMethod = gen.CreateDataConnectorRequestHttpMethod

// DataConnectorUpdateAudience identifies a user type targeted by a connector update.
type DataConnectorUpdateAudience = gen.UpdateDataConnectorRequestAudiences

// DataConnectorUpdateDataInputType identifies the type of an updated connector input.
type DataConnectorUpdateDataInputType = gen.UpdateDataConnectorRequestDataInputsType

// DataConnectorUpdateDataInput configures one input accepted by an updated connector.
type DataConnectorUpdateDataInput = struct {
	DefaultValue *string                           `json:"default_value,omitempty"`
	Description  *string                           `json:"description,omitempty"`
	Name         *string                           `json:"name,omitempty"`
	Required     *bool                             `json:"required,omitempty"`
	Type         *DataConnectorUpdateDataInputType `json:"type,omitempty"`
}

// DataConnectorUpdateHTTPMethod identifies the HTTP method for a connector update.
type DataConnectorUpdateHTTPMethod = gen.UpdateDataConnectorRequestHttpMethod

// DataConnectorUpdateState identifies the desired state for a connector update.
type DataConnectorUpdateState = gen.UpdateDataConnectorRequestState

// DataConnectorsService exposes data-connector configuration and execution-history operations.
type DataConnectorsService struct{ client *Client }

func (s *DataConnectorsService) List(ctx context.Context, params *DataConnectorListParams) (*DataConnectorList, error) {
	res, err := s.client.generated.ListDataConnectorsWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	return requireOK("list data connectors", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

func (s *DataConnectorsService) Create(ctx context.Context, connector DataConnectorCreate) (*DataConnectorDetail, error) {
	res, err := s.client.generated.CreateDataConnectorWithResponse(ctx, nil, connector)
	if err != nil {
		return nil, err
	}
	return requireCreated("create data connector", res.StatusCode(), res.Body, res.JSON201, responseHeaders(res.HTTPResponse))
}

func (s *DataConnectorsService) Get(ctx context.Context, id string) (*DataConnectorDetail, error) {
	res, err := s.client.generated.RetrieveDataConnectorWithResponse(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("get data connector", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

func (s *DataConnectorsService) Update(ctx context.Context, id string, connector DataConnectorUpdate) (*DataConnectorDetail, error) {
	res, err := s.client.generated.UpdateDataConnectorWithResponse(ctx, id, nil, connector)
	if err != nil {
		return nil, err
	}
	return requireOK("update data connector", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

func (s *DataConnectorsService) Delete(ctx context.Context, id string) (*DataConnectorDeleted, error) {
	res, err := s.client.generated.DeleteDataConnectorWithResponse(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("delete data connector", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

func (s *DataConnectorsService) ListExecutionResults(ctx context.Context, connectorID string, params *DataConnectorExecutionListParams) (*DataConnectorExecutionResultList, error) {
	res, err := s.client.generated.ListDataConnectorExecutionResultsWithResponse(ctx, connectorID, params)
	if err != nil {
		return nil, err
	}
	return requireOK("list data connector execution results", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

func (s *DataConnectorsService) GetExecutionResult(ctx context.Context, connectorID, id string) (*DataConnectorExecutionResult, error) {
	res, err := s.client.generated.ShowDataConnectorExecutionResultWithResponse(ctx, connectorID, id, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("get data connector execution result", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}
