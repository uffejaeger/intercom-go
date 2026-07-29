package intercom

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
)

// Error describes one Intercom API error.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse is returned when Intercom responds with a non-2xx status code.
type ErrorResponse struct {
	StatusCode int         `json:"-"`
	Type       string      `json:"type"`
	RequestID  string      `json:"request_id"`
	Errors     []Error     `json:"errors"`
	Body       string      `json:"-"`
	Headers    http.Header `json:"-"`
}

func (e *ErrorResponse) Error() string {
	if e == nil {
		return "intercom: unknown error"
	}

	parts := make([]string, 0, len(e.Errors))
	for _, item := range e.Errors {
		if item.Code == "" && item.Message == "" {
			continue
		}
		if item.Code == "" {
			parts = append(parts, item.Message)
			continue
		}
		if item.Message == "" {
			parts = append(parts, item.Code)
			continue
		}
		parts = append(parts, item.Code+": "+item.Message)
	}

	if len(parts) == 0 {
		return fmt.Sprintf("intercom: API error: status %d", e.StatusCode)
	}

	return fmt.Sprintf("intercom: API error: status %d: %s", e.StatusCode, strings.Join(parts, "; "))
}

// IsStatus reports whether err is an Intercom API response with one of the
// supplied HTTP status codes. Wrapped errors are supported.
func IsStatus(err error, statusCodes ...int) bool {
	var apiErr *ErrorResponse
	if !errors.As(err, &apiErr) {
		return false
	}
	return slices.Contains(statusCodes, apiErr.StatusCode)
}

// IsBadRequest reports whether err is an HTTP 400 Intercom API response.
func IsBadRequest(err error) bool {
	return IsStatus(err, http.StatusBadRequest)
}

// IsUnauthorized reports whether err is an HTTP 401 Intercom API response.
func IsUnauthorized(err error) bool {
	return IsStatus(err, http.StatusUnauthorized)
}

// IsForbidden reports whether err is an HTTP 403 Intercom API response.
func IsForbidden(err error) bool {
	return IsStatus(err, http.StatusForbidden)
}

// IsNotFound reports whether err is an HTTP 404 Intercom API response.
func IsNotFound(err error) bool {
	return IsStatus(err, http.StatusNotFound)
}

// IsConflict reports whether err is an HTTP 409 Intercom API response.
func IsConflict(err error) bool {
	return IsStatus(err, http.StatusConflict)
}

// IsRateLimited reports whether err is an HTTP 429 Intercom API response.
func IsRateLimited(err error) bool {
	return IsStatus(err, http.StatusTooManyRequests)
}

// IsServerError reports whether err is a 5xx Intercom API response.
func IsServerError(err error) bool {
	var apiErr *ErrorResponse
	return errors.As(err, &apiErr) && apiErr.StatusCode >= 500 && apiErr.StatusCode <= 599
}

func parseErrorResponse(statusCode int, body []byte, headers ...http.Header) error {
	responseHeaders := firstHeader(headers)
	apiErr := &ErrorResponse{
		StatusCode: statusCode,
		Body:       string(body),
		Headers:    responseHeaders,
	}

	if err := json.Unmarshal(body, apiErr); err != nil {
		apiErr.Errors = []Error{{
			Code:    "invalid_error_response",
			Message: strings.TrimSpace(string(body)),
		}}
	}
	if apiErr.RequestID == "" {
		apiErr.RequestID = responseHeaders.Get(requestIDHeader)
	}

	return apiErr
}

func firstHeader(headers []http.Header) http.Header {
	if len(headers) == 0 || headers[0] == nil {
		return nil
	}
	return headers[0].Clone()
}
