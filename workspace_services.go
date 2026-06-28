package intercom

import (
	"context"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// IPAllowlist is the workspace IP allowlist configuration.
type IPAllowlist = gen.IpAllowlistSchema

// Job is an Intercom async job.
type Job = gen.JobsSchema

// DataExport is an Intercom content data export job.
type DataExport = gen.DataExportSchema

// DataExportCreate holds the fields for creating a content data export.
type DataExportCreate = gen.CreateDataExportsRequestSchema

// ReportingDatasetAttribute is a reporting dataset attribute.
type ReportingDatasetAttribute struct {
	ID   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

// ReportingDataset is a reporting dataset.
type ReportingDataset struct {
	Attributes             *[]ReportingDatasetAttribute `json:"attributes,omitempty"`
	DefaultTimeAttributeID *string                      `json:"default_time_attribute_id,omitempty"`
	Description            *string                      `json:"description,omitempty"`
	ID                     *string                      `json:"id,omitempty"`
	Name                   *string                      `json:"name,omitempty"`
}

// ReportingDatasetList is a list of reporting datasets.
type ReportingDatasetList struct {
	Data *[]ReportingDataset `json:"data,omitempty"`
	Type *string             `json:"type,omitempty"`
}

// ReportingExportJob is a reporting export job.
type ReportingExportJob struct {
	DownloadExpiresAt *string `json:"download_expires_at,omitempty"`
	DownloadURL       *string `json:"download_url,omitempty"`
	JobIdentifier     *string `json:"job_identifier,omitempty"`
	Status            *string `json:"status,omitempty"`
}

// ReportingExportCreate holds the fields for creating a reporting export job.
type ReportingExportCreate = gen.PostExportReportingDataEnqueueJSONBody

// WorkflowExport is an exported workflow.
type WorkflowExport = gen.WorkflowExportSchema

// WorkspaceService exposes workspace-level Intercom API operations.
type WorkspaceService struct {
	client *Client
}

// GetIPAllowlist returns the workspace IP allowlist configuration.
func (s *WorkspaceService) GetIPAllowlist(ctx context.Context) (*IPAllowlist, error) {
	res, err := s.client.generated.GetIpAllowlistWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("get IP allowlist", res.StatusCode(), res.Body, res.JSON200)
}

// UpdateIPAllowlist updates the workspace IP allowlist configuration.
func (s *WorkspaceService) UpdateIPAllowlist(ctx context.Context, allowlist IPAllowlist) (*IPAllowlist, error) {
	res, err := s.client.generated.UpdateIpAllowlistWithResponse(ctx, nil, gen.UpdateIpAllowlistJSONRequestBody(allowlist))
	if err != nil {
		return nil, err
	}
	return requireOK("update IP allowlist", res.StatusCode(), res.Body, res.JSON200)
}

// JobStatus retrieves the status of an async job.
func (s *WorkspaceService) JobStatus(ctx context.Context, jobID string) (*Job, error) {
	if jobID == "" {
		return nil, errRequiredID("job")
	}
	res, err := s.client.generated.JobsStatusWithResponse(ctx, jobID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("job status", res.StatusCode(), res.Body, res.JSON200)
}

// CreateDataExport creates a content data export job.
func (s *WorkspaceService) CreateDataExport(ctx context.Context, req DataExportCreate) (*DataExport, error) {
	res, err := s.client.generated.CreateDataExportWithResponse(ctx, nil, gen.CreateDataExportJSONRequestBody(req))
	if err != nil {
		return nil, err
	}
	return requireOK("create data export", res.StatusCode(), res.Body, res.JSON200)
}

// DownloadDataExport downloads the raw content data export payload.
func (s *WorkspaceService) DownloadDataExport(ctx context.Context, jobIdentifier string) ([]byte, error) {
	if jobIdentifier == "" {
		return nil, errRequiredID("job identifier")
	}
	res, err := s.client.generated.DownloadDataExportWithResponse(ctx, jobIdentifier, nil)
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != 200 {
		return nil, parseErrorResponse(res.StatusCode(), res.Body)
	}
	return res.Body, nil
}

// GetDataExport retrieves a content data export job by identifier.
func (s *WorkspaceService) GetDataExport(ctx context.Context, jobIdentifier string) (*DataExport, error) {
	if jobIdentifier == "" {
		return nil, errRequiredID("job identifier")
	}
	res, err := s.client.generated.GetDataExportWithResponse(ctx, jobIdentifier, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("get data export", res.StatusCode(), res.Body, res.JSON200)
}

// CancelDataExport cancels a content data export job.
func (s *WorkspaceService) CancelDataExport(ctx context.Context, jobIdentifier string) (*DataExport, error) {
	if jobIdentifier == "" {
		return nil, errRequiredID("job identifier")
	}
	res, err := s.client.generated.CancelDataExportWithResponse(ctx, jobIdentifier, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("cancel data export", res.StatusCode(), res.Body, res.JSON200)
}

// ListReportingDatasets returns reporting datasets available for export.
func (s *WorkspaceService) ListReportingDatasets(ctx context.Context) (*ReportingDatasetList, error) {
	res, err := s.client.generated.GetExportReportingDataGetDatasetsWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return requireJSON[ReportingDatasetList]("list reporting datasets", res.StatusCode(), res.Body)
}

// CreateReportingExport creates a reporting export job.
func (s *WorkspaceService) CreateReportingExport(ctx context.Context, req ReportingExportCreate) (*ReportingExportJob, error) {
	res, err := s.client.generated.PostExportReportingDataEnqueueWithResponse(ctx, nil, gen.PostExportReportingDataEnqueueJSONRequestBody(req))
	if err != nil {
		return nil, err
	}
	return requireJSON[ReportingExportJob]("create reporting export", res.StatusCode(), res.Body)
}

// GetReportingExportJob retrieves a reporting export job by identifier.
func (s *WorkspaceService) GetReportingExportJob(ctx context.Context, jobIdentifier, appID, clientID string) (*ReportingExportJob, error) {
	if jobIdentifier == "" {
		return nil, errRequiredID("job identifier")
	}
	if appID == "" {
		return nil, errRequiredID("app")
	}
	if clientID == "" {
		return nil, errRequiredID("client")
	}
	res, err := s.client.generated.GetExportReportingDataJobIdentifierWithResponse(ctx, jobIdentifier, &gen.GetExportReportingDataJobIdentifierParams{
		AppId:    appID,
		ClientId: clientID,
	})
	if err != nil {
		return nil, err
	}
	return requireJSON[ReportingExportJob]("get reporting export job", res.StatusCode(), res.Body)
}

// DownloadReportingExport downloads the raw reporting export payload.
func (s *WorkspaceService) DownloadReportingExport(ctx context.Context, jobIdentifier, appID string) ([]byte, error) {
	if jobIdentifier == "" {
		return nil, errRequiredID("job identifier")
	}
	if appID == "" {
		return nil, errRequiredID("app")
	}
	res, err := s.client.generated.GetDownloadReportingDataJobIdentifierWithResponse(
		ctx,
		jobIdentifier,
		&gen.GetDownloadReportingDataJobIdentifierParams{AppId: appID, Accept: gen.ApplicationoctetStream},
	)
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != 200 {
		return nil, parseErrorResponse(res.StatusCode(), res.Body)
	}
	return res.Body, nil
}

// ExportWorkflow exports a workflow by ID.
func (s *WorkspaceService) ExportWorkflow(ctx context.Context, workflowID string) (*WorkflowExport, error) {
	if workflowID == "" {
		return nil, errRequiredID("workflow")
	}
	res, err := s.client.generated.ExportWorkflowWithResponse(ctx, workflowID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("export workflow", res.StatusCode(), res.Body, res.JSON200)
}
