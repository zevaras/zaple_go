package zaple

import (
	"context"
	"fmt"
	"net/http"
)

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
//	    Buttons: []zaple.CreateTemplateButton{
//	        {Type: "quick_reply", Replies: []zaple.QuickReplyItem{{Text: "Track"}}},
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
