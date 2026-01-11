# OpenTelemetry Integration Example

This example demonstrates how LogTide automatically integrates with OpenTelemetry for distributed tracing.

## Running the Example

```bash
# Install dependencies
go mod download

# Set your API key (edit main.go or use environment variable)
# Then run the example
go run main.go
```

## What it Demonstrates

- OpenTelemetry tracer setup
- Automatic trace ID and span ID extraction
- Nested spans and trace propagation
- Distributed tracing with LogTide logs
- Parent-child span relationships

## How it Works

When you create a span with OpenTelemetry:

```go
ctx, span := tracer.Start(ctx, "operation-name")
defer span.End()

// Logs sent with this context automatically include trace_id and span_id
client.Info(ctx, "Processing...", metadata)
```

The LogTide SDK automatically:
1. Extracts the trace ID from the OpenTelemetry span context
2. Extracts the span ID
3. Includes them in the log entry

This allows you to:
- Correlate logs with traces in LogTide
- Track requests across services
- Debug distributed systems more effectively

## Expected Output

The example will:
1. Print trace exports to stdout (from OpenTelemetry)
2. Send logs to LogTide with trace IDs included
3. Demonstrate nested operations with parent-child relationships

Check your LogTide dashboard to see logs grouped by trace ID!
