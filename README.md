# zaple-go

[![Go Reference](https://pkg.go.dev/badge/github.com/zevaras/zaple_go.svg)](https://pkg.go.dev/github.com/zevaras/zaple_go)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.22-blue)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Official Go client library for the [Zaple](https://zaple.ai) WhatsApp Business API.

Covers:
- **Messaging API (V3)** — send individual WhatsApp template messages, query delivery status and message counts, inspect template details and approval status.
- **Batch API** — create bulk campaigns, manage recipient lists, trigger dispatch, and track progress.

---

## Table of Contents

- [Requirements](#requirements)
- [Installation](#installation)
- [Authentication](#authentication)
- [Quick Start](#quick-start)
- [Client Configuration](#client-configuration)
- [Messaging API](#messaging-api)
  - [Send Template Message](#send-template-message)
  - [Get Message Status](#get-message-status)
  - [Get Template Details](#get-template-details)
  - [Get Template Status](#get-template-status)
  - [Get Message Count](#get-message-count)
- [Batch API](#batch-api)
  - [Create Batch](#create-batch)
  - [Upsert Contacts](#upsert-contacts)
  - [Send Batch](#send-batch)
  - [Get Batch Status](#get-batch-status)
  - [Get Batch Details](#get-batch-details)
  - [Delete Batch](#delete-batch)
- [Error Handling](#error-handling)
- [Testing & Mocking](#testing--mocking)
- [Examples](#examples)
- [Future Enhancements](#future-enhancements)
- [Contributing](#contributing)
- [License](#license)

---

## Requirements

- Go **1.22** or later
- A Zaple account with API credentials — obtain yours from [app.zaple.ai/settings/api-dev](https://app.zaple.ai/settings/api-dev)

## Installation

```bash
go get github.com/zevaras/zaple_go
```

## Authentication

All API calls require an **API key** and an **API secret**.

```go
import zaple "github.com/zevaras/zaple_go"

client := zaple.NewClient("YOUR_API_KEY", "YOUR_API_SECRET")
```

Store credentials in environment variables — never hard-code them in source files:

```bash
export ZAPLE_API_KEY="your_api_key"
export ZAPLE_API_SECRET="your_api_secret"
```

```go
client := zaple.NewClient(
    os.Getenv("ZAPLE_API_KEY"),
    os.Getenv("ZAPLE_API_SECRET"),
)
```

---

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    zaple "github.com/zevaras/zaple_go"
)

func main() {
    client := zaple.NewClient(
        os.Getenv("ZAPLE_API_KEY"),
        os.Getenv("ZAPLE_API_SECRET"),
    )

    resp, err := client.Messaging.SendTemplate(context.Background(), &zaple.SendTemplateRequest{
        TemplateID:        "475546217187442007",
        CountryCode:       "91",
        SendTo:            "919999999999",
        TemplateArguments: []string{"Alice", "Order #1234"},
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Message ID:", resp.MessageID)
}
```

---

## Client Configuration

`NewClient` accepts functional options to customise its behaviour:

```go
client := zaple.NewClient(apiKey, apiSecret,
    // HTTP request timeout (default: 30s)
    zaple.WithTimeout(15 * time.Second),

    // Number of retries on transient errors — 429, 5xx (default: 3)
    zaple.WithMaxRetries(5),

    // Fine-grained retry config with backoff bounds
    zaple.WithRetryConfig(zaple.RetryConfig{
        MaxRetries: 3,
        WaitMin:    500 * time.Millisecond,
        WaitMax:    10 * time.Second,
    }),

    // Attach a logger (any type with Printf(string, ...any))
    zaple.WithLogger(log.Default()),

    // Bring your own *http.Client (proxy, mTLS, custom transport, etc.)
    zaple.WithHTTPClient(myHTTPClient),

    // Override the base URL (useful for staging or proxies)
    zaple.WithBaseURL("https://staging.zaple.ai"),

    // Append to the User-Agent header
    zaple.WithUserAgent("MyApp/2.0"),
)
```

| Option | Default | Description |
|--------|---------|-------------|
| `WithTimeout` | 30 s | Per-request HTTP timeout |
| `WithMaxRetries` | 3 | Max retries on 429/5xx |
| `WithRetryConfig` | — | Full retry control (overrides `WithMaxRetries`) |
| `WithHTTPClient` | stdlib default | Custom `*http.Client` |
| `WithBaseURL` | `https://app.zaple.ai` | API base URL |
| `WithLogger` | silent | Debug logger |
| `WithUserAgent` | — | Appended to the `User-Agent` header |

---

## Messaging API

### Send Template Message

Send a pre-approved WhatsApp template to a single recipient.

```go
resp, err := client.Messaging.SendTemplate(ctx, &zaple.SendTemplateRequest{
    TemplateID:  "475546217187442007", // required
    CountryCode: "91",                 // required — digits only, no "+"
    SendTo:      "919999999999",       // required — full number with country code
})
```

**With template variables** ({{1}}, {{2}}, … placeholders):

```go
resp, err := client.Messaging.SendTemplate(ctx, &zaple.SendTemplateRequest{
    TemplateID:        "475546217187442007",
    CountryCode:       "91",
    SendTo:            "919999999999",
    TemplateArguments: []string{"Alice", "Order #4567"}, // maps to {{1}}, {{2}}
})
```

**With a media header** (image, video, or document):

```go
resp, err := client.Messaging.SendTemplate(ctx, &zaple.SendTemplateRequest{
    TemplateID:   "YOUR_MEDIA_TEMPLATE_ID",
    CountryCode:  "91",
    SendTo:       "919999999999",
    MediaURL:     "https://example.com/promo.jpg",
    MediaURLType: zaple.MediaURLTypeURL,
    // Use zaple.MediaURLTypeBase64 when providing a base64-encoded payload.
})
```

**With quick-reply buttons and callback metadata**:

```go
resp, err := client.Messaging.SendTemplate(ctx, &zaple.SendTemplateRequest{
    TemplateID:         "YOUR_QUICK_REPLY_TEMPLATE_ID",
    CountryCode:        "91",
    SendTo:             "919999999999",
    QuickReplyPayload1: "confirm_appt_42",
    QuickReplyPayload2: "cancel_appt_42",
    BizOpaqueCallbackData: map[string]any{
        "appointment_id": 42,
        "source":         "booking_system",
    },
})
```

**Response**:

```go
fmt.Println(resp.Status)    // "Message sent successfully."
fmt.Println(resp.MessageID) // "x927831064523907185K9CDMYNReFgHIPWz"
```

---

### Get Message Status

```go
status, err := client.Messaging.GetMessageStatus(ctx, "x927831064523907185K9CDMYNReFgHIPWz")
if err != nil { /* ... */ }

fmt.Println(status.Status)      // "delivered"
fmt.Println(status.DeliveredAt) // "2024-06-10T08:30:00Z"
```

---

### Get Template Details

```go
tpl, err := client.Messaging.GetTemplateDetails(ctx, "475546217187442007")
if err != nil { /* ... */ }

fmt.Println(tpl.Name)     // "order_update"
fmt.Println(tpl.Status)   // "APPROVED"
fmt.Println(tpl.Category) // "UTILITY"
for _, comp := range tpl.Components {
    fmt.Printf("  %s: %s\n", comp.Type, comp.Text)
}
```

---

### Get Template Status

Check the WhatsApp approval status before sending:

```go
ts, err := client.Messaging.GetTemplateStatus(ctx, "475546217187442007")
if err != nil { /* ... */ }

if ts.Status != "APPROVED" {
    log.Fatalf("template not approved: %s — reason: %s", ts.Status, ts.RejectedReason)
}
```

---

### Get Message Count

```go
count, err := client.Messaging.GetMessageCount(ctx, &zaple.MessageCountParams{
    From: "2024-06-01",
    To:   "2024-06-30",
    // TemplateID: "...",  // optional — filter by template
})
if err != nil { /* ... */ }

fmt.Printf("Sent: %d  Delivered: %d  Read: %d  Failed: %d\n",
    count.Total, count.Delivered, count.Read, count.Failed)
```

---

## Batch API

The Batch API is designed for high-volume campaigns. The typical workflow is:

```
Create → UpsertContacts → Send → poll GetStatus → (optionally) Delete
```

### Create Batch

```go
batch, err := client.Batch.Create(ctx, &zaple.CreateBatchRequest{
    Name:       "June Promo Campaign",   // required
    TemplateID: "475546217187442007",    // required
    // ScheduledAt: "2024-06-15T09:00:00Z", // optional future send time
})
```

### Upsert Contacts

Add or update recipients. Each contact can carry its own template arguments and metadata:

```go
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
}

result, err := client.Batch.UpsertContacts(ctx, batch.ID, contacts)
fmt.Printf("Inserted: %d, Updated: %d, Skipped: %d\n",
    result.Inserted, result.Updated, result.Skipped)
```

Call `UpsertContacts` multiple times to load contacts incrementally from large data sources.

### Send Batch

Trigger message dispatch. Pass `nil` to send immediately:

```go
// Send now
resp, err := client.Batch.Send(ctx, batch.ID, nil)

// Schedule for a future time
resp, err := client.Batch.Send(ctx, batch.ID, &zaple.SendBatchRequest{
    ScheduledAt: "2024-06-15T09:00:00Z",
})
```

### Get Batch Status

Poll for real-time campaign progress:

```go
for {
    time.Sleep(5 * time.Second)

    status, err := client.Batch.GetStatus(ctx, batch.ID)
    if err != nil { /* ... */ }

    fmt.Printf("Progress: %d/%d (%.0f%%) — delivered: %d, failed: %d\n",
        status.SentCount, status.TotalContacts,
        status.Progress, status.DeliveredCount, status.FailedCount)

    if status.Status == zaple.BatchStatusCompleted ||
       status.Status == zaple.BatchStatusPartial {
        break
    }
}
```

Available status values: `BatchStatusDraft`, `BatchStatusPending`, `BatchStatusProcessing`, `BatchStatusCompleted`, `BatchStatusFailed`, `BatchStatusPartial`.

### Get Batch Details

```go
batch, err := client.Batch.GetDetails(ctx, batchID)
fmt.Println(batch.Name, batch.Status, batch.SentCount)
```

### Delete Batch

Only draft batches can be deleted:

```go
err := client.Batch.Delete(ctx, batchID)
```

---

## Error Handling

All methods return a `*APIError` on non-2xx responses. Use `errors.As` and `errors.Is` for structured error handling:

```go
resp, err := client.Messaging.SendTemplate(ctx, req)
if err != nil {
    var apiErr *zaple.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("HTTP %d [%s]: %s\n", apiErr.StatusCode, apiErr.Code, apiErr.Message)

        // Validation errors (HTTP 422) carry field-level details:
        for field, msgs := range apiErr.ValidationErrors {
            fmt.Printf("  %s: %v\n", field, msgs)
        }
    }

    // Match specific conditions with sentinel errors:
    switch {
    case errors.Is(err, zaple.ErrUnauthorized):
        // Invalid credentials
    case errors.Is(err, zaple.ErrRateLimited):
        // Back off and retry
    case errors.Is(err, zaple.ErrDailyLimitReached):
        // Upgrade plan or wait for reset
    case errors.Is(err, zaple.ErrPlanExpired):
        // Renew subscription
    case errors.Is(err, zaple.ErrInsufficientBalance):
        // Top up credits
    case errors.Is(err, zaple.ErrInactiveTemplate):
        // Await template approval
    case errors.Is(err, zaple.ErrNumberBlocked):
        // Remove number from send list
    }
}
```

### HTTP status → error mapping

| HTTP Status | Sentinel Error | Code constant |
|-------------|----------------|---------------|
| 400 | `ErrDailyLimitReached` | `ErrCodeDailyLimitReached` |
| 400 | `ErrPlanExpired` | `ErrCodePlanExpired` |
| 400 | `ErrInsufficientBalance` | `ErrCodeInsufficientBalance` |
| 401 | `ErrUnauthorized` | `ErrCodeUnauthorized` |
| 419 | `ErrInactiveTemplate` | `ErrCodeInactiveTemplate` |
| 419 | `ErrNumberBlocked` | `ErrCodeNumberBlocked` |
| 422 | — | `ErrCodeValidation` |
| 429 | `ErrRateLimited` | `ErrCodeRateLimited` |
| 5xx | — | `ErrCodeServerError` |

---

## Testing & Mocking

The library exposes `MessagingAPI` and `BatchAPI` interfaces so you can swap in a mock during tests without reaching the network:

```go
// In your application code, depend on the interface — not the concrete type.
type OrderService struct {
    messaging zaple.MessagingAPI
}

func NewOrderService(m zaple.MessagingAPI) *OrderService {
    return &OrderService{messaging: m}
}
```

```go
// In your tests, implement the interface with a stub.
type mockMessaging struct{}

func (m *mockMessaging) SendTemplate(_ context.Context, req *zaple.SendTemplateRequest) (*zaple.SendTemplateResponse, error) {
    return &zaple.SendTemplateResponse{MessageID: "test-id"}, nil
}

// Satisfy the rest of the interface with no-ops.
func (m *mockMessaging) GetMessageStatus(context.Context, string) (*zaple.MessageStatus, error) { return nil, nil }
func (m *mockMessaging) GetTemplateDetails(context.Context, string) (*zaple.TemplateDetails, error) { return nil, nil }
func (m *mockMessaging) GetTemplateStatus(context.Context, string) (*zaple.TemplateStatus, error) { return nil, nil }
func (m *mockMessaging) GetMessageCount(context.Context, *zaple.MessageCountParams) (*zaple.MessageCount, error) { return nil, nil }

func TestOrderService_sends_on_creation(t *testing.T) {
    svc := NewOrderService(&mockMessaging{})
    // ... test svc
}
```

Alternatively, start a local `httptest.Server` (see [`zaple_test.go`](zaple_test.go)) to test against a realistic HTTP server without hitting the live API.

---

## Examples

Working examples are in the [`examples/`](examples/) directory:

| Example | Description |
|---------|-------------|
| [`01_send_template`](examples/01_send_template/main.go) | Send text, media, variable, and quick-reply templates |
| [`02_batch_messaging`](examples/02_batch_messaging/main.go) | Full batch lifecycle: create → upsert → send → poll |
| [`03_advanced_usage`](examples/03_advanced_usage/main.go) | Custom HTTP client, logger, scheduled batch, interface usage |

Run any example:

```bash
ZAPLE_API_KEY=your_key ZAPLE_API_SECRET=your_secret \
    go run examples/01_send_template/main.go
```

---

## Future Enhancements

The following capabilities are planned for future releases. Contributions are welcome!

| Feature | Description |
|---------|-------------|
| **Webhook verification** | Parse and cryptographically verify incoming Zaple webhook payloads |
| **Service Messages API** | Send transactional/session messages outside template constraints |
| **Templates API** | Create, update, and submit templates for WhatsApp approval |
| **Leads API** | Submit and manage leads via the Zaple CRM integration |
| **Catalog API** | Manage WhatsApp Commerce product catalogs |
| **OpenTelemetry tracing** | First-class distributed tracing via `go.opentelemetry.io/otel` |
| **Rate-limit tracking** | Expose `X-RateLimit-*` headers and provide an automatic token-bucket throttle |
| **Context-aware pagination** | Cursor/page-based iteration helpers for list endpoints |
| **Circuit breaker** | Automatic request suspension when the service is degraded |
| **Template caching** | In-process TTL cache for template details to reduce API calls |
| **Structured logging** | `log/slog` support as an alternative logger interface |

---

## Contributing

Contributions, bug reports, and feature requests are all welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.

```bash
# Clone and set up
git clone https://github.com/zevaras/zaple_go.git
cd zaple_go

# Run tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Lint (requires golangci-lint)
golangci-lint run
```

---

## License

MIT — see [LICENSE](LICENSE).
