// Package voxburst provides a Go client for the VoxBurst (SocialDispatch) API.
package voxburst

import "time"

// Platform represents a social media platform.
type Platform string

const (
	PlatformTwitter   Platform = "twitter"
	PlatformLinkedIn  Platform = "linkedin"
	PlatformInstagram Platform = "instagram"
	PlatformFacebook  Platform = "facebook"
	PlatformBluesky   Platform = "bluesky"
)

// PostStatus represents the status of a post.
type PostStatus string

const (
	PostStatusDraft      PostStatus = "draft"
	PostStatusScheduled  PostStatus = "scheduled"
	PostStatusPublishing PostStatus = "publishing"
	PostStatusPublished  PostStatus = "published"
	PostStatusFailed     PostStatus = "failed"
)

// MediaStatus represents the processing status of uploaded media.
type MediaStatus string

const (
	MediaStatusPending    MediaStatus = "pending"
	MediaStatusProcessing MediaStatus = "processing"
	MediaStatusReady      MediaStatus = "ready"
	MediaStatusFailed     MediaStatus = "failed"
)

// AccountStatus represents the connection status of a social account.
type AccountStatus string

const (
	AccountStatusActive       AccountStatus = "active"
	AccountStatusDisconnected AccountStatus = "disconnected"
	AccountStatusError        AccountStatus = "error"
)

// AccountType represents the type of social account.
type AccountType string

const (
	AccountTypePersonal AccountType = "personal"
	AccountTypeBusiness AccountType = "business"
	AccountTypeCreator  AccountType = "creator"
)

// Granularity represents time granularity for analytics.
type Granularity string

const (
	GranularityHourly  Granularity = "HOURLY"
	GranularityDaily   Granularity = "DAILY"
	GranularityWeekly  Granularity = "WEEKLY"
	GranularityMonthly Granularity = "MONTHLY"
)

// Post represents a social media post.
type Post struct {
	ID           string            `json:"id"`
	Content      string            `json:"content"`
	Status       PostStatus        `json:"status"`
	ScheduledFor *time.Time        `json:"scheduledFor,omitempty"`
	PublishedAt  *time.Time        `json:"publishedAt,omitempty"`
	Platforms    []PostPlatform    `json:"platforms"`
	Media        []MediaAttachment `json:"media,omitempty"`
	Metadata     map[string]any    `json:"metadata,omitempty"`
	CreatedAt    time.Time         `json:"createdAt"`
	UpdatedAt    time.Time         `json:"updatedAt,omitempty"`
}

// PostPlatform represents a post's status on a specific platform.
type PostPlatform struct {
	Platform        Platform   `json:"platform"`
	AccountID       string     `json:"accountId"`
	AccountName     string     `json:"accountName"`
	Status          PostStatus `json:"status"`
	PlatformPostID  string     `json:"platformPostId,omitempty"`
	PlatformPostURL string     `json:"platformPostUrl,omitempty"`
	PublishedAt     *time.Time `json:"publishedAt,omitempty"`
	Error           *APIError  `json:"error,omitempty"`
}

// MediaAttachment represents media attached to a post.
type MediaAttachment struct {
	ID   string `json:"id"`
	URL  string `json:"url"`
	Type string `json:"type"`
}

