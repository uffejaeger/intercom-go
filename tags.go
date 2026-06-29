package intercom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// TagDetail is a single Intercom tag returned by the tag lookup endpoint.
type TagDetail = gen.TagBasicSchema

// TagCreateOrUpdateRequest creates a new tag or renames an existing one.
type TagCreateOrUpdateRequest struct {
	ID   *string `json:"id,omitempty"`
	Name string  `json:"name"`
}

func (TagCreateOrUpdateRequest) isTagCreateRequest() {}

// TagCompanyReference identifies a company in a tag company request.
type TagCompanyReference struct {
	CompanyID *string `json:"company_id,omitempty"`
	ID        *string `json:"id,omitempty"`
}

// TagCompanyRequest tags one or more companies with a tag name.
type TagCompanyRequest struct {
	Companies []TagCompanyReference `json:"companies"`
	Name      string                `json:"name"`
}

func (TagCompanyRequest) isTagCreateRequest() {}

// TagCompanyUntagReference identifies a company to untag.
type TagCompanyUntagReference struct {
	CompanyID *string `json:"company_id,omitempty"`
	ID        *string `json:"id,omitempty"`
	Untag     bool    `json:"untag"`
}

// TagCompanyUntagRequest removes a tag from one or more companies.
type TagCompanyUntagRequest struct {
	Companies []TagCompanyUntagReference `json:"companies"`
	Name      string                     `json:"name"`
}

func (TagCompanyUntagRequest) isTagCreateRequest() {}

// MarshalJSON ensures Intercom's required `untag: true` flag is always present.
func (r TagCompanyUntagRequest) MarshalJSON() ([]byte, error) {
	type company struct {
		CompanyID *string `json:"company_id,omitempty"`
		ID        *string `json:"id,omitempty"`
		Untag     bool    `json:"untag"`
	}
	payload := struct {
		Companies []company `json:"companies"`
		Name      string    `json:"name"`
	}{
		Companies: make([]company, 0, len(r.Companies)),
		Name:      r.Name,
	}
	for _, c := range r.Companies {
		payload.Companies = append(payload.Companies, company{
			CompanyID: c.CompanyID,
			ID:        c.ID,
			Untag:     true,
		})
	}
	return json.Marshal(payload)
}

// TagUserReference identifies a contact in a tag users request.
type TagUserReference struct {
	ID string `json:"id"`
}

// TagUsersRequest tags one or more contacts with a tag name.
type TagUsersRequest struct {
	Name  string             `json:"name"`
	Users []TagUserReference `json:"users"`
}

func (TagUsersRequest) isTagCreateRequest() {}

// TagCreateRequest is a supported request body for the tag create/update endpoint.
type TagCreateRequest interface {
	isTagCreateRequest()
}

// TagsService exposes tag-related Intercom API operations.
type TagsService struct {
	client *Client
}

// List returns all tags for the workspace.
func (s *TagsService) List(ctx context.Context) (*TagList, error) {
	res, err := s.client.generated.ListTagsWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("list tags", res.StatusCode(), res.Body, res.JSON200)
}

// Create creates or updates a tag, or applies tag operations supported by the endpoint.
func (s *TagsService) Create(ctx context.Context, req TagCreateRequest) (*TagDetail, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("intercom: marshal tag request: %w", err)
	}
	res, err := s.client.generated.CreateTagWithBodyWithResponse(ctx, nil, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	return requireOK("create tag", res.StatusCode(), res.Body, res.JSON200)
}

// Retrieve returns a tag by ID.
func (s *TagsService) Retrieve(ctx context.Context, tagID string) (*TagDetail, error) {
	if tagID == "" {
		return nil, fmt.Errorf("intercom: tag ID is required")
	}
	res, err := s.client.generated.FindTagWithResponse(ctx, tagID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("retrieve tag", res.StatusCode(), res.Body, res.JSON200)
}

// Delete deletes a tag by ID.
func (s *TagsService) Delete(ctx context.Context, tagID string) error {
	if tagID == "" {
		return fmt.Errorf("intercom: tag ID is required")
	}
	res, err := s.client.generated.DeleteTagWithResponse(ctx, tagID, nil)
	if err != nil {
		return err
	}
	return requireEmpty(res.StatusCode(), res.Body)
}
