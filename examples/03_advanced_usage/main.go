// Example 03: Advanced client configuration
//
// Shows how to:
//   - Use a custom HTTP client (e.g. with a proxy)
//   - Plug in a custom logger
//   - Disable retries for latency-sensitive paths
//   - Accept the MessagingAPI interface for easy mocking in tests
//   - Check template status before sending
//   - Retrieve message delivery counts for a date range
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
	"net/http"
	"os"
	"time"

	zaple "github.com/zevaras/zaple_go"
)

// ── Custom logger ─────────────────────────────────────────────────────────────

type appLogger struct{ logger *log.Logger }

func (l *appLogger) Printf(format string, v ...any) {
	l.logger.Printf(format, v...)
}

// ── Application using the interface (easily mockable) ─────────────────────────

type NotificationSender struct {
	messaging zaple.MessagingAPI
}

func (ns *NotificationSender) SendOrderUpdate(ctx context.Context, phone, name, orderID string) error {
	resp, err := ns.messaging.SendTemplate(ctx, &zaple.SendTemplateRequest{
		TemplateID:        "YOUR_ORDER_UPDATE_TEMPLATE_ID",
		CountryCode:       "91",
		SendTo:            phone,
		TemplateArguments: []string{name, orderID},
		BizOpaqueCallbackData: map[string]any{
			"order_id": orderID,
		},
	})
	if err != nil {
		return fmt.Errorf("send order update to %s: %w", phone, err)
	}
	log.Printf("Order update sent to %s — message ID: %s", phone, resp.MessageID)
	return nil
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	apiKey := requireEnv("ZAPLE_API_KEY")
	apiSecret := requireEnv("ZAPLE_API_SECRET")

	// ── Option 1: Custom HTTP transport (e.g. proxy, mTLS) ─────────────────
	customTransport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    90 * time.Second,
		DisableCompression: false,
	}
	customHTTPClient := &http.Client{
		Transport: customTransport,
		Timeout:   20 * time.Second,
	}

	// ── Option 2: Custom logger ─────────────────────────────────────────────
	logger := &appLogger{
		logger: log.New(os.Stdout, "[zaple] ", log.LstdFlags|log.Lmicroseconds),
	}

	// ── Create client with all options ──────────────────────────────────────
	client := zaple.NewClient(apiKey, apiSecret,
		zaple.WithHTTPClient(customHTTPClient),
		zaple.WithLogger(logger),
		zaple.WithUserAgent("MyOrderService/2.1"),
		zaple.WithRetryConfig(zaple.RetryConfig{
			MaxRetries: 5,
			WaitMin:    500 * time.Millisecond,
			WaitMax:    10 * time.Second,
		}),
	)

	ctx := context.Background()

	// ── Check template status before sending ────────────────────────────────
	const templateID = "YOUR_TEMPLATE_ID"

	ts, err := client.Messaging.GetTemplateStatus(ctx, templateID)
	if err != nil {
		log.Fatalf("GetTemplateStatus: %v", err)
	}
	if ts.Status != "APPROVED" {
		log.Fatalf("Template %s is not approved (status: %s)", templateID, ts.Status)
	}
	fmt.Printf("✓ Template %s is approved\n", templateID)

	// ── Use the interface-based wrapper ─────────────────────────────────────
	sender := &NotificationSender{messaging: client.Messaging}

	err = sender.SendOrderUpdate(ctx, "919999999999", "Alice", "ORD-9876")
	if err != nil {
		// Detailed inspection using errors.As / errors.Is
		var apiErr *zaple.APIError
		if errors.As(err, &apiErr) {
			fmt.Fprintf(os.Stderr, "API %d [%s]: %s\n",
				apiErr.StatusCode, apiErr.Code, apiErr.Message)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}

	// ── Message count report ────────────────────────────────────────────────
	count, err := client.Messaging.GetMessageCount(ctx, &zaple.MessageCountParams{
		From: "2024-06-01",
		To:   "2024-06-30",
	})
	if err != nil {
		log.Printf("GetMessageCount: %v\n", err)
		return
	}
	fmt.Printf("June stats — total: %d, delivered: %d, read: %d, failed: %d\n",
		count.Total, count.Delivered, count.Read, count.Failed)

	// ── Scheduled batch ─────────────────────────────────────────────────────
	batch, err := client.Batch.Create(ctx, &zaple.CreateBatchRequest{
		Name:       "Weekend Flash Sale",
		TemplateID: templateID,
	})
	if err != nil {
		log.Fatalf("Create batch: %v", err)
	}

	_, err = client.Batch.UpsertContacts(ctx, batch.ID, []zaple.BatchContact{
		{CountryCode: "91", PhoneNumber: "919999999999", TemplateArguments: []string{"Alice", "50%"}},
	})
	if err != nil {
		log.Fatalf("UpsertContacts: %v", err)
	}

	// Schedule for a future time instead of sending immediately.
	_, err = client.Batch.Send(ctx, batch.ID, &zaple.SendBatchRequest{
		ScheduledAt: "2024-06-15T09:00:00Z",
	})
	if err != nil {
		log.Fatalf("Send batch: %v", err)
	}
	fmt.Println("✓ Batch scheduled for 2024-06-15 09:00 UTC")
}

func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Environment variable %s is required", key)
	}
	return val
}
