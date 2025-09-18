# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-09-16

### Added

#### Core Features
- Initial release of Go SDK for MCP Registry API
- Complete client implementation with service-oriented architecture
- Support for all read-only MCP Registry API operations
- HTTP client management with configurable timeouts and custom clients

#### Server Operations
- `List()` method for paginated server listing with filtering options
- `Get()` method for retrieving servers by ID
- `ListAll()` helper method for automatic pagination
- `GetByName()` method returning all versions of a named server
- `GetByNameLatest()` method using API's version=latest filter
- `GetByNameExactVersion()` method with client-side version filtering
- `GetByNameLatestActiveVersion()` method using semantic version comparison

#### Query and Filtering
- Search functionality for server names
- Version filtering support (latest)
- UpdatedSince timestamp filtering
- Cursor-based pagination with automatic next page handling
- Configurable page size limits

#### Error Handling
- Custom `ErrorResponse` type for structured API errors
- `RateLimitError` type with reset time information
- Comprehensive error checking in `CheckResponse()` function
- Context support for request cancellation and timeouts

#### Type System
- Import and reuse of official types from `github.com/modelcontextprotocol/registry`
- Custom `Client`, `Response`, and options types
- Pointer helper functions (`String()`, `Int()`, `Bool()`, etc.)
- Value extraction helpers (`StringValue()`, `IntValue()`, `BoolValue()`)

#### Testing
- Comprehensive unit test suite with table-driven testing patterns
- HTTP mocking using `httptest.Server`
- Integration tests for live API validation (opt-in with `INTEGRATION_TESTS=true`)
- Test coverage: 63% overall, 78-92% for server operations

#### Documentation
- Complete package documentation in `doc.go`
- Usage examples in package documentation
- Comprehensive README with installation and usage guide
- Three working example programs (`list`, `get`, `paginate`)

#### Build and CI
- GitHub Actions workflow for automated testing
- Support for Go versions 1.24 and 1.25
- Automated building of example programs
- Code formatting validation with `gofmt`
- Static analysis with `go vet`

#### Dependencies
- `github.com/google/go-querystring` v1.1.0 for URL parameter encoding
- `github.com/modelcontextprotocol/registry` v1.0.0 for official API types
- `github.com/Masterminds/semver/v3` v3.4.0 for semantic version comparison

### Technical Details

#### Architecture
- Service-oriented design following `google/go-github` patterns
- Separation of concerns with dedicated service structs
- Consistent error handling across all operations
- Rate limit tracking and response metadata parsing

#### API Compatibility
- Full compatibility with MCP Registry API v0
- No authentication required for read operations
- User-Agent string: `go-mcp-registry/v0.1.0`
- Default base URL: `https://registry.modelcontextprotocol.io/`

#### Performance
- Efficient pagination with cursor-based iteration
- Concurrent-safe client with mutex protection
- Memory-efficient handling of large result sets
- Optional automatic pagination for convenience

[0.1.0]: https://github.com/leefowlercu/go-mcp-registry/releases/tag/v0.1.0