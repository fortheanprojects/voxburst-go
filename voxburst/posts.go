package voxburst

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

// PostsService handles post-related API calls.
type PostsService struct {
	client *Client
}

// CreatePostParams are the parameters for creating a post.
type CreatePostParams struct {
	Content           string           `json:"content"`
	AccountIds        []string         `json:"accountIds"`
	ScheduledFor      *time.Time       `json:"scheduledFor,omitempty"`
	ContentType       string           `json:"contentType,omitempty"`
	SaveAsDraft       bool             `json:"saveAsDraft,omitempty"`
	Media             []string         `json:"media,omitempty"`
	FirstComment      string           `json:"firstComment,omitempty"`
	FirstCommentDelay int              `json:"firstCommentDelay,omitempty"`
	PlatformOverrides map[Platform]any `json:"platformOverrides,omitempty"`
	Metadata          map[string]any   `json:"metadata,omitempty"`
}

// UpdatePostParams are the parameters for updating a post.
type UpdatePostParams struct {
	Content           *string          `json:"content,omitempty"`
	AccountIds        []string         `json:"accountIds,omitempty"`
	ScheduledFor      *time.Time       `json:"scheduledFor,omitempty"`
	ContentType       string           `json:"contentType,omitempty"`
	Media             []string         `json:"media,omitempty"`
	FirstComment      *string          `json:"firstComment,omitempty"`
	FirstCommentDelay *int             `json:"firstCommentDelay,omitempty"`
	PlatformOverrides map[Platform]any `json:"platformOverrides,omitempty"`
	Metadata          map[string]any   `json:"metadata,omitempty"`
}

// ListPostsParams are the parameters for listing posts.
type ListPostsParams struct {
	Status   PostStatus
	Platform Platform
	From     *time.Time
	To       *time.Time
	Limit    int
	Cursor   string
}

// Create creates a new post.
func (s *PostsService) Create(ctx context.Context, params CreatePostParams, opts ...RequestOption) (*Post, error) {
	var post Post
	if err := s.client.post(ctx, "/posts", params, &post, opts...); err != nil {
		return nil, err
	}
	return &post, nil
}

// Get retrieves a post by ID.
func (s *PostsService) Get(ctx context.Context, id string, opts ...RequestOption) (*Post, error) {
	var post Post
	if err := s.client.get(ctx, "/posts/"+id, &post, opts...); err != nil {
		return nil, err
	}
	return &post, nil
}

// Update updates a post.
func (s *PostsService) Update(ctx context.Context, id string, params UpdatePostParams, opts ...RequestOption) (*Post, error) {
	var post Post
	if err := s.client.patch(ctx, "/posts/"+id, params, &post, opts...); err != nil {
		return nil, err
	}
	return &post, nil
}

// Delete deletes a post.
func (s *PostsService) Delete(ctx context.Context, id string, opts ...RequestOption) error {
	return s.client.delete(ctx, "/posts/"+id, nil, opts...)
}

