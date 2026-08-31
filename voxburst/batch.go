package voxburst

import (
	"context"
	"time"
)

// BatchService handles the batch API — executing multiple operations in a
// single request, and bulk post creation.
type BatchService struct {
	client *Client
}

// BatchMethod is an HTTP method allowed in a batch operation.
type BatchMethod string

const (
	BatchMethodPost   BatchMethod = "POST"
	BatchMethodPatch  BatchMethod = "PATCH"
	BatchMethodDelete BatchMethod = "DELETE"
)

// BatchOperation is a single operation within a batch request. Only POST,
// PATCH and DELETE are supported, against /v1/posts, /v1/accounts,
// /v1/media, and /v1/webhooks paths.
type BatchOperation struct {
	// ID correlates this operation to its result. Must be unique within the batch.
	ID      string            `json:"id"`
	Method  BatchMethod       `json:"method"`
	Path    string            `json:"path"`
	Body    map[string]any    `json:"body,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// BatchOperationError describes why a single batch operation failed.
type BatchOperationError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// BatchResult is the outcome of a single operation within a batch request.
type BatchResult struct {
	ID     string               `json:"id"`
	Status int                  `json:"status"`
	Body   map[string]any       `json:"body"`
	Error  *BatchOperationError `json:"error,omitempty"`
}

// BatchSummary totals the outcomes of a batch request.
type BatchSummary struct {
	Total     int `json:"total"`
	Succeeded int `json:"succeeded"`
	Failed    int `json:"failed"`
	Skipped   int `json:"skipped"`
}

// BatchResponse is the response from BatchService.Execute.
type BatchResponse struct {
	Results []BatchResult `json:"results"`
	Summary BatchSummary  `json:"summary"`
}

// ExecuteBatchParams are the parameters for BatchService.ExecuteWithOptions.
type ExecuteBatchParams struct {
	Operations []BatchOperation `json:"operations"`
	// ContinueOnError keeps executing after a failed operation (default true
	// via Execute). Set false for all-or-nothing semantics — the first
	// failure stops execution, and later operations are marked skipped.
	ContinueOnError bool `json:"continueOnError"`
	// Parallel executes operations concurrently instead of sequentially.
	// Faster, but gives no ordering guarantees between operations.
	Parallel bool `json:"parallel"`
}

// Execute runs up to 100 operations in a single request (POST /v1/batch),
// continuing past individual failures. Use ExecuteWithOptions for
// all-or-nothing or parallel execution.
func (s *BatchService) Execute(ctx context.Context, operations []BatchOperation, opts ...RequestOption) (*BatchResponse, error) {
	return s.ExecuteWithOptions(ctx, ExecuteBatchParams{
		Operations:      operations,
		ContinueOnError: true,
	}, opts...)
}

// ExecuteWithOptions runs a batch request with explicit continueOnError/parallel settings.
func (s *BatchService) ExecuteWithOptions(ctx context.Context, params ExecuteBatchParams, opts ...RequestOption) (*BatchResponse, error) {
	var result BatchResponse
	if err := s.client.post(ctx, "/batch", params, &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}

// BulkCreatePostItem is a single post within a CreatePosts request.
type BulkCreatePostItem struct {
	Content           string           `json:"content"`
	AccountIds        []string         `json:"accountIds"`
	ScheduledFor      *time.Time       `json:"scheduledFor,omitempty"`
	Media             []string         `json:"media,omitempty"`
	PlatformOverrides map[Platform]any `json:"platformOverrides,omitempty"`
	Metadata          map[string]any   `json:"metadata,omitempty"`
	FirstComment      string           `json:"firstComment,omitempty"`
	FirstCommentDelay int              `json:"firstCommentDelay,omitempty"`
}

// CreatePostsParams are the parameters for BatchService.CreatePostsWithOptions.
type CreatePostsParams struct {
	Posts []BulkCreatePostItem `json:"posts"`
	// DryRun validates every post (accounts exist, media is ready) without
	// creating anything.
	DryRun bool `json:"dryRun"`
}

// BulkCreatePostError describes why one post in a bulk create request failed.
type BulkCreatePostError struct {
	Field  string `json:"field"`
	Issue  string `json:"issue"`
	Detail string `json:"detail,omitempty"`
}

// BulkCreatePostResultItem is the per-post result of a bulk create request.
type BulkCreatePostResultItem struct {
	Index  int                   `json:"index"`
	Status string                `json:"status"` // "created", "failed", or "valid" (dry run)
	PostID string                `json:"postId,omitempty"`
	Errors []BulkCreatePostError `json:"errors,omitempty"`
}

// BulkCreatePostsResult is the response from BatchService.CreatePosts.
type BulkCreatePostsResult struct {
	DryRun  bool                       `json:"dry_run"`
	Total   int                        `json:"total"`
	Created int                        `json:"created"`
	Failed  int                        `json:"failed"`
	Results []BulkCreatePostResultItem `json:"results"`
}

// CreatePosts creates up to 50 posts in a single request (POST /v1/posts/bulk).
func (s *BatchService) CreatePosts(ctx context.Context, posts []BulkCreatePostItem, opts ...RequestOption) (*BulkCreatePostsResult, error) {
	return s.CreatePostsWithOptions(ctx, CreatePostsParams{Posts: posts}, opts...)
}

// CreatePostsWithOptions creates up to 50 posts in a single request, with
// dryRun support for validating before actually creating anything.
func (s *BatchService) CreatePostsWithOptions(ctx context.Context, params CreatePostsParams, opts ...RequestOption) (*BulkCreatePostsResult, error) {
	var result BulkCreatePostsResult
	if err := s.client.post(ctx, "/posts/bulk", params, &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}
