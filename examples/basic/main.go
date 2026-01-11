package main

import (
	"context"
	"log"
	"time"

	"github.com/logtide-dev/logtide-sdk-go"
)

func main() {
	// Create LogTide client
	client, err := logtide.New(
		logtide.WithAPIKey("lp_your_api_key_here"),
		logtide.WithService("example-service"),
		// Optional: customize configuration
		// logtide.WithBatchSize(50),
		// logtide.WithFlushInterval(10*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create LogTide client: %v", err)
	}

	// Ensure logs are flushed on exit
	defer client.Close()

	ctx := context.Background()

	// Debug level - detailed debugging information
	client.Debug(ctx, "Application started", map[string]interface{}{
		"version":     "1.0.0",
		"environment": "production",
	})

	// Info level - general informational messages
	client.Info(ctx, "User logged in", map[string]interface{}{
		"user_id":  12345,
		"username": "john.doe",
		"ip":       "192.168.1.1",
	})

	// Warn level - warning messages
	client.Warn(ctx, "High memory usage detected", map[string]interface{}{
		"memory_usage_percent": 85,
		"threshold":            80,
	})

	// Error level - error events
	client.Error(ctx, "Failed to connect to database", map[string]interface{}{
		"database": "postgres",
		"host":     "db.example.com",
		"error":    "connection timeout after 30s",
		"retries":  3,
	})

	// Critical level - critical system errors
	client.Critical(ctx, "System shutdown initiated", map[string]interface{}{
		"reason": "critical error",
		"uptime": "72h",
	})

	// Logs with nil metadata
	client.Info(ctx, "Simple log without metadata", nil)

	// Simulate some work
	log.Println("Doing some work...")
	time.Sleep(2 * time.Second)

	// Manual flush (optional - Close() will also flush)
	if err := client.Flush(ctx); err != nil {
		log.Printf("Failed to flush logs: %v", err)
	}

	log.Println("Example completed successfully!")
}
