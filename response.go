package intercom

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func requireOK[T any](operation string, statusCode int, body []byte, value *T) (*T, error) {
	if statusCode != http.StatusOK {
		return nil, parseErrorResponse(statusCode, body)
	}
	if value == nil {
		return nil, fmt.Errorf("intercom: %s returned status %d without a response body", operation, statusCode)
	}
	return value, nil
}

func requireEmpty(statusCode int, body []byte) error {
	if statusCode >= 200 && statusCode < 300 {
		return nil
	}
	return parseErrorResponse(statusCode, body)
}

func requireJSON[T any](operation string, statusCode int, body []byte) (*T, error) {
	if statusCode != http.StatusOK {
		return nil, parseErrorResponse(statusCode, body)
	}

	var value T
	if err := json.Unmarshal(body, &value); err != nil {
		return nil, fmt.Errorf("intercom: decode %s response: %w", operation, err)
	}

	return &value, nil
}
