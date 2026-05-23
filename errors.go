package intercom

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Error describes one Intercom API error.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse is returned when Intercom responds with a non-2xx status code.
type ErrorResponse struct {
	StatusCode int
	Type       string  `json:"type"`
	RequestID  string  `json:"request_id"`
	Errors     []Error `json:"errors"`
	Body       string  `json:"-"`
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

func parseErrorResponse(statusCode int, body []byte) error {
	apiErr := &ErrorResponse{
		StatusCode: statusCode,
		Body:       string(body),
	}

	if err := json.Unmarshal(body, apiErr); err != nil {
		return &ErrorResponse{
			StatusCode: statusCode,
			Body:       string(body),
			Errors: []Error{{
				Code:    "invalid_error_response",
				Message: strings.TrimSpace(string(body)),
			}},
		}
	}

	return apiErr
}
