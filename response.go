package intercom

import (
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
