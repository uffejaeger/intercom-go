package intercom

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

func TestConversationIteratorMultiPage(t *testing.T) {
	responses := []string{
		`{"type":"conversation.list","conversations":[{"id":"conv-1"}],"pages":{"next":{"per_page":1,"starting_after":"cursor-1"}}}`,
		`{"type":"conversation.list","conversations":[{"id":"conv-2"}],"pages":{}}`,
	}
	var queries []map[string]string
	client := newIteratorTestClient(t, func(req *http.Request) (*http.Response, error) {
		queries = append(queries, firstQueryValues(req))
		return jsonResponse(req, http.StatusOK, responses[len(queries)-1]), nil
	})

	iter := client.Conversations.ListIter(context.Background(), CursorPageOptions{PerPage: 1})
	got := collectConversationIDs(t, iter)

	if len(got) != 2 || got[0] != "conv-1" || got[1] != "conv-2" {
		t.Fatalf("conversation IDs = %#v, want [conv-1 conv-2]", got)
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("Err() = %v", err)
	}
	assertQueryValues(t, queries[0], map[string]string{"per_page": "1"})
	assertQueryValues(t, queries[1], map[string]string{"per_page": "1", "starting_after": "cursor-1"})
}

func TestConversationIteratorEmptyPage(t *testing.T) {
	client := newIteratorTestClient(t, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(req, http.StatusOK, `{"type":"conversation.list","conversations":[],"pages":{}}`), nil
	})

	iter := client.Conversations.ListIter(context.Background(), CursorPageOptions{})

	if iter.Next() {
		t.Fatal("Next() = true, want false")
	}
	if iter.Next() {
		t.Fatal("second Next() = true, want false")
	}
	if conversation := iter.Conversation(); conversation != nil {
		t.Fatalf("Conversation() = %#v, want nil", conversation)
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("Err() = %v", err)
	}
}

func TestConversationIteratorEarlyStopDoesNotFetchNextPage(t *testing.T) {
	var requests int
	client := newIteratorTestClient(t, func(req *http.Request) (*http.Response, error) {
		requests++
		return jsonResponse(req, http.StatusOK, `{"type":"conversation.list","conversations":[{"id":"conv-1"},{"id":"conv-2"}],"pages":{"next":{"starting_after":"cursor-1"}}}`), nil
	})

	iter := client.Conversations.ListIter(context.Background(), CursorPageOptions{})

	if !iter.Next() {
		t.Fatal("Next() = false, want true")
	}
	if conversation := iter.Conversation(); conversation == nil || conversation.Id == nil || *conversation.Id != "conv-1" {
		t.Fatalf("Conversation() = %#v, want conv-1", conversation)
	}
	if requests != 1 {
		t.Fatalf("requests = %d, want 1", requests)
	}
}

func TestConversationIteratorErrorAfterPartialResults(t *testing.T) {
	var requests int
	client := newIteratorTestClient(t, func(req *http.Request) (*http.Response, error) {
		requests++
		if requests == 1 {
			return jsonResponse(req, http.StatusOK, `{"type":"conversation.list","conversations":[{"id":"conv-1"}],"pages":{"next":{"starting_after":"cursor-1"}}}`), nil
		}
		return jsonResponse(req, http.StatusTooManyRequests, `{"type":"error.list","errors":[{"code":"rate_limit_exceeded"}],"request_id":"req-1"}`), nil
	})

	iter := client.Conversations.ListIter(context.Background(), CursorPageOptions{})

	if !iter.Next() {
		t.Fatal("first Next() = false, want true")
	}
	if conversation := iter.Conversation(); conversation == nil || conversation.Id == nil || *conversation.Id != "conv-1" {
		t.Fatalf("Conversation() = %#v, want conv-1", conversation)
	}
	if iter.Next() {
		t.Fatal("second Next() = true, want false")
	}
	if iter.Next() {
		t.Fatal("third Next() = true, want false")
	}
	if err := iter.Err(); err == nil {
		t.Fatal("Err() = nil, want page fetch error")
	}
	if requests != 2 {
		t.Fatalf("requests = %d, want 2", requests)
	}
}

func TestConversationIteratorPreservesContextCancellationBetweenPages(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var requests int
	client := newIteratorTestClient(t, func(req *http.Request) (*http.Response, error) {
		requests++
		if requests == 1 {
			return jsonResponse(req, http.StatusOK, `{"type":"conversation.list","conversations":[{"id":"conv-1"}],"pages":{"next":{"starting_after":"cursor-1"}}}`), nil
		}
		<-req.Context().Done()
		return nil, req.Context().Err()
	})

	iter := client.Conversations.ListIter(ctx, CursorPageOptions{})

	if !iter.Next() {
		t.Fatal("first Next() = false, want true")
	}
	cancel()
	if iter.Next() {
		t.Fatal("second Next() = true, want false")
	}
	if !errors.Is(iter.Err(), context.Canceled) {
		t.Fatalf("Err() = %v, want context.Canceled", iter.Err())
	}
}

