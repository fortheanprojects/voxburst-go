package voxburst

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	// Version is the SDK version, reported in the default User-Agent header.
	Version = "1.1.0"

	// DefaultBaseURL is the production API base URL.
	DefaultBaseURL = "https://api.voxburst.io/v1"

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
		userAgent:    "voxburst-go/" + Version,
		debug:        false,
	}
}

// WithBaseURL sets a custom base URL.
//
// The URL must use the https scheme: the client sends your API key as a
// Bearer token on every request, and a plaintext transport would expose it.
// Passing a non-https URL panics, consistent with the other options in this
// package (which cannot return errors). Use an https test server (for example
// httptest.NewTLSServer together with WithHTTPClient) when testing locally.
func WithBaseURL(url string) Option {
	if !strings.HasPrefix(url, "https://") {
		panic(fmt.Sprintf("voxburst: WithBaseURL must use HTTPS to protect your API key, got %q", url))
	}
	return func(c *clientConfig) {
		c.baseURL = url
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
