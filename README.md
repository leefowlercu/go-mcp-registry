# go-mcp-registry

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue.svg)](https://golang.org/)
[![Test](https://github.com/leefowlercu/go-mcp-registry/actions/workflows/test.yml/badge.svg)](https://github.com/leefowlercu/go-mcp-registry/actions/workflows/test.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/github.com/leefowlercu/go-mcp-registry.svg)](https://pkg.go.dev/github.com/leefowlercu/go-mcp-registry)

A Go SDK for the [Model Context Protocol (MCP) Registry](https://registry.modelcontextprotocol.io) - the official registry for MCP servers.

## Overview

The Model Context Protocol (MCP) enables applications to integrate with external data sources and tools. The MCP Registry serves as a central hub for discovering and retrieving MCP servers developed by the community.

This Go SDK provides an idiomatic interface to the MCP Registry API, allowing you to:

- 🔍 **Discover MCP servers** with search and filtering capabilities
- 📦 **Retrieve server details** including installation packages and configurations
- 🔄 **Handle pagination** automatically or manually
- ⚡ **Track rate limits** and handle API errors gracefully
- 🎯 **Find specific versions** with flexible version resolution
- 📊 **Access comprehensive metadata** for each server

## Installation

```bash
go get github.com/leefowlercu/go-mcp-registry
```

**Requirements:** Go 1.21 or later

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/leefowlercu/go-mcp-registry/mcp"
)

func main() {
    // Create a client
    client := mcp.NewClient(nil)
    ctx := context.Background()

    // List servers
    servers, _, err := client.Servers.List(ctx, &mcp.ServerListOptions{
        ListOptions: mcp.ListOptions{Limit: 10},
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d servers:\n", len(servers.Servers))
    for _, server := range servers.Servers {
        fmt.Printf("- %s (v%s): %s\n", server.Name, server.Version, server.Description)
    }

    // Get a specific server by name
    gmailServers, err := client.Servers.GetByName(ctx, "ai.waystation/gmail")
    if err != nil {
        log.Fatal(err)
    }
    if len(gmailServers) > 0 {
        fmt.Printf("\nGmail server latest version: %s\n", gmailServers[0].Version)
    }
}
```

## Usage Guide

### Client Configuration

```go
import (
    "net/http"
    "time"
    "github.com/leefowlercu/go-mcp-registry/mcp"
)

// Default client
client := mcp.NewClient(nil)

// Custom HTTP client with timeout
httpClient := &http.Client{
    Timeout: 60 * time.Second,
}
client := mcp.NewClient(httpClient)
```

### Listing Servers

```go
// Basic listing
servers, resp, err := client.Servers.List(ctx, nil)

// With search and filtering
opts := &mcp.ServerListOptions{
    Search: "github",           // Search server names
    Version: "latest",          // Only latest versions
    ListOptions: mcp.ListOptions{
        Limit: 20,              // Page size
    },
}
servers, resp, err := client.Servers.List(ctx, opts)

// Get all servers (handles pagination automatically)
allServers, err := client.Servers.ListAll(ctx, nil)
```

### Getting Servers by Name

```go
// Get all versions of a server
servers, err := client.Servers.GetByName(ctx, "ai.waystation/gmail")

// Get latest version only
server, err := client.Servers.GetByNameLatest(ctx, "ai.waystation/gmail")

// Get specific version
server, err := client.Servers.GetByNameExactVersion(ctx, "ai.waystation/gmail", "0.3.1")

// Get latest active version (uses semantic versioning)
server, err := client.Servers.GetByNameLatestActiveVersion(ctx, "ai.waystation/gmail")
```

### Manual Pagination

```go
opts := &mcp.ServerListOptions{
    ListOptions: mcp.ListOptions{Limit: 50},
}

for {
    resp, _, err := client.Servers.List(ctx, opts)
    if err != nil {
        break
    }

    // Process servers
    for _, server := range resp.Servers {
        fmt.Printf("Server: %s\n", server.Name)
    }

    // Check for more pages
    if resp.Metadata == nil || resp.Metadata.NextCursor == "" {
        break
    }
    opts.Cursor = resp.Metadata.NextCursor
}
```

### Error Handling

```go
servers, resp, err := client.Servers.List(ctx, nil)
if err != nil {
    // Check for rate limiting
    if rateLimitErr, ok := err.(*mcp.RateLimitError); ok {
        fmt.Printf("Rate limited. Reset at: %v\n", rateLimitErr.Rate.Reset)
        return
    }

    // Check for API errors
    if apiErr, ok := err.(*mcp.ErrorResponse); ok {
        fmt.Printf("API error: %v\n", apiErr.Message)
        return
    }

    log.Fatal(err)
}

// Check rate limit info
if resp.Rate.Limit > 0 {
    fmt.Printf("Rate limit: %d/%d remaining\n", resp.Rate.Remaining, resp.Rate.Limit)
}
```

## API Methods Reference

| Method | Description |
|--------|-------------|
| `List(ctx, opts)` | List servers with pagination and filtering |
| `Get(ctx, id)` | Get server by ID |
| `ListAll(ctx, opts)` | Get all servers (automatic pagination) |
| `GetByName(ctx, name)` | Get all versions of a named server |
| `GetByNameLatest(ctx, name)` | Get latest version using API filter |
| `GetByNameExactVersion(ctx, name, version)` | Get specific version |
| `GetByNameLatestActiveVersion(ctx, name)` | Get latest active version by semver |

For detailed documentation, see the [Go Reference](https://pkg.go.dev/github.com/leefowlercu/go-mcp-registry).

## Examples

This repository includes working examples in the `examples/` directory:

- **[examples/list/](examples/list/)** - List servers with basic options
- **[examples/get/](examples/get/)** - Get server details by ID or name
- **[examples/paginate/](examples/paginate/)** - Manual and automatic pagination

Run examples:
```bash
go run ./examples/list/
go run ./examples/get/ "ai.waystation/gmail"
go run ./examples/paginate/
```

## Development

### Running Tests

```bash
# Unit tests
go test ./...

# With coverage
go test -cover ./...

# Integration tests (requires network)
INTEGRATION_TESTS=true go test ./test/integration/

# Specific test
go test -v ./mcp -run TestServersService_GetByName
```

### Building

```bash
# Build all packages
go build ./...

# Build examples
go build ./examples/...

# Format code
gofmt -s -w .

# Lint
go vet ./...
```

## Architecture

This SDK follows the service-oriented architecture pattern established by [google/go-github](https://github.com/google/go-github), organizing API endpoints into logical service groups:

- **Client** - Main entry point with HTTP client management
- **ServersService** - All server-related operations

The SDK imports and reuses official types from the [MCP Registry repository](https://github.com/modelcontextprotocol/registry) to ensure perfect API compatibility without type conversion overhead.

## Links

- **MCP Protocol:** https://modelcontextprotocol.io/
- **MCP Registry:** https://registry.modelcontextprotocol.io/
- **API Documentation:** https://registry.modelcontextprotocol.io/docs
- **Registry Repository:** https://github.com/modelcontextprotocol/registry

## License

MIT License - see [LICENSE](LICENSE) file for details.