func TestConversationSearchIteratorSendsBodyPagination(t *testing.T) {
	responses := []string{
		`{"type":"conversation.list","conversations":[{"id":"conv-1"}],"pages":{"next":{"starting_after":"cursor-2"}}}`,
		`{"type":"conversation.list","conversations":[{"id":"conv-2"}],"pages":{}}`,
	}
	var bodies []map[string]any
	client := newIteratorTestClient(t, func(req *http.Request) (*http.Response, error) {
		var body map[string]any
		decodeJSONBody(t, req, &body)
		bodies = append(bodies, body)
		return jsonResponse(req, http.StatusOK, responses[len(bodies)-1]), nil
	})

	iter := client.Conversations.SearchIter(context.Background(), testSearchQuery(), CursorPageOptions{
		PerPage:       50,
		StartingAfter: "cursor-1",
	})

	got := collectConversationIDs(t, iter)

	if len(got) != 2 || got[0] != "conv-1" || got[1] != "conv-2" {
		t.Fatalf("conversation IDs = %#v, want [conv-1 conv-2]", got)
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("Err() = %v", err)
	}
	if got := nestedFloat(bodies[0], "pagination", "per_page"); got != 50 {
		t.Fatalf("pagination.per_page = %v, want 50", got)
	}
	if got := nestedString(bodies[0], "pagination", "starting_after"); got != "cursor-1" {
		t.Fatalf("pagination.starting_after = %q, want cursor-1", got)
	}
	if got := nestedString(bodies[1], "pagination", "starting_after"); got != "cursor-2" {
		t.Fatalf("second pagination.starting_after = %q, want cursor-2", got)
	}
}

func TestContactIterator(t *testing.T) {
	responses := []string{
		`{"type":"list","data":[{"id":"contact-1"}],"pages":{"next":{"starting_after":"cursor-2"}}}`,
		`{"type":"list","data":[{"id":"contact-2"}],"pages":{}}`,
	}
	var bodies []map[string]any
	client := newIteratorTestClient(t, func(req *http.Request) (*http.Response, error) {
		var body map[string]any
		decodeJSONBody(t, req, &body)
		bodies = append(bodies, body)
		return jsonResponse(req, http.StatusOK, responses[len(bodies)-1]), nil
	})

	iter := client.Contacts.SearchIter(context.Background(), ContactSearch{
		Field:         "email",
		Operator:      ContactSearchEquals,
		Value:         "customer@example.com",
		PerPage:       25,
		StartingAfter: "cursor-1",
	})

	var ids []string
	for iter.Next() {
		contact := iter.Contact()
		if contact == nil || contact.Id == nil {
			t.Fatalf("Contact() = %#v, want ID", contact)
		}
		ids = append(ids, *contact.Id)
	}
	if len(ids) != 2 || ids[0] != "contact-1" || ids[1] != "contact-2" {
		t.Fatalf("contact IDs = %#v, want [contact-1 contact-2]", ids)
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("Err() = %v", err)
	}
	if got := nestedFloat(bodies[0], "pagination", "per_page"); got != 25 {
		t.Fatalf("pagination.per_page = %v, want 25", got)
	}
	if got := nestedString(bodies[0], "pagination", "starting_after"); got != "cursor-1" {
		t.Fatalf("pagination.starting_after = %q, want cursor-1", got)
	}
	if got := nestedString(bodies[1], "pagination", "starting_after"); got != "cursor-2" {
		t.Fatalf("second pagination.starting_after = %q, want cursor-2", got)
	}
}

