# Basic Example

This example demonstrates basic usage of the LogWard Go SDK with all log levels.

## Running the Example

```bash
# Set your API key
export LOGWARD_API_KEY="lp_your_api_key_here"

# Run the example
go run main.go
```

## What it Demonstrates

- Creating a LogWard client
- Using all log levels (Debug, Info, Warn, Error, Critical)
- Adding structured metadata to logs
- Manual flushing
- Graceful shutdown with `defer client.Close()`

## Expected Output

The example will send logs to LogWard and print progress to the console. All logs will be flushed before the program exits.
