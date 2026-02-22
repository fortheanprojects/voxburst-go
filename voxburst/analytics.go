package voxburst

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// AnalyticsService handles analytics-related API calls.
type AnalyticsService struct {
	client *Client
}

// PostMetricsParams are the parameters for getting post metrics.
type PostMetricsParams struct {
	StartDate   *time.Time
	EndDate     *time.Time
	Granularity Granularity
}

// AccountMetricsParams are the parameters for getting account metrics.
type AccountMetricsParams struct {
	StartDate   *time.Time
	EndDate     *time.Time
	Granularity Granularity
}

// AggregateParams are the parameters for getting aggregate metrics.
type AggregateParams struct {
	StartDate  *time.Time
	EndDate    *time.Time
	AccountIDs []string
	Platforms  []Platform
	Limit      int
}

// GetPostMetrics retrieves analytics for a specific post.
func (s *AnalyticsService) GetPostMetrics(ctx context.Context, postID string, params *PostMetricsParams, opts ...RequestOption) ([]PostAnalytics, error) {
	queryParams := make(map[string]string)
	if params != nil {
		if params.StartDate != nil {
			queryParams["startDate"] = params.StartDate.Format(time.RFC3339)
		}
		if params.EndDate != nil {
			queryParams["endDate"] = params.EndDate.Format(time.RFC3339)
		}
		if params.Granularity != "" {
			queryParams["granularity"] = string(params.Granularity)
		}
	}

	path := fmt.Sprintf("/analytics/posts/%s%s", postID, buildQueryString(queryParams))

	var result struct {
		Data []PostAnalytics `json:"data"`
	}
	if err := s.client.get(ctx, path, &result, opts...); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// GetAccountMetrics retrieves analytics for a specific account.
func (s *AnalyticsService) GetAccountMetrics(ctx context.Context, accountID string, params *AccountMetricsParams, opts ...RequestOption) (*AccountAnalytics, error) {
	queryParams := make(map[string]string)
	if params != nil {
		if params.StartDate != nil {
			queryParams["startDate"] = params.StartDate.Format(time.RFC3339)
		}
		if params.EndDate != nil {
			queryParams["endDate"] = params.EndDate.Format(time.RFC3339)
		}
		if params.Granularity != "" {
			queryParams["granularity"] = string(params.Granularity)
		}
	}

	path := fmt.Sprintf("/analytics/accounts/%s%s", accountID, buildQueryString(queryParams))

	var result AccountAnalytics
	if err := s.client.get(ctx, path, &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetAggregate retrieves aggregate metrics across posts/accounts.
func (s *AnalyticsService) GetAggregate(ctx context.Context, params *AggregateParams, opts ...RequestOption) (*AggregateAnalytics, error) {
	queryParams := make(map[string]string)
	if params != nil {
		if params.StartDate != nil {
			queryParams["startDate"] = params.StartDate.Format(time.RFC3339)
		}
		if params.EndDate != nil {
			queryParams["endDate"] = params.EndDate.Format(time.RFC3339)
		}
		if len(params.AccountIDs) > 0 {
			queryParams["accountIds"] = strings.Join(params.AccountIDs, ",")
		}
		if len(params.Platforms) > 0 {
			platforms := make([]string, len(params.Platforms))
			for i, p := range params.Platforms {
				platforms[i] = string(p)
			}
			queryParams["platforms"] = strings.Join(platforms, ",")
		}
		if params.Limit > 0 {
			queryParams["limit"] = strconv.Itoa(params.Limit)
		}
	}

	path := "/analytics/aggregate" + buildQueryString(queryParams)

	var result AggregateAnalytics
	if err := s.client.get(ctx, path, &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetOverview retrieves a high-level analytics summary.
func (s *AnalyticsService) GetOverview(ctx context.Context, opts ...RequestOption) (*AnalyticsOverview, error) {
	var result AnalyticsOverview
	if err := s.client.get(ctx, "/analytics/overview", &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}

// RefreshPostMetrics force-refreshes metrics for a post.
func (s *AnalyticsService) RefreshPostMetrics(ctx context.Context, postID string, opts ...RequestOption) ([]PostAnalytics, error) {
	path := fmt.Sprintf("/analytics/posts/%s/refresh", postID)

	var result struct {
		Data []PostAnalytics `json:"data"`
	}
	if err := s.client.post(ctx, path, nil, &result, opts...); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// RefreshAccountMetrics force-refreshes metrics for an account.
func (s *AnalyticsService) RefreshAccountMetrics(ctx context.Context, accountID string, opts ...RequestOption) (*AccountAnalytics, error) {
	path := fmt.Sprintf("/analytics/accounts/%s/refresh", accountID)

	var result AccountAnalytics
	if err := s.client.post(ctx, path, nil, &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetEngagementRate calculates the engagement rate for a post.
func GetEngagementRate(metrics AnalyticsMetrics) float64 {
	if metrics.Impressions == 0 {
		return 0
	}
	return float64(metrics.Engagements) / float64(metrics.Impressions) * 100
}

// SumMetrics aggregates multiple metrics into a single total.
func SumMetrics(metrics ...AnalyticsMetrics) AnalyticsMetrics {
	var total AnalyticsMetrics
	for _, m := range metrics {
		total.Impressions += m.Impressions
		total.Reach += m.Reach
		total.Engagements += m.Engagements
		total.Likes += m.Likes
		total.Comments += m.Comments
		total.Shares += m.Shares
		total.Clicks += m.Clicks
		total.Saves += m.Saves
	}
	return total
}
