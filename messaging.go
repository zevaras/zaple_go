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
// CreateTemplate
// ──────────────────────────────────────────────────────────────────────────────

// TemplateCategory defines the WhatsApp template category.
type TemplateCategory string

const (
	// TemplateCategoryUtility is for transactional messages (order updates, alerts).
	TemplateCategoryUtility TemplateCategory = "utility"

	// TemplateCategoryMarketing is for promotional messages.
	TemplateCategoryMarketing TemplateCategory = "marketing"

	// TemplateCategoryUtilityMarketing submits as utility but allows Meta category change.
	TemplateCategoryUtilityMarketing TemplateCategory = "utility_marketing"

	// TemplateCategoryAuthentication is for OTP / verification messages.
	TemplateCategoryAuthentication TemplateCategory = "authentication"
)

// TemplateContentType defines the header media type of a template.
type TemplateContentType string

const (
	// TemplateContentTypeNone means no header (text body only).
	TemplateContentTypeNone TemplateContentType = "none"

	// TemplateContentTypeText means a plain-text header.
	TemplateContentTypeText TemplateContentType = "text"

	// TemplateContentTypeImage means an image header.
	TemplateContentTypeImage TemplateContentType = "image"

	// TemplateContentTypeVideo means a video header.
	TemplateContentTypeVideo TemplateContentType = "video"

	// TemplateContentTypeDocument means a document (PDF) header.
	TemplateContentTypeDocument TemplateContentType = "document"

	// TemplateContentTypeLocation means a location header.
	TemplateContentTypeLocation TemplateContentType = "location"
)

// TemplateVariableType controls how body variables are referenced.
type TemplateVariableType string

const (
	// TemplateVariableTypeNumeric uses positional placeholders: {{1}}, {{2}}, …
	TemplateVariableTypeNumeric TemplateVariableType = "numeric"

	// TemplateVariableTypeNamed uses named Meta parameters.
	TemplateVariableTypeNamed TemplateVariableType = "named"
)

// TemplateLocation is the location payload for contentType=location templates.
type TemplateLocation struct {
	// Name is the place name displayed to the recipient.
	Name string `json:"name"`

	// Address is the street address or description.
	Address string `json:"address"`

	// Latitude is the geographic latitude.
	Latitude float64 `json:"latitude"`

	// Longitude is the geographic longitude.
	Longitude float64 `json:"longitude"`
}

// CreateTemplateButton represents a single button in a template.
// Set Type and populate the corresponding field:
//   - quick_reply → Replies
//   - url         → Websites
type CreateTemplateButton struct {
	// Type is one of: quick_reply, url, phone_number, copy_offer_code, catalog.
	Type string `json:"type"`

	// Replies holds the text options for a quick_reply button.
	Replies []QuickReplyItem `json:"replies,omitempty"`

	// Websites holds the URL targets for a url button.
	Websites []URLButtonItem `json:"websites,omitempty"`
}

// QuickReplyItem is a single reply option within a quick_reply button.
type QuickReplyItem struct {
	// Text is the label shown to the recipient.
	Text string `json:"text"`
}

// URLButtonItem is a single link within a url button.
type URLButtonItem struct {
	// Text is the button label shown to the recipient.
	Text string `json:"text"`

	// URL is the destination link.
	URL string `json:"url"`
}

