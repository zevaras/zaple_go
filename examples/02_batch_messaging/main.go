// Example 02: Bulk WhatsApp campaign with the Batch API
//
// This example walks through the full batch lifecycle:
//  1. Create a batch list
//  2. Upsert contacts (personalised per-contact arguments)
//  3. Trigger dispatch
//  4. Poll for completion
//  5. Clean up (delete)
//
// Run:
//
//	ZAPLE_API_KEY=<key> ZAPLE_API_SECRET=<secret> go run main.go
package main

import (
	"context"
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
		zaple.WithLogger(log.Default()),
	)

	ctx := context.Background()

	// ── Step 1: Create a batch ──────────────────────────────────────────────
	batch, err := client.Batch.Create(ctx, &zaple.CreateBatchRequest{
		Name:       "June Promo Campaign",
		TemplateID: "YOUR_TEMPLATE_ID",
	})
	if err != nil {
		log.Fatalf("Create batch: %v", err)
	}
	fmt.Printf("✓ Batch created: %s (status: %s)\n", batch.ID, batch.Status)

	// ── Step 2: Add contacts ────────────────────────────────────────────────
	contacts := []zaple.BatchContact{
		{
			CountryCode:       "91",
			PhoneNumber:       "919999999999",
			TemplateArguments: []string{"Alice", "Gold"},
			Metadata:          map[string]any{"user_id": 101},
		},
		{
			CountryCode:       "91",
			PhoneNumber:       "918888888888",
			TemplateArguments: []string{"Bob", "Silver"},
			Metadata:          map[string]any{"user_id": 102},
		},
		{
			CountryCode:       "1",
			PhoneNumber:       "14155550100",
			TemplateArguments: []string{"Carol", "Platinum"},
			Metadata:          map[string]any{"user_id": 103},
		},
	}

	upsertResp, err := client.Batch.UpsertContacts(ctx, batch.ID, contacts)
	if err != nil {
		log.Fatalf("Upsert contacts: %v", err)
	}
	fmt.Printf("✓ Contacts upserted — inserted: %d, updated: %d, skipped: %d\n",
		upsertResp.Inserted, upsertResp.Updated, upsertResp.Skipped)

	// ── Step 3: Trigger dispatch ────────────────────────────────────────────
	sendResp, err := client.Batch.Send(ctx, batch.ID, nil) // nil = send immediately
	if err != nil {
		log.Fatalf("Send batch: %v", err)
	}
	fmt.Printf("✓ Batch queued: %s\n", sendResp.Status)

	// ── Step 4: Poll for completion ─────────────────────────────────────────
	fmt.Println("Polling for completion…")
	for {
		time.Sleep(5 * time.Second)

		status, err := client.Batch.GetStatus(ctx, batch.ID)
		if err != nil {
			log.Printf("GetStatus error: %v\n", err)
			continue
		}

		fmt.Printf("  Progress: %d/%d sent, %d delivered, %d failed (%.0f%%)\n",
			status.SentCount, status.TotalContacts,
			status.DeliveredCount, status.FailedCount,
			status.Progress)

		switch status.Status {
		case zaple.BatchStatusCompleted, zaple.BatchStatusPartial:
			fmt.Printf("✓ Batch finished with status: %s\n", status.Status)
			goto done
		case zaple.BatchStatusFailed:
			log.Fatalf("✗ Batch failed")
		}
	}

done:
	// ── Step 5 (optional): Retrieve full details ────────────────────────────
	details, err := client.Batch.GetDetails(ctx, batch.ID)
	if err != nil {
		log.Printf("GetDetails: %v\n", err)
		return
	}
	fmt.Printf("Final stats — sent: %d, delivered: %d, failed: %d\n",
		details.SentCount, details.DeliveredCount, details.FailedCount)
}

func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Environment variable %s is required", key)
	}
	return val
}
