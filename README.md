<p align="center">
  <img src="https://raw.githubusercontent.com/logtide-dev/logtide/main/docs/images/logo.png" alt="LogTide Logo" width="400">
</p>

<h1 align="center">LogTide Go SDK</h1>

<p align="center">
  <a href="https://pkg.go.dev/github.com/logtide-dev/logtide-sdk-go"><img src="https://pkg.go.dev/badge/github.com/logtide-dev/logtide-sdk-go.svg" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/logtide-dev/logtide-sdk-go"><img src="https://goreportcard.com/badge/github.com/logtide-dev/logtide-sdk-go" alt="Go Report Card"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License"></a>
  <a href="https://github.com/logtide-dev/logtide-sdk-go/releases"><img src="https://img.shields.io/github/v/release/logtide-dev/logtide-sdk-go" alt="Release"></a>
</p>

<p align="center">
  Official Go SDK for <a href="https://logtide.dev">LogTide</a> with automatic batching, retry logic, circuit breaker, OpenTelemetry integration, and production-ready features.
</p>

---

## Features

- **Leveled Logging** - Debug, Info, Warn, Error, Critical methods
- **Automatic Batching** - Configurable batch size and flush interval
- **Retry Logic** - Exponential backoff with jitter
- **Circuit Breaker** - Prevents cascading failures
- **Graceful Shutdown** - Flushes buffered logs on Close()
- **Context Support** - Respects context cancellation
- **OpenTelemetry Integration** - Automatic trace ID extraction
- **Production Ready** - Thread-safe, well-tested (~87% coverage)

## Requirements

- Go 1.21 or later
- LogTide account and API key

## Installation

```bash
go get github.com/logtide-dev/logtide-sdk-go
```

## Quick Start

```go
package main

import (
    "context"
    logtide "github.com/logtide-dev/logtide-sdk-go"
)

func main() {
    client, _ := logtide.New(
        logtide.WithAPIKey("lp_your_api_key"),
        logtide.WithService("my-service"),
    )
    defer client.Close()

    client.Info(context.Background(), "Hello LogTide!", nil)
}
```

**That's it!** See [Quick Start Guide](./docs/QUICKSTART.md) for detailed tutorial.

---

## Documentation

Complete documentation is available in the [docs](./docs) directory:

- **[Installation Guide](./docs/INSTALLATION.md)** - Detailed installation instructions, API keys, troubleshooting
- **[Quick Start Guide](./docs/QUICKSTART.md)** - Tutorial with patterns and best practices
- **[Framework Integrations](./docs/INTEGRATIONS.md)** - Gin, Echo, Chi, Fiber, Standard Library, OpenTelemetry

---

## Configuration Options

Customize the client behavior:

```go
client, err := logtide.New(
    // Required
    logtide.WithAPIKey("lp_your_api_key"),
    logtide.WithService("my-service"),

    // Optional customization
    logtide.WithBaseURL("https://api.logtide.dev"),
    logtide.WithBatchSize(100),                              // Max logs per batch
    logtide.WithFlushInterval(5*time.Second),                // Flush interval
    logtide.WithTimeout(30*time.Second),                     // HTTP timeout
    logtide.WithRetry(3, 1*time.Second, 60*time.Second),     // Max retries, min/max backoff
    logtide.WithCircuitBreaker(5, 30*time.Second),           // Failure threshold, timeout
)
```

**Defaults:**
- Base URL: `https://api.logtide.dev`
- Batch Size: 100 logs
- Flush Interval: 5 seconds
- Timeout: 30 seconds
- Max Retries: 3 attempts with exponential backoff
- Circuit Breaker: Opens after 5 failures for 30 seconds

---

## Logging Methods

### Basic Logging

```go
ctx := context.Background()

client.Debug(ctx, "Debug message", nil)
client.Info(ctx, "Info message", map[string]any{"userId": 123})
client.Warn(ctx, "Warning message", nil)
client.Error(ctx, "Error message", map[string]any{"custom": "data"})
client.Critical(ctx, "Critical message", nil)
```

### With Metadata

```go
client.Info(ctx, "User logged in", map[string]any{
    "userId":    123,
    "email":     "user@example.com",
    "ip":        "192.168.1.1",
    "userAgent": "Mozilla/5.0...",
})
```

---

## OpenTelemetry Integration

Trace IDs are automatically extracted from context:

```go
ctx, span := tracer.Start(ctx, "operation")
defer span.End()

// trace_id and span_id automatically included!
client.Info(ctx, "Processing", metadata)
```

See [examples/otel](./examples/otel) for complete example.

---

## Error Handling

```go
err := client.Info(ctx, "message", nil)
if err != nil {
    switch {
    case errors.Is(err, logtide.ErrClientClosed):
        // Client was closed
    case errors.Is(err, logtide.ErrCircuitOpen):
        // Circuit breaker is open (too many failures)
    case errors.Is(err, logtide.ErrInvalidAPIKey):
        // Invalid API key
    default:
        // Handle other errors
    }
}
```

---

## Framework Integration

Quick integration examples (full code in [docs/INTEGRATIONS.md](./docs/INTEGRATIONS.md)):

### Gin

```go
r.Use(LogtideMiddleware(client))
```

### Echo

```go
e.Use(LogtideMiddleware(client))
```

### Standard Library

```go
handler := LoggingMiddleware(client, mux)
http.ListenAndServe(":8080", handler)
```

---

## Examples

Complete working examples with full source code:

| Example | Description | Link |
|---------|-------------|------|
| **Basic** | Simple usage with all log levels | [examples/basic](./examples/basic) |
| **Gin** | Gin framework middleware integration | [examples/gin](./examples/gin) |
| **Echo** | Echo framework middleware integration | [examples/echo](./examples/echo) |
| **Standard Library** | net/http middleware | [examples/stdlib](./examples/stdlib) |
| **OpenTelemetry** | Distributed tracing integration | [examples/otel](./examples/otel) |

Each example includes a README with running instructions.

---

## Key Features Explained

### Automatic Batching

Logs are automatically batched for optimal performance:
- Batches flush when size limit is reached (default: 100 logs)
- Batches flush on interval (default: 5 seconds)
- Manual flush with `client.Flush(ctx)`
- All pending logs flushed on `client.Close()`

### Circuit Breaker

Prevents cascading failures when the logging service is unavailable:
- Opens after consecutive failures (default: 5)
- Allows test request after timeout (default: 30s)
- Automatically closes when service recovers

### Performance

- **Non-blocking** - Logging doesn't block your application
- **Automatic batching** - Reduces HTTP overhead
- **Connection pooling** - Reuses HTTP connections
- **Thread-safe** - Safe for concurrent use
- **Context-aware** - Respects cancellation

---

## API Reference

Full API documentation with godoc:

- **Online:** [pkg.go.dev/github.com/logtide-dev/logtide-sdk-go](https://pkg.go.dev/github.com/logtide-dev/logtide-sdk-go)
- **Local:** Run `godoc -http=:6060` and visit http://localhost:6060/pkg/github.com/logtide-dev/logtide-sdk-go/

---

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Links

- [LogTide Website](https://logtide.dev)
- [Documentation](https://logtide.dev/docs/sdks/go/)
- [API Reference](https://pkg.go.dev/github.com/logtide-dev/logtide-sdk-go)
- [GitHub Issues](https://github.com/logtide-dev/logtide-sdk-go/issues)
