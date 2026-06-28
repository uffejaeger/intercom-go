package intercom

import (
	"context"
	"fmt"
	"strconv"
)

// NotesService exposes note-related Intercom API operations.
type NotesService struct {
	client *Client
}

// Retrieve returns a note by ID.
func (s *NotesService) Retrieve(ctx context.Context, noteID string) (*Note, error) {
	if noteID == "" {
		return nil, fmt.Errorf("intercom: note ID is required")
	}
	id, err := strconv.Atoi(noteID)
	if err != nil {
		return nil, fmt.Errorf("intercom: note ID %q is not a valid integer: %w", noteID, err)
	}
	res, err := s.client.generated.RetrieveNoteWithResponse(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("retrieve note", res.StatusCode(), res.Body, res.JSON200)
}
