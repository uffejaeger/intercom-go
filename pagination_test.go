package intercom

import (
	"encoding/json"
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
