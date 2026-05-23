package intercom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	gen "github.com/uffejaeger/intercom-go/internal/generated/intercom"
)

// Call is a phone call in Intercom.
type Call = gen.CallSchema

// CallList is a paginated list of calls.
type CallList = gen.CallListSchema

// FinVoiceCall is the response for a Fin Voice call registration or lookup.
type FinVoiceCall = gen.AiCallResponseSchema

// FinVoiceCallSource identifies the call provider for Fin Voice calls.
type FinVoiceCallSource = gen.RegisterFinVoiceCallRequestSource

const (
	FinVoiceCallSourceAWSConnect FinVoiceCallSource = gen.AwsConnect
	FinVoiceCallSourceFive9      FinVoiceCallSource = gen.Five9
	FinVoiceCallSourceZoomPhone  FinVoiceCallSource = gen.ZoomPhone
)

// RegisterFinVoiceCallRequest holds the fields for registering a Fin Voice call.
type RegisterFinVoiceCallRequest = gen.RegisterFinVoiceCallRequestSchema

// CallWithTranscript is a call with its full transcript data.
type CallWithTranscript struct {
	AdminId             *string                  `json:"admin_id,omitempty"`
	AnsweredAt          *gen.Datetime            `json:"answered_at,omitempty"`
	CallType            *string                  `json:"call_type,omitempty"`
	ContactId           *string                  `json:"contact_id,omitempty"`
	ConversationId      *string                  `json:"conversation_id,omitempty"`
	CreatedAt           *gen.Datetime            `json:"created_at,omitempty"`
	Direction           *string                  `json:"direction,omitempty"`
	EndedAt             *gen.Datetime            `json:"ended_at,omitempty"`
	EndedReason         *string                  `json:"ended_reason,omitempty"`
	FinRecordingUrl     *string                  `json:"fin_recording_url,omitempty"`
	FinTranscriptionUrl *string                  `json:"fin_transcription_url,omitempty"`
	Id                  *string                  `json:"id,omitempty"`
	InitiatedAt         *gen.Datetime            `json:"initiated_at,omitempty"`
	Phone               *string                  `json:"phone,omitempty"`
	RecordingUrl        *string                  `json:"recording_url,omitempty"`
	State               *string                  `json:"state,omitempty"`
	Transcript          []map[string]any `json:"transcript,omitempty"`
	TranscriptStatus    *string                  `json:"transcript_status,omitempty"`
	TranscriptionUrl    *string                  `json:"transcription_url,omitempty"`
	Type                *string                  `json:"type,omitempty"`
	UpdatedAt           *gen.Datetime            `json:"updated_at,omitempty"`
}

// CallWithTranscriptList is the response from ListWithTranscripts.
type CallWithTranscriptList struct {
	Data []CallWithTranscript `json:"data,omitempty"`
	Type *string              `json:"type,omitempty"`
}

// CallsService exposes call-related Intercom API operations.
type CallsService struct {
	client *Client
}

// List returns a paginated list of calls.
func (s *CallsService) List(ctx context.Context) (*CallList, error) {
	res, err := s.client.generated.ListCallsWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("list calls", res.StatusCode(), res.Body, res.JSON200)
}

// Get returns a single call by ID.
func (s *CallsService) Get(ctx context.Context, callID string) (*Call, error) {
	if callID == "" {
		return nil, fmt.Errorf("intercom: call ID is required")
	}
	res, err := s.client.generated.ShowCallWithResponse(ctx, callID, nil)
	if err != nil {
		return nil, err
	}
	return requireOK("get call", res.StatusCode(), res.Body, res.JSON200)
}

// GetRecording downloads the raw recording for a call.
func (s *CallsService) GetRecording(ctx context.Context, callID string) ([]byte, error) {
	if callID == "" {
		return nil, fmt.Errorf("intercom: call ID is required")
	}
	res, err := s.client.generated.ShowCallRecordingWithResponse(ctx, callID, nil)
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		return nil, parseErrorResponse(res.StatusCode(), res.Body)
	}
	return res.Body, nil
}

