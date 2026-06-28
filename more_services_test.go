package intercom

import (
	"context"
	"net/http"
	"testing"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

func TestMoreServicesRequests(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		call       func(context.Context, *Client) error
		wantMethod string
		wantPath   string
	}{
		{
			name:       "list collections",
			statusCode: http.StatusOK,
			response:   `{"type":"list","data":[{"id":"159","name":"English collection title"}],"total_count":1}`,
			call: func(ctx context.Context, client *Client) error {
				collections, err := client.Collections.List(ctx)
				if err != nil {
					return err
				}
				if collections.Data == nil || len(*collections.Data) != 1 {
					t.Fatalf("collections.Data = %#v", collections.Data)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/help_center/collections",
		},
		{
			name:       "create collection",
			statusCode: http.StatusOK,
			response:   `{"id":"165","name":"Thanks for everything"}`,
			call: func(ctx context.Context, client *Client) error {
				name := "Thanks for everything"
				collection, err := client.Collections.Create(ctx, CollectionCreate{Name: name})
				if err != nil {
					return err
				}
				if collection.Id == nil || *collection.Id != "165" {
					t.Fatalf("collection.Id = %v", collection.Id)
				}
				return nil
			},
			wantMethod: http.MethodPost,
			wantPath:   "/help_center/collections",
		},
		{
			name:       "retrieve collection",
			statusCode: http.StatusOK,
			response:   `{"id":"170","name":"English collection title"}`,
			call: func(ctx context.Context, client *Client) error {
				collection, err := client.Collections.Retrieve(ctx, "170")
				if err != nil {
					return err
				}
				if collection.Id == nil || *collection.Id != "170" {
					t.Fatalf("collection.Id = %v", collection.Id)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/help_center/collections/170",
		},
		{
			name:       "update collection",
			statusCode: http.StatusOK,
			response:   `{"id":"176","name":"Update collection name"}`,
			call: func(ctx context.Context, client *Client) error {
				name := "Update collection name"
				collection, err := client.Collections.Update(ctx, "176", CollectionUpdate{Name: &name})
				if err != nil {
					return err
				}
				if collection.Name == nil || *collection.Name != "Update collection name" {
					t.Fatalf("collection.Name = %v", collection.Name)
				}
				return nil
			},
			wantMethod: http.MethodPut,
			wantPath:   "/help_center/collections/176",
		},
		{
			name:       "delete collection",
			statusCode: http.StatusOK,
			response:   `{"id":"182","object":"collection","deleted":true}`,
			call: func(ctx context.Context, client *Client) error {
				return client.Collections.Delete(ctx, "182")
			},
			wantMethod: http.MethodDelete,
			wantPath:   "/help_center/collections/182",
		},
		{
			name:       "list help centers",
			statusCode: http.StatusOK,
			response:   `{"type":"list","data":[{"id":"93","identifier":"help-center-1"}]}`,
			call: func(ctx context.Context, client *Client) error {
				helpCenters, err := client.HelpCenters.List(ctx)
				if err != nil {
					return err
				}
				if helpCenters.Data == nil || len(*helpCenters.Data) != 1 {
					t.Fatalf("helpCenters.Data = %#v", helpCenters.Data)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/help_center/help_centers",
		},
		{
			name:       "retrieve help center",
			statusCode: http.StatusOK,
			response:   `{"id":"93","identifier":"help-center-1"}`,
			call: func(ctx context.Context, client *Client) error {
				helpCenter, err := client.HelpCenters.Retrieve(ctx, "93")
				if err != nil {
					return err
				}
				if helpCenter.Id == nil || *helpCenter.Id != "93" {
					t.Fatalf("helpCenter.Id = %v", helpCenter.Id)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/help_center/help_centers/93",
		},
		{
			name:       "list news items",
			statusCode: http.StatusOK,
			response:   `{"type":"list","data":[],"total_count":0}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.News.ListItems(ctx)
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/news/news_items",
		},
		{
			name:       "create news item",
			statusCode: http.StatusOK,
			response:   `{"id":"33","title":"Halloween is here!"}`,
			call: func(ctx context.Context, client *Client) error {
				title := "Halloween is here!"
				body := "<p>New costumes in store for this spooky season</p>"
				senderID := 991267834
				item, err := client.News.CreateItem(ctx, NewsItemCreate{
					Title:    title,
					Body:     &body,
					SenderId: senderID,
				})
				if err != nil {
					return err
				}
				if item.Id == nil || *item.Id != "33" {
					t.Fatalf("item.Id = %v", item.Id)
				}
				return nil
			},
			wantMethod: http.MethodPost,
			wantPath:   "/news/news_items",
		},
		{
			name:       "retrieve news item",
			statusCode: http.StatusOK,
			response:   `{"id":"34","title":"We have news"}`,
			call: func(ctx context.Context, client *Client) error {
				item, err := client.News.RetrieveItem(ctx, "34")
				if err != nil {
					return err
				}
				if item.Id == nil || *item.Id != "34" {
					t.Fatalf("item.Id = %v", item.Id)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/news/news_items/34",
		},
		{
			name:       "update news item",
			statusCode: http.StatusOK,
			response:   `{"id":"37","title":"Christmas is here!"}`,
			call: func(ctx context.Context, client *Client) error {
				title := "Christmas is here!"
				body := "<p>New gifts in store for the jolly season</p>"
				senderID := 991267845
				item, err := client.News.UpdateItem(ctx, "37", NewsItemUpdate{
					Title:    title,
					Body:     &body,
					SenderId: senderID,
				})
				if err != nil {
					return err
				}
				if item.Title == nil || *item.Title != "Christmas is here!" {
					t.Fatalf("item.Title = %v", item.Title)
				}
				return nil
			},
			wantMethod: http.MethodPut,
			wantPath:   "/news/news_items/37",
		},
		{
			name:       "delete news item",
			statusCode: http.StatusOK,
			response:   `{"id":"40","object":"news-item","deleted":true}`,
			call: func(ctx context.Context, client *Client) error {
				return client.News.DeleteItem(ctx, "40")
			},
			wantMethod: http.MethodDelete,
			wantPath:   "/news/news_items/40",
		},
		{
			name:       "list newsfeeds",
			statusCode: http.StatusOK,
			response:   `{"type":"list","data":[],"total_count":0}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.News.ListFeeds(ctx)
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/news/newsfeeds",
		},
		{
			name:       "retrieve newsfeed",
			statusCode: http.StatusOK,
			response:   `{"id":"72","name":"Visitor Feed"}`,
			call: func(ctx context.Context, client *Client) error {
				feed, err := client.News.RetrieveFeed(ctx, "72")
				if err != nil {
					return err
				}
				if feed.Id == nil || *feed.Id != "72" {
					t.Fatalf("feed.Id = %v", feed.Id)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/news/newsfeeds/72",
		},
		{
			name:       "list newsfeed items",
			statusCode: http.StatusOK,
			response:   `{"type":"list","data":[],"total_count":0}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.News.ListFeedItems(ctx, "72")
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/news/newsfeeds/72/items",
		},
		{
			name:       "create ticket",
			statusCode: http.StatusOK,
			response:   `{"id":"20","type":"ticket"}`,
			call: func(ctx context.Context, client *Client) error {
				var contact gen.CreateTicketRequest_Contacts_Item
				_ = contact.FromCreateTicketRequestContacts0(gen.CreateTicketRequestContacts0{Id: "1234"})
				ticket, err := client.Tickets.Create(ctx, TicketCreate{
					TicketTypeId: "1234",
					Contacts:     []gen.CreateTicketRequest_Contacts_Item{contact},
				})
				if err != nil {
					return err
				}
				if ticket.Id == nil || *ticket.Id != "20" {
					t.Fatalf("ticket.Id = %v", ticket.Id)
				}
				return nil
			},
			wantMethod: http.MethodPost,
			wantPath:   "/tickets",
		},
		{
			name:       "search tickets",
			statusCode: http.StatusOK,
			response:   `{"type":"ticket.list","tickets":[],"total_count":0}`,
			call: func(ctx context.Context, client *Client) error {
				var query TicketSearchQuery
				filter := gen.SingleFilterSearchRequestSchema{}
				field := "state"
				operator := gen.SingleFilterSearchRequestOperator("=")
				var value gen.SingleFilterSearchRequest_Value
				_ = value.FromSingleFilterSearchRequestValue0("open")
				filter.Field = &field
				filter.Operator = &operator
				filter.Value = &value
				_ = query.Query.FromSingleFilterSearchRequestSchema(filter)
				_, err := client.Tickets.Search(ctx, query)
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/tickets/search",
		},
		{
			name:       "get ticket",
			statusCode: http.StatusOK,
			response:   `{"id":"20","type":"ticket"}`,
			call: func(ctx context.Context, client *Client) error {
				ticket, err := client.Tickets.Get(ctx, "20")
				if err != nil {
					return err
				}
				if ticket.Id == nil || *ticket.Id != "20" {
					t.Fatalf("ticket.Id = %v", ticket.Id)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/tickets/20",
		},
		{
			name:       "update ticket",
			statusCode: http.StatusOK,
			response:   `{"id":"20","type":"ticket"}`,
			call: func(ctx context.Context, client *Client) error {
				ticket, err := client.Tickets.Update(ctx, "20", TicketUpdate{})
				if err != nil {
					return err
				}
				if ticket.Id == nil || *ticket.Id != "20" {
					t.Fatalf("ticket.Id = %v", ticket.Id)
				}
				return nil
			},
			wantMethod: http.MethodPut,
			wantPath:   "/tickets/20",
		},
		{
			name:       "delete ticket",
			statusCode: http.StatusOK,
			response:   `{"id":"20","object":"ticket","deleted":true}`,
			call: func(ctx context.Context, client *Client) error {
				return client.Tickets.Delete(ctx, "20")
			},
			wantMethod: http.MethodDelete,
			wantPath:   "/tickets/20",
		},
		{
			name:       "list ticket states",
			statusCode: http.StatusOK,
			response:   `{"type":"list","data":[]}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Tickets.ListStates(ctx)
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/ticket_states",
		},
		{
			name:       "list ticket types",
			statusCode: http.StatusOK,
			response:   `{"type":"list","data":[]}`,
			call: func(ctx context.Context, client *Client) error {
				_, err := client.Tickets.ListTypes(ctx)
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/ticket_types",
		},
		{
			name:       "get ticket type",
			statusCode: http.StatusOK,
			response:   `{"id":"74","type":"ticket_type","name":"Bug Report"}`,
			call: func(ctx context.Context, client *Client) error {
				ticketType, err := client.Tickets.GetType(ctx, "74")
				if err != nil {
					return err
				}
				if ticketType.Id == nil || *ticketType.Id != "74" {
					t.Fatalf("ticketType.Id = %v", ticketType.Id)
				}
				return nil
			},
			wantMethod: http.MethodGet,
			wantPath:   "/ticket_types/74",
		},
		{
			name:       "create ticket type",
			statusCode: http.StatusOK,
			response:   `{"id":"74","type":"ticket_type","name":"Bug"}`,
			call: func(ctx context.Context, client *Client) error {
				ticketType, err := client.Tickets.CreateType(ctx, TicketTypeCreate{Name: "Bug"})
				if err != nil {
					return err
				}
				if ticketType.Id == nil || *ticketType.Id != "74" {
					t.Fatalf("ticketType.Id = %v", ticketType.Id)
				}
				return nil
			},
			wantMethod: http.MethodPost,
			wantPath:   "/ticket_types",
		},
		{
			name:       "update ticket type",
			statusCode: http.StatusOK,
			response:   `{"id":"74","type":"ticket_type","name":"Bug Report 2"}`,
			call: func(ctx context.Context, client *Client) error {
				name := "Bug Report 2"
				ticketType, err := client.Tickets.UpdateType(ctx, "74", TicketTypeUpdate{Name: &name})
				if err != nil {
					return err
				}
				if ticketType.Name == nil || *ticketType.Name != "Bug Report 2" {
					t.Fatalf("ticketType.Name = %v", ticketType.Name)
				}
				return nil
			},
			wantMethod: http.MethodPut,
			wantPath:   "/ticket_types/74",
		},
		{
			name:       "create ticket type attribute",
			statusCode: http.StatusOK,
			response:   `{"id":"185","type":"ticket_type_attribute","name":"Bug Priority"}`,
			call: func(ctx context.Context, client *Client) error {
				attribute, err := client.Tickets.CreateTypeAttribute(ctx, "74", TicketTypeAttributeCreate{
					Name:        "Bug Priority",
					Description: "Priority level of the bug",
					DataType:    "string",
				})
				if err != nil {
					return err
				}
				if attribute.Id == nil || *attribute.Id != "185" {
					t.Fatalf("attribute.Id = %v", attribute.Id)
				}
				return nil
			},
			wantMethod: http.MethodPost,
			wantPath:   "/ticket_types/74/attributes",
		},
		{
			name:       "update ticket type attribute",
			statusCode: http.StatusOK,
			response:   `{"id":"185","type":"ticket_type_attribute","name":"Bug Priority"}`,
			call: func(ctx context.Context, client *Client) error {
				name := "Bug Priority"
				description := "Priority level of the bug"
				attribute, err := client.Tickets.UpdateTypeAttribute(ctx, "74", "185", TicketTypeAttributeUpdate{
					Name:        &name,
					Description: &description,
				})
				if err != nil {
					return err
				}
				if attribute.Id == nil || *attribute.Id != "185" {
					t.Fatalf("attribute.Id = %v", attribute.Id)
				}
				return nil
			},
			wantMethod: http.MethodPut,
			wantPath:   "/ticket_types/74/attributes/185",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod, gotPath string
			transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				gotMethod = req.Method
				gotPath = req.URL.Path
				return jsonResponse(req, tt.statusCode, tt.response), nil
			})

			client := newSupportingServicesTestClient(t, transport)
			if err := tt.call(context.Background(), client); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotMethod != tt.wantMethod {
				t.Fatalf("method = %q, want %q", gotMethod, tt.wantMethod)
			}
			if gotPath != tt.wantPath {
				t.Fatalf("path = %q, want %q", gotPath, tt.wantPath)
			}
		})
	}
}

func TestMoreServicesValidation(t *testing.T) {
	client := newSupportingServicesTestClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected request: %s %s", req.Method, req.URL.Path)
		return nil, nil
	}))

	tests := []struct {
		name string
		call func() error
	}{
		{name: "collection ID required", call: func() error { _, err := client.Collections.Retrieve(context.Background(), ""); return err }},
		{name: "collection ID integer", call: func() error { _, err := client.Collections.Retrieve(context.Background(), "abc"); return err }},
		{name: "help center ID required", call: func() error { _, err := client.HelpCenters.Retrieve(context.Background(), ""); return err }},
		{name: "news item ID required", call: func() error { _, err := client.News.RetrieveItem(context.Background(), ""); return err }},
		{name: "news item ID integer", call: func() error { _, err := client.News.RetrieveItem(context.Background(), "abc"); return err }},
		{name: "newsfeed ID required", call: func() error { _, err := client.News.RetrieveFeed(context.Background(), ""); return err }},
		{name: "ticket ID required", call: func() error { _, err := client.Tickets.Get(context.Background(), ""); return err }},
		{name: "ticket type ID required", call: func() error { _, err := client.Tickets.GetType(context.Background(), ""); return err }},
		{name: "ticket attribute ID required", call: func() error {
			_, err := client.Tickets.UpdateTypeAttribute(context.Background(), "74", "", TicketTypeAttributeUpdate{})
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
