package main

import (
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/logward-dev/logward-sdk-go"
)

func main() {
	// Create LogWard client
	client, err := logward.New(
		logward.WithAPIKey("lp_your_api_key_here"),
		logward.WithService("echo-example"),
	)
	if err != nil {
		log.Fatalf("Failed to create LogWard client: %v", err)
	}
	defer client.Close()

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Recover())
	e.Use(LogwardMiddleware(client))

	// Routes
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Hello from Echo!",
		})
	})

	e.GET("/user/:id", func(c echo.Context) error {
		userID := c.Param("id")

		// Log within handler
		client.Info(c.Request().Context(), "Fetching user details", map[string]interface{}{
			"user_id": userID,
		})

		return c.JSON(http.StatusOK, map[string]interface{}{
			"user_id": userID,
			"name":    "Jane Doe",
		})
	})

	e.POST("/login", func(c echo.Context) error {
		type LoginRequest struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		var req LoginRequest
		if err := c.Bind(&req); err != nil {
			client.Error(c.Request().Context(), "Invalid login request", map[string]interface{}{
				"error": err.Error(),
			})
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		}

		// Simulate login
		client.Info(c.Request().Context(), "User login attempt", map[string]interface{}{
			"username": req.Username,
			"success":  true,
		})

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Login successful",
			"token":   "sample-jwt-token",
		})
	})

	e.GET("/error", func(c echo.Context) error {
		// Simulate an error
		client.Error(c.Request().Context(), "Simulated error endpoint", map[string]interface{}{
			"endpoint": "/error",
			"ip":       c.RealIP(),
		})

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Internal server error",
		})
	})

	// Start server
	log.Println("Starting Echo server on :8080")
	if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// LogwardMiddleware creates an Echo middleware that logs all requests to LogWard
func LogwardMiddleware(client *logward.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Record start time
			start := time.Now()

			// Process request
			err := next(c)

			// Calculate duration
			duration := time.Since(start)

			// Get response status
			statusCode := c.Response().Status

			// Handle error from handler
			if err != nil {
				// Echo's error handler will set the status code
				if he, ok := err.(*echo.HTTPError); ok {
					statusCode = he.Code
				} else {
					statusCode = http.StatusInternalServerError
				}
			}

			// Determine log level based on status code
			logLevel := getLogLevel(statusCode)

			// Prepare metadata
			metadata := map[string]interface{}{
				"method":       c.Request().Method,
				"path":         c.Request().URL.Path,
				"status":       statusCode,
				"duration_ms":  duration.Milliseconds(),
				"ip":           c.RealIP(),
				"user_agent":   c.Request().UserAgent(),
				"query_params": c.QueryParams().Encode(),
			}

			// Add error if present
			if err != nil {
				metadata["error"] = err.Error()
			}

			// Log the request
			message := "HTTP request completed"
			switch logLevel {
			case logward.LogLevelError:
				client.Error(c.Request().Context(), message, metadata)
			case logward.LogLevelWarn:
				client.Warn(c.Request().Context(), message, metadata)
			default:
				client.Info(c.Request().Context(), message, metadata)
			}

			return err
		}
	}
}

// getLogLevel determines the log level based on HTTP status code
func getLogLevel(statusCode int) logward.LogLevel {
	switch {
	case statusCode >= 500:
		return logward.LogLevelError
	case statusCode >= 400:
		return logward.LogLevelWarn
	default:
		return logward.LogLevelInfo
	}
}
