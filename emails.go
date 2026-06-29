package intercom

import (
	"context"
	"fmt"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// EmailSetting is an Intercom sender email setting.
type EmailSetting = gen.EmailSettingSchema

// EmailList is a list of Intercom sender email settings.
type EmailList = gen.EmailListSchema

// EmailsService exposes email-setting Intercom API operations.
type EmailsService struct {
	client *Client
}

// List returns sender email settings for the workspace.
func (s *EmailsService) List(ctx context.Context) (*EmailList, error) {
	res, err := s.client.generated.ListEmailsWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}

	return requireOK("list emails", res.StatusCode(), res.Body, res.JSON200)
}

// Retrieve fetches a sender email setting by ID.
func (s *EmailsService) Retrieve(ctx context.Context, emailID string) (*EmailSetting, error) {
	if emailID == "" {
		return nil, fmt.Errorf("intercom: email ID is required")
	}

	res, err := s.client.generated.RetrieveEmailWithResponse(ctx, emailID, nil)
	if err != nil {
		return nil, err
	}

	return requireOK("retrieve email", res.StatusCode(), res.Body, res.JSON200)
}
