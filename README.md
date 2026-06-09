# zaple-go

[![Go Reference](https://pkg.go.dev/badge/github.com/zevaras/zaple_go.svg)](https://pkg.go.dev/github.com/zevaras/zaple_go)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.22-blue)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Go client library for the [Zaple](https://zaple.ai) WhatsApp Business API.

Covers the **Messaging API (V3)** and the **Batch API**.

## Requirements

- Go 1.22 or later
- Zaple API credentials — [app.zaple.ai/settings/api-dev](https://app.zaple.ai/settings/api-dev)

## Installation

```bash
go get github.com/zevaras/zaple_go
```

## Quick Start

```go
import zaple "github.com/zevaras/zaple_go"

client := zaple.NewClient(
    os.Getenv("ZAPLE_API_KEY"),
    os.Getenv("ZAPLE_API_SECRET"),
)

resp, err := client.Messaging.SendTemplate(ctx, &zaple.SendTemplateRequest{
    TemplateID:  "475546217187442007",
    CountryCode: "91",
    SendTo:      "919999999999",
})
```

## Configuration

```go
client := zaple.NewClient(apiKey, apiSecret,
    zaple.WithTimeout(15 * time.Second),
    zaple.WithMaxRetries(3),
    zaple.WithLogger(log.Default()),
    zaple.WithHTTPClient(myHTTPClient),
    zaple.WithBaseURL("https://app.zaple.ai"),
    zaple.WithUserAgent("MyApp/1.0"),
)
```

## API Coverage

| Service | Methods |
|---------|---------|
| `client.Messaging` | `SendTemplate`, `CreateTemplate`, `GetMessageStatus`, `GetTemplateDetails`, `GetTemplateStatus`, `GetMessageCount` |
| `client.Batch` | `Create`, `UpsertContacts`, `Send`, `GetStatus`, `GetDetails`, `Delete` |

## Error Handling

All methods return `*APIError` on failure. Use `errors.As` / `errors.Is` for structured handling:

```go
var apiErr *zaple.APIError
if errors.As(err, &apiErr) {
    fmt.Println(apiErr.StatusCode, apiErr.Code, apiErr.Message)
}

// Sentinel errors
errors.Is(err, zaple.ErrUnauthorized)
errors.Is(err, zaple.ErrRateLimited)
errors.Is(err, zaple.ErrDailyLimitReached)
errors.Is(err, zaple.ErrInactiveTemplate)
```

## Examples

See the [`examples/`](examples/) directory for runnable code covering:
- [`01_send_template`](examples/01_send_template/main.go) — single message, variables, media, quick replies
- [`02_batch_messaging`](examples/02_batch_messaging/main.go) — full batch campaign lifecycle
- [`03_advanced_usage`](examples/03_advanced_usage/main.go) — custom HTTP client, logger, mocking
- [`04_create_template`](examples/04_create_template/main.go) — create text, image, video, and auth templates

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT — see [LICENSE](LICENSE).
