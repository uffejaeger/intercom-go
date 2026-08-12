package intercom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// The Intercom API has a handful of response fields whose wire representation
// is broader than the pinned OpenAPI schema. Keep the generated public field
// types stable while accepting the representations observed against the live
// API.

func (value *AdminSchema) UnmarshalJSON(data []byte) error {
	type alias AdminSchema
	decoded := struct {
		*alias
		Avatar json.RawMessage `json:"avatar"`
	}{alias: (*alias)(value)}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	if len(decoded.Avatar) == 0 || bytes.Equal(decoded.Avatar, []byte("null")) {
		return nil
	}

	var avatar string
	if err := json.Unmarshal(decoded.Avatar, &avatar); err == nil {
		value.Avatar = &avatar
		return nil
	}
	var object struct {
		ImageURL *string `json:"image_url"`
	}
	if err := json.Unmarshal(decoded.Avatar, &object); err != nil {
		return fmt.Errorf("decode admin avatar: %w", err)
	}
	value.Avatar = object.ImageURL
	return nil
}

func (value *CallSchema) UnmarshalJSON(data []byte) error {
	type alias CallSchema
	decoded := struct {
		*alias
		AdminID        json.RawMessage `json:"admin_id"`
		ContactID      json.RawMessage `json:"contact_id"`
		ConversationID json.RawMessage `json:"conversation_id"`
		ID             json.RawMessage `json:"id"`
	}{alias: (*alias)(value)}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	fields := []struct {
		name string
		raw  json.RawMessage
		dest **string
	}{
		{name: "admin_id", raw: decoded.AdminID, dest: &value.AdminId},
		{name: "contact_id", raw: decoded.ContactID, dest: &value.ContactId},
		{name: "conversation_id", raw: decoded.ConversationID, dest: &value.ConversationId},
		{name: "id", raw: decoded.ID, dest: &value.Id},
	}
	for _, field := range fields {
		parsed, err := flexibleString(field.raw)
		if err != nil {
			return fmt.Errorf("decode call %s: %w", field.name, err)
		}
		*field.dest = parsed
	}
	return nil
}

func (value *ConversationStatisticsSchema) UnmarshalJSON(data []byte) error {
	type alias ConversationStatisticsSchema
	normalized, err := normalizeIntegralNumberFields(data)
	if err != nil {
		return err
	}
	decoded := struct {
		*alias
		LastClosedByID json.RawMessage `json:"last_closed_by_id"`
	}{alias: (*alias)(value)}
	if err := json.Unmarshal(normalized, &decoded); err != nil {
		return err
	}
	lastClosedByID, err := flexibleString(decoded.LastClosedByID)
	if err != nil {
		return fmt.Errorf("decode last_closed_by_id: %w", err)
	}
	value.LastClosedById = lastClosedByID
	return nil
}

