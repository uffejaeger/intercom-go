package intercom

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestAdminsMe(t *testing.T) {
	var method, path, authorization, version, userAgent, accept string

	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		method = req.Method
		path = req.URL.Path
		authorization = req.Header.Get("Authorization")
		version = req.Header.Get("Intercom-Version")
		userAgent = req.Header.Get("User-Agent")
		accept = req.Header.Get("Accept")

		return &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body:    io.NopCloser(strings.NewReader(`{"type":"admin","id":"991267459","email":"admin@example.com","name":"Admin User"}`)),
			Request: req,
		}, nil
	})

	client, err := NewClient(
		"token",
		WithBaseURL("https://example.test"),
		WithHTTPClient(&http.Client{Transport: transport}),
		WithUserAgent("test-agent"),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	admin, err := client.Admins.Me(context.Background())
	if err != nil {
		t.Fatalf("Me returned error: %v", err)
	}

	if method != http.MethodGet {
		t.Fatalf("method = %q", method)
	}
	if path != "/me" {
		t.Fatalf("path = %q", path)
	}
	if authorization != "Bearer token" {
		t.Fatalf("Authorization = %q", authorization)
	}
	if version != defaultAPIVersion {
		t.Fatalf("Intercom-Version = %q", version)
	}
	if userAgent != "test-agent" {
		t.Fatalf("User-Agent = %q", userAgent)
	}
	if accept != "application/json" {
		t.Fatalf("Accept = %q", accept)
	}
	if admin.Id == nil || *admin.Id != "991267459" {
		t.Fatalf("admin.Id = %v", admin.Id)
	}
	if admin.Email == nil || *admin.Email != "admin@example.com" {
		t.Fatalf("admin.Email = %v", admin.Email)
	}
}

func TestAdminsMeError(t *testing.T) {
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body:    io.NopCloser(strings.NewReader(`{"type":"error.list","request_id":"req-1","errors":[{"code":"unauthorized","message":"Access Token Invalid"}]}`)),
			Request: req,
		}, nil
	})

	client, err := NewClient(
		"token",
		WithBaseURL("https://example.test"),
		WithHTTPClient(&http.Client{Transport: transport}),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.Admins.Me(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	apiErr, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("error type = %T, want *ErrorResponse", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("StatusCode = %d", apiErr.StatusCode)
	}
	if apiErr.RequestID != "req-1" {
		t.Fatalf("RequestID = %q", apiErr.RequestID)
	}
}

func TestAdminsMeTransportError(t *testing.T) {
	transport := roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("network down")
	})

	client, err := NewClient(
		"token",
		WithBaseURL("https://example.test"),
		WithHTTPClient(&http.Client{Transport: transport}),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.Admins.Me(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}
