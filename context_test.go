package logtide

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func TestExtractTraceID(t *testing.T) {
	// Create a tracer provider
	provider := trace.NewTracerProvider()
	tracer := provider.Tracer("test")

	t.Run("no span in context", func(t *testing.T) {
		ctx := context.Background()
		traceID := extractTraceID(ctx)
		if traceID != "" {
			t.Errorf("extractTraceID() = %q, want empty string", traceID)
		}
	})

	t.Run("valid span in context", func(t *testing.T) {
		ctx, span := tracer.Start(context.Background(), "test-span")
		defer span.End()

		traceID := extractTraceID(ctx)
		if traceID == "" {
			t.Error("extractTraceID() = empty, want non-empty trace ID")
		}

		// Verify it matches the span's trace ID
		expectedTraceID := span.SpanContext().TraceID().String()
		if traceID != expectedTraceID {
			t.Errorf("extractTraceID() = %q, want %q", traceID, expectedTraceID)
		}
	})

	t.Run("invalid span context", func(t *testing.T) {
		// Create a context with an invalid span
		ctx := oteltrace.ContextWithSpan(context.Background(), oteltrace.SpanFromContext(context.Background()))
		traceID := extractTraceID(ctx)
		if traceID != "" {
			t.Errorf("extractTraceID() with invalid span = %q, want empty string", traceID)
		}
	})
}

func TestExtractSpanID(t *testing.T) {
	// Create a tracer provider
	provider := trace.NewTracerProvider()
	tracer := provider.Tracer("test")

	t.Run("no span in context", func(t *testing.T) {
		ctx := context.Background()
		spanID := extractSpanID(ctx)
		if spanID != "" {
			t.Errorf("extractSpanID() = %q, want empty string", spanID)
		}
	})

	t.Run("valid span in context", func(t *testing.T) {
		ctx, span := tracer.Start(context.Background(), "test-span")
		defer span.End()

		spanID := extractSpanID(ctx)
		if spanID == "" {
			t.Error("extractSpanID() = empty, want non-empty span ID")
		}

		// Verify it matches the span's span ID
		expectedSpanID := span.SpanContext().SpanID().String()
		if spanID != expectedSpanID {
			t.Errorf("extractSpanID() = %q, want %q", spanID, expectedSpanID)
		}

		// Verify span ID is 16 hex characters
		if len(spanID) != 16 {
			t.Errorf("extractSpanID() length = %d, want 16", len(spanID))
		}
	})

	t.Run("invalid span context", func(t *testing.T) {
		// Create a context with an invalid span
		ctx := oteltrace.ContextWithSpan(context.Background(), oteltrace.SpanFromContext(context.Background()))
		spanID := extractSpanID(ctx)
		if spanID != "" {
			t.Errorf("extractSpanID() with invalid span = %q, want empty string", spanID)
		}
	})
}

func TestEnrichLogWithContext(t *testing.T) {
	// Create a tracer provider
	provider := trace.NewTracerProvider()
	tracer := provider.Tracer("test")

	t.Run("enriches log with trace and span IDs", func(t *testing.T) {
		ctx, span := tracer.Start(context.Background(), "test-span")
		defer span.End()

		log := &Log{
			Service: "test",
			Level:   LogLevelInfo,
			Message: "test message",
		}

		enrichLogWithContext(ctx, log)

		if log.TraceID == "" {
			t.Error("TraceID not set")
		}
		if log.SpanID == "" {
			t.Error("SpanID not set")
		}

		// Verify IDs match the span
		if log.TraceID != span.SpanContext().TraceID().String() {
			t.Errorf("TraceID = %q, want %q", log.TraceID, span.SpanContext().TraceID().String())
		}
		if log.SpanID != span.SpanContext().SpanID().String() {
			t.Errorf("SpanID = %q, want %q", log.SpanID, span.SpanContext().SpanID().String())
		}
	})

	t.Run("does not overwrite existing IDs", func(t *testing.T) {
		ctx, span := tracer.Start(context.Background(), "test-span")
		defer span.End()

		existingTraceID := "existing-trace-id"
		existingSpanID := "0123456789abcdef"

		log := &Log{
			Service: "test",
			Level:   LogLevelInfo,
			Message: "test message",
			TraceID: existingTraceID,
			SpanID:  existingSpanID,
		}

		enrichLogWithContext(ctx, log)

		// Should not overwrite
		if log.TraceID != existingTraceID {
			t.Errorf("TraceID = %q, want %q (should not overwrite)", log.TraceID, existingTraceID)
		}
		if log.SpanID != existingSpanID {
			t.Errorf("SpanID = %q, want %q (should not overwrite)", log.SpanID, existingSpanID)
		}
	})

	t.Run("handles context without span", func(t *testing.T) {
		ctx := context.Background()

		log := &Log{
			Service: "test",
			Level:   LogLevelInfo,
			Message: "test message",
		}

		enrichLogWithContext(ctx, log)

		// Should remain empty
		if log.TraceID != "" {
			t.Errorf("TraceID = %q, want empty", log.TraceID)
		}
		if log.SpanID != "" {
			t.Errorf("SpanID = %q, want empty", log.SpanID)
		}
	})
}
