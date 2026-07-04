package intercom

import (
	"context"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

type cursorIterator[T any] struct {
	fetch      func(startingAfter string) ([]T, string, error)
	items      []T
	index      int
	current    T
	started    bool
	exhausted  bool
	nextCursor string
	err        error
}

func (i *cursorIterator[T]) next() bool {
	for {
		if i.err != nil || i.exhausted {
			return false
		}
		if i.index < len(i.items) {
			i.current = i.items[i.index]
			i.index++
			return true
		}
		if i.started && i.nextCursor == "" {
			i.exhausted = true
			return false
		}

		startingAfter := ""
		if i.started {
			startingAfter = i.nextCursor
		}
		i.items, i.nextCursor, i.err = i.fetch(startingAfter)
		i.index = 0
		i.started = true
		if i.err != nil {
			return false
		}
	}
}

func (i *cursorIterator[T]) value() T {
	return i.current
}

func (i *cursorIterator[T]) iterationErr() error {
	return i.err
}

// ContactIterator lazily iterates over contact search results.
type ContactIterator struct {
	iter *cursorIterator[*Contact]
}

// SearchIter returns a lazy iterator for contact search results.
func (s *ContactsService) SearchIter(ctx context.Context, search ContactSearch) *ContactIterator {
	return &ContactIterator{iter: &cursorIterator[*Contact]{
		fetch: func(startingAfter string) ([]*Contact, string, error) {
			pageSearch := search
			if startingAfter != "" {
				pageSearch.StartingAfter = startingAfter
			}
			page, err := s.Search(ctx, pageSearch)
			if err != nil {
				return nil, "", err
			}
			return contactsFromPage(page), nextCursor(page.Pages), nil
		},
	}}
}

// Next advances the iterator to the next contact.
func (i *ContactIterator) Next() bool {
	return i.iter.next()
}

// Contact returns the current contact. It returns nil before the first successful Next call.
func (i *ContactIterator) Contact() *Contact {
	return i.iter.value()
}

// Err returns the first error encountered while fetching pages.
func (i *ContactIterator) Err() error {
	return i.iter.iterationErr()
}

// ConversationIterator lazily iterates over conversation list items.
type ConversationIterator struct {
	iter *cursorIterator[*ConversationListItem]
}

// ListIter returns a lazy iterator for conversations.
func (s *ConversationsService) ListIter(ctx context.Context, options CursorPageOptions) *ConversationIterator {
	return &ConversationIterator{iter: &cursorIterator[*ConversationListItem]{
		fetch: func(startingAfter string) ([]*ConversationListItem, string, error) {
			pageOptions := options
			if startingAfter != "" {
				pageOptions.StartingAfter = startingAfter
			}
			page, err := s.ListWithOptions(ctx, pageOptions)
			if err != nil {
				return nil, "", err
			}
			return conversationsFromPage(page), nextCursor(page.Pages), nil
		},
	}}
}

// SearchIter returns a lazy iterator for conversation search results.
func (s *ConversationsService) SearchIter(ctx context.Context, query ConversationSearchQuery, options CursorPageOptions) *ConversationIterator {
	return &ConversationIterator{iter: &cursorIterator[*ConversationListItem]{
		fetch: func(startingAfter string) ([]*ConversationListItem, string, error) {
			pageOptions := options
			if startingAfter != "" {
				pageOptions.StartingAfter = startingAfter
			}
			page, err := s.SearchWithOptions(ctx, query, pageOptions)
			if err != nil {
				return nil, "", err
			}
			return conversationsFromPage(page), nextCursor(page.Pages), nil
		},
	}}
}

// Next advances the iterator to the next conversation.
func (i *ConversationIterator) Next() bool {
	return i.iter.next()
}

// Conversation returns the current conversation. It returns nil before the first successful Next call.
func (i *ConversationIterator) Conversation() *ConversationListItem {
	return i.iter.value()
}

// Err returns the first error encountered while fetching pages.
func (i *ConversationIterator) Err() error {
	return i.iter.iterationErr()
}

// TicketIterator lazily iterates over ticket search results.
type TicketIterator struct {
	iter *cursorIterator[*Ticket]
}

// SearchIter returns a lazy iterator for ticket search results.
func (s *TicketsService) SearchIter(ctx context.Context, query TicketSearchQuery, options CursorPageOptions) *TicketIterator {
	return &TicketIterator{iter: &cursorIterator[*Ticket]{
		fetch: func(startingAfter string) ([]*Ticket, string, error) {
			pageOptions := options
			if startingAfter != "" {
				pageOptions.StartingAfter = startingAfter
			}
			page, err := s.SearchWithOptions(ctx, query, pageOptions)
			if err != nil {
				return nil, "", err
			}
			return ticketsFromPage(page), nextCursor(page.Pages), nil
		},
	}}
}

// Next advances the iterator to the next ticket.
func (i *TicketIterator) Next() bool {
	return i.iter.next()
}

// Ticket returns the current ticket. It returns nil before the first successful Next call.
func (i *TicketIterator) Ticket() *Ticket {
	return i.iter.value()
}

// Err returns the first error encountered while fetching pages.
func (i *TicketIterator) Err() error {
	return i.iter.iterationErr()
}

func nextCursor(pages *gen.CursorPagesSchema) string {
	if pages == nil || pages.Next == nil || pages.Next.StartingAfter == nil {
		return ""
	}
	return *pages.Next.StartingAfter
}

func contactsFromPage(page *ContactList) []*Contact {
	if page == nil || page.Data == nil {
		return nil
	}
	contacts := make([]*Contact, 0, len(*page.Data))
	for i := range *page.Data {
		contacts = append(contacts, &(*page.Data)[i])
	}
	return contacts
}

func conversationsFromPage(page *ConversationList) []*ConversationListItem {
	if page == nil || page.Conversations == nil {
		return nil
	}
	conversations := make([]*ConversationListItem, 0, len(*page.Conversations))
	for i := range *page.Conversations {
		conversations = append(conversations, &(*page.Conversations)[i])
	}
	return conversations
}

func ticketsFromPage(page *TicketList) []*Ticket {
	if page == nil || page.Tickets == nil {
		return nil
	}
	tickets := make([]*Ticket, 0, len(*page.Tickets))
	for _, ticket := range *page.Tickets {
		if ticket != nil {
			tickets = append(tickets, ticket)
		}
	}
	return tickets
}
