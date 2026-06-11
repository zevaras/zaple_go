package zaple

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

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
