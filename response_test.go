package intercom

import (
	"errors"
	"io"
	"net/http"
	"strings"
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

func TestResponseHeadersNil(t *testing.T) {
	if headers := responseHeaders(nil); headers != nil {
		t.Fatalf("responseHeaders(nil) = %#v", headers)
	}
}

func TestRequireHTTPJSONTransportFailures(t *testing.T) {
	if _, err := requireHTTPJSON[string]("test operation", nil); err == nil || !strings.Contains(err.Error(), "no HTTP response") {
		t.Fatalf("nil response error = %v", err)
	}

	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       requireHTTPJSONErrorReader{err: errors.New("read failed")},
	}
	if _, err := requireHTTPJSON[string]("test operation", response); err == nil || !strings.Contains(err.Error(), "read failed") {
		t.Fatalf("read error = %v", err)
	}

	response = &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("{")),
	}
	if _, err := requireHTTPJSON[string]("test operation", response); err == nil || !strings.Contains(err.Error(), "decode test operation response") {
		t.Fatalf("decode error = %v", err)
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
			value:      ptr("ignored"),
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

func ptr[T any](value T) *T {
	return &value
}

type requireHTTPJSONErrorReader struct{ err error }

func (reader requireHTTPJSONErrorReader) Read([]byte) (int, error) { return 0, reader.err }
func (requireHTTPJSONErrorReader) Close() error                    { return nil }
