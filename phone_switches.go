package intercom

import (
	"context"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// PhoneSwitch is an Intercom phone switch response.
type PhoneSwitch = gen.PhoneSwitchSchema

// PhoneSwitchCreate holds the fields for creating a phone switch.
type PhoneSwitchCreate = gen.CreatePhoneSwitchRequestSchema

// PhoneSwitchesService exposes phone-switch-related Intercom API operations.
type PhoneSwitchesService struct {
	client *Client
}

// Create creates a phone switch.
func (s *PhoneSwitchesService) Create(ctx context.Context, req PhoneSwitchCreate) (*PhoneSwitch, error) {
	res, err := s.client.generated.CreatePhoneSwitchWithResponse(ctx, nil, gen.CreatePhoneSwitchJSONRequestBody(req))
	if err != nil {
		return nil, err
	}
	return requireOK("create phone switch", res.StatusCode(), res.Body, res.JSON200)
}
