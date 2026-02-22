package voxburst

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Error codes returned by the API.
const (
	ErrCodeValidation   = "VALIDATION_ERROR"
	ErrCodeUnauthorized = "UNAUTHORIZED"
	ErrCodeForbidden    = "FORBIDDEN"
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeConflict     = "CONFLICT"
	ErrCodeRateLimited  = "RATE_LIMITED"
	ErrCodeInternal     = "INTERNAL_ERROR"
)

// APIError represents an error returned by the API.
type APIError struct {
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	Details    map[string]any `json:"details,omitempty"`
	StatusCode int            `json:"-"` // HTTP status code
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("voxburst: %s - %s (details: %v)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("voxburst: %s - %s", e.Code, e.Message)
}

// apiErrorResponse is the raw error response from the API.
type apiErrorResponse struct {
	Error *APIError `json:"error"`
}

// parseAPIError parses an API error from the response.
func parseAPIError(resp *http.Response) error {
	var errResp apiErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return &APIError{
			Code:       ErrCodeInternal,
			Message:    fmt.Sprintf("failed to parse error response: %v", err),
			StatusCode: resp.StatusCode,
		}
	}
	if errResp.Error != nil {
		errResp.Error.StatusCode = resp.StatusCode
		return errResp.Error
	}
	return &APIError{
		Code:       ErrCodeInternal,
		Message:    fmt.Sprintf("unexpected error with status %d", resp.StatusCode),
		StatusCode: resp.StatusCode,
	}
}

// IsRateLimited returns true if the error is a rate limit error.
func IsRateLimited(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code == ErrCodeRateLimited || apiErr.StatusCode == http.StatusTooManyRequests
	}
	return false
}

// IsNotFound returns true if the error is a not found error.
func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code == ErrCodeNotFound || apiErr.StatusCode == http.StatusNotFound
	}
	return false
}

// IsUnauthorized returns true if the error is an unauthorized error.
func IsUnauthorized(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code == ErrCodeUnauthorized || apiErr.StatusCode == http.StatusUnauthorized
	}
	return false
}

// IsForbidden returns true if the error is a forbidden error.
func IsForbidden(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code == ErrCodeForbidden || apiErr.StatusCode == http.StatusForbidden
	}
	return false
}

// IsValidationError returns true if the error is a validation error.
func IsValidationError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code == ErrCodeValidation || apiErr.StatusCode == http.StatusBadRequest
	}
	return false
}

// IsConflict returns true if the error is a conflict error.
func IsConflict(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code == ErrCodeConflict || apiErr.StatusCode == http.StatusConflict
	}
	return false
}

// IsServerError returns true if the error is a server error.
func IsServerError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode >= 500
	}
	return false
}

// IsRetryable returns true if the error is likely retryable.
func IsRetryable(err error) bool {
	return IsRateLimited(err) || IsServerError(err)
}

// GetAPIError extracts the APIError from an error if present.
func GetAPIError(err error) (*APIError, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}
