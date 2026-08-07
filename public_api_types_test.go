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
}
