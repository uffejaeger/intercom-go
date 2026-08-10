package intercom

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func requireOK[T any](operation string, statusCode int, body []byte, value *T, headers ...http.Header) (*T, error) {
	if statusCode != http.StatusOK {
		return nil, parseErrorResponse(statusCode, body, headers...)
	}
	if value == nil {
		return nil, fmt.Errorf("intercom: %s returned status %d without a response body", operation, statusCode)
	}
	return value, nil
}

func requireCreated[T any](operation string, statusCode int, body []byte, value *T, headers ...http.Header) (*T, error) {
	return requireStatus(operation, statusCode, http.StatusCreated, body, value, headers...)
}

func requireStatus[T any](operation string, statusCode, wantStatus int, body []byte, value *T, headers ...http.Header) (*T, error) {
	if statusCode != wantStatus {
		return nil, parseErrorResponse(statusCode, body, headers...)
	}
	if value == nil {
		return nil, fmt.Errorf("intercom: %s returned status %d without a response body", operation, statusCode)
	}
	return value, nil
}

func requireEmpty(statusCode int, body []byte, headers ...http.Header) error {
	if statusCode >= 200 && statusCode < 300 {
		return nil
	}
	return parseErrorResponse(statusCode, body, headers...)
}

func requireJSON[T any](operation string, statusCode int, body []byte, headers ...http.Header) (*T, error) {
	if statusCode != http.StatusOK {
		return nil, parseErrorResponse(statusCode, body, headers...)
	}

	var value T
	if err := json.Unmarshal(body, &value); err != nil {
		return nil, fmt.Errorf("intercom: decode %s response: %w", operation, err)
	}

	return &value, nil
}

func responseHeaders(response *http.Response) http.Header {
	if response == nil {
		return nil
	}
	return response.Header
}
