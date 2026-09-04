# VoxBurst Go SDK

Official Go client library for the [VoxBurst (VoxBurst)](https://voxburst.io) API — the unified API for social media scheduling and publishing.

[![Go Reference](https://pkg.go.dev/badge/github.com/fortheanprojects/voxburst-go.svg)](https://pkg.go.dev/github.com/fortheanprojects/voxburst-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/fortheanprojects/voxburst-go)](https://goreportcard.com/report/github.com/fortheanprojects/voxburst-go)

## Installation

```bash
go get github.com/fortheanprojects/voxburst-go
```

Requires Go 1.21 or later.

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/fortheanprojects/voxburst-go/voxburst"
)

func main() {
    // Create client with your API key
    client := voxburst.NewClient("sk_live_xxxxxxxxxxxxx")

    ctx := context.Background()

    // Create and schedule a post
    post, err := client.Posts.Create(ctx, voxburst.CreatePostParams{
        Content:      "Hello from Go! 🚀",
        AccountIds:   []string{"acc_twitter_123", "acc_linkedin_456"},
        ScheduledFor: ptr(time.Now().Add(24 * time.Hour)),
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Created post: %s\n", post.ID)
}

func ptr[T any](v T) *T { return &v }
```

## Features

- **Full API coverage** — Posts, Accounts, Media, Analytics, Webhooks, Batch
- **Automatic retries** — Exponential backoff with jitter for rate limits and server errors
- **Context support** — All methods accept `context.Context` for cancellation
- **Idempotency keys** — Prevent duplicate operations
- **Type-safe** — Strongly typed requests and responses
- **Pagination helpers** — Easy iteration over large result sets

## Configuration

### Client Options

```go
client := voxburst.NewClient(apiKey,
    // Override the API base URL (defaults to https://api.voxburst.io/v1).
    // Must be https:// — the option panics otherwise, to keep your API key
    // off plaintext transports.
    voxburst.WithBaseURL("https://custom.api.url/v1"),
    
    // Custom HTTP client
    voxburst.WithHTTPClient(&http.Client{
        Timeout: 60 * time.Second,
    }),
    
    // Request timeout
    voxburst.WithTimeout(30 * time.Second),
    
    // Retry configuration
    voxburst.WithMaxRetries(5),
    voxburst.WithRetryWait(1*time.Second, 30*time.Second),
    
    // Disable retries
    voxburst.WithNoRetry(),
    
    // Custom User-Agent
    voxburst.WithUserAgent("my-app/1.0.0"),
    
    // Enable debug logging
    voxburst.WithDebug(true),
)
```

### Per-Request Options

```go
// Idempotency key to prevent duplicate operations
post, err := client.Posts.Create(ctx, params,
    voxburst.WithIdempotencyKey(voxburst.GenerateIdempotencyKey()),
)

// Custom headers
post, err := client.Posts.Create(ctx, params,
    voxburst.WithHeader("X-Custom-Header", "value"),
)
```

## Posts

### Create a Post

```go
// Simple draft
post, _ := client.Posts.CreateDraft(ctx, 
    "Hello world!",
    []voxburst.Platform{voxburst.PlatformTwitter},
)

// Scheduled post
post, _ := client.Posts.Schedule(ctx,
    "Scheduled post!",
    []voxburst.Platform{voxburst.PlatformTwitter},
    time.Now().Add(24 * time.Hour),
)

// Full control
post, _ := client.Posts.Create(ctx, voxburst.CreatePostParams{
    Content:   "Cross-platform post!",
    AccountIds: []string{"acc_twitter_123", "acc_linkedin_456"},
    ScheduledFor: &scheduledTime,
    Media:     []string{"media_abc123"},
    PlatformOverrides: map[voxburst.Platform]any{
        voxburst.PlatformTwitter: map[string]any{
            "content": "Twitter-specific content! 🐦",
        },
    },
    Metadata: map[string]any{
        "campaign": "launch",
    },
})
```

### Get, Update, Delete

```go
// Get a post
post, _ := client.Posts.Get(ctx, "post_abc123")

// Update a post
newContent := "Updated content!"
post, _ := client.Posts.Update(ctx, "post_abc123", voxburst.UpdatePostParams{
    Content: &newContent,
})

// Delete a post
err := client.Posts.Delete(ctx, "post_abc123")

// Publish a draft immediately
post, _ := client.Posts.Publish(ctx, "post_abc123")
```

### List and Paginate

```go
// List with filters
posts, _ := client.Posts.List(ctx, &voxburst.ListPostsParams{
    Status:   voxburst.PostStatusScheduled,
    Platform: voxburst.PlatformTwitter,
    From:     &startDate,
    To:       &endDate,
    Limit:    50,
})

// Iterate through all posts (auto-pagination)
iter := client.Posts.ListAll(ctx, &voxburst.ListPostsParams{
    Limit: 100, // Fetch 100 at a time
})

for iter.Next() {
    post := iter.Post()
    fmt.Println(post.ID, post.Content)
}
if err := iter.Err(); err != nil {
    log.Fatal(err)
}
```

### Validate, retry, and fix

```go
// Pre-flight validation — no post is created
result, _ := client.Posts.Validate(ctx, voxburst.ValidatePostParams{
    Content:   "Check out our launch!",
    Platforms: []voxburst.Platform{voxburst.PlatformTwitter, voxburst.PlatformLinkedIn},
})
if !result.Valid {
    log.Printf("validation failed: %+v", result.Platforms)
}

// Cancel a DRAFT or SCHEDULED post
cancelled, _ := client.Posts.Cancel(ctx, "post_abc123")

// Retry a FAILED or PARTIAL post
retried, _ := client.Posts.Retry(ctx, "post_abc123")
fmt.Printf("retried %d, skipped %d\n", len(retried.RetriedPlatforms), len(retried.SkippedPlatforms))

// Fix and republish just one failed platform
fixed, _ := client.Posts.FixPlatform(ctx, "post_abc123", "postplatform_xyz", voxburst.FixPlatformParams{
    Content: "Replacement text for this platform only",
})

// Clone a post as a new draft
clone, _ := client.Posts.Clone(ctx, "post_abc123")
_ = cancelled
_ = clone
_ = fixed
```

## Accounts

```go
// List all connected accounts
accounts, _ := client.Accounts.List(ctx, nil)

// Filter by platform
twitterAccounts, _ := client.Accounts.ListByPlatform(ctx, voxburst.PlatformTwitter)

// Find by username
account, _ := client.Accounts.GetByUsername(ctx, voxburst.PlatformTwitter, "@username")

// Initiate OAuth connection
oauth, _ := client.Accounts.InitiateConnect(ctx, voxburst.PlatformTwitter, "https://yourapp.com/oauth/callback")
// Redirect user to oauth.AuthURL

// Complete OAuth (in your callback handler)
result, _ := client.Accounts.CompleteConnect(ctx, voxburst.PlatformTwitter, voxburst.CallbackParams{
    Code:  authCode,
    State: oauth.State,
})

// Verify a connected account's token still works
test, _ := client.Accounts.Test(ctx, "acc_abc123")
if !test.Success {
    log.Println(test.Message)
}

// Disconnect account
err := client.Accounts.Delete(ctx, "acc_abc123")

// Refresh token
err := client.Accounts.Refresh(ctx, "acc_abc123")
```

## Media

```go
// Upload from file
mediaID, _ := client.Media.UploadFile(ctx, "/path/to/image.jpg")

// Upload from bytes
mediaID, _ := client.Media.Upload(ctx, "photo.jpg", "image/jpeg", imageBytes)

// Upload from reader
mediaID, _ := client.Media.UploadReader(ctx, "video.mp4", "video/mp4", fileSize, reader)

// Get media details
media, _ := client.Media.Get(ctx, "media_abc123")

// List media
mediaList, _ := client.Media.List(ctx, &voxburst.ListMediaParams{
    Status: voxburst.MediaStatusReady,
    Limit:  20,
})

// Delete media
err := client.Media.Delete(ctx, "media_abc123")

// Use in a post
post, _ := client.Posts.Create(ctx, voxburst.CreatePostParams{
    Content:   "Check out this photo!",
    AccountIds: []string{"acc_twitter_123"},
    Media:     []string{mediaID},
})
```

## Analytics

```go
// Get post metrics
metrics, _ := client.Analytics.GetPostMetrics(ctx, "post_abc123", &voxburst.PostMetricsParams{
    StartDate:   &startDate,
    EndDate:     &endDate,
    Granularity: voxburst.GranularityDaily,
})

// Get account metrics
accountMetrics, _ := client.Analytics.GetAccountMetrics(ctx, "acc_abc123", nil)

// Get aggregate metrics
aggregate, _ := client.Analytics.GetAggregate(ctx, &voxburst.AggregateParams{
    AccountIds: []string{"acc_twitter_123", "acc_linkedin_456"},
    Limit:     10,
})

// Get overview
overview, _ := client.Analytics.GetOverview(ctx)

// Force refresh metrics
metrics, _ := client.Analytics.RefreshPostMetrics(ctx, "post_abc123")
```

## Webhooks

```go
// Create webhook
webhook, _ := client.Webhooks.Create(ctx, voxburst.CreateWebhookParams{
    URL:    "https://yourapp.com/webhooks",
    Events: []string{"post.published", "post.failed", "account.disconnected"},
})
// Save webhook.Secret securely!

// List webhooks
webhooks, _ := client.Webhooks.List(ctx)

// Update webhook
enabled := false
webhook, _ := client.Webhooks.Update(ctx, "wh_abc123", voxburst.UpdateWebhookParams{
    Enabled: &enabled,
})

// Enable/disable helpers
webhook, _ := client.Webhooks.Enable(ctx, "wh_abc123")
webhook, _ := client.Webhooks.Disable(ctx, "wh_abc123")

// Delete webhook
err := client.Webhooks.Delete(ctx, "wh_abc123")

// Send a test payload and inspect the delivery result
delivery, _ := client.Webhooks.Test(ctx, "wh_abc123")
fmt.Printf("test delivery: success=%v status=%v\n", delivery.Success, delivery.StatusCode)

// Verify webhook signature (in your webhook handler)
func handleWebhook(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)
    signature := r.Header.Get("X-VoxBurst-Signature")
    
    if !voxburst.VerifySignature(body, signature, webhookSecret) {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }
    
    // Process webhook...
}
```

## Batch

```go
// Execute up to 100 heterogeneous operations in one request
resp, _ := client.Batch.Execute(ctx, []voxburst.BatchOperation{
    {ID: "op1", Method: voxburst.BatchMethodPost, Path: "/v1/posts", Body: map[string]any{
        "content": "Batch post 1", "accountIds": []string{"acc_twitter_123"},
    }},
    {ID: "op2", Method: voxburst.BatchMethodDelete, Path: "/v1/posts/post_old123"},
})
fmt.Printf("%d succeeded, %d failed\n", resp.Summary.Succeeded, resp.Summary.Failed)

// Create up to 50 posts in a single request
bulk, _ := client.Batch.CreatePosts(ctx, []voxburst.BulkCreatePostItem{
    {Content: "Post A", AccountIds: []string{"acc_twitter_123"}},
    {Content: "Post B", AccountIds: []string{"acc_linkedin_456"}},
})
fmt.Printf("created %d of %d\n", bulk.Created, bulk.Total)
```

## Error Handling

```go
post, err := client.Posts.Get(ctx, "post_abc123")
if err != nil {
    // Check error type
    if voxburst.IsNotFound(err) {
        fmt.Println("Post not found")
    } else if voxburst.IsRateLimited(err) {
        fmt.Println("Rate limited, try again later")
    } else if voxburst.IsUnauthorized(err) {
        fmt.Println("Invalid API key")
    } else if voxburst.IsValidationError(err) {
        fmt.Println("Invalid request parameters")
    } else if voxburst.IsServerError(err) {
        fmt.Println("Server error, will be retried automatically")
    }
    
    // Get detailed error information
    if apiErr, ok := voxburst.GetAPIError(err); ok {
        fmt.Printf("Code: %s\n", apiErr.Code)
        fmt.Printf("Message: %s\n", apiErr.Message)
        fmt.Printf("Details: %v\n", apiErr.Details)
        fmt.Printf("HTTP Status: %d\n", apiErr.StatusCode)
    }
    
    return
}
```

## Supported Platforms

| Platform | Constant |
|----------|----------|
| Twitter/X | `voxburst.PlatformTwitter` |
| LinkedIn | `voxburst.PlatformLinkedIn` |
| Instagram | `voxburst.PlatformInstagram` |
| Facebook | `voxburst.PlatformFacebook` |
| Bluesky | `voxburst.PlatformBluesky` |

## API Reference

Full API documentation is available at:
- [VoxBurst API Docs](https://docs.voxburst.io)
- [Go Package Reference](https://pkg.go.dev/github.com/fortheanprojects/voxburst-go)

## Examples

See the [examples](./examples) directory for complete working examples:
- [Create Post](./examples/create_post/main.go) — Creating, scheduling, and managing posts
- [Connect Account](./examples/connect_account/main.go) — OAuth flow and account management

## Contributing

Contributions are welcome! Please read our [contributing guidelines](CONTRIBUTING.md) first.

## License

MIT License - see [LICENSE](LICENSE) for details.
