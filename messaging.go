package zaple

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// MessagingService handles communication with the Zaple Messaging API (V3).
// Access it via Client.Messaging.
type MessagingService struct {
	client *Client
}

// ──────────────────────────────────────────────────────────────────────────────
// SendTemplate
// ──────────────────────────────────────────────────────────────────────────────

// MediaURLType defines how the media_url field is interpreted by the API.
type MediaURLType string

const (
	// MediaURLTypeURL indicates that media_url is a direct HTTP/HTTPS link.
	MediaURLTypeURL MediaURLType = "url"

	// MediaURLTypeBase64 indicates that media_url contains a base64-encoded payload.
	MediaURLTypeBase64 MediaURLType = "base64"
)

// SendTemplateRequest is the payload for sending a single WhatsApp template message.
//
// Only TemplateID, CountryCode, and SendTo are always required. All other
// fields depend on the template's configuration in the Zaple library.
type SendTemplateRequest struct {
	// TemplateID is the unique identifier of the approved WhatsApp template.
	// Find it in your Zaple template library.
	TemplateID string `json:"-"`

	// CountryCode is the recipient's country dialling code without the leading "+".
	// Example: "91" for India, "1" for USA/Canada.
	CountryCode string `json:"-"`

	// SendTo is the recipient's full phone number (digits only, no spaces or dashes).
	// Include the country code prefix, e.g. "919999999999".
	SendTo string `json:"-"`

	// TemplateArguments holds the ordered variable values for {{1}}, {{2}}, …
	// placeholders in the template body. Leave nil for templates with no variables.
	TemplateArguments []string `json:"-"`

	// MediaURL is the URL (or base64 string) of the media attachment for templates
	// that include an image, video, or document header.
	MediaURL string `json:"-"`

	// MediaURLType specifies the format of MediaURL.
	// Use MediaURLTypeBase64 when providing raw base64 data.
	// Defaults to a plain URL when omitted.
	MediaURLType MediaURLType `json:"-"`

	// QuickReplyPayload1 is the developer-defined payload for the first quick-reply
	// button. Returned verbatim in webhook callbacks when the user taps the button.
	QuickReplyPayload1 string `json:"-"`

	// QuickReplyPayload2 is the developer-defined payload for the second quick-reply
	// button.
	QuickReplyPayload2 string `json:"-"`

	// BizOpaqueCallbackData is arbitrary JSON metadata attached to the message.
	// It is returned in delivery webhooks, making it easy to correlate messages
	// with internal records (order IDs, appointment IDs, etc.).
	BizOpaqueCallbackData map[string]any `json:"-"`
}

// MarshalJSON serialises the request into the shape expected by the Zaple V3 API.
// Template arguments are expanded to sequential keys (template_argument1, …).
func (r SendTemplateRequest) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"template_id":  r.TemplateID,
		"country_code": r.CountryCode,
		"send_to":      r.SendTo,
	}
	for i, arg := range r.TemplateArguments {
		m[fmt.Sprintf("template_argument%d", i+1)] = arg
	}
	if r.MediaURL != "" {
		m["media_url"] = r.MediaURL
	}
	if r.MediaURLType != "" {
		m["media_url_type"] = string(r.MediaURLType)
	}
	if r.QuickReplyPayload1 != "" {
		m["quick_reply_payload1"] = r.QuickReplyPayload1
	}
	if r.QuickReplyPayload2 != "" {
		m["quick_reply_payload2"] = r.QuickReplyPayload2
	}
	if len(r.BizOpaqueCallbackData) > 0 {
		m["biz_opaque_callback_data"] = r.BizOpaqueCallbackData
	}
	return json.Marshal(m)
}

// SendTemplateResponse is returned by a successful SendTemplate call.
type SendTemplateResponse struct {
	// Status is a human-readable description, e.g. "Message sent successfully."
	Status string `json:"status"`

	// MessageID is the unique identifier assigned by Zaple/WhatsApp.
	// Use it to query delivery status via GetMessageStatus.
	MessageID string `json:"message_id"`
}

