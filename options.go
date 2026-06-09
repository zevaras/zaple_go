package zaple

import (
	"net/http"
	"time"
)

const (
	defaultBaseURL    = "https://app.zaple.ai"
	defaultTimeout    = 30 * time.Second
	defaultMaxRetries = 3
	defaultRetryWait  = 1 * time.Second
	defaultRetryMax   = 30 * time.Second
	defaultUserAgent  = "zaple-go/" + Version
)

// Logger is the interface that the client uses for debug logging.
// Pass any logger that satisfies this interface via WithLogger.
// The standard library's log.Printf satisfies this signature when wrapped.
type Logger interface {
	Printf(format string, v ...any)
}

// noopLogger silently discards all log output.
type noopLogger struct{}

func (noopLogger) Printf(_ string, _ ...any) {}

// RetryConfig controls retry behaviour on transient failures.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts (not including the
	// initial attempt). Set to 0 to disable retries.
	MaxRetries int

	// WaitMin is the minimum wait duration between retries.
	WaitMin time.Duration

	// WaitMax is the maximum wait duration between retries (exponential backoff
	// is capped at this value).
	WaitMax time.Duration
}

// Option is a functional option for configuring a Client.
type Option func(*clientConfig)

type clientConfig struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
	retry      RetryConfig
	logger     Logger
	userAgent  string
}

func defaultConfig() clientConfig {
	return clientConfig{
		baseURL:   defaultBaseURL,
		timeout:   defaultTimeout,
		userAgent: defaultUserAgent,
		retry: RetryConfig{
			MaxRetries: defaultMaxRetries,
			WaitMin:    defaultRetryWait,
			WaitMax:    defaultRetryMax,
		},
		logger: noopLogger{},
	}
}

// WithBaseURL overrides the API base URL. Useful for testing against a proxy or
// staging environment.
//
//	zaple.WithBaseURL("https://staging.zaple.ai")
func WithBaseURL(url string) Option {
	return func(c *clientConfig) {
		c.baseURL = url
	}
}

// WithHTTPClient replaces the default HTTP client. Use this when you need to
// configure a custom TLS config, proxy, or transport.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *clientConfig) {
		c.httpClient = hc
	}
}

// WithTimeout sets the HTTP request timeout. Defaults to 30 seconds.
func WithTimeout(d time.Duration) Option {
	return func(c *clientConfig) {
		c.timeout = d
	}
}

// WithMaxRetries sets the number of retry attempts on transient errors (5xx,
// 429). Set to 0 to disable retries entirely. Defaults to 3.
func WithMaxRetries(n int) Option {
	return func(c *clientConfig) {
		c.retry.MaxRetries = n
	}
}

// WithRetryConfig fully replaces the retry configuration.
func WithRetryConfig(rc RetryConfig) Option {
	return func(c *clientConfig) {
		c.retry = rc
	}
}

// WithLogger sets the logger used for debug-level output. By default all
// logging is suppressed. Pass nil to disable logging explicitly.
//
//	import "log"
//	zaple.WithLogger(log.Default())
func WithLogger(l Logger) Option {
	return func(c *clientConfig) {
		if l == nil {
			c.logger = noopLogger{}
		} else {
			c.logger = l
		}
	}
}

// WithUserAgent appends a custom string to the default User-Agent header, e.g.
// "MyApp/2.0". The full header becomes "zaple-go/VERSION MyApp/2.0".
func WithUserAgent(ua string) Option {
	return func(c *clientConfig) {
		if ua != "" {
			c.userAgent = defaultUserAgent + " " + ua
		}
	}
}
