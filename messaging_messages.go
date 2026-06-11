package zaple

import (
	"context"
	"fmt"
	"net/http"
)

// MessageStatus represents the delivery state of a sent message.
type MessageStatus struct {
	// MessageID is the Zaple/WhatsApp message identifier.
	MessageID string `json:"message_id"`

	// Status is the current delivery status (e.g. "sent", "delivered", "read", "failed").
	Status string `json:"status"`

	// SentAt is the timestamp when the message was dispatched.
	SentAt string `json:"sent_at,omitempty"`

	// DeliveredAt is the timestamp when the message was delivered to the device.
	DeliveredAt string `json:"delivered_at,omitempty"`

	// ReadAt is the timestamp when the recipient opened the message.
	ReadAt string `json:"read_at,omitempty"`

	// FailedReason explains why delivery failed, if applicable.
	FailedReason string `json:"failed_reason,omitempty"`
}

// GetMessageStatus retrieves the current delivery status for a message.
//
//	status, err := client.Messaging.GetMessageStatus(ctx, "x927831064523907185K9CDMYNReFgHIPWz")
func (s *MessagingService) GetMessageStatus(ctx context.Context, messageID string) (*MessageStatus, error) {
	if messageID == "" {
		return nil, fmt.Errorf("zaple: messageID is required")
	}
	var resp MessageStatus
	path := "/api/v3/message-status?message_id=" + messageID
	if err := s.client.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// MessageCountParams filters the message count query.
type MessageCountParams struct {
	// From is the start date in YYYY-MM-DD format (inclusive).
	From string

	// To is the end date in YYYY-MM-DD format (inclusive).
	To string

	// TemplateID restricts the count to a specific template (optional).
	TemplateID string
}

// MessageCount holds aggregated message delivery statistics.
type MessageCount struct {
	// Total is the number of messages sent in the requested period.
	Total int `json:"total"`

	// Delivered is the number of messages confirmed delivered.
	Delivered int `json:"delivered"`

	// Read is the number of messages opened by the recipient.
	Read int `json:"read"`

	// Failed is the number of messages that could not be delivered.
	Failed int `json:"failed"`
}

// GetMessageCount returns delivery statistics for messages sent within the
// specified date range.
//
//	count, err := client.Messaging.GetMessageCount(ctx, &zaple.MessageCountParams{
//	    From: "2024-01-01",
//	    To:   "2024-01-31",
//	})
func (s *MessagingService) GetMessageCount(ctx context.Context, params *MessageCountParams) (*MessageCount, error) {
	path := "/api/v3/message-count"
	sep := "?"
	if params != nil {
		if params.From != "" {
			path += sep + "from=" + params.From
			sep = "&"
		}
		if params.To != "" {
			path += sep + "to=" + params.To
			sep = "&"
		}
		if params.TemplateID != "" {
			path += sep + "template_id=" + params.TemplateID
		}
	}
	var resp MessageCount
	if err := s.client.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