// SendTemplate sends a pre-approved WhatsApp template message to a single recipient.
//
// The method handles serialisation of template arguments and optional fields,
// and will automatically retry on transient errors unless retries are disabled.
//
//	resp, err := client.Messaging.SendTemplate(ctx, &zaple.SendTemplateRequest{
//	    TemplateID:        "475546217187442007",
//	    CountryCode:       "91",
//	    SendTo:            "919999999999",
//	    TemplateArguments: []string{"Alice", "Order #1234"},
//	    BizOpaqueCallbackData: map[string]any{"order_id": 1234},
//	})
func (s *MessagingService) SendTemplate(ctx context.Context, req *SendTemplateRequest) (*SendTemplateResponse, error) {
	if err := validateSendTemplateRequest(req); err != nil {
		return nil, err
	}
	var resp SendTemplateResponse
	if err := s.client.do(ctx, http.MethodPost, "/api/v3/send-template-message", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func validateSendTemplateRequest(req *SendTemplateRequest) error {
	if req == nil {
		return fmt.Errorf("zaple: SendTemplateRequest must not be nil")
	}
	if req.TemplateID == "" {
		return fmt.Errorf("zaple: TemplateID is required")
	}
	if req.CountryCode == "" {
		return fmt.Errorf("zaple: CountryCode is required")
	}
	if req.SendTo == "" {
		return fmt.Errorf("zaple: SendTo is required")
	}
	return nil
}

// ──────────────────────────────────────────────────────────────────────────────
// GetMessageStatus
// ──────────────────────────────────────────────────────────────────────────────

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

// ──────────────────────────────────────────────────────────────────────────────
// GetTemplateDetails
// ──────────────────────────────────────────────────────────────────────────────

// TemplateComponent describes a section of a WhatsApp message template
// (header, body, footer, or buttons).
type TemplateComponent struct {
	// Type is the component type: "HEADER", "BODY", "FOOTER", or "BUTTONS".
	Type string `json:"type"`

	// Text is the static text for BODY and FOOTER components.
	Text string `json:"text,omitempty"`

	// Format indicates the header media format: "TEXT", "IMAGE", "VIDEO", "DOCUMENT".
	Format string `json:"format,omitempty"`

	// Buttons lists interactive button definitions for the BUTTONS component.
	Buttons []TemplateButton `json:"buttons,omitempty"`
}

// TemplateButton represents a single button within a template.
type TemplateButton struct {
	// Type is "QUICK_REPLY", "URL", or "PHONE_NUMBER".
	Type string `json:"type"`

	// Text is the button label visible to the user.
	Text string `json:"text"`

	// URL is the destination for URL-type buttons.
	URL string `json:"url,omitempty"`

	// PhoneNumber is the dial target for PHONE_NUMBER-type buttons.
	PhoneNumber string `json:"phone_number,omitempty"`
}

// TemplateDetails holds the full definition of a WhatsApp message template.
type TemplateDetails struct {
	// ID is the unique template identifier.
	ID string `json:"id"`

	// Name is the internal template name in Zaple.
	Name string `json:"name"`

	// Status is the approval status: "APPROVED", "PENDING", "REJECTED".
	Status string `json:"status"`

	// Category is the WhatsApp template category (e.g. "MARKETING", "UTILITY", "AUTHENTICATION").
	Category string `json:"category"`

	// Language is the BCP-47 locale code (e.g. "en_US", "hi").
	Language string `json:"language"`

	// Components contains the ordered sections of the template.
	Components []TemplateComponent `json:"components,omitempty"`
}

// GetTemplateDetails retrieves the full definition of a template by its ID.
//
//	tpl, err := client.Messaging.GetTemplateDetails(ctx, "475546217187442007")
func (s *MessagingService) GetTemplateDetails(ctx context.Context, templateID string) (*TemplateDetails, error) {
	if templateID == "" {
		return nil, fmt.Errorf("zaple: templateID is required")
	}
	var resp TemplateDetails
	path := "/api/v3/template-details?template_id=" + templateID
	if err := s.client.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// GetTemplateStatus
// ──────────────────────────────────────────────────────────────────────────────

// TemplateStatus holds the current approval status of a template.
type TemplateStatus struct {
	// TemplateID is the template identifier.
	TemplateID string `json:"template_id"`

	// Status is the WhatsApp review status: "APPROVED", "PENDING", "REJECTED", "PAUSED".
	Status string `json:"status"`

	// RejectedReason is populated when Status is "REJECTED".
	RejectedReason string `json:"rejected_reason,omitempty"`
}

// GetTemplateStatus checks the WhatsApp approval status of a template.
//
//	ts, err := client.Messaging.GetTemplateStatus(ctx, "475546217187442007")
func (s *MessagingService) GetTemplateStatus(ctx context.Context, templateID string) (*TemplateStatus, error) {
	if templateID == "" {
		return nil, fmt.Errorf("zaple: templateID is required")
	}
	var resp TemplateStatus
	path := "/api/v3/template-status?template_id=" + templateID
	if err := s.client.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// GetMessageCount
// ──────────────────────────────────────────────────────────────────────────────

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
