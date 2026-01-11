package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/logtide-dev/logtide-sdk-go"
)

func main() {
	// Create LogTide client
	client, err := logtide.New(
		logtide.WithAPIKey("lp_your_api_key_here"),
		logtide.WithService("gin-example"),
	)
	if err != nil {
		log.Fatalf("Failed to create LogTide client: %v", err)
	}
	defer client.Close()

	// Create Gin router
	r := gin.Default()

	// Add LogTide middleware
	r.Use(LogtideMiddleware(client))

	// Define routes
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello from Gin!",
		})
	})

	r.GET("/user/:id", func(c *gin.Context) {
		userID := c.Param("id")

		// Log within handler
		client.Info(c.Request.Context(), "Fetching user details", map[string]interface{}{
			"user_id": userID,
		})

		c.JSON(200, gin.H{
			"user_id": userID,
			"name":    "John Doe",
		})
	})

	r.POST("/login", func(c *gin.Context) {
		var json struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&json); err != nil {
			client.Error(c.Request.Context(), "Invalid login request", map[string]interface{}{
				"error": err.Error(),
			})
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// Simulate login
		client.Info(c.Request.Context(), "User login attempt", map[string]interface{}{
			"username": json.Username,
			"success":  true,
		})

		c.JSON(200, gin.H{
			"message": "Login successful",
			"token":   "sample-jwt-token",
		})
	})

	r.GET("/error", func(c *gin.Context) {
		// Simulate an error
		client.Error(c.Request.Context(), "Simulated error endpoint", map[string]interface{}{
			"endpoint": "/error",
			"ip":       c.ClientIP(),
		})

		c.JSON(500, gin.H{
			"error": "Internal server error",
		})
	})

	// Start server
	log.Println("Starting Gin server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// LogtideMiddleware creates a Gin middleware that logs all requests to LogTide
func LogtideMiddleware(client *logtide.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record start time
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Determine log level based on status code
		statusCode := c.Writer.Status()
		logLevel := getLogLevel(statusCode)

		// Prepare metadata
		metadata := map[string]interface{}{
			"method":       c.Request.Method,
			"path":         c.Request.URL.Path,
			"status":       statusCode,
			"duration_ms":  duration.Milliseconds(),
			"ip":           c.ClientIP(),
			"user_agent":   c.Request.UserAgent(),
			"query_params": c.Request.URL.RawQuery,
		}

		// Add error if present
		if len(c.Errors) > 0 {
			metadata["errors"] = c.Errors.String()
		}

		// Log the request
		message := "HTTP request completed"
		switch logLevel {
		case logtide.LogLevelError:
			client.Error(c.Request.Context(), message, metadata)
		case logtide.LogLevelWarn:
			client.Warn(c.Request.Context(), message, metadata)
		default:
			client.Info(c.Request.Context(), message, metadata)
		}
	}
}

// getLogLevel determines the log level based on HTTP status code
func getLogLevel(statusCode int) logtide.LogLevel {
	switch {
	case statusCode >= 500:
		return logtide.LogLevelError
	case statusCode >= 400:
		return logtide.LogLevelWarn
	default:
		return logtide.LogLevelInfo
	}
}