// List lists posts with optional filtering.
func (s *PostsService) List(ctx context.Context, params *ListPostsParams, opts ...RequestOption) (*ListResponse[Post], error) {
	queryParams := make(map[string]string)
	if params != nil {
		if params.Status != "" {
			queryParams["status"] = string(params.Status)
		}
		if params.Platform != "" {
			queryParams["platform"] = string(params.Platform)
		}
		if params.From != nil {
			queryParams["from"] = params.From.Format(time.RFC3339)
		}
		if params.To != nil {
			queryParams["to"] = params.To.Format(time.RFC3339)
		}
		if params.Limit > 0 {
			queryParams["limit"] = strconv.Itoa(params.Limit)
		}
		if params.Cursor != "" {
			queryParams["cursor"] = params.Cursor
		}
	}

	path := "/posts" + buildQueryString(queryParams)
	var result ListResponse[Post]
	if err := s.client.get(ctx, path, &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}

// Publish immediately publishes a draft or scheduled post.
func (s *PostsService) Publish(ctx context.Context, id string, opts ...RequestOption) (*Post, error) {
	var post Post
	if err := s.client.post(ctx, fmt.Sprintf("/posts/%s/publish", id), nil, &post, opts...); err != nil {
		return nil, err
	}
	return &post, nil
}

// ListAll returns an iterator for all posts matching the given parameters.
// This handles pagination automatically.
func (s *PostsService) ListAll(ctx context.Context, params *ListPostsParams) *PostsIterator {
	if params == nil {
		params = &ListPostsParams{}
	}
	return &PostsIterator{
		service: s,
		ctx:     ctx,
		params:  *params,
	}
}

// PostsIterator provides an iterator for paginated post results.
type PostsIterator struct {
	service *PostsService
	ctx     context.Context
	params  ListPostsParams

	posts []Post
	index int
	done  bool
	err   error
}

// Next returns the next post. Returns false when there are no more posts.
func (it *PostsIterator) Next() bool {
	if it.err != nil || it.done {
		return false
	}

	// If we have posts remaining in the current page, return true
	if it.index < len(it.posts) {
		return true
	}

	// Fetch the next page
	resp, err := it.service.List(it.ctx, &it.params)
	if err != nil {
		it.err = err
		return false
	}

	it.posts = resp.Data
	it.index = 0

	if len(it.posts) == 0 {
		it.done = true
		return false
	}

	// Update cursor for next fetch
	if resp.Pagination.HasMore && resp.Pagination.NextCursor != "" {
		it.params.Cursor = resp.Pagination.NextCursor
	} else {
		it.done = true
	}

	return true
}

// Post returns the current post.
func (it *PostsIterator) Post() Post {
	if it.index < len(it.posts) {
		post := it.posts[it.index]
		it.index++
		return post
	}
	return Post{}
}

// Err returns any error that occurred during iteration.
func (it *PostsIterator) Err() error {
	return it.err
}

// CreateDraft is a convenience method to create a draft post targeting specific accounts.
func (s *PostsService) CreateDraft(ctx context.Context, content string, accountIds []string, opts ...RequestOption) (*Post, error) {
	return s.Create(ctx, CreatePostParams{
		Content:    content,
		AccountIds: accountIds,
	}, opts...)
}

// Schedule is a convenience method to create a scheduled post targeting specific accounts.
func (s *PostsService) Schedule(ctx context.Context, content string, accountIds []string, scheduledFor time.Time, opts ...RequestOption) (*Post, error) {
	return s.Create(ctx, CreatePostParams{
		Content:      content,
		AccountIds:   accountIds,
		ScheduledFor: &scheduledFor,
	}, opts...)
}

// RetriedPlatform describes a platform target that was re-queued by Retry.
type RetriedPlatform struct {
	Platform    Platform  `json:"platform"`
	AccountID   string    `json:"accountId"`
	RetryCount  int       `json:"retryCount"`
	NextRetryAt time.Time `json:"nextRetryAt"`
}

// SkippedPlatform describes a platform target that Retry could not re-queue
// because it already hit the maximum retry count.
type SkippedPlatform struct {
	Platform   Platform `json:"platform"`
	AccountID  string   `json:"accountId"`
	RetryCount int      `json:"retryCount"`
	Reason     string   `json:"reason"`
}

// RetryResult is the response from PostsService.Retry.
type RetryResult struct {
	ID               string            `json:"id"`
	Status           string            `json:"status"`
	RetriedPlatforms []RetriedPlatform `json:"retriedPlatforms"`
	SkippedPlatforms []SkippedPlatform `json:"skippedPlatforms"`
}

// Retry retries a FAILED or PARTIAL post. Only platform targets that are
// FAILED, PENDING, or PUBLISHING (stuck) and have not exceeded the maximum
// retry count are re-queued; the rest are reported in SkippedPlatforms.
func (s *PostsService) Retry(ctx context.Context, id string, opts ...RequestOption) (*RetryResult, error) {
	var result RetryResult
	if err := s.client.post(ctx, fmt.Sprintf("/posts/%s/retry", id), nil, &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}

// Cancel cancels a DRAFT or SCHEDULED post without deleting it.
func (s *PostsService) Cancel(ctx context.Context, id string, opts ...RequestOption) (*Post, error) {
	var post Post
	if err := s.client.post(ctx, fmt.Sprintf("/posts/%s/cancel", id), nil, &post, opts...); err != nil {
		return nil, err
	}
	return &post, nil
}

// Clone duplicates an existing post as a new DRAFT with the same content,
// accounts, media, and metadata.
func (s *PostsService) Clone(ctx context.Context, id string, opts ...RequestOption) (*Post, error) {
	var post Post
	if err := s.client.post(ctx, fmt.Sprintf("/posts/%s/clone", id), nil, &post, opts...); err != nil {
		return nil, err
	}
	return &post, nil
}

// FixPlatformParams are the optional parameters for PostsService.FixPlatform.
// All fields are optional — an empty FixPlatformParams{} just resets the
// platform to PENDING and re-queues it with its existing content and media.
type FixPlatformParams struct {
	// MediaURL, if set, must be an https URL and replaces the platform's media
	// for the retry.
	MediaURL string `json:"mediaUrl,omitempty"`
	// Content, if set, overrides the post content for this platform only.
	Content string `json:"content,omitempty"`
	// RefreshMedia re-encodes the post's existing first image as a baseline
	// JPEG and republishes it. Image media only — do not set alongside MediaURL.
	RefreshMedia bool `json:"refreshMedia,omitempty"`
}

// FixPlatformResult is the response from PostsService.FixPlatform.
type FixPlatformResult struct {
	Success    bool   `json:"success"`
	PlatformID string `json:"platformId"`
	Status     string `json:"status"`
}

// FixPlatform fixes a single failed or stuck platform on a PARTIAL or FAILED
// post — optionally replacing its media or content — and re-queues just that
// platform for publishing without touching the others. platformID is the
// PostPlatform row ID, e.g. from Post.Platforms[i] once that field carries it,
// or from the REST response of the original create/get call.
func (s *PostsService) FixPlatform(ctx context.Context, postID, platformID string, params FixPlatformParams, opts ...RequestOption) (*FixPlatformResult, error) {
	var result FixPlatformResult
	path := fmt.Sprintf("/posts/%s/platforms/%s/fix", postID, platformID)
	if err := s.client.post(ctx, path, params, &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}

// ValidateMediaItem describes a media attachment for a pre-flight validation check.
type ValidateMediaItem struct {
	Type       string `json:"type"` // "image", "video", or "gif"
	URL        string `json:"url,omitempty"`
	MimeType   string `json:"mimeType,omitempty"`
	SizeBytes  int64  `json:"sizeBytes,omitempty"`
	DurationMs int64  `json:"durationMs,omitempty"`
}

// ValidatePostParams are the parameters for PostsService.Validate.
type ValidatePostParams struct {
	Content          string                      `json:"content"`
	Platforms        []Platform                  `json:"platforms"`
	MediaIds         []string                    `json:"mediaIds,omitempty"`
	Media            []ValidateMediaItem         `json:"media,omitempty"`
	FirstComment     string                      `json:"firstComment,omitempty"`
	PlatformMetadata map[Platform]map[string]any `json:"platformMetadata,omitempty"`
}

// ValidationIssue is a single error or warning returned by PostsService.Validate.
type ValidationIssue struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// PlatformValidationResult is the per-platform result of PostsService.Validate.
type PlatformValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationIssue `json:"errors"`
	Warnings []ValidationIssue `json:"warnings"`
}

// ValidateResult is the response from PostsService.Validate.
type ValidateResult struct {
	Valid     bool                                  `json:"valid"`
	Platforms map[Platform]PlatformValidationResult `json:"platforms"`
}

// Validate runs a pre-flight content validation check — no post is created or
// saved as a draft. Returns per-platform errors (hard failures) and warnings
// (engagement hints). Useful for validating content before calling Create.
func (s *PostsService) Validate(ctx context.Context, params ValidatePostParams, opts ...RequestOption) (*ValidateResult, error) {
	var result ValidateResult
	if err := s.client.post(ctx, "/posts/validate", params, &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}
