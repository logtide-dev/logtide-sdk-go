# Contributing to LogWard Go SDK

Thank you for your interest in contributing to the LogWard Go SDK!

## Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/logward-dev/logward-sdk-go.git
   cd logward-sdk-go
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Run tests**
   ```bash
   go test -v ./...
   ```

4. **Run tests with race detector**
   ```bash
   go test -race ./...
   ```

5. **Check coverage**
   ```bash
   go test -cover ./...
   ```

## Project Structure

```
logward-sdk-go/
├── client.go              # Main client implementation
├── config.go              # Configuration and options
├── types.go               # Core data types
├── validation.go          # Input validation
├── errors.go              # Error types
├── batch.go               # Auto-batching
├── retry.go               # Retry logic
├── circuit_breaker.go     # Circuit breaker
├── context.go             # OpenTelemetry integration
├── internal/http/         # HTTP client wrapper
├── examples/              # Usage examples
└── .github/workflows/     # CI/CD
```

## Testing Guidelines

- Write tests for all new features
- Maintain >80% code coverage
- Include both unit and integration tests
- Use table-driven tests where appropriate
- Test error conditions and edge cases

## Code Style

- Follow standard Go conventions
- Run `gofmt` before committing
- Use meaningful variable names
- Add godoc comments for all exports
- Keep functions focused and small

## Pull Request Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for your changes
5. Ensure all tests pass (`go test ./...`)
6. Ensure no race conditions (`go test -race ./...`)
7. Commit your changes (`git commit -m 'Add amazing feature'`)
8. Push to your branch (`git push origin feature/amazing-feature`)
9. Open a Pull Request

## Commit Message Guidelines

- Use clear, descriptive commit messages
- Start with a verb in imperative mood (Add, Fix, Update, etc.)
- Reference issue numbers when applicable

Examples:
- `Add context cancellation support`
- `Fix race condition in batcher`
- `Update retry backoff algorithm`

## Running Examples

Each example has its own README with instructions:

```bash
cd examples/basic
go run main.go

cd examples/gin
go run main.go

cd examples/echo
go run main.go

cd examples/stdlib
go run main.go

cd examples/otel
go run main.go
```

## Release Process

1. Update version in documentation
2. Update CHANGELOG.md
3. Create a git tag (e.g., `v0.2.0`)
4. Push the tag to trigger release workflow

## Questions?

- Open an issue for bugs or feature requests
- Start a discussion for questions
- Email: support@logward.dev

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
