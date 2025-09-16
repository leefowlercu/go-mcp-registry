package mcp

import (
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Client manages communication with the MCP Registry API.
type Client struct {
	clientMu sync.Mutex   // protects the client during calls
	client   *http.Client // HTTP client used to communicate with the API

	// Base URL for API requests.
	// Defaults to https://registry.modelcontextprotocol.io, but can be
	// overridden to point to another registry instance.
	BaseURL *url.URL

	// User agent used when communicating with the MCP Registry API.
	UserAgent string

	common service // Reuse a single struct instead of allocating one for each service

	// Services used for talking to different parts of the MCP Registry API
	Servers *ServersService

	// Rate limit tracking
	rateMu     sync.Mutex
	rateLimits map[string]Rate
}

// service provides a general service interface for the API.
type service struct {
	client *Client
}

// ServersService handles communication with the server related
// methods of the MCP Registry API.
//
// MCP Registry API docs: https://registry.modelcontextprotocol.io/docs
type ServersService service

// Response wraps the standard http.Response and provides convenient access to
// pagination and rate limit information.
type Response struct {
	*http.Response

	// Pagination cursor extracted from response
	NextCursor string

	// Rate limiting information
	Rate Rate
}

// Rate represents the rate limit information returned in API responses.
type Rate struct {
	// The maximum number of requests that can be made in the current window.
	Limit int

	// The number of requests remaining in the current window.
	Remaining int

	// The time at which the current rate limit window resets.
	Reset time.Time
}

// ListOptions specifies the optional parameters to various List methods that
// support pagination.
type ListOptions struct {
	// Limit specifies the maximum number of items to return.
	// The API may return fewer than this value.
	Limit int `url:"limit,omitempty"`

	// Cursor is an opaque string used for pagination.
	// To get the next page of results, pass the NextCursor from the
	// previous response.
	Cursor string `url:"cursor,omitempty"`
}

// ServerListOptions specifies the optional parameters to the
// ServersService.List method.
type ServerListOptions struct {
	ListOptions

	// UpdatedSince filters servers updated after this timestamp (RFC3339)
	UpdatedSince *time.Time `url:"updated_since,omitempty"`

	// Search performs case-insensitive substring search on server names
	Search string `url:"search,omitempty"`

	// Version filter (supports "latest" for latest versions only)
	Version string `url:"version,omitempty"`
}
