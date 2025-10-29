# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- New example program `examples/version/` demonstrating version-specific retrieval using GetByNameExactVersion
- New example program `examples/updated/` demonstrating timestamp-based filtering with ListByUpdatedSince
- Comprehensive "What's New in v0.6.0" section to README highlighting API v0.1 migration and testing improvements
- "Accessing Registry Metadata" section to README with complete guide on ServerResponse.Meta.Official fields
- Test coverage metric (94.2%) to README Development section

### Changed
- Enhanced `examples/get/` to demonstrate version-specific retrieval and error type checking (RateLimitError, ErrorResponse)
- Enhanced `examples/list/` to demonstrate metadata access (Status, PublishedAt, UpdatedAt, IsLatest)
- Comprehensive README.md update for professional tone, accuracy, and comprehensiveness:
  - Removed emojis from feature list for professional tone
  - Expanded "Getting Servers by Name" section with Get() version options
  - Updated Examples section to include all 5 example programs with descriptions
  - Added `ListByName(ctx, name)` to API Methods Reference table
- Updated all example run commands in README to show version and parameter options
- Updated GitHub Actions workflow to build all 5 examples using matrix strategy:
  - Separated examples into dedicated `build-examples` job that runs after tests
  - Examples now build in parallel with individual failure reporting
  - Uses `fail-fast: false` to show all failing examples at once

### Fixed
- README Quick Start example: corrected `server.Name` to `serverResponse.Server.Name`
- README Manual Pagination example: corrected `server.Name` to `serverResponse.Server.Name`
- Code formatting in `mcp/mcp_test.go` to comply with gofmt standards

## [0.6.0] - 2025-10-28

### Added
- Comprehensive unit tests for pointer utility functions (String, StringValue, Int, IntValue, Bool, BoolValue)
- Comprehensive unit tests for error handling (ErrorResponse, RateLimitError, sanitizeURL, CheckResponse)
- Comprehensive unit tests for core client functionality (NewRequest, parseRate, Do, addOptions, newResponse)
- Edge case tests for servers service methods (nil response handling, missing metadata)
- Increased test coverage from 73.5% to 94.2%

### Changed
- Updated API path prefix from `v0` to `v0.1` to use the stable API version
- All API endpoints now use the `v0.1` path prefix instead of `v0`
- **BREAKING**: Updated `Get()` method to handle removed endpoint `GET /v0/servers/{serverName}`
  - Now uses `GET /v0.1/servers/{serverName}/versions/latest` when no version is specified
  - Uses `GET /v0.1/servers/{serverName}/versions/{version}` when version is specified
- Updated unit tests to use the new `v0.1` API paths and endpoint structures

### Fixed
- Test mock handler in TestServersService_ListAll to use correct v0.1 endpoint

### Migration Guide
- No code changes required for SDK users - the `Get()` method signature remains unchanged
- The method now internally routes to the correct versioned endpoint based on the `ServerGetOptions`
- This update ensures compatibility with the MCP Registry's endpoint simplification

### Notes
- This change aligns with the MCP Registry's introduction of a stable API path prefix and endpoint simplification
- The removed endpoint `GET /v0/servers/{serverName}` has been replaced with versioned endpoints
- Reference: https://github.com/modelcontextprotocol/registry/blob/main/docs/reference/api/CHANGELOG.md

## [0.5.0] - 2025-10-10

### Changed
- **BREAKING**: Updated for MCP Registry API v2 migration (2025-09-29)
- **BREAKING**: `Get()` method parameter changed from `serverID` to `serverName` - all endpoints now use server names instead of UUIDs
- **BREAKING**: `ListByServerID()` method renamed to `ListVersionsByName()` to reflect name-based endpoint
- **BREAKING**: Response schema changes:
  - `ServerListResponse.Servers` changed from `[]ServerJSON` to `[]ServerResponse`
  - `ServerResponse` now wraps `ServerJSON` with metadata: `{Server: ServerJSON, Meta: ResponseMeta}`
  - Status field moved from `ServerJSON` to `ServerResponse.Meta.Official.Status`
  - ServerID field completely removed from API
- **BREAKING**: `GetByNameExactVersion()` now uses dedicated API endpoint `GET /v0/servers/{serverName}/versions/{version}` instead of client-side filtering (performance improvement)
- Updated `github.com/modelcontextprotocol/registry` dependency from v1.1.0 to v1.2.3
- Updated all unit tests for new response schema with ServerResponse wrapper
- Updated all integration tests for name-based endpoints and removed Status field assertions
- Updated all examples to use server names instead of server IDs
- Updated all documentation (README.md, CLAUDE.md, doc.go) with new method signatures and API v2 migration notes
- Fixed pagination tests to use camelCase JSON field names (`nextCursor` instead of `next_cursor`)

### Removed
- **BREAKING**: ServerID-based server retrieval - all operations now require server names (e.g., "ai.waystation/gmail")
- **BREAKING**: Direct access to Status field on `ServerJSON` - Status is now only available through `ServerResponse.Meta.Official.Status`

### Fixed
- URL encoding for server names containing forward slashes
- Response metadata handling for v1.2.3 schema changes
- Test assertions updated to work with unwrapped ServerJSON objects

### Migration Guide
- Replace `client.Servers.Get(ctx, serverID, opts)` with `client.Servers.Get(ctx, serverName, opts)` where `serverName` is in format "publisher/server"
- Replace `client.Servers.ListByServerID(ctx, serverID)` with `client.Servers.ListVersionsByName(ctx, serverName)`
- Update code accessing `ServerListResponse.Servers[i].Name` to `ServerListResponse.Servers[i].Server.Name` (ServerResponse wrapper)
- Remove code accessing `ServerJSON.Status` field - Status is not accessible from unwrapped ServerJSON objects
- Remove code accessing `ServerJSON.Meta.Official` - registry metadata is only in `ServerResponse.Meta.Official`
- Update server identification from UUIDs to qualified names (e.g., "server-uuid-1234" → "ai.waystation/gmail")

## [0.4.0] - 2025-09-25

### Changed
- Updated README.md documentation to reflect current API signatures
- Fixed outdated method examples in Quick Start and Usage Guide sections
- Updated API Methods Reference table with complete method list
- Corrected all code examples to include `*Response` return values

### Fixed
- Documentation inconsistencies with actual API methods
- Missing methods in API reference table (ListByServerID, ListByUpdatedSince)

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

[0.6.0]: https://github.com/leefowlercu/go-mcp-registry/releases/tag/v0.6.0
[0.5.0]: https://github.com/leefowlercu/go-mcp-registry/releases/tag/v0.5.0
[0.4.0]: https://github.com/leefowlercu/go-mcp-registry/releases/tag/v0.4.0
[0.3.0]: https://github.com/leefowlercu/go-mcp-registry/releases/tag/v0.3.0
[0.2.0]: https://github.com/leefowlercu/go-mcp-registry/releases/tag/v0.2.0
[0.1.0]: https://github.com/leefowlercu/go-mcp-registry/releases/tag/v0.1.0
