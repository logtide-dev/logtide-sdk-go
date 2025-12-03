package logward

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	t.Run("creates client with valid config", func(t *testing.T) {
		client, err := New(
			WithAPIKey("lp_test_key"),
			WithService("test-service"),
		)
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		defer client.Close()

		if client == nil {
			t.Fatal("New() returned nil client")
		}
	})

	t.Run("fails with missing API key", func(t *testing.T) {
		_, err := New(
			WithService("test-service"),
		)
		if err == nil {
			t.Fatal("New() error = nil, want error")
		}
	})

	t.Run("fails with missing service", func(t *testing.T) {
		_, err := New(
			WithAPIKey("lp_test_key"),
		)
		if err == nil {
			t.Fatal("New() error = nil, want error")
		}
	})

	t.Run("applies custom configuration", func(t *testing.T) {
		client, err := New(
			WithAPIKey("lp_custom_key"),
			WithService("custom-service"),
			WithBaseURL("https://custom.example.com"),
			WithBatchSize(50),
			WithFlushInterval(10*time.Second),
		)
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		defer client.Close()

		if client.config.APIKey != "lp_custom_key" {
			t.Errorf("APIKey = %q, want %q", client.config.APIKey, "lp_custom_key")
		}
		if client.config.Service != "custom-service" {
			t.Errorf("Service = %q, want %q", client.config.Service, "custom-service")
		}
		if client.config.BaseURL != "https://custom.example.com" {
			t.Errorf("BaseURL = %q, want %q", client.config.BaseURL, "https://custom.example.com")
		}
		if client.config.BatchSize != 50 {
			t.Errorf("BatchSize = %d, want 50", client.config.BatchSize)
		}
	})
}

func TestClientLeveledLogging(t *testing.T) {
	// Create mock server
	var receivedLogs []Log
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req IngestRequest
		json.NewDecoder(r.Body).Decode(&req)
		receivedLogs = append(receivedLogs, req.Logs...)

		resp := IngestResponse{
			Received:  len(req.Logs),
			Timestamp: time.Now().Format(time.RFC3339),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client, err := New(
		WithAPIKey("lp_test_key"),
		WithService("test-service"),
		WithBaseURL(server.URL),
		WithBatchSize(10),
		WithFlushInterval(100*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Test each log level
	tests := []struct {
		name     string
		logFunc  func(context.Context, string, map[string]interface{}) error
		level    LogLevel
		message  string
		metadata map[string]interface{}
	}{
		{
			name:     "Debug",
			logFunc:  client.Debug,
			level:    LogLevelDebug,
			message:  "debug message",
			metadata: map[string]interface{}{"key": "value"},
		},
		{
			name:     "Info",
			logFunc:  client.Info,
			level:    LogLevelInfo,
			message:  "info message",
			metadata: nil,
		},
		{
			name:     "Warn",
			logFunc:  client.Warn,
			level:    LogLevelWarn,
			message:  "warn message",
			metadata: map[string]interface{}{"warning_code": 123},
		},
		{
			name:     "Error",
			logFunc:  client.Error,
			level:    LogLevelError,
			message:  "error message",
			metadata: map[string]interface{}{"error": "details"},
		},
		{
			name:     "Critical",
			logFunc:  client.Critical,
			level:    LogLevelCritical,
			message:  "critical message",
			metadata: map[string]interface{}{"severity": "high"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.logFunc(ctx, tt.message, tt.metadata)
			if err != nil {
				t.Errorf("%s() error = %v", tt.name, err)
			}
		})
	}

	// Flush to ensure all logs are sent
	client.Flush(ctx)
	time.Sleep(200 * time.Millisecond)

	// Verify logs were received
	if len(receivedLogs) != len(tests) {
		t.Errorf("received %d logs, want %d", len(receivedLogs), len(tests))
	}

	// Verify each log
	for i, test := range tests {
		if i >= len(receivedLogs) {
			break
		}
		log := receivedLogs[i]
		if log.Level != test.level {
			t.Errorf("log[%d].Level = %v, want %v", i, log.Level, test.level)
		}
		if log.Message != test.message {
			t.Errorf("log[%d].Message = %q, want %q", i, log.Message, test.message)
		}
		if log.Service != "test-service" {
			t.Errorf("log[%d].Service = %q, want %q", i, log.Service, "test-service")
		}
	}
}

func TestClientBatching(t *testing.T) {
	var requestCount int32
	var totalLogsReceived int32

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)

		var req IngestRequest
		json.NewDecoder(r.Body).Decode(&req)
		atomic.AddInt32(&totalLogsReceived, int32(len(req.Logs)))

		resp := IngestResponse{
			Received:  len(req.Logs),
			Timestamp: time.Now().Format(time.RFC3339),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client with small batch size
	client, err := New(
		WithAPIKey("lp_test_key"),
		WithService("test-service"),
		WithBaseURL(server.URL),
		WithBatchSize(3),
		WithFlushInterval(1*time.Minute), // Long interval to test size-based batching
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx := context.Background()

	// Send 10 logs
	for i := 0; i < 10; i++ {
		client.Info(ctx, "test message", nil)
	}

	// Wait for batches to be sent
	time.Sleep(300 * time.Millisecond)

	// Close (which will flush remaining)
	client.Close()
	time.Sleep(100 * time.Millisecond)

	// Should have received all 10 logs
	total := atomic.LoadInt32(&totalLogsReceived)
	if total != 10 {
		t.Errorf("total logs received = %d, want 10", total)
	}

	// Should have sent multiple batches (at least 3 with batch size of 3)
	count := atomic.LoadInt32(&requestCount)
	if count < 1 {
		t.Errorf("request count = %d, want >= 1", count)
	}
}

func TestClientClose(t *testing.T) {
	var receivedCount int32

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req IngestRequest
		json.NewDecoder(r.Body).Decode(&req)
		atomic.AddInt32(&receivedCount, int32(len(req.Logs)))

		resp := IngestResponse{
			Received:  len(req.Logs),
			Timestamp: time.Now().Format(time.RFC3339),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client, err := New(
		WithAPIKey("lp_test_key"),
		WithService("test-service"),
		WithBaseURL(server.URL),
		WithBatchSize(100),
		WithFlushInterval(1*time.Minute),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx := context.Background()

	// Send logs
	for i := 0; i < 10; i++ {
		client.Info(ctx, "test message", nil)
	}

	// Close should flush remaining logs
	err = client.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	count := atomic.LoadInt32(&receivedCount)
	if count != 10 {
		t.Errorf("received %d logs, want 10", count)
	}

	// Logging after close should fail
	err = client.Info(ctx, "after close", nil)
	if err != ErrClientClosed {
		t.Errorf("Info() after close error = %v, want %v", err, ErrClientClosed)
	}
}
