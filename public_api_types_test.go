package intercom_test

import (
	"testing"

	intercom "github.com/uffejaeger/intercom-go"
)

// TestAPI216WrapperTypesAreImportable verifies that new public wrapper
// signatures can be named and constructed by downstream modules. Generated
// types stay behind the module's internal boundary.
func TestAPI216WrapperTypesAreImportable(t *testing.T) {
	_ = intercom.AdminActivityLogEventTypesParams{}
	_ = intercom.AdminActivityLogSearchParams{}
	_ = intercom.AdminActivityLogSearch{}
	_ = intercom.ArticleVersionListParams{}
	_ = intercom.ConversationDeletedListParams{}
	_ = intercom.ConversationMerge{}
	_ = intercom.ConversationSideListParams{}
	_ = intercom.CustomObjectInstanceListParams{}
	_ = intercom.FinCSATSubmission{}
	_ = intercom.OfficeHoursTimeInterval{}
	_ = intercom.OfficeHoursScheduleCreate{
		Name:          "Weekdays",
		TimeIntervals: []intercom.OfficeHoursTimeInterval{},
		TimeZoneName:  "UTC",
	}
	_ = intercom.OfficeHoursScheduleListParams{}
	_ = intercom.OfficeHoursExceptionListParams{}
	_ = intercom.TeamMetricsParams{}
	_ = intercom.TicketTypeChange{}
	_ = intercom.TicketConversationLink{}
	_ = intercom.WhatsAppMessageStatusParams{}
	_ = intercom.WhatsAppMessageStatusRetrieveParams{}
}

func TestContentAndAudienceRequestItemsAreImportable(t *testing.T) {
	predicate := intercom.AudiencePredicate{Attribute: stringPtr("email"), Comparison: stringPtr("equals"), Value: stringPtr("contact@example.com")}
	_ = intercom.AudienceCreate{Predicates: &[]intercom.AudiencePredicate{predicate}}
	_ = intercom.AudienceUpdate{RolePredicates: &[]intercom.AudiencePredicate{predicate}}

	_ = intercom.ContentBulkActionRequest{
		ContentIds: []intercom.ContentBulkActionContentID{{Id: "content-1", Type: "content_snippet"}},
		Audience:   &intercom.ContentBulkActionAudience{},
		Availability: &intercom.ContentBulkActionAvailability{
			AiAgent: boolPtr(true),
		},
		Tags: &intercom.ContentBulkActionTags{},
	}
}

func TestDataConnectorConfigurationTypesAreImportable(t *testing.T) {
	method := intercom.DataConnectorCreateHTTPMethod("POST")
	inputType := intercom.DataConnectorCreateDataInputType("string")
	createAudience := intercom.DataConnectorCreateAudience("users")
	_ = intercom.DataConnectorCreate{
		Name:       "lookup",
		Audiences:  &[]intercom.DataConnectorCreateAudience{createAudience},
		HttpMethod: &method,
		DataInputs: &[]intercom.DataConnectorCreateDataInput{{Name: stringPtr("query"), Type: &inputType}},
		Headers:    &[]intercom.DataConnectorHeader{{Name: stringPtr("X-API-Key"), Value: stringPtr("token")}},
	}

	updateMethod := intercom.DataConnectorUpdateHTTPMethod("GET")
	updateInputType := intercom.DataConnectorUpdateDataInputType("string")
	updateAudience := intercom.DataConnectorUpdateAudience("leads")
	state := intercom.DataConnectorUpdateState("live")
	_ = intercom.DataConnectorUpdate{
		Audiences:  &[]intercom.DataConnectorUpdateAudience{updateAudience},
		HttpMethod: &updateMethod,
		DataInputs: &[]intercom.DataConnectorUpdateDataInput{{Name: stringPtr("query"), Type: &updateInputType}},
		Headers:    &[]intercom.DataConnectorHeader{{Name: stringPtr("Accept"), Value: stringPtr("application/json")}},
		State:      &state,
	}
}

func TestContactOwnerIDCompatibility(t *testing.T) {
	ownerID := 42
	email := "contact@example.com"
	_ = intercom.ContactCreate{OwnerId: &ownerID}
	_ = intercom.ContactUpdate{OwnerId: &ownerID}
	contact := intercom.Contact{Email: &email, OwnerId: &ownerID}
	var _ *int = contact.OwnerId
	_ = intercom.VisitorConverted{Email: &email, OwnerId: &ownerID}
	_ = intercom.CompanyContacts{Data: &[]intercom.Contact{contact}}
}

func TestTicketAssigneeIDCompatibility(t *testing.T) {
	ticketID := "ticket-1"
	ticket := intercom.Ticket{Id: &ticketID}
	var _ *string = ticket.AdminAssigneeId
	var _ *string = ticket.TeamAssigneeId
}

func TestArticleParentCompatibility(t *testing.T) {
	parentID := 42
	parentType := "collection"
	article := intercom.Article{ParentId: &parentID, ParentType: &parentType}
	_ = intercom.ArticleList{Data: &[]intercom.Article{article}}
}

func TestConversationAttributeConstructorsAreImportable(t *testing.T) {
	constructors := []func() error{
		func() error {
			_, err := intercom.NewConversationAttributeString(intercom.ConversationAttributeStringCreate{Name: "summary"})
			return err
		},
		func() error {
			_, err := intercom.NewConversationAttributeInteger(intercom.ConversationAttributeIntegerCreate{Name: "count"})
			return err
		},
		func() error {
			_, err := intercom.NewConversationAttributeList(intercom.ConversationAttributeListCreate{Name: "status"})
			return err
		},
		func() error {
			_, err := intercom.NewConversationAttributeDecimal(intercom.ConversationAttributeDecimalCreate{Name: "amount"})
			return err
		},
		func() error {
			_, err := intercom.NewConversationAttributeBoolean(intercom.ConversationAttributeBooleanCreate{Name: "enabled"})
			return err
		},
		func() error {
			_, err := intercom.NewConversationAttributeDatetime(intercom.ConversationAttributeDatetimeCreate{Name: "occurred_at"})
			return err
		},
		func() error {
			_, err := intercom.NewConversationAttributeRelationship(intercom.ConversationAttributeRelationshipCreate{Name: "account"})
			return err
		},
		func() error {
			_, err := intercom.NewConversationAttributeFiles(intercom.ConversationAttributeFilesCreate{Name: "attachments"})
			return err
		},
	}
	for _, constructor := range constructors {
		if err := constructor(); err != nil {
			t.Fatalf("constructor returned error: %v", err)
		}
	}
}

func stringPtr(value string) *string { return &value }

func boolPtr(value bool) *bool { return &value }
