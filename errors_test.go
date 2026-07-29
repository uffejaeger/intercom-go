package intercom

import (
	"errors"
	"fmt"
	"net/http"
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
	headers := http.Header{
		requestIDHeader:     []string{"req-header"},
		"X-Custom-Response": []string{"original"},
	}
	err := parseErrorResponse(500, []byte(`not json`), headers)

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
	if apiErr.RequestID != "req-header" {
		t.Fatalf("RequestID = %q", apiErr.RequestID)
	}
	headers.Set("X-Custom-Response", "changed")
	if got := apiErr.Headers.Get("X-Custom-Response"); got != "original" {
		t.Fatalf("Headers = %q, want original", got)
	}
	if len(apiErr.Errors) != 1 {
		t.Fatalf("Errors length = %d", len(apiErr.Errors))
	}
	if apiErr.Errors[0].Code != "invalid_error_response" {
		t.Fatalf("Error code = %q", apiErr.Errors[0].Code)
	}
}

func TestErrorStatusHelpers(t *testing.T) {
	tests := []struct {
		name string
		err  error
		test func(error) bool
		want bool
	}{
		{name: "bad request", err: &ErrorResponse{StatusCode: http.StatusBadRequest}, test: IsBadRequest, want: true},
		{name: "unauthorized", err: &ErrorResponse{StatusCode: http.StatusUnauthorized}, test: IsUnauthorized, want: true},
		{name: "forbidden", err: &ErrorResponse{StatusCode: http.StatusForbidden}, test: IsForbidden, want: true},
		{name: "not found wrapped", err: fmt.Errorf("lookup: %w", &ErrorResponse{StatusCode: http.StatusNotFound}), test: IsNotFound, want: true},
		{name: "conflict", err: &ErrorResponse{StatusCode: http.StatusConflict}, test: IsConflict, want: true},
		{name: "rate limited", err: &ErrorResponse{StatusCode: http.StatusTooManyRequests}, test: IsRateLimited, want: true},
		{name: "server error lower bound", err: &ErrorResponse{StatusCode: 500}, test: IsServerError, want: true},
		{name: "server error upper bound", err: &ErrorResponse{StatusCode: 599}, test: IsServerError, want: true},
		{name: "not server error", err: &ErrorResponse{StatusCode: 499}, test: IsServerError},
		{name: "non API error", err: errors.New("boom"), test: IsNotFound},
		{name: "nil error", test: IsServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.test(tt.err); got != tt.want {
				t.Fatalf("helper returned %v, want %v", got, tt.want)
			}
		})
	}

	if !IsStatus(&ErrorResponse{StatusCode: http.StatusConflict}, http.StatusNotFound, http.StatusConflict) {
		t.Fatal("IsStatus did not match one of multiple statuses")
	}
	if IsStatus(&ErrorResponse{StatusCode: http.StatusConflict}) {
		t.Fatal("IsStatus matched an empty status list")
	}
}
