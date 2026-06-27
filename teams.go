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
	return requireOK("list teams", res.StatusCode(), res.Body, res.JSON200)
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
	return requireOK("retrieve team", res.StatusCode(), res.Body, res.JSON200)
}
