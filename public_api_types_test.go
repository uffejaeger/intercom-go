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
	_ = intercom.OfficeHoursScheduleListParams{}
	_ = intercom.OfficeHoursExceptionListParams{}
	_ = intercom.TeamMetricsParams{}
	_ = intercom.TicketTypeChange{}
	_ = intercom.TicketConversationLink{}
	_ = intercom.WhatsAppMessageStatusParams{}
	_ = intercom.WhatsAppMessageStatusRetrieveParams{}
}

func TestContactOwnerIDCompatibility(t *testing.T) {
	ownerID := 42
	_ = intercom.ContactCreate{OwnerId: &ownerID}
	_ = intercom.ContactUpdate{OwnerId: &ownerID}
	var contact intercom.Contact
	var _ *int = contact.OwnerId
}

func TestTicketAssigneeIDCompatibility(t *testing.T) {
	var ticket intercom.Ticket
	var _ *string = ticket.AdminAssigneeId
	var _ *string = ticket.TeamAssigneeId
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
