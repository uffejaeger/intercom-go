//go:build live

package intercom

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

func TestLiveAPIReadCompatibility(t *testing.T) {
	client := newLiveAPIClient(t)

	t.Run("issue 77 admin activity pagination", func(t *testing.T) {
		ctx, cancel := liveAPIContext(t)
		defer cancel()
		_, err := client.Admins.ListActivityLogs(ctx, strconv.FormatInt(time.Now().Add(-365*24*time.Hour).Unix(), 10))
		liveRequireNoError(t, err)
	})

	t.Run("issue 78 call IDs", func(t *testing.T) {
		ctx, cancel := liveAPIContext(t)
		defer cancel()
		_, err := client.Calls.List(ctx)
		liveRequireNoError(t, err)
	})

	t.Run("issue 79 conversation statistics", func(t *testing.T) {
		ctx, cancel := liveAPIContext(t)
		defer cancel()
		_, err := client.Conversations.List(ctx)
		liveRequireNoError(t, err)
	})

	t.Run("issue 80 email timestamps", func(t *testing.T) {
		ctx, cancel := liveAPIContext(t)
		defer cancel()
		_, err := client.Emails.List(ctx)
		liveRequireNoError(t, err)
	})

	t.Run("issue 84 admin avatar", func(t *testing.T) {
		ctx, cancel := liveAPIContext(t)
		defer cancel()
		me, err := client.Admins.Me(ctx)
		liveRequireNoError(t, err)
		if me.Id == nil || *me.Id == "" {
			t.Fatal("Admins.Me returned no ID")
		}
		_, err = client.Admins.Retrieve(ctx, *me.Id)
		liveRequireNoError(t, err)
	})

	t.Run("issues 88 98 and 100 typed errors", func(t *testing.T) {
		checks := []struct {
			name string
			call func(context.Context) error
		}{
			{name: "brand", call: func(ctx context.Context) error {
				_, err := client.Brands.Retrieve(ctx, "intercom-go-missing")
				return err
			}},
			{name: "macro", call: func(ctx context.Context) error { _, err := client.Macros.Get(ctx, "intercom-go-missing"); return err }},
			{name: "team", call: func(ctx context.Context) error {
				_, err := client.Teams.Retrieve(ctx, "intercom-go-missing")
				return err
			}},
			{name: "job", call: func(ctx context.Context) error {
				_, err := client.Workspace.JobStatus(ctx, "intercom-go-missing")
				return err
			}},
			{name: "visitor", call: func(ctx context.Context) error {
				_, err := client.Visitors.GetByUserID(ctx, "intercom-go-missing")
				return err
			}},
		}
		for _, check := range checks {
			t.Run(check.name, func(t *testing.T) {
				ctx, cancel := liveAPIContext(t)
				defer cancel()
				err := check.call(ctx)
				if !IsNotFound(err) {
					t.Fatalf("error = %T %v, want typed 404", err, err)
				}
			})
		}
	})

	t.Run("issue 83 merge history for a listed user", func(t *testing.T) {
		ctx, cancel := liveAPIContext(t)
		defer cancel()
		contacts, err := client.Contacts.List(ctx)
		liveRequireNoError(t, err)
		if contacts.Data == nil {
			t.Skip("workspace has no contacts")
		}
		for _, contact := range *contacts.Data {
			if contact.Id == nil || contact.Role == nil || *contact.Role != "user" {
				continue
			}
			_, err = client.Contacts.ListMergeHistory(ctx, *contact.Id)
			liveRequireNoError(t, err)
			return
		}
		t.Skip("workspace has no listed user contact")
	})

	t.Run("issue 85 audience list detail consistency", func(t *testing.T) {
		ctx, cancel := liveAPIContext(t)
		defer cancel()
		list, err := client.Audiences.List(ctx, nil)
		liveRequireNoError(t, err)
		if list.Data == nil || len(*list.Data) == 0 || (*list.Data)[0].Id == nil {
			t.Skip("workspace has no audiences")
		}
		_, err = client.Audiences.Get(ctx, *(*list.Data)[0].Id)
		liveRequireNoError(t, err)
	})

	t.Run("issue 86 content snippet list detail consistency", func(t *testing.T) {
		ctx, cancel := liveAPIContext(t)
		defer cancel()
		list, err := client.Content.ListSnippets(ctx, nil)
		liveRequireNoError(t, err)
		if list.Data == nil || len(*list.Data) == 0 || (*list.Data)[0].Id == nil {
			t.Skip("workspace has no content snippets")
		}
		_, err = client.Content.GetSnippet(ctx, *(*list.Data)[0].Id)
		liveRequireNoError(t, err)
	})

	t.Run("issue 87 news item list detail consistency", func(t *testing.T) {
		ctx, cancel := liveAPIContext(t)
		defer cancel()
		list, err := client.News.ListItems(ctx)
		liveRequireNoError(t, err)
		if list.Data == nil || len(*list.Data) == 0 || (*list.Data)[0].Id == nil {
			t.Skip("workspace has no news items")
		}
		_, err = client.News.RetrieveItem(ctx, *(*list.Data)[0].Id)
		liveRequireNoError(t, err)
	})

	t.Run("issue 89 empty article draft", func(t *testing.T) {
		ctx, cancel := liveAPIContext(t)
		defer cancel()
		list, err := client.Articles.List(ctx)
		liveRequireNoError(t, err)
		if list.Data == nil || len(*list.Data) == 0 || (*list.Data)[0].Id == nil {
			t.Skip("workspace has no articles")
		}
		_, err = client.Articles.GetDraft(ctx, *(*list.Data)[0].Id)
		if err != nil && strings.Contains(err.Error(), "unexpected end of JSON input") {
			t.Fatalf("generated JSON error leaked: %v", err)
		}
	})

	t.Run("issue 90 office hours exception list detail consistency", func(t *testing.T) {
		ctx, cancel := liveAPIContext(t)
		defer cancel()
		schedules, err := client.OfficeHours.ListSchedules(ctx, nil)
		liveRequireNoError(t, err)
		if schedules.Data == nil || len(*schedules.Data) == 0 || (*schedules.Data)[0].Id == nil {
			t.Skip("workspace has no office-hours schedules")
		}
		scheduleID := *(*schedules.Data)[0].Id
		exceptions, err := client.OfficeHours.ListExceptions(ctx, scheduleID, nil)
		liveRequireNoError(t, err)
		if exceptions.Data == nil || len(*exceptions.Data) == 0 || (*exceptions.Data)[0].Id == nil {
			t.Skip("workspace has no office-hours exceptions")
		}
		_, err = client.OfficeHours.GetException(ctx, scheduleID, *(*exceptions.Data)[0].Id)
		liveRequireNoError(t, err)
	})

	t.Run("issue 91 content import source list detail consistency", func(t *testing.T) {
		ctx, cancel := liveAPIContext(t)
		defer cancel()
		list, err := client.AIContent.ListContentImportSources(ctx)
		liveRequireNoError(t, err)
		if list.Data == nil || len(*list.Data) == 0 {
			t.Skip("workspace has no content import sources")
		}
		_, err = client.AIContent.GetContentImportSource(ctx, strconv.Itoa((*list.Data)[0].Id))
		liveRequireNoError(t, err)
	})

	t.Run("issue 92 external page list detail consistency", func(t *testing.T) {
		ctx, cancel := liveAPIContext(t)
		defer cancel()
		list, err := client.AIContent.ListExternalPages(ctx)
		liveRequireNoError(t, err)
		if list.Data == nil || len(*list.Data) == 0 {
			t.Skip("workspace has no external pages")
		}
		_, err = client.AIContent.GetExternalPage(ctx, (*list.Data)[0].Id)
		liveRequireNoError(t, err)
	})

	t.Run("issue 93 help center redirect list detail consistency", func(t *testing.T) {
		ctx, cancel := liveAPIContext(t)
		defer cancel()
		helpCenters, err := client.HelpCenters.List(ctx)
		liveRequireNoError(t, err)
		if helpCenters.Data == nil || len(*helpCenters.Data) == 0 || (*helpCenters.Data)[0].Id == nil {
			t.Skip("workspace has no help centers")
		}
		helpCenterID := *(*helpCenters.Data)[0].Id
		redirects, err := client.HelpCenterRedirects.List(ctx, helpCenterID, nil)
		liveRequireNoError(t, err)
		if redirects.Data == nil || len(*redirects.Data) == 0 || (*redirects.Data)[0].Id == nil {
			t.Skip("workspace has no help-center redirects")
		}
		_, err = client.HelpCenterRedirects.Get(ctx, helpCenterID, *(*redirects.Data)[0].Id)
		liveRequireNoError(t, err)
	})

	t.Run("issue 94 data connector execution list detail consistency", func(t *testing.T) {
		ctx, cancel := liveAPIContext(t)
		defer cancel()
		connectors, err := client.DataConnectors.List(ctx, nil)
		liveRequireNoError(t, err)
		if connectors.Data == nil || len(*connectors.Data) == 0 || (*connectors.Data)[0].Id == nil {
			t.Skip("workspace has no data connectors")
		}
		connectorID := *(*connectors.Data)[0].Id
		results, err := client.DataConnectors.ListExecutionResults(ctx, connectorID, nil)
		liveRequireNoError(t, err)
		if results.Data == nil || len(*results.Data) == 0 || (*results.Data)[0].Id == nil {
			t.Skip("workspace has no data-connector executions")
		}
		_, err = client.DataConnectors.GetExecutionResult(ctx, connectorID, *(*results.Data)[0].Id)
		liveRequireNoError(t, err)
	})

	t.Run("issues 95 and 96 newsfeed list detail consistency", func(t *testing.T) {
		ctx, cancel := liveAPIContext(t)
		defer cancel()
		feeds, err := client.News.ListFeeds(ctx)
		liveRequireNoError(t, err)
		if feeds.Data == nil || len(*feeds.Data) == 0 || (*feeds.Data)[0].Id == nil {
			t.Skip("workspace has no newsfeeds")
		}
		feedID := *(*feeds.Data)[0].Id
		_, err = client.News.RetrieveFeed(ctx, feedID)
		liveRequireNoError(t, err)
		_, err = client.News.ListFeedItems(ctx, feedID)
		liveRequireNoError(t, err)
	})

	t.Run("issue 97 team metrics", func(t *testing.T) {
		ctx, cancel := liveAPIContext(t)
		defer cancel()
		teams, err := client.Teams.List(ctx)
		liveRequireNoError(t, err)
		if teams.Teams == nil || len(*teams.Teams) == 0 || (*teams.Teams)[0].Id == nil {
			t.Skip("workspace has no teams")
		}
		idleThreshold := 1800
		_, err = client.Teams.Metrics(ctx, *(*teams.Teams)[0].Id, &TeamMetricsParams{IdleThreshold: &idleThreshold})
		if err != nil && !IsServerError(err) {
			t.Fatalf("live API call failed: %T %v", err, err)
		}
	})

	t.Run("issue 101 existing company contact association", func(t *testing.T) {
		ctx, cancel := liveAPIContext(t)
		defer cancel()
		contacts, err := client.Contacts.List(ctx)
		liveRequireNoError(t, err)
		if contacts.Data == nil {
			t.Skip("workspace has no contacts")
		}
		for _, contact := range *contacts.Data {
			if contact.Id == nil || contact.Companies == nil || contact.Companies.Data == nil {
				continue
			}
			for _, company := range *contact.Companies.Data {
				if company.Id == nil {
					continue
				}
				companies, companiesErr := client.Companies.ListForContact(ctx, *contact.Id)
				contacts, contactsErr := client.Companies.ListContacts(ctx, *company.Id)
				liveRequireNoError(t, companiesErr)
				liveRequireNoError(t, contactsErr)
				if !liveHasCompany(companies, *company.Id) || !liveHasContact(contacts, *contact.Id) {
					t.Fatalf("known association missing from list endpoints: contact=%s company=%s", *contact.Id, *company.Id)
				}
				return
			}
		}
		t.Skip("workspace has no listed contact/company associations")
	})
}

