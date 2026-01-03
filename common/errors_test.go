package common

import (
	"errors"
	"testing"
)

func TestAPIError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		code       string
		message    string
		expected   string
	}{
		{
			name:       "with message",
			statusCode: 400,
			code:       "BAD_REQUEST",
			message:    "Invalid parameters",
			expected:   "API error [400]: BAD_REQUEST - Invalid parameters",
		},
		{
			name:       "without message",
			statusCode: 500,
			code:       "INTERNAL_ERROR",
			message:    "",
			expected:   "API error [500]: INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAPIError(tt.statusCode, tt.code, tt.message)
			if err.Error() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, err.Error())
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ErrNotFound",
			err:      ErrNotFound,
			expected: true,
		},
		{
			name:     "API 404 error",
			err:      &APIError{StatusCode: 404, Code: "NOT_FOUND"},
			expected: true,
		},
		{
			name:     "API 500 error",
			err:      &APIError{StatusCode: 500, Code: "SERVER_ERROR"},
			expected: false,
		},
		{
			name:     "other error",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.expected {
				t.Errorf("IsNotFound() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestIsUnauthorized(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ErrUnauthorized",
			err:      ErrUnauthorized,
			expected: true,
		},
		{
			name:     "API 401 error",
			err:      &APIError{StatusCode: 401, Code: "UNAUTHORIZED"},
			expected: true,
		},
		{
			name:     "API 403 error",
			err:      &APIError{StatusCode: 403, Code: "FORBIDDEN"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUnauthorized(tt.err); got != tt.expected {
				t.Errorf("IsUnauthorized() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestIsRateLimited(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ErrRateLimited",
			err:      ErrRateLimited,
			expected: true,
		},
		{
			name:     "API 429 error",
			err:      &APIError{StatusCode: 429, Code: "RATE_LIMITED"},
			expected: true,
		},
		{
			name:     "API 500 error",
			err:      &APIError{StatusCode: 500, Code: "SERVER_ERROR"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRateLimited(tt.err); got != tt.expected {
				t.Errorf("IsRateLimited() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestErrorVariables(t *testing.T) {
	// Test that error variables are defined
	errorVars := []error{
		ErrNotInitialized,
		ErrInvalidConfig,
		ErrInvalidCredentials,
		ErrInvalidPrivateKey,
		ErrUnauthorized,
		ErrForbidden,
		ErrNotFound,
		ErrRateLimited,
		ErrServerError,
		ErrTimeout,
		ErrBadRequest,
		ErrInvalidOrder,
		ErrInsufficientBalance,
		ErrOrderNotFound,
		ErrOrderAlreadyCanceled,
		ErrInvalidOrderType,
		ErrInvalidOrderSide,
		ErrInvalidPrice,
		ErrInvalidSize,
		ErrMarketNotFound,
		ErrMarketClosed,
		ErrMarketNotActive,
		ErrSigningFailed,
		ErrInvalidSignature,
		ErrInvalidAddress,
		ErrInvalidTimestamp,
	}

	for _, err := range errorVars {
		if err == nil {
			t.Error("Error variable should not be nil")
		}
		if err.Error() == "" {
			t.Error("Error message should not be empty")
		}
	}
}
