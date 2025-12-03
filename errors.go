package logward

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidAPIKey is returned when the API key is missing or invalid.
	ErrInvalidAPIKey = errors.New("invalid or missing API key")

	// ErrCircuitOpen is returned when the circuit breaker is in the open state.
	ErrCircuitOpen = errors.New("circuit breaker is open")

	// ErrTimeout is returned when an operation times out.
	ErrTimeout = errors.New("operation timed out")

	// ErrClientClosed is returned when attempting to use a closed client.
	ErrClientClosed = errors.New("client is closed")
)

// ValidationError represents a validation error for log data.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// Is allows errors.Is to match ValidationError types.
func (e *ValidationError) Is(target error) bool {
	_, ok := target.(*ValidationError)
	return ok
}

// HTTPError represents an HTTP error response from the LogWard API.
type HTTPError struct {
	StatusCode int
	Message    string
	Body       string
}

// Error implements the error interface.
func (e *HTTPError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("HTTP %d", e.StatusCode)
}

// Is allows errors.Is to match HTTPError types.
func (e *HTTPError) Is(target error) bool {
	_, ok := target.(*HTTPError)
	return ok
}

// IsRetryable returns true if the HTTP error indicates a retryable condition.
func (e *HTTPError) IsRetryable() bool {
	return e.StatusCode == 429 || // Too Many Requests
		e.StatusCode == 500 || // Internal Server Error
		e.StatusCode == 502 || // Bad Gateway
		e.StatusCode == 503 || // Service Unavailable
		e.StatusCode == 504 // Gateway Timeout
}
