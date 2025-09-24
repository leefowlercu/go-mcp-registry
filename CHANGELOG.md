# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [0.3.0] - 2025-09-24

### Added
- `ListByServerID()` method to retrieve all versions of a server by server ID
- `ServerGetOptions` type to specify version parameter for `Get()` method
- Support for version-specific server retrieval using the `version` query parameter
- Comprehensive unit tests for new `ListByServerID()` method
- Integration tests for `ListByServerID()` method against live API
- Enhanced documentation with new method signatures and usage examples

### Changed
- **BREAKING**: Renamed `GetByName()` method to `ListByName()` for consistent naming convention
- **BREAKING**: Renamed `ListVersions()` method to `ListByServerID()` for consistent naming convention
- **BREAKING**: Renamed `ListUpdatedSince()` method to `ListByUpdatedSince()` for consistent naming convention
- **BREAKING**: Updated `Get()` method signature to support version parameter: `Get(ctx, serverID, opts)`
- **BREAKING**: Changed `Get()` method parameter from `id` to `serverID` for clarity
- **BREAKING**: Updated all helper method signatures to return `*Response`:
  - `ListAll(ctx, opts) ([]ServerJSON, *Response, error)` - now returns Response
  - `ListByName(ctx, name) ([]ServerJSON, *Response, error)` - now returns Response
  - `ListByUpdatedSince(ctx, since) ([]ServerJSON, *Response, error)` - now returns Response
  - `GetByNameLatest(ctx, name) (*ServerJSON, *Response, error)` - now returns Response
  - `GetByNameExactVersion(ctx, name, version) (*ServerJSON, *Response, error)` - now returns Response
  - `GetByNameLatestActiveVersion(ctx, name) (*ServerJSON, *Response, error)` - now returns Response
- Updated all unit tests to reflect new method signatures
- Updated integration tests with renamed methods
- Updated examples to use new method names and signatures
- Enhanced package documentation with new API patterns
- Helper methods now follow the same pattern as core SDK methods (List, Get, ListVersions) for consistent rate limit handling
- Updated `github.com/modelcontextprotocol/registry` dependency from v1.0.0 to v1.1.0
- Updated code to use `RegistryExtensions.ServerID` field instead of deprecated `ID` field
- Updated Metadata handling for new value type structure in v1.1.0

### Migration Guide
- Replace `client.Servers.GetByName()` calls with `client.Servers.ListByName()`
- Replace `client.Servers.ListVersions()` calls with `client.Servers.ListByServerID()`
- Replace `client.Servers.ListUpdatedSince()` calls with `client.Servers.ListByUpdatedSince()`
- Replace `client.Servers.Get(ctx, id)` calls with `client.Servers.Get(ctx, serverID, nil)`
- Use `client.Servers.Get(ctx, serverID, &ServerGetOptions{Version: "1.0.0"})` for version-specific retrieval
- Use `client.Servers.ListByServerID(ctx, serverID)` to get all versions of a server
- Update all helper method calls to handle the additional `*Response` return value:
  - `servers, err := client.Servers.ListAll(ctx, opts)` → `servers, _, err := client.Servers.ListAll(ctx, opts)`
  - `servers, err := client.Servers.ListByName(ctx, name)` → `servers, _, err := client.Servers.ListByName(ctx, name)`
  - `servers, err := client.Servers.ListByUpdatedSince(ctx, since)` → `servers, _, err := client.Servers.ListByUpdatedSince(ctx, since)`
  - `server, err := client.Servers.GetByNameLatest(ctx, name)` → `server, _, err := client.Servers.GetByNameLatest(ctx, name)`
  - `server, err := client.Servers.GetByNameExactVersion(ctx, name, version)` → `server, _, err := client.Servers.GetByNameExactVersion(ctx, name, version)`
  - `server, err := client.Servers.GetByNameLatestActiveVersion(ctx, name)` → `server, _, err := client.Servers.GetByNameLatestActiveVersion(ctx, name)`

## [0.2.0] - 2025-09-17

### Added
- `ListByUpdatedSince()` method to retrieve all servers updated since a specific timestamp
- Automatic pagination handling in timestamp-based server filtering
- Enhanced documentation with timestamp filtering usage examples
- Comprehensive unit tests for `ListByUpdatedSince()` functionality
- Integration tests for timestamp-based filtering against live API

### Changed
- Updated Go version requirement documentation to align with module constraints
- Enhanced package documentation with additional usage examples

### Fixed
- Corrected Go module version constraints in go.mod file

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
- `github.com/modelcontextprotocol/registry` v1.1.0 for official API types
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

[0.3.0]: https://github.com/leefowlercu/go-mcp-registry/releases/tag/v0.3.0
[0.2.0]: https://github.com/leefowlercu/go-mcp-registry/releases/tag/v0.2.0
[0.1.0]: https://github.com/leefowlercu/go-mcp-registry/releases/tag/v0.1.0
