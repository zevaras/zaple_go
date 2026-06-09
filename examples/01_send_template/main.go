// Example 01: Send a WhatsApp template message
//
// Demonstrates the most common use case: sending a single pre-approved
// template to one recipient.
//
// Run:
//
//	ZAPLE_API_KEY=<key> ZAPLE_API_SECRET=<secret> go run main.go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	zaple "github.com/zevaras/zaple_go"
)

func main() {
	apiKey := requireEnv("ZAPLE_API_KEY")
	apiSecret := requireEnv("ZAPLE_API_SECRET")

	// Create the client with a 15-second timeout.
	client := zaple.NewClient(apiKey, apiSecret,
		zaple.WithTimeout(15*time.Second),
	)

	ctx := context.Background()

	// ── Example 1: Simple text template (no variables) ──────────────────────
	resp, err := client.Messaging.SendTemplate(ctx, &zaple.SendTemplateRequest{
		TemplateID:  "YOUR_TEMPLATE_ID",
		CountryCode: "91",
		SendTo:      "919999999999",
	})
	if err != nil {
		handleError(err)
		return
	}
	fmt.Printf("✓ Message sent! ID: %s\n", resp.MessageID)

	// ── Example 2: Template with body variables ─────────────────────────────
	resp2, err := client.Messaging.SendTemplate(ctx, &zaple.SendTemplateRequest{
		TemplateID:  "YOUR_TEMPLATE_ID",
		CountryCode: "91",
		SendTo:      "919999999999",
		// {{1}} → "Alice", {{2}} → "Order #4567"
		TemplateArguments: []string{"Alice", "Order #4567"},
		// Attach metadata that will be echoed back in delivery webhooks.
		BizOpaqueCallbackData: map[string]any{
			"user_id":  42,
			"order_id": 4567,
		},
	})
	if err != nil {
		handleError(err)
		return
	}
	fmt.Printf("✓ Variable template sent! ID: %s\n", resp2.MessageID)

	// ── Example 3: Media template (image header) ────────────────────────────
	resp3, err := client.Messaging.SendTemplate(ctx, &zaple.SendTemplateRequest{
		TemplateID:   "YOUR_MEDIA_TEMPLATE_ID",
		CountryCode:  "91",
		SendTo:       "919999999999",
		MediaURL:     "https://example.com/promo-banner.jpg",
		MediaURLType: zaple.MediaURLTypeURL,
	})
	if err != nil {
		handleError(err)
		return
	}
	fmt.Printf("✓ Media template sent! ID: %s\n", resp3.MessageID)

	// ── Example 4: Template with quick-reply buttons ────────────────────────
	resp4, err := client.Messaging.SendTemplate(ctx, &zaple.SendTemplateRequest{
		TemplateID:         "YOUR_QUICK_REPLY_TEMPLATE_ID",
		CountryCode:        "91",
		SendTo:             "919999999999",
		QuickReplyPayload1: "confirm_appt_123",
		QuickReplyPayload2: "cancel_appt_123",
		BizOpaqueCallbackData: map[string]any{
			"appointment_id": 123,
		},
	})
	if err != nil {
		handleError(err)
		return
	}
	fmt.Printf("✓ Quick-reply template sent! ID: %s\n", resp4.MessageID)

	// ── Check message delivery status ───────────────────────────────────────
	status, err := client.Messaging.GetMessageStatus(ctx, resp.MessageID)
	if err != nil {
		log.Printf("Could not fetch status: %v\n", err)
		return
	}
	fmt.Printf("Status: %s (delivered at: %s)\n", status.Status, status.DeliveredAt)
}

func handleError(err error) {
	var apiErr *zaple.APIError
	if errors.As(err, &apiErr) {
		fmt.Fprintf(os.Stderr, "API error %d [%s]: %s\n",
			apiErr.StatusCode, apiErr.Code, apiErr.Message)

		// Check for specific conditions using sentinel errors.
		switch {
		case errors.Is(err, zaple.ErrUnauthorized):
			fmt.Fprintln(os.Stderr, "→ Check your API key and secret.")
		case errors.Is(err, zaple.ErrRateLimited):
			fmt.Fprintln(os.Stderr, "→ You have been rate limited; slow down requests.")
		case errors.Is(err, zaple.ErrDailyLimitReached):
			fmt.Fprintln(os.Stderr, "→ Daily message limit reached; consider upgrading.")
		case errors.Is(err, zaple.ErrInactiveTemplate):
			fmt.Fprintln(os.Stderr, "→ The template is not approved or has been paused.")
		}

		if len(apiErr.ValidationErrors) > 0 {
			fmt.Fprintln(os.Stderr, "Validation errors:")
			for field, msgs := range apiErr.ValidationErrors {
				for _, m := range msgs {
					fmt.Fprintf(os.Stderr, "  %s: %s\n", field, m)
				}
			}
		}
		return
	}
	fmt.Fprintf(os.Stderr, "Unexpected error: %v\n", err)
}

func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Environment variable %s is required", key)
	}
	return val
}
