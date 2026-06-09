package zaple_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	zaple "github.com/zevaras/zaple_go"
)

// ──────────────────────────────────────────────────────────────────────────────
// Test helpers
// ──────────────────────────────────────────────────────────────────────────────

// newTestServer starts an httptest.Server and returns both the server and a
// Client wired to talk to it.
func newTestServer(t *testing.T, handler http.Handler) (*httptest.Server, *zaple.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client := zaple.NewClient("test-key", "test-secret",
		zaple.WithBaseURL(srv.URL),
		zaple.WithTimeout(5*time.Second),
		zaple.WithMaxRetries(0), // disable retries for deterministic tests
	)
	return srv, client
}

func jsonHandler(t *testing.T, statusCode int, body any) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if err := json.NewEncoder(w).Encode(body); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Client construction
// ──────────────────────────────────────────────────────────────────────────────

func TestNewClient_defaults(t *testing.T) {
	c := zaple.NewClient("key", "secret")
	if c == nil {
		t.Fatal("expected non-nil Client")
	}
	if c.Messaging == nil {
		t.Error("expected Messaging service to be initialised")
	}
	if c.Batch == nil {
		t.Error("expected Batch service to be initialised")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Messaging.SendTemplate
// ──────────────────────────────────────────────────────────────────────────────

func TestSendTemplate_success(t *testing.T) {
	_, client := newTestServer(t, jsonHandler(t, http.StatusOK, map[string]any{
		"status":     "Message sent successfully.",
		"message_id": "abc123",
	}))

	resp, err := client.Messaging.SendTemplate(context.Background(), &zaple.SendTemplateRequest{
		TemplateID:  "tpl1",
		CountryCode: "91",
		SendTo:      "919999999999",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.MessageID != "abc123" {
		t.Errorf("MessageID: got %q, want %q", resp.MessageID, "abc123")
	}
}

func TestSendTemplate_withArguments(t *testing.T) {
	var captured map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":     "Message sent successfully.",
			"message_id": "xyz",
		})
	}))
	defer srv.Close()

	client := zaple.NewClient("k", "s",
		zaple.WithBaseURL(srv.URL),
		zaple.WithMaxRetries(0),
	)

	_, err := client.Messaging.SendTemplate(context.Background(), &zaple.SendTemplateRequest{
		TemplateID:        "tpl1",
		CountryCode:       "91",
		SendTo:            "919999999999",
		TemplateArguments: []string{"Alice", "ORD-99"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if captured["template_argument1"] != "Alice" {
		t.Errorf("template_argument1: got %v, want Alice", captured["template_argument1"])
	}
	if captured["template_argument2"] != "ORD-99" {
		t.Errorf("template_argument2: got %v, want ORD-99", captured["template_argument2"])
	}
}

func TestSendTemplate_validatesRequired(t *testing.T) {
	client := zaple.NewClient("k", "s")

	tests := []struct {
		name string
		req  *zaple.SendTemplateRequest
	}{
		{"nil request", nil},
		{"missing TemplateID", &zaple.SendTemplateRequest{CountryCode: "91", SendTo: "91999"}},
		{"missing CountryCode", &zaple.SendTemplateRequest{TemplateID: "t1", SendTo: "91999"}},
		{"missing SendTo", &zaple.SendTemplateRequest{TemplateID: "t1", CountryCode: "91"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.Messaging.SendTemplate(context.Background(), tt.req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestSendTemplate_401(t *testing.T) {
	_, client := newTestServer(t, jsonHandler(t, http.StatusUnauthorized, map[string]any{
		"success": false,
		"error": map[string]any{
			"code":    "unauthorized",
			"message": "Unauthorized",
		},
	}))

	_, err := client.Messaging.SendTemplate(context.Background(), &zaple.SendTemplateRequest{
		TemplateID:  "t1",
		CountryCode: "91",
		SendTo:      "91999",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, zaple.ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
	var apiErr *zaple.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError")
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode: got %d, want 401", apiErr.StatusCode)
	}
}

func TestSendTemplate_400_dailyLimit(t *testing.T) {
	_, client := newTestServer(t, jsonHandler(t, http.StatusBadRequest, map[string]any{
		"success": false,
		"error": map[string]any{
			"code":    "daily_limit_reached",
			"message": "You've hit your daily message limit.",
		},
	}))

	_, err := client.Messaging.SendTemplate(context.Background(), &zaple.SendTemplateRequest{
		TemplateID: "t1", CountryCode: "91", SendTo: "91999",
	})
	if !errors.Is(err, zaple.ErrDailyLimitReached) {
		t.Errorf("expected ErrDailyLimitReached, got %v", err)
	}
}

func TestSendTemplate_422_validation(t *testing.T) {
	_, client := newTestServer(t, jsonHandler(t, http.StatusUnprocessableEntity, map[string]any{
		"success": false,
		"message": "Validation errors",
		"data": map[string]any{
			"template_id": []string{"The template id field is required."},
		},
	}))

	_, err := client.Messaging.SendTemplate(context.Background(), &zaple.SendTemplateRequest{
		TemplateID: "t1", CountryCode: "91", SendTo: "91999",
	})

	var apiErr *zaple.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Code != zaple.ErrCodeValidation {
		t.Errorf("Code: got %q, want %q", apiErr.Code, zaple.ErrCodeValidation)
	}
	if len(apiErr.ValidationErrors) == 0 {
		t.Error("expected ValidationErrors to be populated")
	}
}

func TestSendTemplate_authHeaders(t *testing.T) {
	var gotKey, gotSecret string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("Zaple-Api-Key")
		gotSecret = r.Header.Get("Zaple-Api-Secret")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok", "message_id": "1"})
	}))
	defer srv.Close()

	client := zaple.NewClient("my-key", "my-secret",
		zaple.WithBaseURL(srv.URL),
		zaple.WithMaxRetries(0),
	)
	client.Messaging.SendTemplate(context.Background(), &zaple.SendTemplateRequest{ //nolint:errcheck
		TemplateID: "t", CountryCode: "91", SendTo: "9199",
	})

	if gotKey != "my-key" {
		t.Errorf("Zaple-Api-Key: got %q, want %q", gotKey, "my-key")
	}
	if gotSecret != "my-secret" {
		t.Errorf("Zaple-Api-Secret: got %q, want %q", gotSecret, "my-secret")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Batch service
// ──────────────────────────────────────────────────────────────────────────────

func TestBatch_Create_success(t *testing.T) {
	_, client := newTestServer(t, jsonHandler(t, http.StatusOK, map[string]any{
		"id":         "batch-001",
		"name":       "Test Campaign",
		"template_id": "tpl1",
		"status":     "draft",
	}))

	batch, err := client.Batch.Create(context.Background(), &zaple.CreateBatchRequest{
		Name:       "Test Campaign",
		TemplateID: "tpl1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if batch.ID != "batch-001" {
		t.Errorf("ID: got %q, want batch-001", batch.ID)
	}
	if batch.Status != zaple.BatchStatusDraft {
		t.Errorf("Status: got %q, want draft", batch.Status)
	}
}

func TestBatch_Create_validatesRequired(t *testing.T) {
	client := zaple.NewClient("k", "s")

	tests := []struct {
		name string
		req  *zaple.CreateBatchRequest
	}{
		{"nil", nil},
		{"missing Name", &zaple.CreateBatchRequest{TemplateID: "t"}},
		{"missing TemplateID", &zaple.CreateBatchRequest{Name: "n"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.Batch.Create(context.Background(), tt.req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestBatch_UpsertContacts_serialisesArguments(t *testing.T) {
	var captured map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&captured) //nolint:errcheck
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"inserted": 1, "updated": 0, "skipped": 0})
	}))
	defer srv.Close()

	client := zaple.NewClient("k", "s",
		zaple.WithBaseURL(srv.URL),
		zaple.WithMaxRetries(0),
	)

	_, err := client.Batch.UpsertContacts(context.Background(), "batch-001", []zaple.BatchContact{
		{CountryCode: "91", PhoneNumber: "9199", TemplateArguments: []string{"Alice", "Gold"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	contacts, _ := captured["contacts"].([]any)
	if len(contacts) != 1 {
		t.Fatalf("expected 1 contact, got %d", len(contacts))
	}
	c := contacts[0].(map[string]any)
	if c["template_argument1"] != "Alice" {
		t.Errorf("template_argument1: got %v, want Alice", c["template_argument1"])
	}
	if c["template_argument2"] != "Gold" {
		t.Errorf("template_argument2: got %v, want Gold", c["template_argument2"])
	}
}

func TestBatch_Delete_success(t *testing.T) {
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := client.Batch.Delete(context.Background(), "batch-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Error type
// ──────────────────────────────────────────────────────────────────────────────

func TestAPIError_Error(t *testing.T) {
	e := &zaple.APIError{StatusCode: 400, Code: "daily_limit_reached", Message: "limit hit"}
	want := "zaple API error 400 (daily_limit_reached): limit hit"
	if e.Error() != want {
		t.Errorf("got %q, want %q", e.Error(), want)
	}
}

func TestAPIError_sentinels(t *testing.T) {
	tests := []struct {
		statusCode int
		code       string
		sentinel   error
	}{
		{http.StatusUnauthorized, zaple.ErrCodeUnauthorized, zaple.ErrUnauthorized},
		{http.StatusTooManyRequests, zaple.ErrCodeRateLimited, zaple.ErrRateLimited},
		{http.StatusBadRequest, zaple.ErrCodeDailyLimitReached, zaple.ErrDailyLimitReached},
		{http.StatusBadRequest, zaple.ErrCodePlanExpired, zaple.ErrPlanExpired},
		{http.StatusBadRequest, zaple.ErrCodeInsufficientBalance, zaple.ErrInsufficientBalance},
	}

	for _, tt := range tests {
		_, client := newTestServer(t, jsonHandler(t, tt.statusCode, map[string]any{
			"success": false,
			"error":   map[string]any{"code": tt.code, "message": "test"},
		}))

		_, err := client.Messaging.SendTemplate(context.Background(), &zaple.SendTemplateRequest{
			TemplateID: "t", CountryCode: "91", SendTo: "9199",
		})

		if !errors.Is(err, tt.sentinel) {
			t.Errorf("code=%s: expected errors.Is(err, %v), got %v", tt.code, tt.sentinel, err)
		}
	}
}
