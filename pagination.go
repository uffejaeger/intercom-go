package intercom

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