func TestLiveAPIFixtureCompatibility(t *testing.T) {
	client := newLiveAPIClient(t)
	unique := liveUnique("intercom-go-live")
	role := "user"
	name := "Intercom Go Live Regression"
	email := unique + "@example.com"
	externalID := unique

	ctx, cancel := liveAPIContext(t)
	contact, err := client.Contacts.Create(ctx, ContactCreate{
		Email:      &email,
		ExternalId: &externalID,
		Name:       &name,
		Role:       &role,
	})
	cancel()
	liveRequireNoError(t, err)
	if contact.Id == nil || *contact.Id == "" {
		t.Fatal("created contact has no ID")
	}
	contactID := *contact.Id
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = client.Contacts.Delete(ctx, contactID)
	})

	companyExternalID := unique + "-company"
	ctx, cancel = liveAPIContext(t)
	company, err := client.Companies.CreateOrUpdate(ctx, CompanyCreate{
		CompanyId: &companyExternalID,
		Name:      &name,
	})
	cancel()
	liveRequireNoError(t, err)
	if company.Id == nil || *company.Id == "" {
		t.Fatal("created company has no ID")
	}
	companyID := *company.Id
	liveWaitForCompany(t, client, companyID)

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = client.Companies.DetachContact(ctx, contactID, companyID)
		_, _ = client.Companies.Delete(ctx, companyID)
	})

	t.Run("issue 99 create note with Intercom contact ID", func(t *testing.T) {
		ctx, cancel := liveAPIContext(t)
		defer cancel()
		admin, err := client.Admins.Me(ctx)
		liveRequireNoError(t, err)
		if admin.Id == nil {
			t.Fatal("Admins.Me returned no ID")
		}
		_, err = client.Contacts.CreateNote(ctx, contactID, "Intercom Go live regression note", *admin.Id)
		liveRequireNoError(t, err)
	})

	t.Run("issue 82 create conversation with Intercom contact ID", func(t *testing.T) {
		ctx, cancel := liveAPIContext(t)
		defer cancel()
		_, err := client.Conversations.Create(ctx, ConversationCreate{
			Body: "Intercom Go live regression conversation",
			From: struct {
				Id   string                                `json:"id"`
				Type gen.CreateConversationRequestFromType `json:"type"`
			}{Id: contactID, Type: "user"},
		})
		liveRequireNoError(t, err)
	})

	t.Run("issue 83 merge history for valid contact", func(t *testing.T) {
		ctx, cancel := liveAPIContext(t)
		defer cancel()
		_, err := client.Contacts.ListMergeHistory(ctx, contactID)
		liveRequireNoError(t, err)
	})

	t.Run("issue 101 company contact association", func(t *testing.T) {
		ctx, cancel := liveAPIContext(t)
		attached, err := client.Companies.AttachContact(ctx, contactID, companyID)
		cancel()
		liveRequireNoError(t, err)
		if attached == nil || attached.Id == nil {
			t.Fatal("attach response has no company ID")
		}
		t.Logf("attach response company ID=%s; created company ID=%s", *attached.Id, companyID)

		deadline := time.Now().Add(5 * time.Minute)
		var lastCompanies *ContactCompanies
		var lastContacts *CompanyContacts
		for {
			ctx, cancel := liveAPIContext(t)
			companies, companiesErr := client.Companies.ListForContact(ctx, contactID)
			contacts, contactsErr := client.Companies.ListContacts(ctx, companyID)
			cancel()
			lastCompanies = companies
			lastContacts = contacts
			if companiesErr == nil && contactsErr == nil && liveHasCompany(companies, companyID) && liveHasContact(contacts, contactID) {
				return
			}
			if time.Now().After(deadline) {
				if companiesErr != nil || contactsErr != nil {
					t.Fatalf("association lists failed: companies=%v contacts=%v", companiesErr, contactsErr)
				}
				t.Fatalf("attached resources absent from association lists: company IDs=%v contact IDs=%v raw contact-company response=%s", liveCompanyIDs(lastCompanies), liveContactIDs(lastContacts), liveRawCompaniesForContact(t, client, contactID))
			}
			time.Sleep(5 * time.Second)
		}
	})
}

