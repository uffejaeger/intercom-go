package intercom

import (
	"encoding/json"
	"strconv"
	"testing"
)

func FuzzParseErrorResponse(f *testing.F) {
	f.Add(400, `{"type":"error.list","request_id":"req-1","errors":[{"code":"bad_request","message":"Bad request"}]}`)
	f.Add(401, `{"type":"error.list","errors":[{"code":"unauthorized"}]}`)
	f.Add(500, `not json`)
	f.Add(503, ``)

	f.Fuzz(func(t *testing.T, statusCode int, body string) {
		err := parseErrorResponse(statusCode, []byte(body))
		apiErr, ok := err.(*ErrorResponse)
		if !ok {
			t.Fatalf("error type = %T, want *ErrorResponse", err)
		}
		if apiErr.StatusCode != statusCode {
			t.Fatalf("StatusCode = %d, want %d", apiErr.StatusCode, statusCode)
		}
		if apiErr.Body != body {
			t.Fatalf("Body = %q, want %q", apiErr.Body, body)
		}
		if apiErr.Error() == "" {
			t.Fatal("Error() returned empty string")
		}
	})
}

func FuzzIntegerIDConversionHelpers(f *testing.F) {
	f.Add("1")
	f.Add("0")
	f.Add("-1")
	f.Add("00123")
	f.Add("")
	f.Add("contact-1")

	f.Fuzz(func(t *testing.T, id string) {
		want, wantErr := strconv.Atoi(id)

		contactID, contactErr := contactIDToInt(id)
		assertFuzzIDConversion(t, "contactIDToInt", contactID, contactErr, want, wantErr)

		conversationID, conversationErr := conversationIDToInt(id)
		assertFuzzIDConversion(t, "conversationIDToInt", conversationID, conversationErr, want, wantErr)
	})
}

func FuzzDataEventListQuery(f *testing.F) {
	f.Add("user-1", "", "", false)
	f.Add("", "ic-1", "", true)
	f.Add("", "", "user@example.com", false)
	f.Add("", "", "", false)
	f.Add("user-1", "", "user@example.com", true)

	f.Fuzz(func(t *testing.T, userID, intercomUserID, email string, summary bool) {
		filter := DataEventListFilter{
			UserID:         userID,
			IntercomUserID: intercomUserID,
			Email:          email,
			Summary:        summary,
		}
		values, err := dataEventListQuery(filter)

		identifiers := countNonEmpty(userID, intercomUserID, email)
		if identifiers != 1 {
			if err == nil {
				t.Fatalf("dataEventListQuery returned nil error for %d identifiers", identifiers)
			}
			return
		}
		if err != nil {
			t.Fatalf("dataEventListQuery returned error: %v", err)
		}
		if got := values.Get("type"); got != "user" {
			t.Fatalf("type = %q, want user", got)
		}
		if summary && values.Get("summary") != "true" {
			t.Fatalf("summary = %q, want true", values.Get("summary"))
		}
		if !summary && values.Get("summary") != "" {
			t.Fatalf("summary = %q, want empty", values.Get("summary"))
		}
		if userID != "" && values.Get("user_id") != userID {
			t.Fatalf("user_id = %q, want %q", values.Get("user_id"), userID)
		}
		if intercomUserID != "" && values.Get("intercom_user_id") != intercomUserID {
			t.Fatalf("intercom_user_id = %q, want %q", values.Get("intercom_user_id"), intercomUserID)
		}
		if email != "" && values.Get("email") != email {
			t.Fatalf("email = %q, want %q", values.Get("email"), email)
		}
	})
}

func FuzzContactSearchToGenerated(f *testing.F) {
	f.Add("email", string(ContactSearchEquals), "user@example.com", 25, "cursor-1")
	f.Add("created_at", string(ContactSearchGreaterThan), "1700000000", 0, "")
	f.Add("", string(ContactSearchEquals), "user@example.com", 10, "")
	f.Add("email", "", "user@example.com", 10, "")

	f.Fuzz(func(t *testing.T, field, operator, value string, perPage int, startingAfter string) {
		search := ContactSearch{
			Field:         field,
			Operator:      ContactSearchOperator(operator),
			Value:         value,
			PerPage:       perPage,
			StartingAfter: startingAfter,
		}
		body, err := search.toGenerated()
		if field == "" || operator == "" {
			if err == nil {
				t.Fatal("toGenerated returned nil error for missing field or operator")
			}
			return
		}
		if err != nil {
			t.Fatalf("toGenerated returned error: %v", err)
		}

		var decoded map[string]any
		mustFuzzJSONRoundTrip(t, body, &decoded)
		if got := fuzzNestedString(decoded, "query", "field"); got != field {
			t.Fatalf("query.field = %q, want %q", got, field)
		}
		if got := fuzzNestedString(decoded, "query", "operator"); got != operator {
			t.Fatalf("query.operator = %q, want %q", got, operator)
		}
		if got := fuzzNestedString(decoded, "query", "value"); got != value {
			t.Fatalf("query.value = %q, want %q", got, value)
		}
		if perPage > 0 {
			if got := fuzzNestedFloat(decoded, "pagination", "per_page"); got != float64(perPage) {
				t.Fatalf("pagination.per_page = %v, want %d", got, perPage)
			}
		}
		if startingAfter != "" {
			if got := fuzzNestedString(decoded, "pagination", "starting_after"); got != startingAfter {
				t.Fatalf("pagination.starting_after = %q, want %q", got, startingAfter)
			}
		}
	})
}

