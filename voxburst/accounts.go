package voxburst

import (
	"context"
	"fmt"
	"strconv"
)

// AccountsService handles account-related API calls.
type AccountsService struct {
	client *Client
}

// ConnectParams are the parameters for initiating OAuth connection.
type ConnectParams struct {
	CallbackURL string `json:"callbackUrl"`
}

// CallbackParams are the parameters for completing OAuth.
type CallbackParams struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// ListAccountsParams are the parameters for listing accounts.
type ListAccountsParams struct {
	// Status filters by connection status: "ACTIVE", "INACTIVE", "ERROR", or "all".
	Status   string
	Platform Platform
	// IncludeArchived includes DISCONNECTED (archived) accounts. Ignored when
	// Status is set explicitly.
	IncludeArchived bool
	Limit           int
	Cursor          string
}

// List lists connected accounts, optionally filtered and paginated. Pass nil
// for params to list with default filters (excludes archived accounts).
func (s *AccountsService) List(ctx context.Context, params *ListAccountsParams, opts ...RequestOption) (*ListResponse[Account], error) {
	queryParams := make(map[string]string)
	if params != nil {
		if params.Status != "" {
			queryParams["status"] = params.Status
		}
		if params.Platform != "" {
			queryParams["platform"] = string(params.Platform)
		}
		if params.IncludeArchived {
			queryParams["includeArchived"] = "true"
		}
		if params.Limit > 0 {
			queryParams["limit"] = strconv.Itoa(params.Limit)
		}
		if params.Cursor != "" {
			queryParams["cursor"] = params.Cursor
		}
	}

	path := "/accounts" + buildQueryString(queryParams)
	var result ListResponse[Account]
	if err := s.client.get(ctx, path, &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListAll returns an iterator for all accounts matching the given parameters.
// This handles pagination automatically.
func (s *AccountsService) ListAll(ctx context.Context, params *ListAccountsParams) *AccountsIterator {
	if params == nil {
		params = &ListAccountsParams{}
	}
	return &AccountsIterator{
		service: s,
		ctx:     ctx,
		params:  *params,
	}
}

// AccountsIterator provides an iterator for paginated account results.
type AccountsIterator struct {
	service *AccountsService
	ctx     context.Context
	params  ListAccountsParams

	accounts []Account
	index    int
	done     bool
	err      error
}

// Next returns the next account. Returns false when there are no more accounts.
func (it *AccountsIterator) Next() bool {
	if it.err != nil || it.done {
		return false
	}

	if it.index < len(it.accounts) {
		return true
	}

	resp, err := it.service.List(it.ctx, &it.params)
	if err != nil {
		it.err = err
		return false
	}

	it.accounts = resp.Data
	it.index = 0

	if len(it.accounts) == 0 {
		it.done = true
		return false
	}

	if resp.Pagination.HasMore && resp.Pagination.NextCursor != "" {
		it.params.Cursor = resp.Pagination.NextCursor
	} else {
		it.done = true
	}

	return true
}

// Account returns the current account.
func (it *AccountsIterator) Account() Account {
	if it.index < len(it.accounts) {
		account := it.accounts[it.index]
		it.index++
		return account
	}
	return Account{}
}

// Err returns any error that occurred during iteration.
func (it *AccountsIterator) Err() error {
	return it.err
}

// Get retrieves a specific account by ID.
func (s *AccountsService) Get(ctx context.Context, id string, opts ...RequestOption) (*Account, error) {
	var account Account
	if err := s.client.get(ctx, "/accounts/"+id, &account, opts...); err != nil {
		return nil, err
	}
	return &account, nil
}

// Connect initiates the OAuth flow for a platform.
// Returns the authorization URL to redirect the user to.
func (s *AccountsService) Connect(ctx context.Context, platform Platform, params ConnectParams, opts ...RequestOption) (*OAuthResponse, error) {
	var result OAuthResponse
	if err := s.client.post(ctx, fmt.Sprintf("/accounts/connect/%s", platform), params, &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}

// InitiateConnect is a convenience wrapper around Connect for the common case
// of a plain callback URL with no extra options.
func (s *AccountsService) InitiateConnect(ctx context.Context, platform Platform, callbackURL string, opts ...RequestOption) (*OAuthResponse, error) {
	return s.Connect(ctx, platform, ConnectParams{CallbackURL: callbackURL}, opts...)
}

// Callback completes the OAuth flow after the user authorizes.
func (s *AccountsService) Callback(ctx context.Context, platform Platform, params CallbackParams, opts ...RequestOption) (*OAuthCallbackResponse, error) {
	var result OAuthCallbackResponse
	if err := s.client.post(ctx, fmt.Sprintf("/accounts/callback/%s", platform), params, &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}

// CompleteConnect is an alias for Callback, matching the InitiateConnect/CompleteConnect naming pair.
func (s *AccountsService) CompleteConnect(ctx context.Context, platform Platform, params CallbackParams, opts ...RequestOption) (*OAuthCallbackResponse, error) {
	return s.Callback(ctx, platform, params, opts...)
}

// Disconnect removes a connected account.
func (s *AccountsService) Disconnect(ctx context.Context, id string, opts ...RequestOption) error {
	return s.client.delete(ctx, "/accounts/"+id, nil, opts...)
}

// Delete is an alias for Disconnect, matching the REST verb (DELETE /accounts/:id).
func (s *AccountsService) Delete(ctx context.Context, id string, opts ...RequestOption) error {
	return s.Disconnect(ctx, id, opts...)
}

// TestAccountDetails carries the platform's own identifiers for the account
// that was verified, when the test succeeds.
type TestAccountDetails struct {
	PlatformUserID string `json:"platformUserId"`
	Username       string `json:"username"`
	DisplayName    string `json:"displayName"`
}

// TestAccountResult is the response from AccountsService.Test.
type TestAccountResult struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Details *TestAccountDetails `json:"details,omitempty"`
}

// Test verifies a connected account's stored token still works by calling the
// platform's own API. On failure it does not return a Go error for an
// unhealthy account — check Success and Message. A Go error is only returned
// for transport/HTTP-level failures (account not found, auth failure, etc).
func (s *AccountsService) Test(ctx context.Context, id string, opts ...RequestOption) (*TestAccountResult, error) {
	var result TestAccountResult
	if err := s.client.post(ctx, fmt.Sprintf("/accounts/%s/test", id), nil, &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}

// Refresh manually triggers a token refresh for an account.
// SelectPageParams are the parameters for AccountsService.SelectPage.
type SelectPageParams struct {
	// PageID is the platform's Page/organization identifier. Pass an empty
	// string to clear the selection and publish as the personal profile
	// (LinkedIn).
	PageID string `json:"pageId"`
	// PageName is optional and used only for display.
	PageName string `json:"pageName,omitempty"`
}

// SelectPage chooses which Page or organization a connected account publishes
// as. Requires the accounts:write scope.
//
// On failure the returned error is an *APIError with HTTP status 400 and a Code
// of ErrCodePageSelectionNotSupported, ErrCodeAccountPageMismatch, or
// ErrCodeDestinationsNotLoaded. Use GetAPIError to inspect it.
func (s *AccountsService) SelectPage(ctx context.Context, id string, params SelectPageParams, opts ...RequestOption) error {
	var result struct {
		Success bool `json:"success"`
	}
	if err := s.client.post(ctx, fmt.Sprintf("/accounts/%s/select-page", id), params, &result, opts...); err != nil {
		return err
	}
	return nil
}

func (s *AccountsService) Refresh(ctx context.Context, id string, opts ...RequestOption) error {
	var result struct {
		Success bool `json:"success"`
	}
	if err := s.client.post(ctx, fmt.Sprintf("/accounts/%s/refresh", id), nil, &result, opts...); err != nil {
		return err
	}
	return nil
}

// ListByPlatform returns all accounts for a specific platform.
func (s *AccountsService) ListByPlatform(ctx context.Context, platform Platform, opts ...RequestOption) ([]Account, error) {
	result, err := s.List(ctx, nil, opts...)
	if err != nil {
		return nil, err
	}

	var filtered []Account
	for _, acc := range result.Data {
		if acc.Platform == platform {
			filtered = append(filtered, acc)
		}
	}
	return filtered, nil
}

// GetByUsername finds an account by platform and username.
func (s *AccountsService) GetByUsername(ctx context.Context, platform Platform, username string, opts ...RequestOption) (*Account, error) {
	result, err := s.List(ctx, nil, opts...)
	if err != nil {
		return nil, err
	}

	for _, acc := range result.Data {
		if acc.Platform == platform && acc.Username == username {
			return &acc, nil
		}
	}
	return nil, &APIError{
		Code:    ErrCodeNotFound,
		Message: fmt.Sprintf("account not found: %s on %s", username, platform),
	}
}