// CreateTemplateRequest is the payload for submitting a new WhatsApp template
// for Meta review via POST /api/v3/create-template.
type CreateTemplateRequest struct {
	// Title is the internal label shown in Zaple. Maximum 255 characters. (required)
	Title string `json:"title"`

	// Category is the WhatsApp template category. (required)
	Category TemplateCategory `json:"category"`

	// Language is the WhatsApp language code, e.g. "en_US" or "en". (required)
	// Authentication templates must use "en_US".
	Language string `json:"language"`

	// Content is the template body text. Use {{1}}, {{2}} for numeric variables
	// or named variables when VariableType is TemplateVariableTypeNamed.
	// Not required for authentication templates.
	Content string `json:"content,omitempty"`

	// ContentType sets the header media format. Defaults to TemplateContentTypeNone.
	ContentType TemplateContentType `json:"contentType,omitempty"`

	// HeaderText is the text shown in the header when ContentType is
	// TemplateContentTypeText. Maximum 60 characters.
	HeaderText string `json:"headerText,omitempty"`

	// TemplateImage is a public image URL or base64 data URI.
	// Required when ContentType is TemplateContentTypeImage.
	TemplateImage string `json:"templateImage,omitempty"`

	// TemplateVideo is a public video URL.
	// Required when ContentType is TemplateContentTypeVideo.
	// For file uploads use multipart/form-data and set WithHTTPClient accordingly.
	TemplateVideo string `json:"templateVideo,omitempty"`

	// TemplateDocument is a public PDF URL.
	// Required when ContentType is TemplateContentTypeDocument.
	TemplateDocument string `json:"templateDocument,omitempty"`

	// TemplateLocation is required when ContentType is TemplateContentTypeLocation.
	TemplateLocation *TemplateLocation `json:"templateLocation,omitempty"`

	// FooterText is the text shown below the message body. Maximum 60 characters.
	FooterText string `json:"footerText,omitempty"`

	// VariableType controls how body variables are referenced.
	// Defaults to TemplateVariableTypeNumeric.
	VariableType TemplateVariableType `json:"variable_type,omitempty"`

	// VariableSamples provides example values for template variables.
	// Each item can be up to 1000 characters.
	VariableSamples []string `json:"variable_samples,omitempty"`

	// Buttons defines the interactive buttons attached to the template.
	// Supported types: quick_reply, url, phone_number, copy_offer_code, catalog.
	// Button labels must be unique.
	Buttons []CreateTemplateButton `json:"buttons,omitempty"`

	// ── Authentication-only fields ────────────────────────────────────────────

	// AddSecurityRecommendation adds the standard security disclaimer to the
	// message body. Only used when Category is TemplateCategoryAuthentication.
	AddSecurityRecommendation bool `json:"addSecurityRecommendation,omitempty"`

	// CopyOtpButtonText is the label for the Meta copy-code button.
	// Only used when Category is TemplateCategoryAuthentication.
	CopyOtpButtonText string `json:"copyOtpButtonText,omitempty"`

	// EnableCodeExpiration shows an expiry notice in the message.
	// Only used when Category is TemplateCategoryAuthentication.
	EnableCodeExpiration bool `json:"enableCodeExpiration,omitempty"`

	// CodeExpirationMinutes sets the OTP validity window in minutes.
	// Only used when Category is TemplateCategoryAuthentication.
	CodeExpirationMinutes string `json:"codeExpirationMinutes,omitempty"`
}

// CreateTemplateResponse is returned by a successful CreateTemplate call.
type CreateTemplateResponse struct {
	// Success indicates whether the submission was accepted.
	Success bool `json:"success"`

	// StatusCode is the HTTP status echoed in the response body.
	StatusCode int `json:"status_code"`

	// Message is a human-readable description, e.g. "Template submitted for review."
	Message string `json:"message"`

	// TemplateID is the Zaple-assigned identifier for the new template.
	// Use this value in SendTemplate once the template is approved.
	TemplateID int64 `json:"template_id"`
}

// CreateTemplate submits a new WhatsApp template for Meta review.
//
// Zaple creates the template record, uploads any optional header media, submits
// it to Meta, and returns the local template_id. The template becomes usable
// with SendTemplate once its status reaches "APPROVED".
//
//	resp, err := client.Messaging.CreateTemplate(ctx, &zaple.CreateTemplateRequest{
//	    Title:    "order_update",
//	    Category: zaple.TemplateCategoryUtility,
//	    Language: "en_US",
//	    Content:  "Hi {{1}}, your order {{2}} is ready for pickup.",
//	    ContentType: zaple.TemplateContentTypeText,
//	    HeaderText:  "Order update",
//	    FooterText:  "Reply STOP to opt out",
//	    VariableType:    zaple.TemplateVariableTypeNumeric,
//	    VariableSamples: []string{"Priya", "ORD-1007"},
//	    Buttons: []zaple.CreateTemplateButton{
//	        {
//	            Type:    "quick_reply",
//	            Replies: []zaple.QuickReplyItem{{Text: "Track order"}, {Text: "Contact support"}},
//	        },
//	    },
//	})
func (s *MessagingService) CreateTemplate(ctx context.Context, req *CreateTemplateRequest) (*CreateTemplateResponse, error) {
	if err := validateCreateTemplateRequest(req); err != nil {
		return nil, err
	}
	var resp CreateTemplateResponse
	if err := s.client.do(ctx, http.MethodPost, "/api/v3/create-template", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func validateCreateTemplateRequest(req *CreateTemplateRequest) error {
	if req == nil {
		return fmt.Errorf("zaple: CreateTemplateRequest must not be nil")
	}
	if req.Title == "" {
		return fmt.Errorf("zaple: Title is required")
	}
	if req.Category == "" {
		return fmt.Errorf("zaple: Category is required")
	}
	if req.Language == "" {
		return fmt.Errorf("zaple: Language is required")
	}
	if req.Category != TemplateCategoryAuthentication && req.Content == "" {
		return fmt.Errorf("zaple: Content is required for non-authentication templates")
	}
	return nil
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
