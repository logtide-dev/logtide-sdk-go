package logward

import "time"

// Config holds the configuration for the LogWard client.
type Config struct {
	// APIKey is the LogWard API key (required).
	APIKey string

	// BaseURL is the LogWard API base URL.
	// Default: "https://api.logward.dev"
	BaseURL string

	// Service is the default service name for all logs (required).
	Service string

	// Timeout is the HTTP request timeout.
	// Default: 30 seconds
	Timeout time.Duration

	// BatchSize is the maximum number of logs per batch.
	// Default: 100
	BatchSize int

	// FlushInterval is the maximum time to wait before flushing a batch.
	// Default: 5 seconds
	FlushInterval time.Duration

	// RetryConfig holds the retry configuration.
	RetryConfig *RetryConfig

	// CircuitBreakerConfig holds the circuit breaker configuration.
	CircuitBreakerConfig *CircuitBreakerConfig
}

// Option is a functional option for configuring the Client.
type Option func(*Config)

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		BaseURL:              "https://api.logward.dev",
		Timeout:              30 * time.Second,
		BatchSize:            100,
		FlushInterval:        5 * time.Second,
		RetryConfig:          DefaultRetryConfig(),
		CircuitBreakerConfig: DefaultCircuitBreakerConfig(),
	}
}

// WithAPIKey sets the API key.
func WithAPIKey(apiKey string) Option {
	return func(c *Config) {
		c.APIKey = apiKey
	}
}

// WithBaseURL sets the base URL.
func WithBaseURL(baseURL string) Option {
	return func(c *Config) {
		c.BaseURL = baseURL
	}
}

// WithService sets the default service name.
func WithService(service string) Option {
	return func(c *Config) {
		c.Service = service
	}
}

// WithTimeout sets the HTTP timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithBatchSize sets the maximum batch size.
func WithBatchSize(size int) Option {
	return func(c *Config) {
		c.BatchSize = size
	}
}

// WithFlushInterval sets the flush interval.
func WithFlushInterval(interval time.Duration) Option {
	return func(c *Config) {
		c.FlushInterval = interval
	}
}

// WithRetry sets the retry configuration.
func WithRetry(maxRetries int, minBackoff, maxBackoff time.Duration) Option {
	return func(c *Config) {
		c.RetryConfig = &RetryConfig{
			MaxRetries: maxRetries,
			MinBackoff: minBackoff,
			MaxBackoff: maxBackoff,
		}
	}
}

// WithCircuitBreaker sets the circuit breaker configuration.
func WithCircuitBreaker(failureThreshold int, timeout time.Duration) Option {
	return func(c *Config) {
		c.CircuitBreakerConfig = &CircuitBreakerConfig{
			FailureThreshold: failureThreshold,
			Timeout:          timeout,
		}
	}
}

// validate validates the configuration.
func (c *Config) validate() error {
	if c.APIKey == "" {
		return ErrInvalidAPIKey
	}
	if c.Service == "" {
		return &ValidationError{Field: "service", Message: "service name is required"}
	}
	if len(c.Service) > 100 {
		return &ValidationError{Field: "service", Message: "service name must be 100 characters or less"}
	}
	if c.BaseURL == "" {
		return &ValidationError{Field: "baseURL", Message: "base URL is required"}
	}
	return nil
}