func (value *ConversationResponseTimeSchema) UnmarshalJSON(data []byte) error {
	type alias ConversationResponseTimeSchema
	normalized, err := normalizeIntegralNumberFields(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(normalized, (*alias)(value))
}

func (value *CursorPagesSchema) UnmarshalJSON(data []byte) error {
	type alias CursorPagesSchema
	decoded := struct {
		*alias
		Next    json.RawMessage `json:"next"`
		PerPage json.RawMessage `json:"per_page"`
	}{alias: (*alias)(value)}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	perPage, err := flexibleInt(decoded.PerPage)
	if err != nil {
		return fmt.Errorf("decode cursor per_page: %w", err)
	}
	value.PerPage = perPage

	if len(decoded.Next) == 0 || bytes.Equal(decoded.Next, []byte("null")) {
		value.Next = nil
		return nil
	}
	var cursor string
	if err := json.Unmarshal(decoded.Next, &cursor); err == nil {
		value.Next = &StartingAfterPagingSchema{StartingAfter: &cursor}
		return nil
	}
	var next StartingAfterPagingSchema
	if err := json.Unmarshal(decoded.Next, &next); err != nil {
		return fmt.Errorf("decode cursor next: %w", err)
	}
	value.Next = &next
	return nil
}

func (value *EmailSettingSchema) UnmarshalJSON(data []byte) error {
	type alias EmailSettingSchema
	decoded := struct {
		*alias
		CreatedAt                    json.RawMessage `json:"created_at"`
		ForwardedEmailLastReceivedAt json.RawMessage `json:"forwarded_email_last_received_at"`
		UpdatedAt                    json.RawMessage `json:"updated_at"`
	}{alias: (*alias)(value)}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	fields := []struct {
		name string
		raw  json.RawMessage
		dest **int
	}{
		{name: "created_at", raw: decoded.CreatedAt, dest: &value.CreatedAt},
		{name: "forwarded_email_last_received_at", raw: decoded.ForwardedEmailLastReceivedAt, dest: &value.ForwardedEmailLastReceivedAt},
		{name: "updated_at", raw: decoded.UpdatedAt, dest: &value.UpdatedAt},
	}
	for _, field := range fields {
		parsed, err := flexibleTimestamp(field.raw)
		if err != nil {
			return fmt.Errorf("decode email %s: %w", field.name, err)
		}
		*field.dest = parsed
	}
	return nil
}

func (value *PartAttachmentSchema) UnmarshalJSON(data []byte) error {
	type alias PartAttachmentSchema
	decoded := struct {
		*alias
		Filesize json.RawMessage `json:"filesize"`
		Height   json.RawMessage `json:"height"`
		Width    json.RawMessage `json:"width"`
	}{alias: (*alias)(value)}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	fields := []struct {
		name string
		raw  json.RawMessage
		dest **int
	}{
		{name: "filesize", raw: decoded.Filesize, dest: &value.Filesize},
		{name: "height", raw: decoded.Height, dest: &value.Height},
		{name: "width", raw: decoded.Width, dest: &value.Width},
	}
	for _, field := range fields {
		parsed, err := flexibleInt(field.raw)
		if err != nil {
			return fmt.Errorf("decode attachment %s: %w", field.name, err)
		}
		*field.dest = parsed
	}
	return nil
}

func (value *WhatsappMessageStatusListSchema) UnmarshalJSON(data []byte) error {
	type alias WhatsappMessageStatusListSchema
	var document map[string]json.RawMessage
	if err := json.Unmarshal(data, &document); err != nil {
		return err
	}
	if rawPages := document["pages"]; len(rawPages) != 0 {
		var pages map[string]json.RawMessage
		if err := json.Unmarshal(rawPages, &pages); err != nil {
			return err
		}
		if rawPerPage := pages["per_page"]; len(rawPerPage) != 0 {
			perPage, err := flexibleInt(rawPerPage)
			if err != nil {
				return fmt.Errorf("decode WhatsApp per_page: %w", err)
			}
			if perPage != nil {
				pages["per_page"] = json.RawMessage(strconv.Itoa(*perPage))
			}
		}
		normalized, err := json.Marshal(pages)
		if err != nil {
			return err
		}
		document["pages"] = normalized
	}
	normalized, err := json.Marshal(document)
	if err != nil {
		return err
	}
	return json.Unmarshal(normalized, (*alias)(value))
}

func flexibleString(raw json.RawMessage) (*string, error) {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil, nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return &text, nil
	}
	var number json.Number
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&number); err != nil {
		return nil, err
	}
	text = number.String()
	return &text, nil
}

func flexibleInt(raw json.RawMessage) (*int, error) {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil, nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		integer, err := strconv.Atoi(text)
		if err != nil {
			return nil, err
		}
		return &integer, nil
	}
	var number float64
	if err := json.Unmarshal(raw, &number); err != nil {
		return nil, err
	}
	integer := int(number)
	if float64(integer) != number {
		return nil, fmt.Errorf("%v is not an integer", number)
	}
	return &integer, nil
}

func flexibleTimestamp(raw json.RawMessage) (*int, error) {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil, nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		if integer, err := strconv.Atoi(text); err == nil {
			return &integer, nil
		}
		parsed, err := time.Parse(time.RFC3339Nano, text)
		if err != nil {
			return nil, err
		}
		seconds := int(parsed.Unix())
		return &seconds, nil
	}
	return flexibleInt(raw)
}

func normalizeIntegralNumberFields(data []byte) ([]byte, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		return nil, err
	}
	for name, raw := range object {
		trimmed := bytes.TrimSpace(raw)
		if len(trimmed) == 0 || (!bytes.ContainsAny(trimmed, ".eE")) {
			continue
		}
		number, err := strconv.ParseFloat(string(trimmed), 64)
		if err != nil || number != float64(int64(number)) {
			continue
		}
		object[name] = json.RawMessage(strconv.FormatInt(int64(number), 10))
	}
	return json.Marshal(object)
}