func FuzzTagCompanyUntagRequestMarshalJSON(f *testing.F) {
	f.Add("enterprise", "company-id", "", "", "external-id")
	f.Add("", "", "", "", "")

	f.Fuzz(func(t *testing.T, name, firstCompanyID, firstID, secondCompanyID, secondID string) {
		req := TagCompanyUntagRequest{
			Name: name,
			Companies: []TagCompanyUntagReference{
				{CompanyID: optionalFuzzString(firstCompanyID), ID: optionalFuzzString(firstID)},
				{CompanyID: optionalFuzzString(secondCompanyID), ID: optionalFuzzString(secondID)},
			},
		}

		var decoded struct {
			Companies []struct {
				CompanyID *string `json:"company_id,omitempty"`
				ID        *string `json:"id,omitempty"`
				Untag     bool    `json:"untag"`
			} `json:"companies"`
			Name string `json:"name"`
		}
		mustFuzzJSONRoundTrip(t, req, &decoded)

		if decoded.Name != name {
			t.Fatalf("name = %q, want %q", decoded.Name, name)
		}
		if len(decoded.Companies) != len(req.Companies) {
			t.Fatalf("companies length = %d, want %d", len(decoded.Companies), len(req.Companies))
		}
		for i, company := range decoded.Companies {
			if !company.Untag {
				t.Fatalf("companies[%d].untag = false, want true", i)
			}
		}
	})
}

func FuzzTicketReplyContactPayload(f *testing.F) {
	f.Add("user@example.com", "", "", true, false, false)
	f.Add("", "ic-1", "", false, true, false)
	f.Add("", "", "user-1", false, false, true)
	f.Add("user@example.com", "ic-1", "", true, true, false)
	f.Add("", "", "", false, false, false)

	f.Fuzz(func(t *testing.T, email, intercomUserID, userID string, useEmail, useIntercomUserID, useUserID bool) {
		contact := TicketReplyContact{
			Email:          optionalFuzzStringWhen(useEmail, email),
			IntercomUserID: optionalFuzzStringWhen(useIntercomUserID, intercomUserID),
			UserID:         optionalFuzzStringWhen(useUserID, userID),
		}
		payload, err := contact.payload()

		identifiers := countTrue(useEmail, useIntercomUserID, useUserID)
		if identifiers != 1 {
			if err == nil {
				t.Fatalf("payload returned nil error for %d identifiers", identifiers)
			}
			return
		}
		if err != nil {
			t.Fatalf("payload returned error: %v", err)
		}
		if len(payload) != 1 {
			t.Fatalf("payload length = %d, want 1", len(payload))
		}
		switch {
		case useEmail:
			if payload["email"] != email {
				t.Fatalf("email = %#v, want %q", payload["email"], email)
			}
		case useIntercomUserID:
			if payload["intercom_user_id"] != intercomUserID {
				t.Fatalf("intercom_user_id = %#v, want %q", payload["intercom_user_id"], intercomUserID)
			}
		case useUserID:
			if payload["user_id"] != userID {
				t.Fatalf("user_id = %#v, want %q", payload["user_id"], userID)
			}
		}
	})
}

func assertFuzzIDConversion(t *testing.T, name string, got int, gotErr error, want int, wantErr error) {
	t.Helper()

	if wantErr != nil {
		if gotErr == nil {
			t.Fatalf("%s returned nil error for invalid ID", name)
		}
		return
	}
	if gotErr != nil {
		t.Fatalf("%s returned error: %v", name, gotErr)
	}
	if got != want {
		t.Fatalf("%s = %d, want %d", name, got, want)
	}
}

func mustFuzzJSONRoundTrip(t *testing.T, value any, target any) {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("json.Unmarshal(%q) returned error: %v", string(data), err)
	}
}

func optionalFuzzString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func optionalFuzzStringWhen(enabled bool, value string) *string {
	if !enabled {
		return nil
	}
	return &value
}

func countNonEmpty(values ...string) int {
	count := 0
	for _, value := range values {
		if value != "" {
			count++
		}
	}
	return count
}

func countTrue(values ...bool) int {
	count := 0
	for _, value := range values {
		if value {
			count++
		}
	}
	return count
}

func fuzzNestedString(body map[string]any, keys ...string) string {
	value := fuzzNestedValue(body, keys...)
	text, _ := value.(string)
	return text
}

func fuzzNestedFloat(body map[string]any, keys ...string) float64 {
	value := fuzzNestedValue(body, keys...)
	number, _ := value.(float64)
	return number
}

func fuzzNestedValue(body map[string]any, keys ...string) any {
	var value any = body
	for _, key := range keys {
		next, ok := value.(map[string]any)
		if !ok {
			return nil
		}
		value = next[key]
	}
	return value
}
