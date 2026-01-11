package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/logtide-dev/logtide-sdk-go"
)

func main() {
	// Create LogTide client
	client, err := logtide.New(
		logtide.WithAPIKey("lp_your_api_key_here"),
		logtide.WithService("stdlib-example"),
	)
	if err != nil {
		log.Fatalf("Failed to create LogTide client: %v", err)
	}
	defer client.Close()

	// Create router
	mux := http.NewServeMux()

	// Define routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Hello from standard library!",
		})
	})

	mux.HandleFunc("/user/", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Path[len("/user/"):]

		// Log within handler
		client.Info(r.Context(), "Fetching user details", map[string]interface{}{
			"user_id": userID,
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id": userID,
			"name":    "Alice Smith",
		})
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		type LoginRequest struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			client.Error(r.Context(), "Invalid login request", map[string]interface{}{
				"error": err.Error(),
			})
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Simulate login
		client.Info(r.Context(), "User login attempt", map[string]interface{}{
			"username": req.Username,
			"success":  true,
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Login successful",
			"token":   "sample-jwt-token",
		})
	})

	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		// Simulate an error
		client.Error(r.Context(), "Simulated error endpoint", map[string]interface{}{
			"endpoint": "/error",
			"ip":       r.RemoteAddr,
		})

		http.Error(w, "Internal server error", http.StatusInternalServerError)
	})

	// Wrap with logging middleware
	handler := LoggingMiddleware(client, mux)

	// Start server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("Starting HTTP server on :8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// LoggingMiddleware creates a middleware that logs all requests to LogTide
func LoggingMiddleware(client *logtide.Client, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Record start time
		start := time.Now()

		// Wrap response writer to capture status code
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Process request
		next.ServeHTTP(rw, r)

		// Calculate duration
		duration := time.Since(start)

		// Determine log level based on status code
		logLevel := getLogLevel(rw.statusCode)

		// Prepare metadata
		metadata := map[string]interface{}{
			"method":       r.Method,
			"path":         r.URL.Path,
			"status":       rw.statusCode,
			"duration_ms":  duration.Milliseconds(),
			"ip":           r.RemoteAddr,
			"user_agent":   r.UserAgent(),
			"query_params": r.URL.RawQuery,
		}

		// Log the request
		message := "HTTP request completed"
		switch logLevel {
		case logtide.LogLevelError:
			client.Error(r.Context(), message, metadata)
		case logtide.LogLevelWarn:
			client.Warn(r.Context(), message, metadata)
		default:
			client.Info(r.Context(), message, metadata)
		}
	})
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
