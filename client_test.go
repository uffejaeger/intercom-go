package intercom

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestNewClientRequiresToken(t *testing.T) {
	_, err := NewClient(" ")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWithRegion(t *testing.T) {
	client, err := NewClient("token", WithRegion(EU))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	if got, want := client.BaseURL(), "https://api.eu.intercom.io"; got != want {
		t.Fatalf("BaseURL() = %q, want %q", got, want)
	}
}

func TestDoAddsHeaders(t *testing.T) {
	var authorization, version, userAgent string

	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		authorization = req.Header.Get("Authorization")
		version = req.Header.Get("Intercom-Version")
		userAgent = req.Header.Get("User-Agent")
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     make(http.Header),
			Request:    req,
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

	req, err := client.NewRequest(context.Background(), http.MethodGet, "/me", nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}

	res, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	defer res.Body.Close()

	if authorization != "Bearer token" {
		t.Fatalf("Authorization = %q", authorization)
	}
	if version != defaultAPIVersion {
		t.Fatalf("Intercom-Version = %q", version)
	}
	if userAgent != "test-agent" {
		t.Fatalf("User-Agent = %q", userAgent)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
