package voxburst

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// WebhooksService handles webhook-related API calls.
type WebhooksService struct {
	client *Client
}

// CreateWebhookParams are the parameters for creating a webhook.
type CreateWebhookParams struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
}

// UpdateWebhookParams are the parameters for updating a webhook.
type UpdateWebhookParams struct {
	URL     *string  `json:"url,omitempty"`
	Events  []string `json:"events,omitempty"`
	Enabled *bool    `json:"enabled,omitempty"`
}

// List lists all webhooks.
func (s *WebhooksService) List(ctx context.Context, opts ...RequestOption) (*ListResponse[Webhook], error) {
	var result ListResponse[Webhook]
	if err := s.client.get(ctx, "/webhooks", &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get retrieves a webhook by ID.
func (s *WebhooksService) Get(ctx context.Context, id string, opts ...RequestOption) (*Webhook, error) {
	var webhook Webhook
	if err := s.client.get(ctx, "/webhooks/"+id, &webhook, opts...); err != nil {
		return nil, err
	}
	return &webhook, nil
}

// Create creates a new webhook.
// Returns the webhook with the secret (only available on create).
func (s *WebhooksService) Create(ctx context.Context, params CreateWebhookParams, opts ...RequestOption) (*Webhook, error) {
	var webhook Webhook
	if err := s.client.post(ctx, "/webhooks", params, &webhook, opts...); err != nil {
		return nil, err
	}
	return &webhook, nil
}

// Update updates a webhook.
func (s *WebhooksService) Update(ctx context.Context, id string, params UpdateWebhookParams, opts ...RequestOption) (*Webhook, error) {
	var webhook Webhook
	if err := s.client.patch(ctx, "/webhooks/"+id, params, &webhook, opts...); err != nil {
		return nil, err
	}
	return &webhook, nil
}

// Delete deletes a webhook.
func (s *WebhooksService) Delete(ctx context.Context, id string, opts ...RequestOption) error {
	return s.client.delete(ctx, "/webhooks/"+id, nil, opts...)
}

// Enable enables a webhook.
func (s *WebhooksService) Enable(ctx context.Context, id string, opts ...RequestOption) (*Webhook, error) {
	enabled := true
	return s.Update(ctx, id, UpdateWebhookParams{Enabled: &enabled}, opts...)
}

// Disable disables a webhook.
func (s *WebhooksService) Disable(ctx context.Context, id string, opts ...RequestOption) (*Webhook, error) {
	enabled := false
	return s.Update(ctx, id, UpdateWebhookParams{Enabled: &enabled}, opts...)
}

// WebhookTestDelivery is the result of a single test delivery attempt sent by
// WebhooksService.Test.
type WebhookTestDelivery struct {
	ID           string     `json:"id"`
	WebhookID    string     `json:"webhookId"`
	EventType    string     `json:"eventType"`
	URL          string     `json:"url"`
	StatusCode   *int       `json:"statusCode"`
	LatencyMs    *int       `json:"latencyMs"`
	ResponseBody *string    `json:"responseBody"`
	RetryCount   int        `json:"retryCount"`
	Success      bool       `json:"success"`
	Error        *string    `json:"error"`
	CreatedAt    time.Time  `json:"createdAt"`
	DeliveredAt  *time.Time `json:"deliveredAt"`
}

// Test sends a test payload to a registered, enabled webhook and returns the
// delivery result. Unlike Test failures on other resources, a delivery that
// the receiving endpoint rejects is still a successful API call — check
// Success and StatusCode on the returned delivery, not just the error return.
func (s *WebhooksService) Test(ctx context.Context, webhookID string, opts ...RequestOption) (*WebhookTestDelivery, error) {
	var result struct {
		Data WebhookTestDelivery `json:"data"`
	}
	body := map[string]string{"webhookId": webhookID}
	if err := s.client.post(ctx, "/webhooks/test", body, &result, opts...); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// VerifySignature verifies the webhook signature.
// payload is the raw request body, signature is the X-SocialDispatch-Signature header value,
// and secret is the webhook secret returned when the webhook was created.
func VerifySignature(payload []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSig := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expectedSig), []byte(signature))
}

// ParseSignature extracts the hash from a signature header.
func ParseSignature(header string) (string, error) {
	if !strings.HasPrefix(header, "sha256=") {
		return "", fmt.Errorf("invalid signature format: expected sha256=...")
	}
	return strings.TrimPrefix(header, "sha256="), nil
}

// WebhookPayload represents a parsed webhook payload.
type WebhookPayload struct {
	Event     string         `json:"event"`
	Timestamp string         `json:"timestamp"`
	Data      map[string]any `json:"data"`
}

// PostPublishedData represents the data for a post.published event.
type PostPublishedData struct {
	PostID          string `json:"postId"`
	Platform        string `json:"platform"`
	PlatformPostID  string `json:"platformPostId"`
	PlatformPostURL string `json:"platformPostUrl"`
}

// PostFailedData represents the data for a post.failed event.
type PostFailedData struct {
	PostID   string `json:"postId"`
	Platform string `json:"platform"`
	Error    struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// PostScheduledData represents the data for a post.scheduled event.
type PostScheduledData struct {
	PostID       string   `json:"postId"`
	ScheduledFor string   `json:"scheduledFor"`
	Platforms    []string `json:"platforms"`
}

// AccountConnectedData represents the data for an account.connected event.
type AccountConnectedData struct {
	AccountID string `json:"accountId"`
	Platform  string `json:"platform"`
	Username  string `json:"username"`
}

// AccountDisconnectedData represents the data for an account.disconnected event.
type AccountDisconnectedData struct {
	AccountID string `json:"accountId"`
	Platform  string `json:"platform"`
	Reason    string `json:"reason"`
}

// AccountErrorData represents the data for an account.error event.
type AccountErrorData struct {
	AccountID string `json:"accountId"`
	Platform  string `json:"platform"`
	Error     struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