func TestLiveAPIContentExportJobCompatibility(t *testing.T) {
	client := newLiveAPIClient(t)
	now := time.Now()

	ctx, cancel := liveAPIContext(t)
	export, err := client.Workspace.CreateDataExport(ctx, DataExportCreate{
		CreatedAtAfter:  int(now.Add(-2 * time.Hour).Unix()),
		CreatedAtBefore: int(now.Add(-time.Hour).Unix()),
	})
	cancel()
	liveRequireNoError(t, err)
	if export.JobIdentifier == nil || *export.JobIdentifier == "" {
		t.Fatal("created data export has no job identifier")
	}
	jobID := *export.JobIdentifier
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = client.Workspace.CancelDataExport(ctx, jobID)
	})

	t.Run("issue 98 real export identifier has typed job status", func(t *testing.T) {
		ctx, cancel := liveAPIContext(t)
		defer cancel()
		_, err := client.Workspace.JobStatus(ctx, jobID)
		if err == nil {
			return
		}
		var apiErr *ErrorResponse
		if !errors.As(err, &apiErr) {
			t.Fatalf("JobStatus error = %T %v, want job or typed API response", err, err)
		}
	})

	ctx, cancel = liveAPIContext(t)
	got, err := client.Workspace.GetDataExport(ctx, jobID)
	cancel()
	liveRequireNoError(t, err)
	if got.JobIdentifier == nil || *got.JobIdentifier != jobID {
		t.Fatalf("GetDataExport job identifier = %v, want %q", got.JobIdentifier, jobID)
	}

	ctx, cancel = liveAPIContext(t)
	_, cancelErr := client.Workspace.CancelDataExport(ctx, jobID)
	cancel()
	if cancelErr != nil {
		t.Logf("content export could not be cancelled after verification: %v", cancelErr)
	}
}

