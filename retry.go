package logtide

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"time"
)

// RetryConfig holds the configuration for retry logic.
type RetryConfig struct {
	MaxRetries int
	MinBackoff time.Duration
	MaxBackoff time.Duration
}

// DefaultRetryConfig returns the default retry configuration.
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries: 3,
		MinBackoff: 1 * time.Second,
		MaxBackoff: 60 * time.Second,
	}
}

// shouldRetry determines if a request should be retried based on the response.
func shouldRetry(resp *http.Response, err error) bool {
	// Retry on network errors
	if err != nil {
		return true
	}

	// Retry on specific HTTP status codes
	switch resp.StatusCode {
	case 429: // Too Many Requests
		return true
	case 500: // Internal Server Error
		return true
	case 502: // Bad Gateway
		return true
	case 503: // Service Unavailable
		return true
	case 504: // Gateway Timeout
		return true
	default:
		return false
	}
}

// calculateBackoff calculates the backoff duration for a retry attempt with exponential backoff and jitter.
func calculateBackoff(attempt int, config *RetryConfig) time.Duration {
	// Calculate exponential backoff: min_backoff * 2^attempt
	backoff := float64(config.MinBackoff) * math.Pow(2, float64(attempt))

	// Cap at max backoff
	if backoff > float64(config.MaxBackoff) {
		backoff = float64(config.MaxBackoff)
	}

	// Add jitter (random value between 0 and 25% of backoff)
	jitter := rand.Float64() * 0.25 * backoff
	backoff += jitter

	return time.Duration(backoff)
}

// retryableFunc is a function that can be retried.
type retryableFunc func(ctx context.Context) (*http.Response, error)

// withRetry executes a function with retry logic.
func withRetry(ctx context.Context, config *RetryConfig, fn retryableFunc) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Execute the function
		resp, err = fn(ctx)

		// Check if we should retry
		if !shouldRetry(resp, err) {
			// Success or non-retryable error
			return resp, err
		}

		// Check if we've exhausted retries
		if attempt == config.MaxRetries {
			// Last attempt failed
			if err != nil {
				return nil, fmt.Errorf("max retries exceeded: %w", err)
			}
			return resp, nil
		}

		// Calculate backoff
		backoff := calculateBackoff(attempt, config)

		// Wait before retrying, respecting context cancellation
		select {
		case <-time.After(backoff):
			// Continue to next attempt
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		// Close response body if it exists before retrying
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}

	return resp, err
}
