package voxburst

import (
	"context"
	"fmt"
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

// List lists all connected accounts.
func (s *AccountsService) List(ctx context.Context, opts ...RequestOption) (*ListResponse[Account], error) {
	var result ListResponse[Account]
	if err := s.client.get(ctx, "/accounts", &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
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

// Callback completes the OAuth flow after the user authorizes.
func (s *AccountsService) Callback(ctx context.Context, platform Platform, params CallbackParams, opts ...RequestOption) (*OAuthCallbackResponse, error) {
	var result OAuthCallbackResponse
	if err := s.client.post(ctx, fmt.Sprintf("/accounts/callback/%s", platform), params, &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}

// Disconnect removes a connected account.
func (s *AccountsService) Disconnect(ctx context.Context, id string, opts ...RequestOption) error {
	return s.client.delete(ctx, "/accounts/"+id, nil, opts...)
}

// Refresh manually triggers a token refresh for an account.
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
	result, err := s.List(ctx, opts...)
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
	result, err := s.List(ctx, opts...)
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
