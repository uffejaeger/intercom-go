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

func TestErrorResponseError(t *testing.T) {
	tests := []struct {
		name string
		err  *ErrorResponse
		want string
	}{
		{
			name: "nil",
			want: "intercom: unknown error",
		},
		{
			name: "message only",
			err: &ErrorResponse{
				StatusCode: 400,
				Errors:     []Error{{Message: "Bad request"}},
			},
			want: "intercom: API error: status 400: Bad request",
		},
		{
			name: "code only",
			err: &ErrorResponse{
				StatusCode: 404,
				Errors:     []Error{{Code: "not_found"}},
			},
			want: "intercom: API error: status 404: not_found",
		},
		{
			name: "empty errors",
			err: &ErrorResponse{
				StatusCode: 500,
				Errors:     []Error{{}},
			},
			want: "intercom: API error: status 500",
		},
		{
			name: "multiple errors",
			err: &ErrorResponse{
				StatusCode: 422,
				Errors: []Error{
					{Code: "invalid", Message: "Invalid field"},
					{Code: "missing", Message: "Missing field"},
				},
			},
			want: "intercom: API error: status 422: invalid: Invalid field; missing: Missing field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Fatalf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseInvalidErrorResponse(t *testing.T) {
	err := parseErrorResponse(500, []byte(`not json`))

	apiErr, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("error type = %T, want *ErrorResponse", err)
	}
	if apiErr.StatusCode != 500 {
		t.Fatalf("StatusCode = %d", apiErr.StatusCode)
	}
	if apiErr.Body != "not json" {
		t.Fatalf("Body = %q", apiErr.Body)
	}
	if len(apiErr.Errors) != 1 {
		t.Fatalf("Errors length = %d", len(apiErr.Errors))
	}
	if apiErr.Errors[0].Code != "invalid_error_response" {
		t.Fatalf("Error code = %q", apiErr.Errors[0].Code)
	}
}
