// Package zaple provides a Go client library for the Zaple (https://zaple.ai) WhatsApp Business API.
//
// # Overview
//
// Zaple (https://zaple.ai) is a WhatsApp Business API platform. This library
// covers the Messaging API (V3) and the Batch API, giving you a type-safe,
// idiomatic Go interface to send template messages and run bulk campaigns.
//
// # Authentication
//
// Every API call requires an API key and an API secret, which you can obtain
// from your Zaple dashboard at https://app.zaple.ai/settings/api-dev.
//
// # Quick Start
//
//	client := zaple.NewClient("YOUR_API_KEY", "YOUR_API_SECRET")
//
//	resp, err := client.Messaging.SendTemplate(ctx, &zaple.SendTemplateRequest{
//	    TemplateID:  "475546217187442007",
//	    CountryCode: "91",
//	    SendTo:      "919999999999",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Message ID:", resp.MessageID)
//
// # Error Handling
//
// All methods return a typed *APIError when the server responds with a non-2xx
// status. You can inspect the error with errors.As or switch on its Code field:
//
//	_, err := client.Messaging.SendTemplate(ctx, req)
//	var apiErr *zaple.APIError
//	if errors.As(err, &apiErr) {
//	    switch apiErr.Code {
//	    case zaple.ErrCodeDailyLimitReached:
//	        // handle limit
//	    case zaple.ErrCodeUnauthorized:
//	        // handle auth
//	    }
//	}
//
// # Configuring the Client
//
// Use functional options to customise the client:
//
//	client := zaple.NewClient(apiKey, apiSecret,
//	    zaple.WithTimeout(30*time.Second),
//	    zaple.WithMaxRetries(3),
//	    zaple.WithLogger(myLogger),
//	)
//
// # Concurrency
//
// The Client is safe for concurrent use across goroutines.
package zaple
