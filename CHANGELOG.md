# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-01-13

### Added

- Initial release of LogTide Go SDK
- Leveled logging: Debug, Info, Warn, Error, Critical
- Automatic batching with configurable size and interval
- Retry logic with exponential backoff and jitter
- Circuit breaker pattern for fault tolerance
- Graceful shutdown with context support
- OpenTelemetry integration for automatic trace ID extraction
- Thread-safe operations
- ~87% test coverage
- Framework integration examples:
  - Gin middleware
  - Echo middleware
  - Chi middleware
  - Standard library net/http middleware
- Complete documentation:
  - Installation guide
  - Quick start guide
  - Framework integrations guide

[0.1.0]: https://github.com/logtide-dev/logtide-sdk-go/releases/tag/v0.1.0
