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
