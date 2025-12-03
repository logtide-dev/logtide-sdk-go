package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

// Client wraps an HTTP client with LogWard-specific configuration.
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	timeout    time.Duration
}

// Config holds the configuration for the HTTP client.
type Config struct {
	BaseURL        string
	APIKey         string
	Timeout        time.Duration
	MaxIdleConns   int
	IdleConnTimeout time.Duration
	TLSMinVersion  uint16
}

// NewClient creates a new HTTP client with the specified configuration.
func NewClient(cfg *Config) *Client {
	// Set defaults
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 10
	}
	if cfg.IdleConnTimeout == 0 {
		cfg.IdleConnTimeout = 90 * time.Second
	}
	if cfg.TLSMinVersion == 0 {
		cfg.TLSMinVersion = tls.VersionTLS12
	}

	// Create transport with custom settings
	transport := &http.Transport{
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConns,
		IdleConnTimeout:     cfg.IdleConnTimeout,
		TLSClientConfig: &tls.Config{
			MinVersion: cfg.TLSMinVersion,
		},
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	return &Client{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   cfg.Timeout,
		},
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		timeout: cfg.Timeout,
	}
}

// Post sends a POST request to the specified path with the given payload.
func (c *Client) Post(ctx context.Context, path string, payload interface{}) (*http.Response, error) {
	// Marshal payload to JSON
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create request
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("User-Agent", "logward-sdk-go/0.1.0")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil
}

// DecodeResponse decodes the JSON response body into the provided target.
func DecodeResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// ReadResponseBody reads the entire response body and returns it as a string.
func ReadResponseBody(resp *http.Response) (string, error) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}
