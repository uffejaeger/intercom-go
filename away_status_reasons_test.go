package intercom

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func newAwayStatusReasonsTestClient(t *testing.T, transport http.RoundTripper) *Client {
	t.Helper()
	client, err := NewClient(
		"token",
		WithBaseURL("https://example.test"),
		WithHTTPClient(&http.Client{Transport: transport}),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	return client
}

func TestAwayStatusReasonsServiceRequests(t *testing.T) {
	t.Run("list away status reasons", func(t *testing.T) {
		var gotMethod, gotPath string
		transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
			gotMethod = req.Method
			gotPath = req.URL.Path
			return jsonResponse(req, http.StatusOK, `{"type":"list","data":[{"id":"1","label":"Lunch"}]}`), nil
		})
		client := newAwayStatusReasonsTestClient(t, transport)
		reasons, err := client.AwayStatusReasons.List(context.Background())
		if err != nil {
			t.Fatalf("List returned error: %v", err)
		}
		if len(reasons) != 1 {
			t.Fatalf("len(reasons) = %d, want 1", len(reasons))
		}
		if reasons[0].Label == nil || *reasons[0].Label != "Lunch" {
			t.Fatalf("Label = %v", reasons[0].Label)
		}
		if gotMethod != http.MethodGet {
			t.Fatalf("method = %q, want GET", gotMethod)
		}
		if gotPath != "/away_status_reasons" {
			t.Fatalf("path = %q, want /away_status_reasons", gotPath)
		}
	})

	t.Run("list away status reasons without data", func(t *testing.T) {
		client := newAwayStatusReasonsTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return jsonResponse(req, http.StatusOK, `{"type":"list"}`), nil
		}))
		reasons, err := client.AwayStatusReasons.List(context.Background())
		if err != nil {
			t.Fatalf("List returned error: %v", err)
		}
		if reasons != nil {
			t.Fatalf("reasons = %#v, want nil", reasons)
		}
	})
}

func TestAwayStatusReasonsErrors(t *testing.T) {
	errBody := `{"type":"error.list","errors":[{"code":"unauthorized"}],"request_id":"12a938a3-314e-4939-b773-5cd45738bd21"}`
	t.Run("api error", func(t *testing.T) {
		client := newAwayStatusReasonsTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return jsonResponse(req, http.StatusUnauthorized, errBody), nil
		}))
		if _, err := client.AwayStatusReasons.List(context.Background()); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("transport error", func(t *testing.T) {
		client := newAwayStatusReasonsTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("connection refused")
		}))
		if _, err := client.AwayStatusReasons.List(context.Background()); err == nil {
			t.Fatal("expected transport error")
		}
	})
}
