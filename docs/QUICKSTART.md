# Quick Start Guide

Get started with the LogTide Go SDK in minutes!

## Basic Setup

### 1. Install the SDK

```bash
go get github.com/logtide-dev/logtide-sdk-go
```

### 2. Import and Initialize

```go
package main

import (
    "context"
    "log"

    "github.com/logtide-dev/logtide-sdk-go"
)

func main() {
    // Create client
    client, err := logtide.New(
        logtide.WithAPIKey("lp_your_api_key"),
        logtide.WithService("my-service"),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close() // Important: flushes buffered logs

    ctx := context.Background()

    // Send logs
    client.Info(ctx, "Application started", nil)
}
```

### 3. Run Your Application

```bash
go run main.go
```

Check your [LogTide dashboard](https://app.logtide.dev) to see your logs!

## Leveled Logging

The SDK provides five log levels:

```go
// Debug - Detailed debugging information
client.Debug(ctx, "Debug message", map[string]interface{}{
    "variable": value,
})

// Info - General informational messages
client.Info(ctx, "User logged in", map[string]interface{}{
    "user_id": 123,
})

// Warn - Warning messages
client.Warn(ctx, "High memory usage", map[string]interface{}{
    "usage_mb": 850,
})

// Error - Error events
client.Error(ctx, "Database connection failed", map[string]interface{}{
    "error": err.Error(),
})

// Critical - Critical system errors
client.Critical(ctx, "System shutdown", nil)
```

## Structured Logging

Add metadata to provide context:

```go
client.Info(ctx, "User action", map[string]interface{}{
    "user_id":   12345,
    "action":    "purchase",
    "amount":    99.99,
    "currency":  "USD",
    "timestamp": time.Now(),
})
```

## Configuration Options

Customize the client behavior:

```go
client, err := logtide.New(
    // Required
    logtide.WithAPIKey("lp_your_api_key"),
    logtide.WithService("my-service"),

    // Optional
    logtide.WithBaseURL("https://api.logtide.dev"),
    logtide.WithBatchSize(100),                    // Max logs per batch
    logtide.WithFlushInterval(5*time.Second),      // Flush interval
    logtide.WithTimeout(30*time.Second),           // HTTP timeout
    logtide.WithRetry(3, 1*time.Second, 60*time.Second), // Retry config
    logtide.WithCircuitBreaker(5, 30*time.Second), // Circuit breaker
)
```

## Common Patterns

### Pattern 1: Application Logging

```go
func main() {
    client := setupLogtide()
    defer client.Close()

    client.Info(ctx, "Application starting", map[string]interface{}{
        "version": "1.0.0",
        "env":     "production",
    })

    // Your application logic

    client.Info(ctx, "Application stopping", nil)
}
```

### Pattern 2: Error Handling

```go
result, err := doSomething()
if err != nil {
    client.Error(ctx, "Operation failed", map[string]interface{}{
        "operation": "doSomething",
        "error":     err.Error(),
    })
    return err
}

client.Info(ctx, "Operation succeeded", map[string]interface{}{
    "result": result,
})
```

### Pattern 3: Request Logging

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    start := time.Now()

    // Process request

    client.Info(r.Context(), "Request completed", map[string]interface{}{
        "method":      r.Method,
        "path":        r.URL.Path,
        "duration_ms": time.Since(start).Milliseconds(),
        "status":      200,
    })
}
```

### Pattern 4: Distributed Tracing

```go
import "go.opentelemetry.io/otel"

func processOrder(ctx context.Context, orderID string) {
    // Create span
    ctx, span := otel.Tracer("my-service").Start(ctx, "process-order")
    defer span.End()

    // Trace ID automatically included!
    client.Info(ctx, "Processing order", map[string]interface{}{
        "order_id": orderID,
    })
}
```

## Best Practices

### 1. Always Close the Client

```go
client, err := logtide.New(...)
if err != nil {
    return err
}
defer client.Close() // Ensures buffered logs are flushed
```

### 2. Use Context

```go
// Pass context for cancellation support
client.Info(ctx, "message", metadata)

// Not recommended
client.Info(context.Background(), "message", metadata)
```

### 3. Structure Your Metadata

```go
// Good: Structured data
client.Info(ctx, "User registered", map[string]interface{}{
    "user_id": 123,
    "email":   "user@example.com",
    "source":  "web",
})

// Bad: Everything in the message
client.Info(ctx, "User 123 (user@example.com) registered from web", nil)
```

### 4. Choose Appropriate Log Levels

- **Debug**: Development debugging only
- **Info**: Normal application flow
- **Warn**: Potentially problematic situations
- **Error**: Error events that don't stop the application
- **Critical**: Severe errors requiring immediate attention

### 5. Don't Log Sensitive Data

```go
// Bad: Logging passwords
client.Info(ctx, "Login attempt", map[string]interface{}{
    "password": password, // ❌ Never log passwords
})

// Good: Log only safe information
client.Info(ctx, "Login attempt", map[string]interface{}{
    "username": username,
    "success":  true,
})
```

## Performance Tips

### 1. Automatic Batching

Logs are automatically batched for performance. No need to batch manually.

### 2. Non-Blocking

All logging operations are non-blocking. The SDK uses background goroutines for flushing.

### 3. Manual Flush (Optional)

For critical logs, you can force an immediate flush:

```go
client.Critical(ctx, "System error", metadata)
client.Flush(ctx) // Ensure it's sent immediately
```

## Error Handling

```go
err := client.Info(ctx, "message", metadata)
if err != nil {
    switch {
    case errors.Is(err, logtide.ErrClientClosed):
        // Client was closed
    case errors.Is(err, logtide.ErrCircuitOpen):
        // Circuit breaker is open
    case errors.Is(err, logtide.ErrInvalidAPIKey):
        // Invalid API key
    default:
        // Other errors
        log.Printf("Logging error: %v", err)
    }
}
```

## Testing

### Mock for Testing

```go
// In tests, you can skip logging or use a mock
func TestMyFunction(t *testing.T) {
    // Option 1: Use a test client that doesn't send logs
    client, _ := logtide.New(
        logtide.WithAPIKey("lp_test_key"),
        logtide.WithService("test"),
        logtide.WithBaseURL("http://localhost:9999"), // Non-existent
    )
    defer client.Close()

    // Run your tests
}
```

## Next Steps

- Explore [Framework Integrations](INTEGRATIONS.md) for Gin, Echo, and stdlib
- Check out complete [Examples](../examples/)
- Review the [API Reference](https://pkg.go.dev/github.com/logtide-dev/logtide-sdk-go)
- Learn about [Advanced Configuration](../README.md#configuration)
