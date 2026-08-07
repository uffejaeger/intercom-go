package intercom

import (
	"context"
	"fmt"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// Team is an Intercom team.
type Team = gen.TeamSchema

// TeamList is a list of Intercom teams.
type TeamList = gen.TeamListSchema

// TeamMetrics is a list of performance metrics for a team.
type TeamMetrics = gen.TeamMetricListSchema

// TeamMetricsParams configures a team metrics request.
type TeamMetricsParams = gen.GetTeamMetricsParams

// TeamsService exposes team-related Intercom API operations.
type TeamsService struct {
	client *Client
}

// List returns all teams for the workspace.
func (s *TeamsService) List(ctx context.Context) (*TeamList, error) {
	res, err := s.client.generated.ListTeamsWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("list teams", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

// Retrieve returns a team by ID.
func (s *TeamsService) Retrieve(ctx context.Context, teamID string) (*Team, error) {
	if teamID == "" {
		return nil, fmt.Errorf("intercom: team ID is required")
	}
	res, err := s.client.generated.RetrieveTeamWithResponse(ctx, teamID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("retrieve team", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}

// Metrics returns performance metrics for a team.
func (s *TeamsService) Metrics(ctx context.Context, teamID string, params *TeamMetricsParams) (*TeamMetrics, error) {
	res, err := s.client.generated.GetTeamMetricsWithResponse(ctx, teamID, params)
	if err != nil {
		return nil, err
	}
	return requireOK("get team metrics", res.StatusCode(), res.Body, res.JSON200, responseHeaders(res.HTTPResponse))
}
