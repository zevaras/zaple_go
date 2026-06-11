package zaple

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ListTemplatesParams holds optional filters for listing templates.
type ListTemplatesParams struct {
	// Search filters by template name (partial match).
	Search string

	// Page is the 1-based page number. Defaults to 1.
	Page int

	// PerPage is the number of results per page. Defaults to the API default (typically 10).
	PerPage int

	// Category filters by template category: UTILITY, MARKETING, AUTHENTICATION, CAROUSEL.
	Category string

	// Status filters by approval status: APPROVED, PENDING, REJECTED.
	Status string

	// Active filters by the template's active/inactive state.
	// Nil means no filter; pointer to true/false applies the filter.
	Active *bool

	// Favorite filters to show only favourite templates.
	// Nil means no filter; pointer to true/false applies the filter.
	Favorite *bool
}

// TemplateListItem is a single entry in the templates list.
type TemplateListItem struct {
	// Name is the internal template name.
	Name string `json:"name"`

	// TemplateID is the unique template identifier.
	TemplateID string `json:"template_id"`

	// Category is the WhatsApp template category (e.g. "UTILITY", "MARKETING").
	Category string `json:"category"`

	// Status is the approval status: "APPROVED", "PENDING", "REJECTED".
	Status string `json:"status"`

	// HeaderType indicates the header media format (e.g. "TEXT", "IMAGE", "NONE").
	HeaderType string `json:"header_type"`

	// IsFavorite indicates whether the template is marked as a favourite.
	IsFavorite bool `json:"is_favorite"`

	// IsDefault indicates whether this is the account's default template.
	IsDefault bool `json:"is_default"`

	// VariableCount is the number of dynamic variables in the template body.
	VariableCount int `json:"variable_count"`

	// CreatedAt is the ISO-8601 timestamp when the template was created.
	CreatedAt string `json:"created_at"`
}

// TemplateListStat is a labelled counter returned alongside the template list.
type TemplateListStat struct {
	// Label is the human-readable stat name (e.g. "Templates Created").
	Label string `json:"label"`

	// Value is the count as a string (e.g. "15").
	Value string `json:"value"`
}

// TemplateListMeta holds pagination information for a list response.
type TemplateListMeta struct {
	// CurrentPage is the page number returned.
	CurrentPage int `json:"current_page"`

	// LastPage is the index of the final page.
	LastPage int `json:"last_page"`

	// PerPage is the number of items per page.
	PerPage int `json:"per_page"`

	// Total is the total number of templates matching the filter.
	Total int `json:"total"`
}

// ListTemplatesResponse is returned by a successful ListTemplates call.
type ListTemplatesResponse struct {
	// Templates is the slice of template summaries for the current page.
	Templates []TemplateListItem `json:"templates"`

	// Stats contains aggregate counters (e.g. total templates created).
	Stats []TemplateListStat `json:"stats"`

	// Meta holds pagination details.
	Meta TemplateListMeta `json:"meta"`
}

// listTemplatesEnvelope unwraps the "data" key in the API response.
type listTemplatesEnvelope struct {
	Data *ListTemplatesResponse `json:"data"`
}

// ListTemplates retrieves a paginated list of templates with optional filters.
//
//	resp, err := client.Messaging.ListTemplates(ctx, &zaple.ListTemplatesParams{
//	    Status:  "APPROVED",
//	    PerPage: 20,
//	})
//	for _, t := range resp.Templates {
//	    fmt.Printf("%s — %s\n", t.TemplateID, t.Name)
//	}
func (s *MessagingService) ListTemplates(ctx context.Context, params *ListTemplatesParams) (*ListTemplatesResponse, error) {
	q := url.Values{}
	if params != nil {
		if params.Search != "" {
			q.Set("search", params.Search)
		}
		if params.Page > 0 {
			q.Set("page", fmt.Sprintf("%d", params.Page))
		}
		if params.PerPage > 0 {
			q.Set("per_page", fmt.Sprintf("%d", params.PerPage))
		}
		if params.Category != "" {
			q.Set("category", params.Category)
		}
		if params.Status != "" {
			q.Set("status", params.Status)
		}
		if params.Active != nil {
			if *params.Active {
				q.Set("active", "true")
			} else {
				q.Set("active", "false")
			}
		}
		if params.Favorite != nil {
			if *params.Favorite {
				q.Set("favorite", "true")
			} else {
				q.Set("favorite", "false")
			}
		}
	}

	path := "/api/v3/templates"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	var envelope listTemplatesEnvelope
	if err := s.client.do(ctx, http.MethodGet, path, nil, &envelope); err != nil {
		return nil, err
	}
	if envelope.Data == nil {
		return &ListTemplatesResponse{}, nil
	}
	return envelope.Data, nil
}
