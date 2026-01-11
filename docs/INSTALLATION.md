# Installation Guide

This guide covers how to install and set up the LogTide Go SDK in your project.

## Requirements

- Go 1.21 or later
- A LogTide account with an API key

## Installation

### Using `go get`

```bash
go get github.com/logtide-dev/logtide-sdk-go
```

### Using Go Modules

Add to your `go.mod`:

```go
require github.com/logtide-dev/logtide-sdk-go v0.1.0
```

Then run:

```bash
go mod download
```

### Verifying Installation

Create a simple test file to verify the installation:

```go
package main

import (
    "context"
    "fmt"
    "github.com/logtide-dev/logtide-sdk-go"
)

func main() {
    client, err := logtide.New(
        logtide.WithAPIKey("lp_your_api_key"),
        logtide.WithService("test-service"),
    )
    if err != nil {
        panic(err)
    }
    defer client.Close()

    client.Info(context.Background(), "Installation successful!", nil)
    fmt.Println("LogTide SDK installed successfully!")
}
```

Run it:

```bash
go run main.go
```

## Getting Your API Key

1. Sign up at [https://logtide.dev](https://logtide.dev)
2. Create a project
3. Navigate to **Project Settings** → **API Keys**
4. Generate a new API key (starts with `lp_`)
5. Copy your API key for use in your application

## Environment Variables (Recommended)

Instead of hardcoding your API key, use environment variables:

```bash
export LOGTIDE_API_KEY="lp_your_api_key_here"
export LOGTIDE_SERVICE="my-service"
```

Then in your code:

```go
import "os"

client, err := logtide.New(
    logtide.WithAPIKey(os.Getenv("LOGTIDE_API_KEY")),
    logtide.WithService(os.Getenv("LOGTIDE_SERVICE")),
)
```

## Dependencies

The SDK has minimal dependencies:

- **Core SDK**: Only Go standard library
- **OpenTelemetry Support**: `go.opentelemetry.io/otel/trace` (for trace extraction)

All dependencies are automatically managed by Go modules.

## Updating

To update to the latest version:

```bash
go get -u github.com/logtide-dev/logtide-sdk-go
```

Or specify a version:

```bash
go get github.com/logtide-dev/logtide-sdk-go@v0.2.0
```

## Troubleshooting

### Import Issues

If you encounter import issues, try:

```bash
go mod tidy
go clean -modcache
go mod download
```

### API Key Errors

If you see "invalid or missing API key":
- Verify your API key starts with `lp_`
- Check that the API key is not expired
- Ensure you're using the correct project's API key

### Connection Issues

If logs aren't reaching LogTide:
- Check your network connectivity
- Verify the base URL (default: `https://api.logtide.dev`)
- Check firewall settings
- Look for error logs

## Next Steps

- Read the [Quick Start Guide](QUICKSTART.md)
- Explore [Framework Integrations](INTEGRATIONS.md)
- Check out the [Examples](../examples/)
- Review the [API Documentation](https://pkg.go.dev/github.com/logtide-dev/logtide-sdk-go)
