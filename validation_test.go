package logward

import (
	"strings"
	"testing"
	"time"
)

func TestValidateLog(t *testing.T) {
	tests := []struct {
		name    string
		log     *Log
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid log",
			log: &Log{
				Time:    time.Now(),
				Service: "test-service",
				Level:   LogLevelInfo,
				Message: "test message",
			},
			wantErr: false,
		},
		{
			name: "valid log with metadata",
			log: &Log{
				Time:    time.Now(),
				Service: "test-service",
				Level:   LogLevelError,
				Message: "error occurred",
				Metadata: map[string]interface{}{
					"user_id": 123,
					"action":  "login",
				},
			},
			wantErr: false,
		},
		{
			name: "valid log with trace and span IDs",
			log: &Log{
				Time:    time.Now(),
				Service: "test-service",
				Level:   LogLevelDebug,
				Message: "debug message",
				TraceID: "trace-123",
				SpanID:  "0123456789abcdef",
			},
			wantErr: false,
		},
		{
			name: "missing service",
			log: &Log{
				Time:    time.Now(),
				Level:   LogLevelInfo,
				Message: "test message",
			},
			wantErr: true,
			errMsg:  "service name is required",
		},
		{
			name: "service too long",
			log: &Log{
				Time:    time.Now(),
				Service: strings.Repeat("a", 101),
				Level:   LogLevelInfo,
				Message: "test message",
			},
			wantErr: true,
			errMsg:  "service name must be 100 characters or less",
		},
		{
			name: "missing message",
			log: &Log{
				Time:    time.Now(),
				Service: "test-service",
				Level:   LogLevelInfo,
			},
			wantErr: true,
			errMsg:  "message is required",
		},
		{
			name: "invalid log level",
			log: &Log{
				Time:    time.Now(),
				Service: "test-service",
				Level:   LogLevel("invalid"),
				Message: "test message",
			},
			wantErr: true,
			errMsg:  "invalid log level",
		},
		{
			name: "invalid span ID - too short",
			log: &Log{
				Time:    time.Now(),
				Service: "test-service",
				Level:   LogLevelInfo,
				Message: "test message",
				SpanID:  "abc123",
			},
			wantErr: true,
			errMsg:  "span_id must be exactly 16 hexadecimal characters",
		},
		{
			name: "invalid span ID - invalid characters",
			log: &Log{
				Time:    time.Now(),
				Service: "test-service",
				Level:   LogLevelInfo,
				Message: "test message",
				SpanID:  "0123456789abcdez",
			},
			wantErr: true,
			errMsg:  "span_id must be exactly 16 hexadecimal characters",
		},
		{
			name: "valid span ID - uppercase",
			log: &Log{
				Time:    time.Now(),
				Service: "test-service",
				Level:   LogLevelInfo,
				Message: "test message",
				SpanID:  "0123456789ABCDEF",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLog(tt.log)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateLog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateLog() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestValidateBatch(t *testing.T) {
	validLog := &Log{
		Time:    time.Now(),
		Service: "test-service",
		Level:   LogLevelInfo,
		Message: "test message",
	}

	tests := []struct {
		name    string
		logs    []Log
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid batch with one log",
			logs:    []Log{*validLog},
			wantErr: false,
		},
		{
			name:    "valid batch with multiple logs",
			logs:    []Log{*validLog, *validLog, *validLog},
			wantErr: false,
		},
		{
			name:    "empty batch",
			logs:    []Log{},
			wantErr: true,
			errMsg:  "at least one log is required",
		},
		{
			name:    "batch too large",
			logs:    make([]Log, 1001),
			wantErr: true,
			errMsg:  "batch size must be 1000 logs or less",
		},
		{
			name: "batch with invalid log",
			logs: []Log{
				*validLog,
				{
					Time:    time.Now(),
					Service: "",
					Level:   LogLevelInfo,
					Message: "test",
				},
			},
			wantErr: true,
			errMsg:  "log at index 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBatch(tt.logs)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateBatch() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}
