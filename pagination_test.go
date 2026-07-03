package intercom

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestPageInfoHasNext(t *testing.T) {
	var page PageInfo
	if err := json.Unmarshal([]byte(`{"next":{"per_page":2,"starting_after":"cursor"}}`), &page); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}

	if !page.HasNext() {
		t.Fatal("HasNext() = false, want true")
	}
	if page.Next.StartingAfter != "cursor" {
		t.Fatalf("StartingAfter = %q", page.Next.StartingAfter)
	}
}

func TestPageInfoHasNextWithoutCursor(t *testing.T) {
	page := PageInfo{Next: &StartingAfterPage{}}

	if page.HasNext() {
		t.Fatal("HasNext() = true, want false")
	}
}

func TestNewSearchPagination(t *testing.T) {
	pagination := NewSearchPagination(CursorPageOptions{
		PerPage:       25,
		StartingAfter: "cursor-1",
	})

	if pagination == nil {
		t.Fatal("NewSearchPagination returned nil")
	}
	if pagination.PerPage == nil || *pagination.PerPage != 25 {
		t.Fatalf("PerPage = %v, want 25", pagination.PerPage)
	}
	if pagination.StartingAfter == nil || *pagination.StartingAfter != "cursor-1" {
		t.Fatalf("StartingAfter = %v, want cursor-1", pagination.StartingAfter)
	}
}

func TestNewSearchPaginationEmpty(t *testing.T) {
	if pagination := NewSearchPagination(CursorPageOptions{}); pagination != nil {
		t.Fatalf("NewSearchPagination = %#v, want nil", pagination)
	}
}

func firstQueryValues(req *http.Request) map[string]string {
	values := req.URL.Query()
	query := make(map[string]string, len(values))
	for key := range values {
		query[key] = values.Get(key)
	}
	return query
}

func assertQueryValues(t *testing.T, got map[string]string, want map[string]string) {
	t.Helper()
	if want == nil {
		want = map[string]string{}
	}
	if len(got) != len(want) {
		t.Fatalf("query = %#v, want %#v", got, want)
	}
	for key, wantValue := range want {
		if gotValue := got[key]; gotValue != wantValue {
			t.Fatalf("query[%q] = %q, want %q; full query = %#v", key, gotValue, wantValue, got)
		}
	}
}
