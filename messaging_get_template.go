package zaple

import (
	"context"
	"fmt"
	"net/http"
)

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
