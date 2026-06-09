// Example 04: Create WhatsApp templates via the API
//
// Demonstrates all four template types:
//  1. Text template with quick-reply buttons
//  2. Image template with URL button
//  3. Video template
//  4. Authentication (OTP) template
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
	client := zaple.NewClient(
		requireEnv("ZAPLE_API_KEY"),
		requireEnv("ZAPLE_API_SECRET"),
		zaple.WithTimeout(30*time.Second),
	)

	ctx := context.Background()

	// ── 1. Text template with quick-reply buttons ────────────────────────────
	resp1, err := client.Messaging.CreateTemplate(ctx, &zaple.CreateTemplateRequest{
		Title:       "order_update",
		Category:    zaple.TemplateCategoryUtility,
		Language:    "en_US",
		ContentType: zaple.TemplateContentTypeText,
		HeaderText:  "Order update",
		Content:     "Hi {{1}}, your order {{2}} is ready for pickup.",
		FooterText:  "Reply STOP to opt out",
		VariableType:    zaple.TemplateVariableTypeNumeric,
		VariableSamples: []string{"Priya", "ORD-1007"},
		Buttons: []zaple.CreateTemplateButton{
			{
				Type: "quick_reply",
				Replies: []zaple.QuickReplyItem{
					{Text: "Track order"},
					{Text: "Contact support"},
				},
			},
		},
	})
	if err != nil {
		handleError("text template", err)
	} else {
		fmt.Printf("✓ Text template created — ID: %d (%s)\n", resp1.TemplateID, resp1.Message)
	}

	// ── 2. Image template with URL button ────────────────────────────────────
	resp2, err := client.Messaging.CreateTemplate(ctx, &zaple.CreateTemplateRequest{
		Title:         "new_collection_launch",
		Category:      zaple.TemplateCategoryMarketing,
		Language:      "en_US",
		ContentType:   zaple.TemplateContentTypeImage,
		TemplateImage: "https://example.com/new-collection.jpg",
		Content:       "Hi {{1}}, our new collection is live. Tap below to explore.",
		VariableSamples: []string{"Aarav"},
		Buttons: []zaple.CreateTemplateButton{
			{
				Type: "url",
				Websites: []zaple.URLButtonItem{
					{Text: "Shop now", URL: "https://example.com/collections"},
				},
			},
		},
	})
	if err != nil {
		handleError("image template", err)
	} else {
		fmt.Printf("✓ Image template created — ID: %d (%s)\n", resp2.TemplateID, resp2.Message)
	}

	// ── 3. Video template ────────────────────────────────────────────────────
	resp3, err := client.Messaging.CreateTemplate(ctx, &zaple.CreateTemplateRequest{
		Title:         "product_demo",
		Category:      zaple.TemplateCategoryMarketing,
		Language:      "en_US",
		ContentType:   zaple.TemplateContentTypeVideo,
		TemplateVideo: "https://example.com/demo.mp4",
		Content:       "Hi {{1}}, watch this quick demo of your new product.",
		VariableSamples: []string{"Riya"},
	})
	if err != nil {
		handleError("video template", err)
	} else {
		fmt.Printf("✓ Video template created — ID: %d (%s)\n", resp3.TemplateID, resp3.Message)
	}

	// ── 4. Authentication (OTP) template ─────────────────────────────────────
	resp4, err := client.Messaging.CreateTemplate(ctx, &zaple.CreateTemplateRequest{
		Title:                     "login_otp",
		Category:                  zaple.TemplateCategoryAuthentication,
		Language:                  "en_US",
		// Content is not required for authentication templates.
		AddSecurityRecommendation: true,
		CopyOtpButtonText:         "Copy code",
		EnableCodeExpiration:      true,
		CodeExpirationMinutes:     "10",
	})
	if err != nil {
		handleError("auth template", err)
	} else {
		fmt.Printf("✓ Auth template created — ID: %d (%s)\n", resp4.TemplateID, resp4.Message)
	}
}

func handleError(label string, err error) {
	var apiErr *zaple.APIError
	if errors.As(err, &apiErr) {
		fmt.Fprintf(os.Stderr, "✗ %s — API %d [%s]: %s\n",
			label, apiErr.StatusCode, apiErr.Code, apiErr.Message)
		for field, msgs := range apiErr.ValidationErrors {
			for _, m := range msgs {
				fmt.Fprintf(os.Stderr, "   %s: %s\n", field, m)
			}
		}
		return
	}
	fmt.Fprintf(os.Stderr, "✗ %s — %v\n", label, err)
}

func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Environment variable %s is required", key)
	}
	return val
}
