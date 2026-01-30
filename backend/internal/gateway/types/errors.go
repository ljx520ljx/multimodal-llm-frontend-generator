package types

import (
	"errors"
	"fmt"
)

// Common gateway errors
var (
	ErrInvalidAPIKey     = errors.New("invalid or missing API key")
	ErrInvalidModel      = errors.New("invalid or unsupported model")
	ErrRateLimited       = errors.New("rate limited by API")
	ErrContextCanceled   = errors.New("context canceled")
	ErrTimeout           = errors.New("request timeout")
	ErrServerError       = errors.New("server error")
	ErrInvalidRequest    = errors.New("invalid request")
	ErrUnsupportedProvider = errors.New("unsupported provider")
)

// GatewayError represents an error from the LLM gateway
type GatewayError struct {
	Provider   string
	StatusCode int
	Message    string
	Retryable  bool
	Err        error
}

func (e *GatewayError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Provider, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s (status: %d)", e.Provider, e.Message, e.StatusCode)
}

func (e *GatewayError) Unwrap() error {
	return e.Err
}

// IsRetryable returns true if the error is retryable
func (e *GatewayError) IsRetryable() bool {
	return e.Retryable
}

// NewGatewayError creates a new gateway error
func NewGatewayError(provider string, statusCode int, message string, retryable bool, err error) *GatewayError {
	return &GatewayError{
		Provider:   provider,
		StatusCode: statusCode,
		Message:    message,
		Retryable:  retryable,
		Err:        err,
	}
}

// IsRetryableStatusCode returns true if the HTTP status code is retryable
func IsRetryableStatusCode(statusCode int) bool {
	switch statusCode {
	case 429, 500, 502, 503, 504:
		return true
	default:
		return false
	}
}

// IsRetryableError checks if an error is retryable
func IsRetryableError(err error) bool {
	var gatewayErr *GatewayError
	if errors.As(err, &gatewayErr) {
		return gatewayErr.IsRetryable()
	}
	return false
}
