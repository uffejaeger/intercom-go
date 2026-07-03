package intercom

import gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"

// PageInfo describes cursor pagination metadata returned by many Intercom list endpoints.
type PageInfo struct {
	Type       string             `json:"type,omitempty"`
	Page       int                `json:"page,omitempty"`
	PerPage    int                `json:"per_page,omitempty"`
	TotalPages int                `json:"total_pages,omitempty"`
	Next       *StartingAfterPage `json:"next,omitempty"`
}

// StartingAfterPage describes the next cursor page in Intercom responses.
type StartingAfterPage struct {
	PerPage       int    `json:"per_page,omitempty"`
	StartingAfter string `json:"starting_after,omitempty"`
}

// HasNext reports whether another cursor page is available.
func (p PageInfo) HasNext() bool {
	return p.Next != nil && p.Next.StartingAfter != ""
}

// PageOptions controls page-number pagination for endpoints that use page and per_page query parameters.
type PageOptions struct {
	Page    int
	PerPage int
}

// CursorPageOptions controls cursor pagination for endpoints that use per_page and starting_after query parameters.
type CursorPageOptions struct {
	PerPage       int
	StartingAfter string
}

// SearchPagination controls cursor pagination in Intercom search request bodies.
type SearchPagination = gen.StartingAfterPagingSchema

// NewSearchPagination converts cursor options into a search request pagination body.
func NewSearchPagination(options CursorPageOptions) *SearchPagination {
	if options.PerPage <= 0 && options.StartingAfter == "" {
		return nil
	}
	pagination := &SearchPagination{}
	if options.PerPage > 0 {
		pagination.PerPage = &options.PerPage
	}
	if options.StartingAfter != "" {
		pagination.StartingAfter = &options.StartingAfter
	}
	return pagination
}
