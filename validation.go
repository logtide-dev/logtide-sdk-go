package logtide

import (
	"fmt"
	"regexp"
)

var (
	// spanIDRegex validates that span IDs are exactly 16 hexadecimal characters.
	spanIDRegex = regexp.MustCompile(`^[a-fA-F0-9]{16}$`)

	// validLogLevels contains the set of acceptable log levels.
	validLogLevels = map[LogLevel]bool{
		LogLevelDebug:    true,
		LogLevelInfo:     true,
		LogLevelWarn:     true,
		LogLevelError:    true,
		LogLevelCritical: true,
	}
)

// validateLog validates a single log entry according to LogTide's requirements.
func validateLog(log *Log) error {
	// Validate service name
	if len(log.Service) == 0 {
		return &ValidationError{Field: "service", Message: "service name is required"}
	}
	if len(log.Service) > 100 {
		return &ValidationError{Field: "service", Message: "service name must be 100 characters or less"}
	}

	// Validate message
	if len(log.Message) == 0 {
		return &ValidationError{Field: "message", Message: "message is required"}
	}

	// Validate log level
	if !validLogLevels[log.Level] {
		return &ValidationError{
			Field:   "level",
			Message: fmt.Sprintf("invalid log level: %s (must be one of: debug, info, warn, error, critical)", log.Level),
		}
	}

	// Validate span ID format if provided
	if log.SpanID != "" && !spanIDRegex.MatchString(log.SpanID) {
		return &ValidationError{
			Field:   "span_id",
			Message: "span_id must be exactly 16 hexadecimal characters",
		}
	}

	return nil
}

// validateBatch validates a batch of logs according to LogTide's requirements.
func validateBatch(logs []Log) error {
	if len(logs) == 0 {
		return &ValidationError{Field: "logs", Message: "at least one log is required"}
	}
	if len(logs) > 1000 {
		return &ValidationError{Field: "logs", Message: "batch size must be 1000 logs or less"}
	}

	// Validate each log in the batch
	for i, log := range logs {
		if err := validateLog(&log); err != nil {
			return fmt.Errorf("log at index %d: %w", i, err)
		}
	}

	return nil
}
