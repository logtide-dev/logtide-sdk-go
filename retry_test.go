package logward

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		err        error
		want       bool
	}{
		{
			name:       "network error",
			statusCode: 0,
			err:        errors.New("network error"),
			want:       true,
		},
		{
			name:       "429 Too Many Requests",
			statusCode: 429,
			err:        nil,
			want:       true,
		},
		{
			name:       "500 Internal Server Error",
			statusCode: 500,
			err:        nil,
			want:       true,
		},
		{
			name:       "502 Bad Gateway",
			statusCode: 502,
			err:        nil,
			want:       true,
		},
		{
			name:       "503 Service Unavailable",
			statusCode: 503,
			err:        nil,
			want:       true,
		},
		{
			name:       "504 Gateway Timeout",
			statusCode: 504,
			err:        nil,
			want:       true,
		},
		{
			name:       "200 OK",
			statusCode: 200,
			err:        nil,
			want:       false,
		},
		{
			name:       "400 Bad Request",
			statusCode: 400,
			err:        nil,
			want:       false,
		},
		{
			name:       "401 Unauthorized",
			statusCode: 401,
			err:        nil,
			want:       false,
		},
		{
			name:       "403 Forbidden",
			statusCode: 403,
			err:        nil,
			want:       false,
		},
		{
			name:       "404 Not Found",
			statusCode: 404,
			err:        nil,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp *http.Response
			if tt.statusCode > 0 {
				resp = &http.Response{StatusCode: tt.statusCode}
			}
			got := shouldRetry(resp, tt.err)
			if got != tt.want {
				t.Errorf("shouldRetry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateBackoff(t *testing.T) {
	config := &RetryConfig{
		MinBackoff: 1 * time.Second,
		MaxBackoff: 10 * time.Second,
	}

	tests := []struct {
		name     string
		attempt  int
		wantMin  time.Duration
		wantMax  time.Duration
	}{
		{
			name:     "first retry",
			attempt:  0,
			wantMin:  1 * time.Second,
			wantMax:  1500 * time.Millisecond, // 1s + 25% jitter
		},
		{
			name:     "second retry",
			attempt:  1,
			wantMin:  2 * time.Second,
			wantMax:  2500 * time.Millisecond, // 2s + 25% jitter
		},
		{
			name:     "third retry",
			attempt:  2,
			wantMin:  4 * time.Second,
			wantMax:  5 * time.Second, // 4s + 25% jitter
		},
		{
			name:     "capped at max",
			attempt:  10,
			wantMin:  10 * time.Second,
			wantMax:  12500 * time.Millisecond, // 10s + 25% jitter
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backoff := calculateBackoff(tt.attempt, config)
			if backoff < tt.wantMin || backoff > tt.wantMax {
				t.Errorf("calculateBackoff(%d) = %v, want between %v and %v", tt.attempt, backoff, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestWithRetry(t *testing.T) {
	t.Run("success on first attempt", func(t *testing.T) {
		attempts := 0
		config := DefaultRetryConfig()

		fn := func(ctx context.Context) (*http.Response, error) {
			attempts++
			return &http.Response{StatusCode: 200}, nil
		}

		resp, err := withRetry(context.Background(), config, fn)
		if err != nil {
			t.Errorf("withRetry() error = %v, want nil", err)
		}
		if resp.StatusCode != 200 {
			t.Errorf("withRetry() status = %d, want 200", resp.StatusCode)
		}
		if attempts != 1 {
			t.Errorf("withRetry() attempts = %d, want 1", attempts)
		}
	})

	t.Run("retries on 500 error", func(t *testing.T) {
		attempts := 0
		config := &RetryConfig{
			MaxRetries: 2,
			MinBackoff: 10 * time.Millisecond,
			MaxBackoff: 100 * time.Millisecond,
		}

		fn := func(ctx context.Context) (*http.Response, error) {
			attempts++
			if attempts <= 2 {
				return &http.Response{StatusCode: 500}, nil
			}
			return &http.Response{StatusCode: 200}, nil
		}

		resp, err := withRetry(context.Background(), config, fn)
		if err != nil {
			t.Errorf("withRetry() error = %v, want nil", err)
		}
		if resp.StatusCode != 200 {
			t.Errorf("withRetry() status = %d, want 200", resp.StatusCode)
		}
		if attempts != 3 {
			t.Errorf("withRetry() attempts = %d, want 3", attempts)
		}
	})

	t.Run("exhausts retries", func(t *testing.T) {
		attempts := 0
		config := &RetryConfig{
			MaxRetries: 2,
			MinBackoff: 10 * time.Millisecond,
			MaxBackoff: 100 * time.Millisecond,
		}

		fn := func(ctx context.Context) (*http.Response, error) {
			attempts++
			return &http.Response{StatusCode: 500}, nil
		}

		resp, err := withRetry(context.Background(), config, fn)
		if err != nil {
			t.Errorf("withRetry() error = %v, want nil", err)
		}
		if resp.StatusCode != 500 {
			t.Errorf("withRetry() status = %d, want 500", resp.StatusCode)
		}
		if attempts != 3 {
			t.Errorf("withRetry() attempts = %d, want 3 (initial + 2 retries)", attempts)
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		attempts := 0
		config := &RetryConfig{
			MaxRetries: 5,
			MinBackoff: 100 * time.Millisecond,
			MaxBackoff: 1 * time.Second,
		}

		ctx, cancel := context.WithCancel(context.Background())

		fn := func(ctx context.Context) (*http.Response, error) {
			attempts++
			if attempts == 2 {
				cancel() // Cancel after second attempt
			}
			return &http.Response{StatusCode: 500}, nil
		}

		_, err := withRetry(ctx, config, fn)
		if err == nil {
			t.Error("withRetry() error = nil, want context.Canceled")
		}
		if attempts > 2 {
			t.Errorf("withRetry() attempts = %d, want <= 2 (should stop after context cancellation)", attempts)
		}
	})
}