// Media represents uploaded media.
type Media struct {
	ID          string            `json:"id"`
	Filename    string            `json:"filename"`
	ContentType string            `json:"contentType"`
	SizeBytes   int64             `json:"sizeBytes"`
	Status      MediaStatus       `json:"status"`
	URL         string            `json:"url,omitempty"`
	Variants    map[string]string `json:"variants,omitempty"`
	Dimensions  *MediaDimensions  `json:"dimensions,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
}

// MediaDimensions represents the dimensions of media.
type MediaDimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Account represents a connected social media account.
type Account struct {
	ID          string        `json:"id"`
	Platform    Platform      `json:"platform"`
	Username    string        `json:"username"`
	DisplayName string        `json:"displayName"`
	AvatarURL   string        `json:"avatarUrl,omitempty"`
	AccountType AccountType   `json:"accountType"`
	Status      AccountStatus `json:"status"`
	ConnectedAt time.Time     `json:"connectedAt"`
}

// Webhook represents a webhook subscription.
type Webhook struct {
	ID             string     `json:"id"`
	URL            string     `json:"url"`
	Secret         string     `json:"secret,omitempty"` // Only returned on create
	Events         []string   `json:"events"`
	Enabled        bool       `json:"enabled"`
	LastDeliveryAt *time.Time `json:"lastDeliveryAt,omitempty"`
	FailureCount   int        `json:"failureCount"`
	CreatedAt      time.Time  `json:"createdAt"`
}

// WebhookEvent represents the types of webhook events.
type WebhookEvent string

const (
	WebhookEventPostPublished       WebhookEvent = "post.published"
	WebhookEventPostFailed          WebhookEvent = "post.failed"
	WebhookEventPostScheduled       WebhookEvent = "post.scheduled"
	WebhookEventAccountConnected    WebhookEvent = "account.connected"
	WebhookEventAccountDisconnected WebhookEvent = "account.disconnected"
	WebhookEventAccountError        WebhookEvent = "account.error"
)

// PlatformInfo represents information about a social platform.
type PlatformInfo struct {
	Platform     Platform             `json:"platform"`
	Name         string               `json:"name"`
	Capabilities PlatformCapabilities `json:"capabilities"`
	RateLimit    PlatformRateLimit    `json:"rateLimit"`
}

// PlatformCapabilities describes what a platform supports.
type PlatformCapabilities struct {
	MaxTextLength    int      `json:"maxTextLength"`
	MaxImages        int      `json:"maxImages"`
	MaxVideoLength   int      `json:"maxVideoLength"`
	SupportsThreads  bool     `json:"supportsThreads"`
	SupportsCarousel bool     `json:"supportsCarousels"`
	SupportsStories  bool     `json:"supportsStories"`
	MediaTypes       []string `json:"mediaTypes"`
}

// PlatformRateLimit describes a platform's rate limits.
type PlatformRateLimit struct {
	PostsPerDay  int `json:"postsPerDay"`
	PostsPerHour int `json:"postsPerHour"`
}

// Pagination holds pagination information for list responses.
type Pagination struct {
	HasMore    bool   `json:"hasMore"`
	NextCursor string `json:"nextCursor,omitempty"`
}

// ListResponse is a generic response for list endpoints.
type ListResponse[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// AnalyticsMetrics represents engagement metrics.
type AnalyticsMetrics struct {
	Impressions int `json:"impressions"`
	Reach       int `json:"reach"`
	Engagements int `json:"engagements"`
	Likes       int `json:"likes"`
	Comments    int `json:"comments"`
	Shares      int `json:"shares"`
	Clicks      int `json:"clicks"`
	Saves       int `json:"saves"`
}

// PostAnalytics represents analytics for a specific post.
type PostAnalytics struct {
	PostID          string           `json:"postId"`
	PostPlatformID  string           `json:"postPlatformId"`
	Platform        Platform         `json:"platform"`
	PlatformPostID  string           `json:"platformPostId"`
	PlatformPostURL string           `json:"platformPostUrl"`
	Metrics         AnalyticsMetrics `json:"metrics"`
	CollectedAt     time.Time        `json:"collectedAt"`
	PeriodStart     time.Time        `json:"periodStart"`
	PeriodEnd       time.Time        `json:"periodEnd"`
	Granularity     Granularity      `json:"granularity"`
}

// AccountAnalytics represents analytics for an account.
type AccountAnalytics struct {
	AccountID   string   `json:"accountId"`
	Platform    Platform `json:"platform"`
	Username    string   `json:"username"`
	DisplayName string   `json:"displayName"`
	Metrics     struct {
		Followers        int     `json:"followers"`
		Following        int     `json:"following"`
		TotalImpressions int     `json:"totalImpressions"`
		TotalReach       int     `json:"totalReach"`
		TotalEngagements int     `json:"totalEngagements"`
		TotalPosts       int     `json:"totalPosts"`
		EngagementRate   float64 `json:"engagementRate"`
	} `json:"metrics"`
	CollectedAt time.Time   `json:"collectedAt"`
	PeriodStart time.Time   `json:"periodStart"`
	PeriodEnd   time.Time   `json:"periodEnd"`
	Granularity Granularity `json:"granularity"`
}

// AggregateAnalytics represents aggregated analytics across posts/accounts.
type AggregateAnalytics struct {
	TotalPosts            int                 `json:"totalPosts"`
	TotalImpressions      int                 `json:"totalImpressions"`
	TotalReach            int                 `json:"totalReach"`
	TotalEngagements      int                 `json:"totalEngagements"`
	TotalLikes            int                 `json:"totalLikes"`
	TotalComments         int                 `json:"totalComments"`
	TotalShares           int                 `json:"totalShares"`
	TotalClicks           int                 `json:"totalClicks"`
	TotalSaves            int                 `json:"totalSaves"`
	AverageEngagementRate float64             `json:"averageEngagementRate"`
	TopPerformingPosts    []TopPost           `json:"topPerformingPosts"`
	ByPlatform            []PlatformAggregate `json:"byPlatform"`
	PeriodStart           time.Time           `json:"periodStart"`
	PeriodEnd             time.Time           `json:"periodEnd"`
}

// TopPost represents a top-performing post in analytics.
type TopPost struct {
	PostID      string   `json:"postId"`
	Platform    Platform `json:"platform"`
	Engagements int      `json:"engagements"`
	Impressions int      `json:"impressions"`
}

// PlatformAggregate represents aggregated metrics for a platform.
type PlatformAggregate struct {
	Platform    Platform `json:"platform"`
	Posts       int      `json:"posts"`
	Impressions int      `json:"impressions"`
	Engagements int      `json:"engagements"`
}

// AnalyticsOverview represents a high-level analytics summary.
type AnalyticsOverview struct {
	Summary struct {
		TotalPosts            int     `json:"totalPosts"`
		TotalImpressions      int     `json:"totalImpressions"`
		TotalEngagements      int     `json:"totalEngagements"`
		AverageEngagementRate float64 `json:"averageEngagementRate"`
	} `json:"summary"`
	TopPosts   []TopPost           `json:"topPosts"`
	ByPlatform []PlatformAggregate `json:"byPlatform"`
	Period     struct {
		Start time.Time `json:"start"`
		End   time.Time `json:"end"`
	} `json:"period"`
}

// UploadURL represents a presigned upload URL response.
type UploadURL struct {
	MediaID       string            `json:"mediaId"`
	UploadURL     string            `json:"uploadUrl"`
	UploadHeaders map[string]string `json:"uploadHeaders"`
	ExpiresAt     time.Time         `json:"expiresAt"`
}

// OAuthResponse represents the response from initiating OAuth.
type OAuthResponse struct {
	AuthURL string `json:"authUrl"`
	State   string `json:"state"`
}

// OAuthCallbackResponse represents the response from completing OAuth.
type OAuthCallbackResponse struct {
	Account Account `json:"account"`
}
