# Standard Library Example

This example demonstrates how to integrate LogWard with Go's standard library `net/http` package.

## Running the Example

```bash
# Install dependencies
go mod download

# Set your API key (edit main.go or use environment variable)
# Then run the server
go run main.go
```

## What it Demonstrates

- Standard library HTTP server
- Custom logging middleware
- Response writer wrapping to capture status codes
- Request/response logging without third-party frameworks
- Timeouts and server configuration

## Testing the Server

Once running, you can test the endpoints:

```bash
# GET request
curl http://localhost:8080/

# GET with parameter
curl http://localhost:8080/user/789

# POST request
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"stdlib","password":"test"}'

# Error endpoint
curl http://localhost:8080/error
```

The middleware automatically logs all requests with detailed information.
