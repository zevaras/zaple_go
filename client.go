package zaple

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"
)

// Version is the current release of the library.
// Follows Semantic Versioning (https://semver.org).
const Version = "0.1.0"

// Client is the entry point for all Zaple API operations.
// Create one with NewClient and reuse it across your application — it is safe
// for concurrent use by multiple goroutines.
//
//	client := zaple.NewClient(apiKey, apiSecret)
//	resp, err := client.Messaging.SendTemplate(ctx, req)
type Client struct {
	cfg       clientConfig
	apiKey    string
	apiSecret string
	http      *http.Client

	// Messaging exposes the Zaple Messaging API (V3).
	Messaging *MessagingService

	// Batch exposes the Zaple Batch API for bulk campaigns.
	Batch *BatchService
}

// NewClient creates a new Zaple API client.
//
// apiKey and apiSecret are required and can be obtained from
// https://app.zaple.ai/settings/api-dev.
//
// Apply functional options to customise timeouts, retries, logging, etc.
//
//	client := zaple.NewClient(apiKey, apiSecret,
//	    zaple.WithTimeout(15*time.Second),
//	    zaple.WithMaxRetries(5),
//	    zaple.WithLogger(log.Default()),
//	)
func NewClient(apiKey, apiSecret string, opts ...Option) *Client {
	cfg := defaultConfig()
	for _, o := range opts {
		o(&cfg)
	}

	hc := cfg.httpClient
	if hc == nil {
		hc = &http.Client{Timeout: cfg.timeout}
	}

	c := &Client{
		cfg:       cfg,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		http:      hc,
	}

	c.Messaging = &MessagingService{client: c}
	c.Batch = &BatchService{client: c}

	return c
}

// ──────────────────────────────────────────────────────────────────────────────
// Internal HTTP plumbing
// ──────────────────────────────────────────────────────────────────────────────

// do executes an authenticated HTTP request with retry and error handling.
//
// method   — HTTP verb (GET, POST, DELETE, …)
// path     — path relative to the base URL, must start with "/"
// body     — request payload, marshalled to JSON (may be nil)
// result   — pointer to a struct that receives the decoded response (may be nil)
func (c *Client) do(ctx context.Context, method, path string, body, result any) error {
	url := strings.TrimRight(c.cfg.baseURL, "/") + path

	var attempt int
	for {
		attempt++

		err := c.executeOnce(ctx, method, url, body, result)
		if err == nil {
			return nil
		}

		if !c.shouldRetry(err, attempt) {
			return err
		}

		wait := c.retryWait(attempt)
		c.cfg.logger.Printf("[zaple] attempt %d failed: %v — retrying in %s", attempt, err, wait)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
}

func (c *Client) executeOnce(ctx context.Context, method, url string, body, result any) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("zaple: marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("zaple: build request: %w", err)
	}

	c.applyHeaders(req)

	c.cfg.logger.Printf("[zaple] %s %s", method, url)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("zaple: execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MiB cap
	if err != nil {
		return fmt.Errorf("zaple: read response body: %w", err)
	}

	c.cfg.logger.Printf("[zaple] %s %s → %d", method, url, resp.StatusCode)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if result != nil && len(respBody) > 0 {
			if err := json.Unmarshal(respBody, result); err != nil {
				return fmt.Errorf("zaple: decode response: %w", err)
			}
		}
		return nil
	}

	return parseAPIError(resp.StatusCode, respBody)
}

func (c *Client) applyHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.cfg.userAgent)
	req.Header.Set("Zaple-Api-Key", c.apiKey)
	req.Header.Set("Zaple-Api-Secret", c.apiSecret)
}

// shouldRetry returns true when the error is transient and we have attempts left.
func (c *Client) shouldRetry(err error, attempt int) bool {
	if attempt > c.cfg.retry.MaxRetries {
		return false
	}
	apiErr, ok := toAPIError(err)
	if !ok {
		// Network-level error — always retry.
		return true
	}
	// Retry on rate-limit and server errors only.
	return apiErr.StatusCode == http.StatusTooManyRequests || apiErr.StatusCode >= 500
}

// retryWait computes exponential backoff with full jitter.
func (c *Client) retryWait(attempt int) time.Duration {
	base := float64(c.cfg.retry.WaitMin)
	exp := base * math.Pow(2, float64(attempt-1))
	jitter := rand.Float64() * exp
	wait := time.Duration(jitter)
	if wait > c.cfg.retry.WaitMax {
		wait = c.cfg.retry.WaitMax
	}
	return wait
}

// ──────────────────────────────────────────────────────────────────────────────
// Error parsing
// ──────────────────────────────────────────────────────────────────────────────

// apiErrorBody covers the various error shapes the Zaple API returns.
type apiErrorBody struct {
	// Shape 1: {"success":false, "error":{"code":"...", "message":"..."}}
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`

	// Shape 2: {"status":"error", "message":"..."}
	Status  string `json:"status"`
	Message string `json:"message"`

	// Shape 3: {"success":false, "message":"...", "data":{...}}
	Success any                         `json:"success"`
	Data    map[string]json.RawMessage  `json:"data"`

	// Shape 4: {"success": 429}  (rate limit)
}

func parseAPIError(statusCode int, body []byte) *APIError {
	var parsed apiErrorBody
	_ = json.Unmarshal(body, &parsed) // best-effort; fall through on failure

	switch {
	case statusCode == http.StatusUnauthorized:
		return newAPIError(statusCode, ErrCodeUnauthorized, "Unauthorized", nil)

	case statusCode == http.StatusTooManyRequests:
		return newAPIError(statusCode, ErrCodeRateLimited, "Too many requests", nil)

	case statusCode == http.StatusUnprocessableEntity:
		validation := extractValidation(parsed.Data)
		msg := parsed.Message
		if msg == "" {
			msg = "Validation errors"
		}
		return newAPIError(statusCode, ErrCodeValidation, msg, validation)

	case parsed.Error != nil:
		return newAPIError(statusCode, parsed.Error.Code, parsed.Error.Message, nil)

	case parsed.Status == "error" && parsed.Message != "":
		return newAPIError(statusCode, codeFromMessage(parsed.Message), parsed.Message, nil)

	case parsed.Message != "":
		return newAPIError(statusCode, ErrCodeServerError, parsed.Message, nil)

	default:
		return newAPIError(statusCode, ErrCodeUnknown, fmt.Sprintf("HTTP %d", statusCode), nil)
	}
}

func extractValidation(data map[string]json.RawMessage) ValidationErrors {
	if len(data) == 0 {
		return nil
	}
	out := make(ValidationErrors, len(data))
	for field, raw := range data {
		var msgs []string
		if err := json.Unmarshal(raw, &msgs); err == nil {
			out[field] = msgs
		}
	}
	return out
}

func codeFromMessage(msg string) string {
	switch {
	case strings.Contains(msg, "inactive"):
		return ErrCodeInactiveTemplate
	case strings.Contains(msg, "blocked"):
		return ErrCodeNumberBlocked
	default:
		return ErrCodeUnknown
	}
}

func toAPIError(err error) (*APIError, bool) {
	if err == nil {
		return nil, false
	}
	apiErr, ok := err.(*APIError)
	return apiErr, ok
}
