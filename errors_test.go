package logtide

import (
	"errors"
	"testing"
)

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "service",
		Message: "service name is required",
	}

	// Test Error() method
	expected := "validation error on field 'service': service name is required"
	if err.Error() != expected {
		t.Errorf("ValidationError.Error() = %q, want %q", err.Error(), expected)
	}

	// Test errors.Is() compatibility
	var target *ValidationError
	if !errors.As(err, &target) {
		t.Error("errors.As() should match ValidationError type")
	}
}

func TestHTTPError(t *testing.T) {
	tests := []struct {
		name         string
		err          *HTTPError
		expectedMsg  string
		isRetryable  bool
	}{
		{
			name: "429 Too Many Requests",
			err: &HTTPError{
				StatusCode: 429,
				Message:    "Rate limit exceeded",
			},
			expectedMsg: "HTTP 429: Rate limit exceeded",
			isRetryable: true,
		},
		{
			name: "500 Internal Server Error",
			err: &HTTPError{
				StatusCode: 500,
				Message:    "Internal server error",
			},
			expectedMsg: "HTTP 500: Internal server error",
			isRetryable: true,
		},
		{
			name: "502 Bad Gateway",
			err: &HTTPError{
				StatusCode: 502,
			},
			expectedMsg: "HTTP 502",
			isRetryable: true,
		},
		{
			name: "503 Service Unavailable",
			err: &HTTPError{
				StatusCode: 503,
			},
			expectedMsg: "HTTP 503",
			isRetryable: true,
		},
		{
			name: "504 Gateway Timeout",
			err: &HTTPError{
				StatusCode: 504,
			},
			expectedMsg: "HTTP 504",
			isRetryable: true,
		},
		{
			name: "400 Bad Request",
			err: &HTTPError{
				StatusCode: 400,
				Message:    "Invalid request",
			},
			expectedMsg: "HTTP 400: Invalid request",
			isRetryable: false,
		},
		{
			name: "401 Unauthorized",
			err: &HTTPError{
				StatusCode: 401,
				Message:    "Unauthorized",
			},
			expectedMsg: "HTTP 401: Unauthorized",
			isRetryable: false,
		},
		{
			name: "403 Forbidden",
			err: &HTTPError{
				StatusCode: 403,
				Message:    "Forbidden",
			},
			expectedMsg: "HTTP 403: Forbidden",
			isRetryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Error() method
			if tt.err.Error() != tt.expectedMsg {
				t.Errorf("HTTPError.Error() = %q, want %q", tt.err.Error(), tt.expectedMsg)
			}

			// Test IsRetryable() method
			if tt.err.IsRetryable() != tt.isRetryable {
				t.Errorf("HTTPError.IsRetryable() = %v, want %v", tt.err.IsRetryable(), tt.isRetryable)
			}

			// Test errors.Is() compatibility
			var target *HTTPError
			if !errors.As(tt.err, &target) {
				t.Error("errors.As() should match HTTPError type")
			}
		})
	}
}

func TestSentinelErrors(t *testing.T) {
	// Test that sentinel errors are defined
	if ErrInvalidAPIKey == nil {
		t.Error("ErrInvalidAPIKey should not be nil")
	}
	if ErrCircuitOpen == nil {
		t.Error("ErrCircuitOpen should not be nil")
	}
	if ErrTimeout == nil {
		t.Error("ErrTimeout should not be nil")
	}
	if ErrClientClosed == nil {
		t.Error("ErrClientClosed should not be nil")
	}

	// Test errors.Is() with sentinel errors
	err1 := ErrInvalidAPIKey
	if !errors.Is(err1, ErrInvalidAPIKey) {
		t.Error("errors.Is() should match ErrInvalidAPIKey")
	}

	err2 := ErrCircuitOpen
	if !errors.Is(err2, ErrCircuitOpen) {
		t.Error("errors.Is() should match ErrCircuitOpen")
	}
}
