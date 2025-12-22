# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1] - 2025-12-22

### Fixed
- Added missing `fmt` import in `observation.go`
- Removed unused `err` variable in `observation.go` StartObservation function
- Removed unused `expectedStatus` variable in `client.go` handleResponse function
- Fixed integration test missing context parameter in CreateTrace call
- Fixed examples compilation errors:
  - `getting_started`: Added missing StartTime field to Generation
  - `error_handling`: Fixed taking address of constant (LevelError)
  - `observe_wrapper`: Fixed generic type mismatch in observeSpan function
- Fixed redundant newline warnings in example files

## [0.1.0] - 2025-12-22

### Added
- Initial release of Langfuse Go SDK
- Client initialization with configurable base URL and HTTP client
- Trace creation and updates
- Span creation and updates
- Generation creation and updates
- Event creation
- Score creation
- Helper functions for creating entities with auto-generated IDs
- Comprehensive error handling with APIError type
- W3C Trace Context support with deterministic trace ID generation
- Context manager pattern with StartObservation and StartChildObservation
- Trace context propagation across services
- Logger interface for debugging API requests/responses
- Example code demonstrating:
  - Basic usage
  - Advanced features (nested observations, context propagation)
  - Observe wrapper pattern
  - Error handling
  - Logger integration
  - Context propagation across services

