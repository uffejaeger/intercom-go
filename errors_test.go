package intercom

import (
	"strings"
	"testing"
)

func TestParseErrorResponse(t *testing.T) {
	err := parseErrorResponse(401, []byte(`{"type":"error.list","request_id":"req-1","errors":[{"code":"unauthorized","message":"Access Token Invalid"}]}`))

	apiErr, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("error type = %T, want *ErrorResponse", err)
	}
	if apiErr.StatusCode != 401 {
		t.Fatalf("StatusCode = %d", apiErr.StatusCode)
	}
	if apiErr.RequestID != "req-1" {
		t.Fatalf("RequestID = %q", apiErr.RequestID)
	}
	if !strings.Contains(apiErr.Error(), "unauthorized: Access Token Invalid") {
		t.Fatalf("Error() = %q", apiErr.Error())
	}
}
