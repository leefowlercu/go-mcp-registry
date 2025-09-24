package mcp

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Masterminds/semver/v3"
	registryv0 "github.com/modelcontextprotocol/registry/pkg/api/v0"
	"github.com/modelcontextprotocol/registry/pkg/model"
)

// List retrieves a paginated list of servers from the MCP Registry.
//
// MCP Registry API docs: https://registry.modelcontextprotocol.io/docs#/servers/get_servers_v0_servers_get
func (s *ServersService) List(ctx context.Context, opts *ServerListOptions) (*registryv0.ServerListResponse, *Response, error) {
	u := "v0/servers"
	u, err := addOptions(u, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var servers *registryv0.ServerListResponse
	resp, err := s.client.Do(ctx, req, &servers)
	if err != nil {
		return nil, resp, err
	}

	// Extract NextCursor from the response metadata
	if servers != nil && servers.Metadata.NextCursor != "" {
		resp.NextCursor = servers.Metadata.NextCursor
	}

	return servers, resp, nil
}

// Get retrieves a specific server by its server ID.
// Optionally specify a version to retrieve a specific version instead of the latest.
//
// MCP Registry API docs: https://registry.modelcontextprotocol.io/docs#/operations/get-server
func (s *ServersService) Get(ctx context.Context, serverID string, opts *ServerGetOptions) (*registryv0.ServerJSON, *Response, error) {
	u := fmt.Sprintf("v0/servers/%s", serverID)
	u, err := addOptions(u, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var server *registryv0.ServerJSON
	resp, err := s.client.Do(ctx, req, &server)
	if err != nil {
		return nil, resp, err
	}

	return server, resp, nil
}

// ListByServerID retrieves all available versions for a specific server by its server ID.
// Returns all versions of the server in a slice.
//
// MCP Registry API docs: https://registry.modelcontextprotocol.io/docs#/operations/get-server-versions
func (s *ServersService) ListByServerID(ctx context.Context, serverID string) ([]registryv0.ServerJSON, *Response, error) {
	u := fmt.Sprintf("v0/servers/%s/versions", serverID)

	req, err := s.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var serverResp *registryv0.ServerListResponse
	resp, err := s.client.Do(ctx, req, &serverResp)
	if err != nil {
		return nil, resp, err
	}

	// Extract servers from the response
	var servers []registryv0.ServerJSON
	if serverResp != nil && serverResp.Servers != nil {
		servers = serverResp.Servers
	}

	return servers, resp, nil
}

// ListAll fetches all pages of results for servers.
// This is a convenience method that handles pagination automatically.
func (s *ServersService) ListAll(ctx context.Context, opts *ServerListOptions) ([]registryv0.ServerJSON, *Response, error) {
	if opts == nil {
		opts = &ServerListOptions{}
	}

	var allServers []registryv0.ServerJSON
	var lastResp *Response

	for {
		resp, httpResp, err := s.List(ctx, opts)
		if err != nil {
			return allServers, httpResp, err
		}

		lastResp = httpResp

		if resp.Servers != nil {
			allServers = append(allServers, resp.Servers...)
		}

		// Check if there are more pages
		if resp.Metadata.NextCursor == "" {
			break
		}

		// Update cursor for next request
		opts.Cursor = resp.Metadata.NextCursor
	}

	return allServers, lastResp, nil
}

// ListByName retrieves all servers with the specified name.
// Since each server can have multiple versions in the registry,
// this method returns a slice containing all matching servers.
// Returns an empty slice if no matches are found.
func (s *ServersService) ListByName(ctx context.Context, name string) ([]registryv0.ServerJSON, *Response, error) {
	opts := &ServerListOptions{
		Search: name,
		ListOptions: ListOptions{
			Limit: 100,
		},
	}

	var matchingServers []registryv0.ServerJSON
	var lastResp *Response

	for {
		resp, httpResp, err := s.List(ctx, opts)
		if err != nil {
			return nil, httpResp, err
		}

		lastResp = httpResp

		// Collect all exact matches
		for _, server := range resp.Servers {
			if server.Name == name {
				matchingServers = append(matchingServers, server)
			}
		}

		// Check if there are more pages
		if resp.Metadata.NextCursor == "" {
			break
		}

		opts.Cursor = resp.Metadata.NextCursor
	}

	return matchingServers, lastResp, nil
}

// GetByNameLatest retrieves the latest version of a server with the specified name.
// This method uses the version=latest query parameter to filter results to only
// the latest version, then returns the match.
// Returns nil if no latest version is found.
func (s *ServersService) GetByNameLatest(ctx context.Context, name string) (*registryv0.ServerJSON, *Response, error) {
	opts := &ServerListOptions{
		Search:  name,
		Version: "latest",
		ListOptions: ListOptions{
			Limit: 100,
		},
	}

	var lastResp *Response

	for {
		resp, httpResp, err := s.List(ctx, opts)
		if err != nil {
			return nil, httpResp, err
		}

		lastResp = httpResp

		// Look for exact match
		for _, server := range resp.Servers {
			if server.Name == name {
				return &server, lastResp, nil
			}
		}

		// Check if there are more pages
		if resp.Metadata.NextCursor == "" {
			break
		}

		opts.Cursor = resp.Metadata.NextCursor
	}

	return nil, lastResp, nil
}

// GetByNameExactVersion retrieves a specific version of a server with the specified name.
// Since the API only supports version="latest", this method performs client-side
// filtering to find the exact name and version match.
// Returns nil if no matching version is found.
func (s *ServersService) GetByNameExactVersion(ctx context.Context, name, version string) (*registryv0.ServerJSON, *Response, error) {
	opts := &ServerListOptions{
		Search: name,
		ListOptions: ListOptions{
			Limit: 100,
		},
	}

	var lastResp *Response

	for {
		resp, httpResp, err := s.List(ctx, opts)
		if err != nil {
			return nil, httpResp, err
		}

		lastResp = httpResp

		// Look for exact name and version match
		for _, server := range resp.Servers {
			if server.Name == name && server.Version == version {
				return &server, lastResp, nil
			}
		}

		// Check if there are more pages
		if resp.Metadata.NextCursor == "" {
			break
		}

		opts.Cursor = resp.Metadata.NextCursor
	}

	return nil, lastResp, nil
}

// GetByNameLatestActiveVersion retrieves the latest active version of a server with the specified name.
// This method performs client-side filtering to find servers with Status == "active",
// then uses semantic version comparison to determine the latest version.
// Returns nil if no active versions are found.
func (s *ServersService) GetByNameLatestActiveVersion(ctx context.Context, name string) (*registryv0.ServerJSON, *Response, error) {
	opts := &ServerListOptions{
		Search: name,
		ListOptions: ListOptions{
			Limit: 100,
		},
	}

	var latestServer *registryv0.ServerJSON
	var latestVersion *semver.Version
	var lastResp *Response

	for {
		resp, httpResp, err := s.List(ctx, opts)
		if err != nil {
			return nil, httpResp, err
		}

		lastResp = httpResp

		// Look for active servers with exact name match
		for _, server := range resp.Servers {
			if server.Name == name && server.Status == model.StatusActive {
				// Try to parse the version as semantic version
				version, err := semver.NewVersion(server.Version)
				if err != nil {
					// Skip servers with invalid semantic versions
					continue
				}

				// Keep track of the latest version
				if latestVersion == nil || version.GreaterThan(latestVersion) {
					latestVersion = version
					serverCopy := server // Create a copy to avoid pointer issues
					latestServer = &serverCopy
				}
			}
		}

		// Check if there are more pages
		if resp.Metadata.NextCursor == "" {
			break
		}

		opts.Cursor = resp.Metadata.NextCursor
	}

	return latestServer, lastResp, nil
}

// ListByUpdatedSince retrieves all servers that have been updated since the specified timestamp.
// This method automatically handles pagination to return all matching servers.
// The timestamp should be in RFC3339 format.
// Returns an empty slice if no servers have been updated since the timestamp.
func (s *ServersService) ListByUpdatedSince(ctx context.Context, since time.Time) ([]registryv0.ServerJSON, *Response, error) {
	opts := &ServerListOptions{
		UpdatedSince: &since,
		ListOptions: ListOptions{
			Limit: 100,
		},
	}

	var updatedServers []registryv0.ServerJSON
	var lastResp *Response

	for {
		resp, httpResp, err := s.List(ctx, opts)
		if err != nil {
			return updatedServers, httpResp, err
		}

		lastResp = httpResp

		if resp.Servers != nil {
			updatedServers = append(updatedServers, resp.Servers...)
		}

		// Check if there are more pages
		if resp.Metadata.NextCursor == "" {
			break
		}

		opts.Cursor = resp.Metadata.NextCursor
	}

	return updatedServers, lastResp, nil
}
