// Package mcp provides a Go client library for the MCP Server Registry API.
//
// The MCP Registry API allows you to discover and retrieve information about
// Model Context Protocol servers. This client library provides an idiomatic
// Go interface to the registry API, following architectural patterns established
// by popular Go libraries like google/go-github.
//
// # Features
//
// The SDK currently provides read-only operations for the MCP Registry:
//
//   - List servers with pagination, search, and filtering
//   - Get server details by ID or name
//   - Cursor-based pagination support
//   - Rate limit tracking from response headers
//   - Context support for all API calls
//   - Comprehensive error handling
//   - Helper methods for common operations
//
// # Authentication
//
// The MCP Registry API currently supports read-only operations without
// authentication. Future versions of this SDK will support authentication
// for write operations such as publishing and updating servers.
//
// # Usage
//
// Import the package:
//
//	import "github.com/leefowlercu/go-mcp-registry/mcp"
//
// Create a new client:
//
//	client := mcp.NewClient(nil)
//
// You can provide a custom HTTP client for advanced configuration:
//
//	httpClient := &http.Client{
//		Timeout: 60 * time.Second,
//	}
//	client := mcp.NewClient(httpClient)
//
// List servers:
//
//	servers, resp, err := client.Servers.List(context.Background(), nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Found %d servers\n", len(servers.Servers))
//
// List servers with options:
//
//	opts := &mcp.ServerListOptions{
//		Search: "github",
//		Version: "latest",
//		ListOptions: mcp.ListOptions{
//			Limit: 20,
//		},
//	}
//	servers, resp, err := client.Servers.List(context.Background(), opts)
//
// Get a specific server by ID:
//
//	server, resp, err := client.Servers.Get(context.Background(), "server-uuid", nil)
//
// Get a specific version of a server by ID:
//
//	opts := &mcp.ServerGetOptions{Version: "1.0.0"}
//	server, resp, err := client.Servers.Get(context.Background(), "server-uuid", opts)
//
// List all versions of a server by ID:
//
//	versions, resp, err := client.Servers.ListByServerID(context.Background(), "server-uuid")
//
// List servers by name (returns all versions):
//
//	servers, _, err := client.Servers.ListByName(context.Background(), "ai.waystation/gmail")
//	if len(servers) > 0 {
//		fmt.Printf("Found %d versions of %s\n", len(servers), servers[0].Name)
//	}
//
// Get latest version of a server by name:
//
//	server, _, err := client.Servers.GetByNameLatest(context.Background(), "ai.waystation/gmail")
//	if server != nil {
//		fmt.Printf("Latest version: %s (v%s)\n", server.Name, server.Version)
//	}
//
// Get specific version of a server by name:
//
//	server, _, err := client.Servers.GetByNameExactVersion(context.Background(), "ai.waystation/gmail", "0.3.1")
//	if server != nil {
//		fmt.Printf("Found version: %s (v%s)\n", server.Name, server.Version)
//	}
//
// Get latest active version of a server by name (semantic version comparison):
//
//	server, _, err := client.Servers.GetByNameLatestActiveVersion(context.Background(), "ai.waystation/gmail")
//	if server != nil {
//		fmt.Printf("Latest active: %s (v%s) - %s\n", server.Name, server.Version, server.Status)
//	}
//
// Get servers updated since a specific timestamp:
//
//	since := time.Now().AddDate(0, 0, -7) // Last 7 days
//	servers, _, err := client.Servers.ListByUpdatedSince(context.Background(), since)
//	if err == nil {
//		fmt.Printf("Found %d servers updated since %s\n", len(servers), since.Format("2006-01-02"))
//		for _, server := range servers {
//			fmt.Printf("  %s (v%s) - updated: %s\n",
//				server.Name, server.Version,
//				server.Meta.Official.UpdatedAt.Format("2006-01-02"))
//		}
//	}
//
// # Pagination
//
// The API uses cursor-based pagination following the MCP Protocol specification.
// Use ListOptions to control pagination:
//
//	var allServers []registryv0.ServerJSON
//	opts := &mcp.ServerListOptions{
//		ListOptions: mcp.ListOptions{Limit: 50},
//	}
//
//	for {
//		resp, _, err := client.Servers.List(context.Background(), opts)
//		if err != nil {
//			break
//		}
//		allServers = append(allServers, resp.Servers...)
//
//		if resp.Metadata.NextCursor == "" {
//			break // No more pages
//		}
//		opts.Cursor = resp.Metadata.NextCursor
//	}
//
// Or use the convenience method to fetch all pages automatically:
//
//	servers, _, err := client.Servers.ListAll(context.Background(), nil)
//
// # Error Handling
//
// The library provides structured error handling with custom error types:
//
//	servers, resp, err := client.Servers.List(context.Background(), nil)
//	if err != nil {
//		if rateLimitErr, ok := err.(*mcp.RateLimitError); ok {
//			fmt.Printf("Rate limited. Reset at: %v\n", rateLimitErr.Rate.Reset)
//			return
//		}
//		if apiErr, ok := err.(*mcp.ErrorResponse); ok {
//			fmt.Printf("API error: %v\n", apiErr.Message)
//			return
//		}
//		// Handle other errors
//		log.Fatal(err)
//	}
//
// # Rate Limiting
//
// Rate limit information is tracked and available in response objects:
//
//	servers, resp, err := client.Servers.List(context.Background(), nil)
//	if err == nil && resp.Rate.Limit > 0 {
//		fmt.Printf("Rate Limit: %d/%d remaining\n",
//			resp.Rate.Remaining, resp.Rate.Limit)
//		fmt.Printf("Reset at: %v\n", resp.Rate.Reset)
//	}
//
// # Service Architecture
//
// The client follows a service-oriented architecture where different API
// endpoints are organized into service structs:
//
//	// Available services
//	client.Servers  // Server-related operations
//
// Each service provides methods for different operations:
//
//	// ServersService methods
//	List(ctx, opts) (*ServerListResponse, *Response, error)
//	Get(ctx, serverID, opts) (*ServerJSON, *Response, error)
//	ListByServerID(ctx, serverID) ([]ServerJSON, *Response, error)
//	ListAll(ctx, opts) ([]ServerJSON, *Response, error)                     // Helper
//	ListByName(ctx, name) ([]ServerJSON, *Response, error)                  // Helper - returns all versions
//	ListByUpdatedSince(ctx, since) ([]ServerJSON, *Response, error)           // Helper - returns servers updated since timestamp
//	GetByNameLatest(ctx, name) (*ServerJSON, *Response, error)              // Helper - returns latest version
//	GetByNameExactVersion(ctx, name, version) (*ServerJSON, *Response, error)     // Helper - returns specific version
//	GetByNameLatestActiveVersion(ctx, name) (*ServerJSON, *Response, error)       // Helper - latest active by semver
//
// # Type Reuse
//
// This SDK imports and uses official types from the MCP Registry repository
// to ensure perfect API compatibility:
//
//	import registryv0 "github.com/modelcontextprotocol/registry/pkg/api/v0"
//
// Key types include:
//
//   - registryv0.ServerJSON - Complete server information
//   - registryv0.ServerListResponse - Paginated server list response
//   - registryv0.Metadata - Pagination metadata with NextCursor
//   - model.Repository, model.Package, model.Transport - Supporting types
//
// # Helper Functions
//
// The package provides helper functions for working with pointer types,
// following Go API conventions:
//
//	mcp.String("value")    // Returns *string
//	mcp.Int(42)           // Returns *int
//	mcp.Bool(true)        // Returns *bool
//
//	mcp.StringValue(ptr)  // Returns string value or ""
//	mcp.IntValue(ptr)     // Returns int value or 0
//	mcp.BoolValue(ptr)    // Returns bool value or false
//
// # Examples
//
// See the examples/ directory for complete working examples:
//
//   - examples/list/     - List servers with pagination
//   - examples/get/      - Get server details by ID or name
//   - examples/paginate/ - Handle pagination manually and automatically
//
// # See Also
//
// Related resources:
//
//   - MCP Registry API: https://registry.modelcontextprotocol.io
//   - API Documentation: https://registry.modelcontextprotocol.io/docs
//   - MCP Protocol: https://modelcontextprotocol.io/specification
//   - Registry Repository: https://github.com/modelcontextprotocol/registry
package mcp