// GetTranscript downloads the raw transcript for a call.
func (s *CallsService) GetTranscript(ctx context.Context, callID string) ([]byte, error) {
	if callID == "" {
		return nil, fmt.Errorf("intercom: call ID is required")
	}
	res, err := s.client.generated.ShowCallTranscriptWithResponse(ctx, callID, nil)
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		return nil, parseErrorResponse(res.StatusCode(), res.Body)
	}
	return res.Body, nil
}

// ListWithTranscripts returns calls with transcripts for up to 20 conversation IDs.
func (s *CallsService) ListWithTranscripts(ctx context.Context, conversationIDs []string) (*CallWithTranscriptList, error) {
	if len(conversationIDs) == 0 {
		return nil, fmt.Errorf("intercom: at least one conversation ID is required")
	}
	if len(conversationIDs) > 20 {
		return nil, fmt.Errorf("intercom: at most 20 conversation IDs are allowed")
	}
	body := gen.ListCallsWithTranscriptsJSONRequestBody{ConversationIds: conversationIDs}
	res, err := s.client.generated.ListCallsWithTranscriptsWithResponse(ctx, nil, body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		return nil, parseErrorResponse(res.StatusCode(), res.Body)
	}
	var result CallWithTranscriptList
	if err := json.Unmarshal(res.Body, &result); err != nil {
		return nil, fmt.Errorf("intercom: list calls with transcripts: %w", err)
	}
	return &result, nil
}

// RegisterFinVoiceCall registers a new Fin Voice call.
func (s *CallsService) RegisterFinVoiceCall(ctx context.Context, req RegisterFinVoiceCallRequest) (*FinVoiceCall, error) {
	res, err := s.client.generated.RegisterFinVoiceCallWithResponse(ctx, gen.RegisterFinVoiceCallJSONRequestBody(req))
	if err != nil {
		return nil, err
	}
	return requireOK("register fin voice call", res.StatusCode(), res.Body, res.JSON200)
}

// CollectFinVoiceCallByID returns a registered Fin Voice call by its Intercom ID.
func (s *CallsService) CollectFinVoiceCallByID(ctx context.Context, id int) (*FinVoiceCall, error) {
	res, err := s.client.generated.CollectFinVoiceCallByIdWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	return requireOK("collect fin voice call by ID", res.StatusCode(), res.Body, res.JSON200)
}

// CollectFinVoiceCallByExternalID returns a registered Fin Voice call by external call ID.
func (s *CallsService) CollectFinVoiceCallByExternalID(ctx context.Context, externalID string) (*FinVoiceCall, error) {
	if externalID == "" {
		return nil, fmt.Errorf("intercom: external ID is required")
	}
	res, err := s.client.generated.CollectFinVoiceCallByExternalIdWithResponse(ctx, externalID)
	if err != nil {
		return nil, err
	}
	return requireOK("collect fin voice call by external ID", res.StatusCode(), res.Body, res.JSON200)
}

// CollectFinVoiceCallByPhoneNumber returns a registered Fin Voice call by phone number (E.164 format).
func (s *CallsService) CollectFinVoiceCallByPhoneNumber(ctx context.Context, phone string) (*FinVoiceCall, error) {
	if phone == "" {
		return nil, fmt.Errorf("intercom: phone number is required")
	}
	res, err := s.client.generated.CollectFinVoiceCallByPhoneNumberWithResponse(ctx, phone)
	if err != nil {
		return nil, err
	}
	// The generated client has no JSON200 for this endpoint; parse the body directly.
	if res.StatusCode() != http.StatusOK {
		return nil, parseErrorResponse(res.StatusCode(), res.Body)
	}
	var result FinVoiceCall
	if err := json.Unmarshal(res.Body, &result); err != nil {
		return nil, fmt.Errorf("intercom: collect fin voice call by phone number: %w", err)
	}
	return &result, nil
}

// CollectFinVoiceCallsByConversationID returns all Fin Voice calls for a conversation.
func (s *CallsService) CollectFinVoiceCallsByConversationID(ctx context.Context, conversationID string) ([]FinVoiceCall, error) {
	if conversationID == "" {
		return nil, fmt.Errorf("intercom: conversation ID is required")
	}
	res, err := s.client.generated.CollectFinVoiceCallsByConversationIdWithResponse(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	list, err := requireOK("collect fin voice calls by conversation ID", res.StatusCode(), res.Body, res.JSON200)
	if err != nil {
		return nil, err
	}
	return *list, nil
}
