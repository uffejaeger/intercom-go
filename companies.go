package intercom

import (
	"context"
	"fmt"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// Company is an Intercom company.
type Company = gen.CompanySchema

// CompanyList is a page of Intercom companies.
type CompanyList = gen.CompanyListSchema

// CompanyScroll is a scroll page of Intercom companies.
type CompanyScroll = gen.CompanyScrollSchema

// CompanyDeleted is the result of deleting a company.
type CompanyDeleted = gen.DeletedCompanyObjectSchema

// CompanyContacts is a list of contacts attached to a company.
type CompanyContacts = gen.CompanyAttachedContactsSchema

// CompanySegmentsAttached is a list of segments a company belongs to.
type CompanySegmentsAttached = gen.CompanyAttachedSegmentsSchema

// CompanyCreate holds the fields for creating or updating a company.
type CompanyCreate = gen.CreateOrUpdateCompanyRequestSchema

// CompanyUpdate holds the fields for updating a company.
type CompanyUpdate = gen.UpdateCompanyRequestSchema

// ContactCompanies is the list of companies a contact belongs to.
type ContactCompanies = gen.ContactAttachedCompaniesSchema

// CompaniesService exposes company-related Intercom API operations.
type CompaniesService struct {
	client *Client
}

// CreateOrUpdate creates or updates a company.
func (s *CompaniesService) CreateOrUpdate(ctx context.Context, company CompanyCreate) (*Company, error) {
	res, err := s.client.generated.CreateOrUpdateCompanyWithResponse(ctx, nil, gen.CreateOrUpdateCompanyJSONRequestBody(company))
	if err != nil {
		return nil, err
	}
	return requireOK("create or update company", res.StatusCode(), res.Body, res.JSON200)
}

// Retrieve retrieves a company by its Intercom-assigned ID.
func (s *CompaniesService) Retrieve(ctx context.Context, companyID string) (*Company, error) {
	if companyID == "" {
		return nil, fmt.Errorf("intercom: company ID is required")
	}
	res, err := s.client.generated.RetrieveACompanyByIdWithResponse(ctx, companyID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("retrieve company", res.StatusCode(), res.Body, res.JSON200)
}

// RetrieveByID retrieves companies matching the given company_id (external ID set by the caller).
func (s *CompaniesService) RetrieveByID(ctx context.Context, companyID string) (*CompanyList, error) {
	if companyID == "" {
		return nil, fmt.Errorf("intercom: company ID is required")
	}
	res, err := s.client.generated.RetrieveCompanyWithResponse(ctx, &gen.RetrieveCompanyParams{CompanyId: &companyID})
	if err != nil {
		return nil, err
	}
	return requireOK("retrieve company by ID", res.StatusCode(), res.Body, res.JSON200)
}

// Update updates a company by its Intercom-assigned ID.
func (s *CompaniesService) Update(ctx context.Context, companyID string, update CompanyUpdate) (*Company, error) {
	if companyID == "" {
		return nil, fmt.Errorf("intercom: company ID is required")
	}
	res, err := s.client.generated.UpdateCompanyWithResponse(ctx, companyID, nil, gen.UpdateCompanyJSONRequestBody(update))
	if err != nil {
		return nil, err
	}
	return requireOK("update company", res.StatusCode(), res.Body, res.JSON200)
}

// Delete deletes a company by its Intercom-assigned ID.
func (s *CompaniesService) Delete(ctx context.Context, companyID string) (*CompanyDeleted, error) {
	if companyID == "" {
		return nil, fmt.Errorf("intercom: company ID is required")
	}
	res, err := s.client.generated.DeleteCompanyWithResponse(ctx, companyID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("delete company", res.StatusCode(), res.Body, res.JSON200)
}

// ListAll returns all companies using cursor-based pagination.
func (s *CompaniesService) ListAll(ctx context.Context) (*CompanyList, error) {
	res, err := s.client.generated.ListAllCompaniesWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("list all companies", res.StatusCode(), res.Body, res.JSON200)
}

// Scroll returns companies using the scroll API for large datasets.
func (s *CompaniesService) Scroll(ctx context.Context) (*CompanyScroll, error) {
	res, err := s.client.generated.ScrollOverAllCompaniesWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("scroll companies", res.StatusCode(), res.Body, res.JSON200)
}

// ListNotes returns notes for a company.
func (s *CompaniesService) ListNotes(ctx context.Context, companyID string) (*NoteList, error) {
	if companyID == "" {
		return nil, fmt.Errorf("intercom: company ID is required")
	}
	res, err := s.client.generated.ListCompanyNotesWithResponse(ctx, companyID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("list company notes", res.StatusCode(), res.Body, res.JSON200)
}

// ListContacts returns contacts attached to a company.
func (s *CompaniesService) ListContacts(ctx context.Context, companyID string) (*CompanyContacts, error) {
	if companyID == "" {
		return nil, fmt.Errorf("intercom: company ID is required")
	}
	res, err := s.client.generated.ListAttachedContactsWithResponse(ctx, companyID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("list company contacts", res.StatusCode(), res.Body, res.JSON200)
}

// ListSegments returns segments a company belongs to.
func (s *CompaniesService) ListSegments(ctx context.Context, companyID string) (*CompanySegmentsAttached, error) {
	if companyID == "" {
		return nil, fmt.Errorf("intercom: company ID is required")
	}
	res, err := s.client.generated.ListAttachedSegmentsForCompaniesWithResponse(ctx, companyID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("list company segments", res.StatusCode(), res.Body, res.JSON200)
}

// AttachContact attaches a contact to a company.
func (s *CompaniesService) AttachContact(ctx context.Context, contactID, companyID string) (*Company, error) {
	if contactID == "" {
		return nil, fmt.Errorf("intercom: contact ID is required")
	}
	if companyID == "" {
		return nil, fmt.Errorf("intercom: company ID is required")
	}
	res, err := s.client.generated.AttachContactToACompanyWithResponse(ctx, contactID, nil, gen.AttachContactToACompanyJSONRequestBody{Id: companyID})
	if err != nil {
		return nil, err
	}
	return requireOK("attach contact to company", res.StatusCode(), res.Body, res.JSON200)
}

// DetachContact detaches a contact from a company.
func (s *CompaniesService) DetachContact(ctx context.Context, contactID, companyID string) (*Company, error) {
	if contactID == "" {
		return nil, fmt.Errorf("intercom: contact ID is required")
	}
	if companyID == "" {
		return nil, fmt.Errorf("intercom: company ID is required")
	}
	res, err := s.client.generated.DetachContactFromACompanyWithResponse(ctx, contactID, companyID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("detach contact from company", res.StatusCode(), res.Body, res.JSON200)
}

// ListForContact returns companies a contact belongs to.
func (s *CompaniesService) ListForContact(ctx context.Context, contactID string) (*ContactCompanies, error) {
	if contactID == "" {
		return nil, fmt.Errorf("intercom: contact ID is required")
	}
	res, err := s.client.generated.ListCompaniesForAContactWithResponse(ctx, contactID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("list companies for contact", res.StatusCode(), res.Body, res.JSON200)
}
