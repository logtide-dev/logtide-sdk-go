package logward

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	internalhttp "github.com/logward-dev/logward-sdk-go/internal/http"
)

// Client is the LogWard SDK client for sending logs.
type Client struct {
	config         *Config
	httpClient     *internalhttp.Client
	batcher        *Batcher
	circuitBreaker *CircuitBreaker
	retryConfig    *RetryConfig

	mu     sync.RWMutex
	closed bool
}

// New creates a new LogWard client with the specified options.
func New(opts ...Option) (*Client, error) {
	// Start with default config
	config := DefaultConfig()

	// Apply options
	for _, opt := range opts {
		opt(config)
	}

	// Validate config
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create HTTP client
	httpClient := internalhttp.NewClient(&internalhttp.Config{
		BaseURL:   config.BaseURL,
		APIKey:    config.APIKey,
		Timeout:   config.Timeout,
	})

	// Create circuit breaker
	circuitBreaker := NewCircuitBreaker(config.CircuitBreakerConfig)

	// Create client
	client := &Client{
		config:         config,
		httpClient:     httpClient,
		circuitBreaker: circuitBreaker,
		retryConfig:    config.RetryConfig,
	}

	// Create batcher with flush function
	batcherConfig := &BatcherConfig{
		MaxSize:       config.BatchSize,
		FlushInterval: config.FlushInterval,
		FlushFunc:     client.sendBatch,
	}
	client.batcher = NewBatcher(batcherConfig)

	return client, nil
}

// Debug sends a debug-level log.
func (c *Client) Debug(ctx context.Context, message string, metadata map[string]interface{}) error {
	return c.log(ctx, LogLevelDebug, message, metadata)
}

// Info sends an info-level log.
func (c *Client) Info(ctx context.Context, message string, metadata map[string]interface{}) error {
	return c.log(ctx, LogLevelInfo, message, metadata)
}

// Warn sends a warn-level log.
func (c *Client) Warn(ctx context.Context, message string, metadata map[string]interface{}) error {
	return c.log(ctx, LogLevelWarn, message, metadata)
}

// Error sends an error-level log.
func (c *Client) Error(ctx context.Context, message string, metadata map[string]interface{}) error {
	return c.log(ctx, LogLevelError, message, metadata)
}

// Critical sends a critical-level log.
func (c *Client) Critical(ctx context.Context, message string, metadata map[string]interface{}) error {
	return c.log(ctx, LogLevelCritical, message, metadata)
}

// log creates and adds a log entry to the batcher.
func (c *Client) log(ctx context.Context, level LogLevel, message string, metadata map[string]interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrClientClosed
	}

	// Create log entry
	log := Log{
		Time:     time.Now(),
		Service:  c.config.Service,
		Level:    level,
		Message:  message,
		Metadata: metadata,
	}

	// Enrich with context (OpenTelemetry trace/span IDs)
	enrichLogWithContext(ctx, &log)

	// Validate log
	if err := validateLog(&log); err != nil {
		return fmt.Errorf("invalid log: %w", err)
	}

	// Add to batcher
	return c.batcher.Add(log)
}

// sendBatch sends a batch of logs to the LogWard API.
func (c *Client) sendBatch(ctx context.Context, logs []Log) error {
	// Validate batch
	if err := validateBatch(logs); err != nil {
		return fmt.Errorf("invalid batch: %w", err)
	}

	// Check circuit breaker
	if err := c.circuitBreaker.Allow(); err != nil {
		return err
	}

	// Create request
	req := &IngestRequest{
		Logs: logs,
	}

	// Send with retry
	resp, err := withRetry(ctx, c.retryConfig, func(ctx context.Context) (*http.Response, error) {
		return c.httpClient.Post(ctx, "/api/v1/ingest", req)
	})

	// Record circuit breaker result
	if err != nil || (resp != nil && resp.StatusCode >= 500) {
		c.circuitBreaker.RecordFailure()
	} else {
		c.circuitBreaker.RecordSuccess()
	}

	if err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := internalhttp.ReadResponseBody(resp)
		return &HTTPError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
			Body:       body,
		}
	}

	// Decode response
	var ingestResp IngestResponse
	if err := internalhttp.DecodeResponse(resp, &ingestResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// Flush immediately flushes all pending logs.
func (c *Client) Flush(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrClientClosed
	}

	return c.batcher.Flush(ctx)
}

// Close stops the client and flushes all pending logs.
func (c *Client) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	c.mu.Unlock()

	// Stop batcher (will flush remaining logs)
	return c.batcher.Stop()
}
