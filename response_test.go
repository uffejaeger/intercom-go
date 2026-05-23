package intercom

import (
	"net/http"
	"testing"
)

func TestRequireOK(t *testing.T) {
	value := "ok"

	got, err := requireOK("test operation", http.StatusOK, nil, &value)
	if err != nil {
		t.Fatalf("requireOK returned error: %v", err)
	}
	if got == nil || *got != value {
		t.Fatalf("value = %v, want %q", got, value)
	}
}

func TestRequireOKErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       []byte
		value      *string
		wantType   any
	}{
		{
			name:       "non ok",
			statusCode: http.StatusUnauthorized,
			body:       []byte(`{"type":"error.list","errors":[{"code":"unauthorized"}]}`),
			value:      new("ignored"),
			wantType:   &ErrorResponse{},
		},
		{
			name:       "missing body",
			statusCode: http.StatusOK,
			wantType:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := requireOK("test operation", tt.statusCode, tt.body, tt.value)
			if err == nil {
				t.Fatal("expected error")
			}
			switch tt.wantType.(type) {
			case *ErrorResponse:
				if _, ok := err.(*ErrorResponse); !ok {
					t.Fatalf("error type = %T, want *ErrorResponse", err)
				}
			case string:
				if err.Error() != "intercom: test operation returned status 200 without a response body" {
					t.Fatalf("error = %q", err)
				}
			}
		})
	}
}

//go:fix inline
func ptr[T any](value T) *T {
	return new(value)
}