func newLiveAPIClient(t *testing.T) *Client {
	t.Helper()
	if os.Getenv("INTERCOM_LIVE_TESTS") != "1" {
		t.Skip("set INTERCOM_LIVE_TESTS=1 to run live API tests")
	}
	client, err := NewClientFromEnv(WithRetry(RetryConfig{
		MaxAttempts:    3,
		InitialBackoff: 250 * time.Millisecond,
		MaxBackoff:     2 * time.Second,
		Jitter:         0.1,
	}))
	if err != nil {
		t.Fatalf("NewClientFromEnv: %v", err)
	}
	return client
}

func liveAPIContext(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), 30*time.Second)
}

func liveRequireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("live API call failed: %T %v", err, err)
	}
}

func liveUnique(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func liveHasCompany(companies *ContactCompanies, companyID string) bool {
	if companies == nil || companies.Companies == nil {
		return false
	}
	for _, company := range *companies.Companies {
		if company.Id != nil && *company.Id == companyID {
			return true
		}
	}
	return false
}

func liveHasContact(contacts *CompanyContacts, contactID string) bool {
	if contacts == nil || contacts.Data == nil {
		return false
	}
	for _, contact := range *contacts.Data {
		if contact.Id != nil && *contact.Id == contactID {
			return true
		}
	}
	return false
}

func liveCompanyIDs(companies *ContactCompanies) []string {
	if companies == nil || companies.Companies == nil {
		return nil
	}
	ids := make([]string, 0, len(*companies.Companies))
	for _, company := range *companies.Companies {
		if company.Id != nil {
			ids = append(ids, *company.Id)
		}
	}
	return ids
}

func liveContactIDs(contacts *CompanyContacts) []string {
	if contacts == nil || contacts.Data == nil {
		return nil
	}
	ids := make([]string, 0, len(*contacts.Data))
	for _, contact := range *contacts.Data {
		if contact.Id != nil {
			ids = append(ids, *contact.Id)
		}
	}
	return ids
}

func liveWaitForCompany(t *testing.T, client *Client, companyID string) {
	t.Helper()
	started := time.Now()
	deadline := started.Add(60 * time.Second)
	for {
		ctx, cancel := liveAPIContext(t)
		_, err := client.Companies.Retrieve(ctx, companyID)
		cancel()
		if err == nil {
			t.Logf("company became readable after %s", time.Since(started).Round(time.Millisecond))
			return
		}
		if !IsNotFound(err) || time.Now().After(deadline) {
			t.Fatalf("created company did not become readable: %T %v", err, err)
		}
		time.Sleep(time.Second)
	}
}

func liveRawCompaniesForContact(t *testing.T, client *Client, contactID string) string {
	t.Helper()
	ctx, cancel := liveAPIContext(t)
	defer cancel()
	response, err := client.generated.ListCompaniesForAContactWithResponse(ctx, contactID, nil)
	if err != nil {
		t.Fatalf("raw list companies for contact: %v", err)
	}
	return string(response.Body)
}
