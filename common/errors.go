package common

import (
	"errors"
	"fmt"
)

// 通用错误
var (
	ErrNotInitialized     = errors.New("not initialized")
	ErrInvalidConfig      = errors.New("invalid configuration")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidPrivateKey  = errors.New("invalid private key")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrNotFound           = errors.New("not found")
	ErrRateLimited        = errors.New("rate limited")
	ErrServerError        = errors.New("server error")
	ErrTimeout            = errors.New("request timeout")
	ErrBadRequest         = errors.New("bad request")
)

// 订单相关错误
var (
	ErrInvalidOrder         = errors.New("invalid order")
	ErrInsufficientBalance  = errors.New("insufficient balance")
	ErrOrderNotFound        = errors.New("order not found")
	ErrOrderAlreadyCanceled = errors.New("order already canceled")
	ErrInvalidOrderType     = errors.New("invalid order type")
	ErrInvalidOrderSide     = errors.New("invalid order side")
	ErrInvalidPrice         = errors.New("invalid price")
	ErrInvalidSize          = errors.New("invalid size")
)

// 市场相关错误
var (
	ErrMarketNotFound  = errors.New("market not found")
	ErrMarketClosed    = errors.New("market is closed")
	ErrMarketNotActive = errors.New("market is not active")
)

// 签名相关错误
var (
	ErrSigningFailed     = errors.New("signing failed")
	ErrInvalidSignature  = errors.New("invalid signature")
	ErrInvalidAddress    = errors.New("invalid address")
	ErrInvalidTimestamp  = errors.New("invalid timestamp")
)

// APIError API 错误响应
type APIError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"error,omitempty"`
	Message    string `json:"message,omitempty"`
	Details    string `json:"details,omitempty"`
}

// Error 实现 error 接口
func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("API error [%d]: %s - %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("API error [%d]: %s", e.StatusCode, e.Code)
}

// NewAPIError 创建 API 错误
func NewAPIError(statusCode int, code, message string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}

// IsNotFound 判断是否为 404 错误
func IsNotFound(err error) bool {
	if errors.Is(err, ErrNotFound) {
		return true
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 404
	}
	return false
}

// IsUnauthorized 判断是否为 401 错误
func IsUnauthorized(err error) bool {
	if errors.Is(err, ErrUnauthorized) {
		return true
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 401
	}
	return false
}

// IsRateLimited 判断是否为速率限制错误
func IsRateLimited(err error) bool {
	if errors.Is(err, ErrRateLimited) {
		return true
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 429
	}
	return false
}
