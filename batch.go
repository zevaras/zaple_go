package zaple

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// BatchService handles communication with the Zaple Batch API for bulk
// WhatsApp campaigns. Access it via Client.Batch.
type BatchService struct {
	client *Client
}

// ──────────────────────────────────────────────────────────────────────────────
// Common batch types
// ──────────────────────────────────────────────────────────────────────────────

// BatchStatus represents the lifecycle state of a batch.
type BatchStatus string

const (
	// BatchStatusDraft means the batch has been created but not yet sent.
	BatchStatusDraft BatchStatus = "draft"

	// BatchStatusPending means the batch is queued for processing.
	BatchStatusPending BatchStatus = "pending"

	// BatchStatusProcessing means messages are currently being dispatched.
	BatchStatusProcessing BatchStatus = "processing"

	// BatchStatusCompleted means all messages have been processed.
	BatchStatusCompleted BatchStatus = "completed"

	// BatchStatusFailed means the batch encountered a fatal error.
	BatchStatusFailed BatchStatus = "failed"

	// BatchStatusPartial means some messages succeeded and some failed.
	BatchStatusPartial BatchStatus = "partial"
)

// Batch represents a bulk messaging campaign (batch list).
type Batch struct {
	// ID is the unique batch identifier, used in all subsequent API calls.
	ID string `json:"id"`

	// Name is the human-readable label you provided when creating the batch.
	Name string `json:"name"`

	// TemplateID is the WhatsApp template applied to every contact in the batch.
	TemplateID string `json:"template_id"`

	// Status is the current lifecycle state of the batch.
	Status BatchStatus `json:"status"`

	// TotalContacts is the number of contacts currently in the batch list.
	TotalContacts int `json:"total_contacts"`

	// SentCount is the number of messages dispatched so far.
	SentCount int `json:"sent_count"`

	// DeliveredCount is the number of messages confirmed delivered.
	DeliveredCount int `json:"delivered_count"`

	// FailedCount is the number of messages that could not be delivered.
	FailedCount int `json:"failed_count"`

	// CreatedAt is the ISO 8601 timestamp when the batch was created.
	CreatedAt string `json:"created_at,omitempty"`

	// UpdatedAt is the ISO 8601 timestamp of the last status change.
	UpdatedAt string `json:"updated_at,omitempty"`

	// ScheduledAt is the ISO 8601 timestamp for a future send, if scheduled.
	ScheduledAt string `json:"scheduled_at,omitempty"`
}

// BatchContact represents a single recipient within a batch, along with their
// personalisation variables and optional metadata.
type BatchContact struct {
	// CountryCode is the dialling code without "+" (e.g. "91" for India).
	CountryCode string `json:"country_code"`

	// PhoneNumber is the recipient's phone number including the country prefix.
	PhoneNumber string `json:"phone_number"`

	// TemplateArguments are the ordered values for the {{1}}, {{2}}, … placeholders
	// in the template body, specific to this contact.
	TemplateArguments []string `json:"-"`

	// MediaURL is an optional per-contact media attachment URL, used when the
	// template has a dynamic media header.
	MediaURL string `json:"media_url,omitempty"`

	// MediaURLType specifies the format of MediaURL.
	MediaURLType MediaURLType `json:"media_url_type,omitempty"`

	// Metadata is arbitrary key-value data stored with the contact record.
	// It is not sent to WhatsApp but is available in reporting and webhooks.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// MarshalJSON serialises BatchContact, expanding TemplateArguments to
// sequential keys matching the Zaple API convention.
func (c BatchContact) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"country_code": c.CountryCode,
		"phone_number": c.PhoneNumber,
	}
	for i, arg := range c.TemplateArguments {
		m[fmt.Sprintf("template_argument%d", i+1)] = arg
	}
	if c.MediaURL != "" {
		m["media_url"] = c.MediaURL
	}
	if c.MediaURLType != "" {
		m["media_url_type"] = string(c.MediaURLType)
	}
	if len(c.Metadata) > 0 {
		m["metadata"] = c.Metadata
	}
	return json.Marshal(m)
}

// ──────────────────────────────────────────────────────────────────────────────
// Create
// ──────────────────────────────────────────────────────────────────────────────

// CreateBatchRequest is the payload for creating a new batch list.
type CreateBatchRequest struct {
	// Name is a human-readable label for the campaign (required).
	Name string `json:"name"`

	// TemplateID is the approved WhatsApp template to use for all messages (required).
	TemplateID string `json:"template_id"`

	// ScheduledAt is an optional ISO 8601 datetime to schedule a future send.
	// Leave empty to send immediately after calling SendBatch.
	ScheduledAt string `json:"scheduled_at,omitempty"`
}

