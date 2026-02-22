package voxburst

import (
	"net/http"
	"time"
)

const (
	// DefaultBaseURL is the production API base URL.
	DefaultBaseURL = "https://api.socialdispatch.io/v1"

	// StagingBaseURL is the staging API base URL.
	StagingBaseURL = "https://api-staging.socialdispatch.io/v1"

	// DefaultTimeout is the default HTTP client timeout.
	DefaultTimeout = 30 * time.Second

	// DefaultMaxRetries is the default number of retries for failed requests.
	DefaultMaxRetries = 3

	// DefaultRetryWaitMin is the minimum wait time between retries.
	DefaultRetryWaitMin = 1 * time.Second

	// DefaultRetryWaitMax is the maximum wait time between retries.
	DefaultRetryWaitMax = 30 * time.Second
)

// Option is a functional option for configuring the client.
type Option func(*clientConfig)

// clientConfig holds the client configuration.
type clientConfig struct {
	baseURL      string
	httpClient   *http.Client
	timeout      time.Duration
	maxRetries   int
	retryWaitMin time.Duration
	retryWaitMax time.Duration
	userAgent    string
	debug        bool
}

// defaultConfig returns the default client configuration.
func defaultConfig() *clientConfig {
	return &clientConfig{
		baseURL:      DefaultBaseURL,
		timeout:      DefaultTimeout,
		maxRetries:   DefaultMaxRetries,
		retryWaitMin: DefaultRetryWaitMin,
		retryWaitMax: DefaultRetryWaitMax,
		userAgent:    "voxburst-go/1.0.0",
		debug:        false,
	}
}

// WithBaseURL sets a custom base URL.
func WithBaseURL(url string) Option {
	return func(c *clientConfig) {
		c.baseURL = url
	}
}

// WithStaging configures the client to use the staging environment.
func WithStaging() Option {
	return func(c *clientConfig) {
		c.baseURL = StagingBaseURL
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *clientConfig) {
		c.httpClient = client
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *clientConfig) {
		c.timeout = timeout
	}
}

// WithMaxRetries sets the maximum number of retries for failed requests.
func WithMaxRetries(n int) Option {
	return func(c *clientConfig) {
		c.maxRetries = n
	}
}

// WithRetryWait sets the retry wait time bounds.
func WithRetryWait(min, max time.Duration) Option {
	return func(c *clientConfig) {
		c.retryWaitMin = min
		c.retryWaitMax = max
	}
}

// WithUserAgent sets a custom User-Agent header.
func WithUserAgent(ua string) Option {
	return func(c *clientConfig) {
		c.userAgent = ua
	}
}

// WithDebug enables debug logging.
func WithDebug(enable bool) Option {
	return func(c *clientConfig) {
		c.debug = enable
	}
}

// WithNoRetry disables automatic retries.
func WithNoRetry() Option {
	return func(c *clientConfig) {
		c.maxRetries = 0
	}
}
