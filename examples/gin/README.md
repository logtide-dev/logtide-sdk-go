# Gin Framework Example

This example demonstrates how to integrate LogTide with the Gin web framework.

## Running the Example

```bash
# Install dependencies
go mod download

# Set your API key (edit main.go or use environment variable)
# Then run the server
go run main.go
```

## What it Demonstrates

- Gin middleware integration
- Automatic request/response logging
- Log level determination based on HTTP status codes
- Logging within route handlers
- Structured metadata (method, path, status, duration, IP, user agent)

## Testing the Server

Once running, you can test the endpoints:

```bash
# GET request
curl http://localhost:8080/

# GET with parameter
curl http://localhost:8080/user/123

# POST request
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"pass"}'

# Error endpoint
curl http://localhost:8080/error
```

Each request will be automatically logged to LogTide with detailed metadata.
