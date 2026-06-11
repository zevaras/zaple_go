// Example 05: List and filter WhatsApp templates
//
// Demonstrates three common listing patterns:
//  1. Fetch all approved templates (paginated)
//  2. Search by name
//  3. Filter by category and active state
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
		zaple.WithTimeout(15*time.Second),
	)

	ctx := context.Background()

	// ── 1. All approved templates (first page) ───────────────────────────────
	approved, err := client.Messaging.ListTemplates(ctx, &zaple.ListTemplatesParams{
		Status:  "APPROVED",
		PerPage: 20,
	})
	if err != nil {
		handleError("list approved", err)
	} else {
		fmt.Printf("Approved templates: %d / %d total (page %d of %d)\n",
			len(approved.Templates),
			approved.Meta.Total,
			approved.Meta.CurrentPage,
			approved.Meta.LastPage,
		)
		for _, t := range approved.Templates {
			favorite := ""
			if t.IsFavorite {
				favorite = " ★"
			}
			fmt.Printf("  [%s] %s — %s%s\n", t.TemplateID, t.Name, t.Category, favorite)
		}
		if len(approved.Stats) > 0 {
			fmt.Println("Stats:")
			for _, s := range approved.Stats {
				fmt.Printf("  %s: %s\n", s.Label, s.Value)
			}
		}
	}

	fmt.Println()

	// ── 2. Search by name ────────────────────────────────────────────────────
	results, err := client.Messaging.ListTemplates(ctx, &zaple.ListTemplatesParams{
		Search: "order",
	})
	if err != nil {
		handleError("search templates", err)
	} else {
		fmt.Printf("Search 'order': %d result(s)\n", len(results.Templates))
		for _, t := range results.Templates {
			fmt.Printf("  %s — %s (%s)\n", t.Name, t.Status, t.HeaderType)
		}
	}

	fmt.Println()

	// ── 3. Active utility templates only ─────────────────────────────────────
	active := true
	utility, err := client.Messaging.ListTemplates(ctx, &zaple.ListTemplatesParams{
		Category: "UTILITY",
		Active:   &active,
		Page:     1,
		PerPage:  10,
	})
	if err != nil {
		handleError("list utility", err)
	} else {
		fmt.Printf("Active utility templates: %d\n", len(utility.Templates))
		for _, t := range utility.Templates {
			fmt.Printf("  %s — vars: %d, created: %s\n",
				t.Name, t.VariableCount, t.CreatedAt)
		}
	}
}

func handleError(label string, err error) {
	var apiErr *zaple.APIError
	if errors.As(err, &apiErr) {
		fmt.Fprintf(os.Stderr, "✗ %s — API %d [%s]: %s\n",
			label, apiErr.StatusCode, apiErr.Code, apiErr.Message)
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