// Create creates a new batch list.
//
// After creation, add contacts with UpsertContacts and then trigger delivery
// with Send.
//
//	batch, err := client.Batch.Create(ctx, &zaple.CreateBatchRequest{
//	    Name:       "June Promo",
//	    TemplateID: "475546217187442007",
//	})
func (s *BatchService) Create(ctx context.Context, req *CreateBatchRequest) (*Batch, error) {
	if req == nil {
		return nil, fmt.Errorf("zaple: CreateBatchRequest must not be nil")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("zaple: batch Name is required")
	}
	if req.TemplateID == "" {
		return nil, fmt.Errorf("zaple: batch TemplateID is required")
	}
	var resp Batch
	if err := s.client.do(ctx, http.MethodPost, "/api/v3/batch", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// UpsertContacts
// ──────────────────────────────────────────────────────────────────────────────

// UpsertContactsResponse is returned after upserting contacts into a batch.
type UpsertContactsResponse struct {
	// Inserted is the number of new contacts added.
	Inserted int `json:"inserted"`

	// Updated is the number of existing contacts whose data was refreshed.
	Updated int `json:"updated"`

	// Skipped is the number of records that were ignored (e.g. invalid numbers).
	Skipped int `json:"skipped"`

	// Errors lists any per-contact validation issues.
	Errors []string `json:"errors,omitempty"`
}

// upsertContactsPayload wraps the contacts slice for the API request body.
type upsertContactsPayload struct {
	Contacts []BatchContact `json:"contacts"`
}

// UpsertContacts adds or updates contacts in an existing batch list.
//
// Contacts are matched by phone number. Existing contacts are updated in-place;
// new ones are appended.  Call this multiple times to add contacts in chunks.
//
//	res, err := client.Batch.UpsertContacts(ctx, batchID, []zaple.BatchContact{
//	    {CountryCode: "91", PhoneNumber: "919999999999", TemplateArguments: []string{"Alice"}},
//	    {CountryCode: "91", PhoneNumber: "918888888888", TemplateArguments: []string{"Bob"}},
//	})
func (s *BatchService) UpsertContacts(ctx context.Context, batchID string, contacts []BatchContact) (*UpsertContactsResponse, error) {
	if batchID == "" {
		return nil, fmt.Errorf("zaple: batchID is required")
	}
	if len(contacts) == 0 {
		return nil, fmt.Errorf("zaple: at least one contact is required")
	}
	payload := upsertContactsPayload{Contacts: contacts}
	var resp UpsertContactsResponse
	path := "/api/v3/batch/" + batchID + "/contacts"
	if err := s.client.do(ctx, http.MethodPost, path, payload, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Send
// ──────────────────────────────────────────────────────────────────────────────

// SendBatchRequest is the optional payload for triggering batch dispatch.
type SendBatchRequest struct {
	// ScheduledAt overrides any schedule set at creation time.
	// Use an ISO 8601 datetime string (e.g. "2024-06-15T10:00:00Z") to schedule
	// future delivery, or leave empty to send immediately.
	ScheduledAt string `json:"scheduled_at,omitempty"`
}

// SendBatchResponse is returned after successfully triggering batch dispatch.
type SendBatchResponse struct {
	// Status describes the outcome, e.g. "Batch queued for sending."
	Status string `json:"status"`

	// BatchID is the identifier of the dispatched batch.
	BatchID string `json:"batch_id"`
}

// Send triggers message dispatch for a batch that has been created and populated
// with contacts.
//
// Pass a nil request to send immediately with no schedule override.
//
//	resp, err := client.Batch.Send(ctx, batchID, nil) // send now
//	resp, err := client.Batch.Send(ctx, batchID, &zaple.SendBatchRequest{
//	    ScheduledAt: "2024-06-15T10:00:00Z",
//	})
func (s *BatchService) Send(ctx context.Context, batchID string, req *SendBatchRequest) (*SendBatchResponse, error) {
	if batchID == "" {
		return nil, fmt.Errorf("zaple: batchID is required")
	}
	var resp SendBatchResponse
	path := "/api/v3/batch/" + batchID + "/send"
	if err := s.client.do(ctx, http.MethodPost, path, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// GetStatus
// ──────────────────────────────────────────────────────────────────────────────

// BatchStatusResponse holds real-time dispatch statistics for a batch.
type BatchStatusResponse struct {
	// BatchID is the identifier of the batch.
	BatchID string `json:"batch_id"`

	// Status is the current lifecycle state.
	Status BatchStatus `json:"status"`

	// TotalContacts is the total number of contacts in the batch.
	TotalContacts int `json:"total_contacts"`

	// SentCount is the number of messages dispatched so far.
	SentCount int `json:"sent_count"`

	// DeliveredCount is the number of messages confirmed delivered.
	DeliveredCount int `json:"delivered_count"`

	// FailedCount is the number of messages that failed.
	FailedCount int `json:"failed_count"`

	// Progress is the percentage of contacts processed (0–100).
	Progress float64 `json:"progress,omitempty"`
}

// GetStatus retrieves real-time delivery statistics for a batch.
//
// Poll this method after calling Send to track campaign progress.
//
//	status, err := client.Batch.GetStatus(ctx, batchID)
//	fmt.Printf("Delivered: %d/%d\n", status.DeliveredCount, status.TotalContacts)
func (s *BatchService) GetStatus(ctx context.Context, batchID string) (*BatchStatusResponse, error) {
	if batchID == "" {
		return nil, fmt.Errorf("zaple: batchID is required")
	}
	var resp BatchStatusResponse
	path := "/api/v3/batch/" + batchID + "/status"
	if err := s.client.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// GetDetails
// ──────────────────────────────────────────────────────────────────────────────

// GetDetails retrieves the full details of a batch, including its configuration
// and current delivery statistics.
//
//	batch, err := client.Batch.GetDetails(ctx, batchID)
func (s *BatchService) GetDetails(ctx context.Context, batchID string) (*Batch, error) {
	if batchID == "" {
		return nil, fmt.Errorf("zaple: batchID is required")
	}
	var resp Batch
	path := "/api/v3/batch/" + batchID
	if err := s.client.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Delete
// ──────────────────────────────────────────────────────────────────────────────

// Delete permanently removes a batch and all its contacts.
// This action is irreversible. Only draft batches can be deleted; attempting to
// delete a batch that is processing or completed will return an API error.
//
//	err := client.Batch.Delete(ctx, batchID)
func (s *BatchService) Delete(ctx context.Context, batchID string) error {
	if batchID == "" {
		return fmt.Errorf("zaple: batchID is required")
	}
	path := "/api/v3/batch/" + batchID
	return s.client.do(ctx, http.MethodDelete, path, nil, nil)
}
