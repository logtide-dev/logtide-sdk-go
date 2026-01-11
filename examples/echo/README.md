# Echo Framework Example

This example demonstrates how to integrate LogTide with the Echo web framework.

## Running the Example

```bash
# Install dependencies
go mod download

# Set your API key (edit main.go or use environment variable)
# Then run the server
go run main.go
```

## What it Demonstrates

- Echo middleware integration
- Automatic request/response logging
- Error handling and logging
- Log level determination based on HTTP status codes
- Real IP extraction
- Structured metadata

## Testing the Server

Once running, you can test the endpoints:

```bash
# GET request
curl http://localhost:8080/

# GET with parameter
curl http://localhost:8080/user/456

# POST request
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"echo","password":"test"}'

# Error endpoint
curl http://localhost:8080/error
```

All requests are automatically logged with comprehensive metadata.
