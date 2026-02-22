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
	Platforms         []Platform       `json:"platforms"`
	ScheduledFor      *time.Time       `json:"scheduledFor,omitempty"`
	Media             []string         `json:"media,omitempty"`
	PlatformOverrides map[Platform]any `json:"platformOverrides,omitempty"`
	Metadata          map[string]any   `json:"metadata,omitempty"`
}

// UpdatePostParams are the parameters for updating a post.
type UpdatePostParams struct {
	Content           *string          `json:"content,omitempty"`
	Platforms         []Platform       `json:"platforms,omitempty"`
	ScheduledFor      *time.Time       `json:"scheduledFor,omitempty"`
	Media             []string         `json:"media,omitempty"`
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

// CreateDraft is a convenience method to create a draft post (no scheduled time).
func (s *PostsService) CreateDraft(ctx context.Context, content string, platforms []Platform, opts ...RequestOption) (*Post, error) {
	return s.Create(ctx, CreatePostParams{
		Content:   content,
		Platforms: platforms,
	}, opts...)
}

// Schedule is a convenience method to create a scheduled post.
func (s *PostsService) Schedule(ctx context.Context, content string, platforms []Platform, scheduledFor time.Time, opts ...RequestOption) (*Post, error) {
	return s.Create(ctx, CreatePostParams{
		Content:      content,
		Platforms:    platforms,
		ScheduledFor: &scheduledFor,
	}, opts...)
}
