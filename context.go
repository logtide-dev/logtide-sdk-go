package logtide

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// extractTraceID extracts the trace ID from the context if an OpenTelemetry span is present.
func extractTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span == nil || !span.SpanContext().IsValid() {
		return ""
	}

	return span.SpanContext().TraceID().String()
}

// extractSpanID extracts the span ID from the context if an OpenTelemetry span is present.
func extractSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span == nil || !span.SpanContext().IsValid() {
		return ""
	}

	return span.SpanContext().SpanID().String()
}

// enrichLogWithContext enriches a log entry with trace and span IDs from the context.
func enrichLogWithContext(ctx context.Context, log *Log) {
	// Only extract if not already set
	if log.TraceID == "" {
		log.TraceID = extractTraceID(ctx)
	}
	if log.SpanID == "" {
		log.SpanID = extractSpanID(ctx)
	}
}