func TestTicketIterator(t *testing.T) {
	responses := []string{
		`{"type":"ticket.list","tickets":[{"id":"ticket-1"}],"pages":{"next":{"starting_after":"cursor-2"}}}`,
		`{"type":"ticket.list","tickets":[{"id":"ticket-2"}],"pages":{}}`,
	}
	var bodies []map[string]any
	client := newIteratorTestClient(t, func(req *http.Request) (*http.Response, error) {
		var body map[string]any
		decodeJSONBody(t, req, &body)
		bodies = append(bodies, body)
		return jsonResponse(req, http.StatusOK, responses[len(bodies)-1]), nil
	})

	iter := client.Tickets.SearchIter(context.Background(), testSearchQuery(), CursorPageOptions{
		PerPage:       50,
		StartingAfter: "cursor-1",
	})

	var ids []string
	for iter.Next() {
		ticket := iter.Ticket()
		if ticket == nil || ticket.Id == nil {
			t.Fatalf("Ticket() = %#v, want ID", ticket)
		}
		ids = append(ids, *ticket.Id)
	}
	if len(ids) != 2 || ids[0] != "ticket-1" || ids[1] != "ticket-2" {
		t.Fatalf("ticket IDs = %#v, want [ticket-1 ticket-2]", ids)
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("Err() = %v", err)
	}
	if got := nestedFloat(bodies[0], "pagination", "per_page"); got != 50 {
		t.Fatalf("pagination.per_page = %v, want 50", got)
	}
	if got := nestedString(bodies[0], "pagination", "starting_after"); got != "cursor-1" {
		t.Fatalf("pagination.starting_after = %q, want cursor-1", got)
	}
	if got := nestedString(bodies[1], "pagination", "starting_after"); got != "cursor-2" {
		t.Fatalf("second pagination.starting_after = %q, want cursor-2", got)
	}
}

func TestSearchIteratorsExposeFirstPageErrors(t *testing.T) {
	tests := []struct {
		name string
		call func(context.Context, *Client) error
	}{
		{
			name: "contacts",
			call: func(ctx context.Context, client *Client) error {
				iter := client.Contacts.SearchIter(ctx, ContactSearch{
					Field:    "email",
					Operator: ContactSearchEquals,
					Value:    "customer@example.com",
				})
				if iter.Next() {
					t.Fatal("Next() = true, want false")
				}
				return iter.Err()
			},
		},
		{
			name: "conversations",
			call: func(ctx context.Context, client *Client) error {
				iter := client.Conversations.SearchIter(ctx, testSearchQuery(), CursorPageOptions{})
				if iter.Next() {
					t.Fatal("Next() = true, want false")
				}
				return iter.Err()
			},
		},
		{
			name: "tickets",
			call: func(ctx context.Context, client *Client) error {
				iter := client.Tickets.SearchIter(ctx, testSearchQuery(), CursorPageOptions{})
				if iter.Next() {
					t.Fatal("Next() = true, want false")
				}
				return iter.Err()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newIteratorTestClient(t, func(req *http.Request) (*http.Response, error) {
				return jsonResponse(req, http.StatusUnauthorized, `{"type":"error.list","errors":[{"code":"unauthorized"}],"request_id":"req-1"}`), nil
			})

			if err := tt.call(context.Background(), client); err == nil {
				t.Fatal("Err() = nil, want first page error")
			}
		})
	}
}

func TestIteratorNilPages(t *testing.T) {
	if got := contactsFromPage(nil); got != nil {
		t.Fatalf("contactsFromPage(nil) = %#v, want nil", got)
	}
	if got := conversationsFromPage(nil); got != nil {
		t.Fatalf("conversationsFromPage(nil) = %#v, want nil", got)
	}
	if got := ticketsFromPage(nil); got != nil {
		t.Fatalf("ticketsFromPage(nil) = %#v, want nil", got)
	}
	if got := nextCursor(nil); got != "" {
		t.Fatalf("nextCursor(nil) = %q, want empty", got)
	}
}

func collectConversationIDs(t *testing.T, iter *ConversationIterator) []string {
	t.Helper()
	var ids []string
	for iter.Next() {
		conversation := iter.Conversation()
		if conversation == nil || conversation.Id == nil {
			t.Fatalf("Conversation() = %#v, want ID", conversation)
		}
		ids = append(ids, *conversation.Id)
	}
	return ids
}

func newIteratorTestClient(t *testing.T, roundTrip roundTripFunc) *Client {
	t.Helper()
	client, err := NewClient(
		"token",
		WithBaseURL("https://example.test"),
		WithHTTPClient(&http.Client{Transport: roundTrip}),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	return client
}

func decodeJSONBody(t *testing.T, req *http.Request, v any) {
	t.Helper()
	if err := json.NewDecoder(req.Body).Decode(v); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
}

func testSearchQuery() gen.SearchRequestSchema {
	var query gen.SearchRequestSchema
	filter := gen.SingleFilterSearchRequestSchema{}
	field := "state"
	operator := gen.SingleFilterSearchRequestOperator("=")
	var value gen.SingleFilterSearchRequest_Value
	_ = value.FromSingleFilterSearchRequestValue0("open")
	filter.Field = &field
	filter.Operator = &operator
	filter.Value = &value
	_ = query.Query.FromSingleFilterSearchRequestSchema(filter)
	return query
}
