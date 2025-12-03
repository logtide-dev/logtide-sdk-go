package logward

import "time"

// LogLevel represents the severity level of a log entry.
type LogLevel string

const (
	// LogLevelDebug represents debug-level logs for detailed debugging information.
	LogLevelDebug LogLevel = "debug"

	// LogLevelInfo represents informational messages that highlight the progress of the application.
	LogLevelInfo LogLevel = "info"

	// LogLevelWarn represents potentially harmful situations.
	LogLevelWarn LogLevel = "warn"

	// LogLevelError represents error events that might still allow the application to continue running.
	LogLevelError LogLevel = "error"

	// LogLevelCritical represents very severe error events that will presumably lead the application to abort.
	LogLevelCritical LogLevel = "critical"
)

// Log represents a single log entry to be sent to LogWard.
type Log struct {
	// Time is the timestamp of the log entry. If not set, the current time will be used.
	Time time.Time `json:"time"`

	// Service is the name of the service generating the log (1-100 characters, required).
	Service string `json:"service"`

	// Level is the severity level of the log entry (required).
	Level LogLevel `json:"level"`

	// Message is the log message (minimum 1 character, required).
	Message string `json:"message"`

	// Metadata contains additional structured data associated with the log entry (optional).
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// TraceID is the W3C trace ID for distributed tracing (optional).
	TraceID string `json:"trace_id,omitempty"`

	// SpanID is the W3C span ID, must be exactly 16 hex characters if provided (optional).
	SpanID string `json:"span_id,omitempty"`
}

// IngestRequest represents the request payload for batch log ingestion.
type IngestRequest struct {
	// Logs is the array of log entries to ingest (1-1000 logs per request).
	Logs []Log `json:"logs"`
}

// IngestResponse represents the response from the log ingestion API.
type IngestResponse struct {
	// Received is the number of logs successfully received.
	Received int `json:"received"`

	// Timestamp is the server timestamp when the logs were processed.
	Timestamp string `json:"timestamp"`
}
