# Framework Integrations

This guide shows how to integrate the LogWard Go SDK with popular Go web frameworks.

## Table of Contents

- [Gin](#gin)
- [Echo](#echo)
- [Standard Library](#standard-library-nethttp)
- [Chi](#chi)
- [Fiber](#fiber)
- [OpenTelemetry](#opentelemetry)

---

## Gin

### Installation

```bash
go get github.com/gin-gonic/gin
go get github.com/logward-dev/logward-sdk-go
```

### Middleware Implementation

```go
package main

import (
    "time"
    "github.com/gin-gonic/gin"
    "github.com/logward-dev/logward-sdk-go"
)

func LogwardMiddleware(client *logward.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()

        // Process request
        c.Next()

        // Log after request
        duration := time.Since(start)

        client.Info(c.Request.Context(), "Request completed", map[string]interface{}{
            "method":      c.Request.Method,
            "path":        c.Request.URL.Path,
            "status":      c.Writer.Status(),
            "duration_ms": duration.Milliseconds(),
            "ip":          c.ClientIP(),
        })
    }
}

func main() {
    client, _ := logward.New(
        logward.WithAPIKey("lp_your_api_key"),
        logward.WithService("gin-api"),
    )
    defer client.Close()

    r := gin.Default()
    r.Use(LogwardMiddleware(client))

    r.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "Hello!"})
    })

    r.Run(":8080")
}
```

**Complete Example:** [examples/gin/](../examples/gin/)

---

## Echo

### Installation

```bash
go get github.com/labstack/echo/v4
go get github.com/logward-dev/logward-sdk-go
```

### Middleware Implementation

```go
package main

import (
    "time"
    "github.com/labstack/echo/v4"
    "github.com/logward-dev/logward-sdk-go"
)

func LogwardMiddleware(client *logward.Client) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            start := time.Now()

            err := next(c)

            duration := time.Since(start)

            client.Info(c.Request().Context(), "Request completed", map[string]interface{}{
                "method":      c.Request().Method,
                "path":        c.Request().URL.Path,
                "status":      c.Response().Status,
                "duration_ms": duration.Milliseconds(),
                "ip":          c.RealIP(),
            })

            return err
        }
    }
}

func main() {
    client, _ := logward.New(
        logward.WithAPIKey("lp_your_api_key"),
        logward.WithService("echo-api"),
    )
    defer client.Close()

    e := echo.New()
    e.Use(LogwardMiddleware(client))

    e.GET("/", func(c echo.Context) error {
        return c.JSON(200, map[string]string{"message": "Hello!"})
    })

    e.Start(":8080")
}
```

**Complete Example:** [examples/echo/](../examples/echo/)

---

## Standard Library (net/http)

### Basic Middleware

```go
package main

import (
    "net/http"
    "time"
    "github.com/logward-dev/logward-sdk-go"
)

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(client *logward.Client, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        rw := &responseWriter{ResponseWriter: w, statusCode: 200}
        next.ServeHTTP(rw, r)

        duration := time.Since(start)

        client.Info(r.Context(), "Request completed", map[string]interface{}{
            "method":      r.Method,
            "path":        r.URL.Path,
            "status":      rw.statusCode,
            "duration_ms": duration.Milliseconds(),
        })
    })
}

func main() {
    client, _ := logward.New(
        logward.WithAPIKey("lp_your_api_key"),
        logward.WithService("http-api"),
    )
    defer client.Close()

    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello!"))
    })

    handler := LoggingMiddleware(client, mux)
    http.ListenAndServe(":8080", handler)
}
```

**Complete Example:** [examples/stdlib/](../examples/stdlib/)

---

## Chi

### Installation

```bash
go get github.com/go-chi/chi/v5
go get github.com/logward-dev/logward-sdk-go
```

### Middleware Implementation

```go
package main

import (
    "net/http"
    "time"
    "github.com/go-chi/chi/v5"
    "github.com/logward-dev/logward-sdk-go"
)

func LogwardMiddleware(client *logward.Client) func(next http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()

            ww := chi.NewWrapResponseWriter(w, r.ProtoMajor)
            next.ServeHTTP(ww, r)

            duration := time.Since(start)

            client.Info(r.Context(), "Request completed", map[string]interface{}{
                "method":      r.Method,
                "path":        r.URL.Path,
                "status":      ww.Status(),
                "duration_ms": duration.Milliseconds(),
                "bytes":       ww.BytesWritten(),
            })
        })
    }
}

func main() {
    client, _ := logward.New(
        logward.WithAPIKey("lp_your_api_key"),
        logward.WithService("chi-api"),
    )
    defer client.Close()

    r := chi.NewRouter()
    r.Use(LogwardMiddleware(client))

    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello!"))
    })

    http.ListenAndServe(":8080", r)
}
```

---

## Fiber

### Installation

```bash
go get github.com/gofiber/fiber/v2
go get github.com/logward-dev/logward-sdk-go
```

### Middleware Implementation

```go
package main

import (
    "time"
    "github.com/gofiber/fiber/v2"
    "github.com/logward-dev/logward-sdk-go"
)

func LogwardMiddleware(client *logward.Client) fiber.Handler {
    return func(c *fiber.Ctx) error {
        start := time.Now()

        err := c.Next()

        duration := time.Since(start)

        client.Info(c.UserContext(), "Request completed", map[string]interface{}{
            "method":      c.Method(),
            "path":        c.Path(),
            "status":      c.Response().StatusCode(),
            "duration_ms": duration.Milliseconds(),
            "ip":          c.IP(),
        })

        return err
    }
}

func main() {
    client, _ := logward.New(
        logward.WithAPIKey("lp_your_api_key"),
        logward.WithService("fiber-api"),
    )
    defer client.Close()

    app := fiber.New()
    app.Use(LogwardMiddleware(client))

    app.Get("/", func(c *fiber.Ctx) error {
        return c.SendString("Hello!")
    })

    app.Listen(":8080")
}
```

---

## OpenTelemetry

### Automatic Trace ID Extraction

The SDK automatically extracts trace and span IDs from OpenTelemetry contexts:

```go
package main

import (
    "context"
    "go.opentelemetry.io/otel"
    "github.com/logward-dev/logward-sdk-go"
)

func main() {
    // Setup OpenTelemetry (tracer provider, etc.)
    tracer := otel.Tracer("my-service")

    client, _ := logward.New(
        logward.WithAPIKey("lp_your_api_key"),
        logward.WithService("traced-api"),
    )
    defer client.Close()

    ctx, span := tracer.Start(context.Background(), "operation")
    defer span.End()

    // Trace ID and Span ID automatically included!
    client.Info(ctx, "Processing request", map[string]interface{}{
        "user_id": 123,
    })
}
```

### Manual Trace ID Injection

If you're not using OpenTelemetry but have trace IDs:

```go
client.Info(ctx, "Processing", map[string]interface{}{
    "trace_id": myTraceID,
    "span_id":  mySpanID,
})
```

**Complete Example:** [examples/otel/](../examples/otel/)

---

## Advanced Patterns

### Request ID Middleware

```go
func RequestIDMiddleware(client *logward.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := c.GetHeader("X-Request-ID")
        if requestID == "" {
            requestID = generateRequestID()
        }

        c.Set("request_id", requestID)
        c.Next()

        client.Info(c.Request.Context(), "Request processed", map[string]interface{}{
            "request_id": requestID,
            "path":       c.Request.URL.Path,
        })
    }
}
```

### Error Recovery Middleware

```go
func RecoveryMiddleware(client *logward.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                client.Critical(c.Request.Context(), "Panic recovered", map[string]interface{}{
                    "error": err,
                    "path":  c.Request.URL.Path,
                })
                c.AbortWithStatus(500)
            }
        }()
        c.Next()
    }
}
```

### Authentication Logging

```go
func AuthMiddleware(client *logward.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")

        user, err := authenticate(token)
        if err != nil {
            client.Warn(c.Request.Context(), "Authentication failed", map[string]interface{}{
                "error": err.Error(),
                "ip":    c.ClientIP(),
            })
            c.AbortWithStatus(401)
            return
        }

        client.Info(c.Request.Context(), "User authenticated", map[string]interface{}{
            "user_id": user.ID,
        })

        c.Set("user", user)
        c.Next()
    }
}
```

---

## Best Practices

1. **One Client Per Application**: Create a single client and reuse it
2. **Use Context**: Always pass request context for trace propagation
3. **Log After Processing**: Log after the request is processed to include duration
4. **Structured Metadata**: Use metadata for searchable fields
5. **Appropriate Log Levels**: Use Warn/Error for failed requests
6. **Performance**: The SDK is non-blocking and batches automatically

## Next Steps

- Check out the [Examples Directory](../examples/) for complete working code
- Read the [Quick Start Guide](QUICKSTART.md)
- Review the [API Reference](https://pkg.go.dev/github.com/logward-dev/logward-sdk-go)
