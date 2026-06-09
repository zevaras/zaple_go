package zaple

import (
	"errors"
	"fmt"
	"net/http"
)

// Error code constants returned by the Zaple API.
const (
	ErrCodeUnauthorized        = "unauthorized"
	ErrCodeDailyLimitReached   = "daily_limit_reached"
	ErrCodePlanExpired         = "plan_expired"
	ErrCodeInsufficientBalance = "insufficient_balance"
	ErrCodeRateLimited         = "rate_limited"
	ErrCodeInactiveTemplate    = "inactive_template"
	ErrCodeNumberBlocked       = "number_blocked"
	ErrCodeValidation          = "validation_error"
	ErrCodeServerError         = "server_error"
	ErrCodeUnknown             = "unknown"
)

// Sentinel errors for use with errors.Is.
var (
	ErrUnauthorized        = errors.New("unauthorized")
	ErrDailyLimitReached   = errors.New("daily message limit reached")
	ErrPlanExpired         = errors.New("plan has expired")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrRateLimited         = errors.New("rate limited")
	ErrInactiveTemplate    = errors.New("template is inactive")
	ErrNumberBlocked       = errors.New("recipient number is blocked")
)

// ValidationErrors holds per-field validation messages returned by the API.
type ValidationErrors map[string][]string

func (v ValidationErrors) Error() string {
	var msgs []string
	for field, errs := range v {
		for _, e := range errs {
			msgs = append(msgs, field+": "+e)
		}
	}
	if len(msgs) == 0 {
		return "validation error"
	}
	return fmt.Sprintf("validation errors: %v", msgs)
}

// APIError represents an error response from the Zaple API.
// Use errors.As to unwrap it from a returned error value.
type APIError struct {
	// StatusCode is the HTTP status code returned by the server.
	StatusCode int

	// Code is the machine-readable error code from the API (see ErrCode* constants).
	Code string

	// Message is the human-readable error message from the API.
	Message string

	// ValidationErrors holds field-level validation messages (only populated
	// when StatusCode is 422).
	ValidationErrors ValidationErrors

	// sentinel is the matching sentinel error (used by Unwrap / errors.Is).
	sentinel error
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("zaple API error %d (%s): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("zaple API error %d: %s", e.StatusCode, e.Message)
}

// Unwrap allows errors.Is to match against sentinel errors.
func (e *APIError) Unwrap() error {
	return e.sentinel
}

// newAPIError builds an *APIError from an HTTP response and parsed body.
func newAPIError(statusCode int, code, message string, validation ValidationErrors) *APIError {
	err := &APIError{
		StatusCode:       statusCode,
		Code:             code,
		Message:          message,
		ValidationErrors: validation,
	}
	switch {
	case statusCode == http.StatusUnauthorized, code == ErrCodeUnauthorized:
		err.sentinel = ErrUnauthorized
	case statusCode == http.StatusTooManyRequests, code == ErrCodeRateLimited:
		err.sentinel = ErrRateLimited
	case code == ErrCodeDailyLimitReached:
		err.sentinel = ErrDailyLimitReached
	case code == ErrCodePlanExpired:
		err.sentinel = ErrPlanExpired
	case code == ErrCodeInsufficientBalance:
		err.sentinel = ErrInsufficientBalance
	case code == ErrCodeInactiveTemplate:
		err.sentinel = ErrInactiveTemplate
	case code == ErrCodeNumberBlocked:
		err.sentinel = ErrNumberBlocked
	}
	return err
}
