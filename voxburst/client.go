package voxburst

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// Client is the VoxBurst API client.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	config     *clientConfig

	// Services
	Posts     *PostsService
	Accounts  *AccountsService
	Analytics *AnalyticsService
	Webhooks  *WebhooksService
	Media     *MediaService
	Batch     *BatchService
}

// NewClient creates a new VoxBurst API client.
func NewClient(apiKey string, opts ...Option) *Client {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	httpClient := cfg.httpClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: cfg.timeout,
		}
	}

	c := &Client{
		apiKey:     apiKey,
		baseURL:    cfg.baseURL,
		httpClient: httpClient,
		config:     cfg,
	}

	// Initialize services
	c.Posts = &PostsService{client: c}
	c.Accounts = &AccountsService{client: c}
	c.Analytics = &AnalyticsService{client: c}
	c.Webhooks = &WebhooksService{client: c}
	c.Media = &MediaService{client: c}
	c.Batch = &BatchService{client: c}

	return c
}

// RequestOption is an option for a single request.
type RequestOption func(*requestConfig)

// requestConfig holds configuration for a single request.
type requestConfig struct {
	idempotencyKey string
	headers        map[string]string
}

// WithIdempotencyKey sets an idempotency key for the request.
func WithIdempotencyKey(key string) RequestOption {
	return func(rc *requestConfig) {
		rc.idempotencyKey = key
	}
}

// WithHeader adds a custom header to the request.
func WithHeader(key, value string) RequestOption {
	return func(rc *requestConfig) {
		if rc.headers == nil {
			rc.headers = make(map[string]string)
		}
		rc.headers[key] = value
	}
}

// GenerateIdempotencyKey generates a new UUID for use as an idempotency key.
func GenerateIdempotencyKey() string {
	return uuid.New().String()
}

// request makes an HTTP request with retry logic.
func (c *Client) request(ctx context.Context, method, path string, body any, result any, opts ...RequestOption) error {
	reqCfg := &requestConfig{}
	for _, opt := range opts {
		opt(reqCfg)
	}

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	fullURL := c.baseURL + path

	var lastErr error
	for attempt := 0; attempt <= c.config.maxRetries; attempt++ {
		if attempt > 0 {
			// Calculate backoff with jitter
			wait := c.calculateBackoff(attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}

			// Reset body reader for retry
			if body != nil {
				jsonBody, _ := json.Marshal(body)
				bodyReader = bytes.NewReader(jsonBody)
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", c.config.userAgent)

		if reqCfg.idempotencyKey != "" {
			req.Header.Set("Idempotency-Key", reqCfg.idempotencyKey)
		}

		for k, v := range reqCfg.headers {
			req.Header.Set(k, v)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			if IsRetryable(lastErr) {
				continue
			}
			return lastErr
		}

		defer resp.Body.Close()

		// Check for successful response
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if result != nil {
				if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
					return fmt.Errorf("failed to decode response: %w", err)
				}
			}
			return nil
		}

		// Parse error
		lastErr = parseAPIError(resp)

		// Retry on rate limit or server error
		if IsRetryable(lastErr) && attempt < c.config.maxRetries {
			// Check for Retry-After header
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if seconds, err := strconv.Atoi(retryAfter); err == nil {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(time.Duration(seconds) * time.Second):
					}
				}
			}
			continue
		}

		return lastErr
	}

	return lastErr
}

// calculateBackoff calculates the backoff duration with exponential backoff and jitter.
func (c *Client) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff: min * 2^attempt
	backoff := float64(c.config.retryWaitMin) * math.Pow(2, float64(attempt-1))

	// Add jitter (±25%)
	jitter := backoff * 0.25 * (rand.Float64()*2 - 1)
	backoff += jitter

	// Clamp to max
	if backoff > float64(c.config.retryWaitMax) {
		backoff = float64(c.config.retryWaitMax)
	}

	return time.Duration(backoff)
}

// get makes a GET request.
func (c *Client) get(ctx context.Context, path string, result any, opts ...RequestOption) error {
	return c.request(ctx, http.MethodGet, path, nil, result, opts...)
}

// post makes a POST request.
func (c *Client) post(ctx context.Context, path string, body, result any, opts ...RequestOption) error {
	return c.request(ctx, http.MethodPost, path, body, result, opts...)
}

// patch makes a PATCH request.
func (c *Client) patch(ctx context.Context, path string, body, result any, opts ...RequestOption) error {
	return c.request(ctx, http.MethodPatch, path, body, result, opts...)
}

// delete makes a DELETE request.
func (c *Client) delete(ctx context.Context, path string, result any, opts ...RequestOption) error {
	return c.request(ctx, http.MethodDelete, path, nil, result, opts...)
}

// buildQueryString builds a query string from parameters.
func buildQueryString(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}
	values := url.Values{}
	for k, v := range params {
		if v != "" {
			values.Set(k, v)
		}
	}
	if len(values) == 0 {
		return ""
	}
	return "?" + values.Encode()
}
