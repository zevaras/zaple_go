package zaple

import (
	"context"
	"fmt"
	"net/http"
)

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
