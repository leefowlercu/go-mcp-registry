//go:build integration
// +build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	mcp "github.com/leefowlercu/go-mcp-registry/mcp"
)

func TestServersService_List_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=true to run.")
	}

	client := mcp.NewClient(nil)
	ctx := context.Background()

	// Test basic list
	opts := &mcp.ServerListOptions{
		ListOptions: mcp.ListOptions{
			Limit: 10,
		},
	}

	resp, _, err := client.Servers.List(ctx, opts)
	if err != nil {
		t.Fatalf("Servers.List returned error: %v", err)
	}

	if len(resp.Servers) == 0 {
		t.Error("Expected at least one server in the registry")
	}

	t.Logf("Found %d servers", len(resp.Servers))
	for i, server := range resp.Servers {
		if i < 3 { // Log first 3 servers
			t.Logf("  - %s (v%s): %s", server.Name, server.Version, server.Description)
		}
	}
}

func TestServersService_Search_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=true to run.")
	}

	client := mcp.NewClient(nil)
	ctx := context.Background()

	// Search for MCP-related servers
	opts := &mcp.ServerListOptions{
		Search: "mcp",
		ListOptions: mcp.ListOptions{
			Limit: 5,
		},
	}

	resp, _, err := client.Servers.List(ctx, opts)
	if err != nil {
		t.Fatalf("Servers.List with search returned error: %v", err)
	}

	t.Logf("Found %d servers matching 'mcp'", len(resp.Servers))
	for _, server := range resp.Servers {
		t.Logf("  - %s: %s", server.Name, server.Description)
	}
}

func TestServersService_GetByName_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=true to run.")
	}

	client := mcp.NewClient(nil)
	ctx := context.Background()

	// First, get a list to find a valid server name
	resp, _, err := client.Servers.List(ctx, &mcp.ServerListOptions{
		ListOptions: mcp.ListOptions{Limit: 1},
	})
	if err != nil {
		t.Fatalf("Failed to list servers: %v", err)
	}

	if len(resp.Servers) == 0 {
		t.Skip("No servers available to test")
	}

	serverName := resp.Servers[0].Name
	t.Logf("Testing GetByName with server: %s", serverName)

	// Get the server by name
	servers, err := client.Servers.GetByName(ctx, serverName)
	if err != nil {
		t.Fatalf("GetByName returned error: %v", err)
	}

	if len(servers) == 0 {
		t.Fatalf("GetByName returned no servers for name: %s", serverName)
	}

	// Verify all returned servers have the expected name
	for _, server := range servers {
		if server.Name != serverName {
			t.Errorf("Expected server name %s, got %s", serverName, server.Name)
		}
	}

	t.Logf("Successfully retrieved %d server(s) by name: %s", len(servers), serverName)
	if len(servers) > 1 {
		t.Logf("Multiple versions found:")
		for _, server := range servers {
			t.Logf("  - %s (v%s)", server.Name, server.Version)
		}
	} else {
		t.Logf("  - %s (v%s)", servers[0].Name, servers[0].Version)
	}
}

func TestServersService_GetByNameLatest_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=true to run.")
	}

	client := mcp.NewClient(nil)
	ctx := context.Background()

	// First, get a list to find a valid server name
	resp, _, err := client.Servers.List(ctx, &mcp.ServerListOptions{
		ListOptions: mcp.ListOptions{Limit: 1},
	})
	if err != nil {
		t.Fatalf("Failed to list servers: %v", err)
	}

	if len(resp.Servers) == 0 {
		t.Skip("No servers available to test")
	}

	serverName := resp.Servers[0].Name
	t.Logf("Testing GetByNameLatest with server: %s", serverName)

	// Get the latest version of the server by name
	server, err := client.Servers.GetByNameLatest(ctx, serverName)
	if err != nil {
		t.Fatalf("GetByNameLatest returned error: %v", err)
	}

	if server == nil {
		t.Fatalf("GetByNameLatest returned nil for name: %s", serverName)
	}

	if server.Name != serverName {
		t.Errorf("Expected server name %s, got %s", serverName, server.Name)
	}

	t.Logf("Successfully retrieved latest version: %s (v%s)", server.Name, server.Version)

	// Compare with GetByName to verify we get the latest
	allVersions, err := client.Servers.GetByName(ctx, serverName)
	if err != nil {
		t.Fatalf("GetByName returned error: %v", err)
	}

	if len(allVersions) > 1 {
		t.Logf("Server has %d versions total", len(allVersions))
		// Note: We can't easily verify which is "latest" without version comparison logic
		// but we can verify that GetLatestByName returned one of the versions
		found := false
		for _, v := range allVersions {
			if v.Version == server.Version {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetByNameLatest version %s not found in GetByName results", server.Version)
		}
	}
}

func TestServersService_GetByNameExactVersion_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=true to run.")
	}

	client := mcp.NewClient(nil)
	ctx := context.Background()

	// First, get a list to find a server with multiple versions
	allVersions, err := client.Servers.GetByName(ctx, "io.github.containers/kubernetes-mcp-server")
	if err != nil {
		t.Fatalf("Failed to get server versions: %v", err)
	}

	if len(allVersions) == 0 {
		t.Skip("No kubernetes-mcp-server available to test")
	}

	// Test with the first version found
	targetVersion := allVersions[0].Version
	serverName := allVersions[0].Name
	t.Logf("Testing GetByNameExactVersion with server: %s, version: %s", serverName, targetVersion)

	// Get the specific version
	server, err := client.Servers.GetByNameExactVersion(ctx, serverName, targetVersion)
	if err != nil {
		t.Fatalf("GetByNameExactVersion returned error: %v", err)
	}

	if server == nil {
		t.Fatalf("GetByNameExactVersion returned nil for name: %s, version: %s", serverName, targetVersion)
	}

	if server.Name != serverName {
		t.Errorf("Expected server name %s, got %s", serverName, server.Name)
	}

	if server.Version != targetVersion {
		t.Errorf("Expected server version %s, got %s", targetVersion, server.Version)
	}

	t.Logf("Successfully retrieved specific version: %s (v%s)", server.Name, server.Version)

	// Test with a non-existent version
	nonExistentVersion := "999.999.999"
	server, err = client.Servers.GetByNameExactVersion(ctx, serverName, nonExistentVersion)
	if err != nil {
		t.Fatalf("GetByNameExactVersion returned error for non-existent version: %v", err)
	}

	if server != nil {
		t.Errorf("Expected nil for non-existent version %s, got %+v", nonExistentVersion, server)
	}

	t.Logf("Correctly returned nil for non-existent version: %s", nonExistentVersion)
}

func TestServersService_GetByNameLatestActiveVersion_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=true to run.")
	}

	client := mcp.NewClient(nil)
	ctx := context.Background()

	// Test with kubernetes server which has multiple versions
	serverName := "io.github.containers/kubernetes-mcp-server"
	t.Logf("Testing GetByNameLatestActiveVersion with server: %s", serverName)

	// Get the latest active version
	server, err := client.Servers.GetByNameLatestActiveVersion(ctx, serverName)
	if err != nil {
		t.Fatalf("GetByNameLatestActiveVersion returned error: %v", err)
	}

	if server == nil {
		t.Fatalf("GetByNameLatestActiveVersion returned nil for name: %s", serverName)
	}

	if server.Name != serverName {
		t.Errorf("Expected server name %s, got %s", serverName, server.Name)
	}

	if server.Status != "active" {
		t.Errorf("Expected active status, got %s", server.Status)
	}

	t.Logf("Successfully retrieved latest active version: %s (v%s) - %s", server.Name, server.Version, server.Status)

	// Compare with GetByName to ensure we got a valid version
	allVersions, err := client.Servers.GetByName(ctx, serverName)
	if err != nil {
		t.Fatalf("GetByName returned error: %v", err)
	}

	if len(allVersions) > 1 {
		t.Logf("Server has %d total versions", len(allVersions))

		// Verify that the returned version exists in all versions
		found := false
		for _, v := range allVersions {
			if v.Version == server.Version && v.Status == "active" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetByNameLatestActiveVersion version %s (status: %s) not found in GetByName results", server.Version, server.Status)
		}

		// Log all versions for debugging
		t.Logf("All versions:")
		for _, v := range allVersions {
			t.Logf("  - %s (v%s) - %s", v.Name, v.Version, v.Status)
		}
	}

	// Test with a non-existent server
	nonExistentServer := "nonexistent/test-server"
	server, err = client.Servers.GetByNameLatestActiveVersion(ctx, nonExistentServer)
	if err != nil {
		t.Fatalf("GetByNameLatestActiveVersion returned error for non-existent server: %v", err)
	}

	if server != nil {
		t.Errorf("Expected nil for non-existent server %s, got %+v", nonExistentServer, server)
	}

	t.Logf("Correctly returned nil for non-existent server: %s", nonExistentServer)
}

func TestServersService_Pagination_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=true to run.")
	}

	client := mcp.NewClient(nil)
	ctx := context.Background()

	// Test pagination with small page size
	opts := &mcp.ServerListOptions{
		ListOptions: mcp.ListOptions{
			Limit: 2,
		},
	}

	// Get first page
	page1, _, err := client.Servers.List(ctx, opts)
	if err != nil {
		t.Fatalf("Failed to get first page: %v", err)
	}

	if len(page1.Servers) == 0 {
		t.Skip("No servers available to test pagination")
	}

	t.Logf("Page 1: Got %d servers", len(page1.Servers))

	// If there's a next page, fetch it
	if page1.Metadata != nil && page1.Metadata.NextCursor != "" {
		opts.Cursor = page1.Metadata.NextCursor
		page2, _, err := client.Servers.List(ctx, opts)
		if err != nil {
			t.Fatalf("Failed to get second page: %v", err)
		}

		t.Logf("Page 2: Got %d servers", len(page2.Servers))

		// Ensure pages have different content
		if len(page2.Servers) > 0 && len(page1.Servers) > 0 {
			if page1.Servers[0].Name == page2.Servers[0].Name {
				t.Error("Expected different servers on different pages")
			}
		}
	} else {
		t.Log("No second page available")
	}
}

func TestServersService_ListUpdatedSince_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=true to run.")
	}

	client := mcp.NewClient(nil)
	ctx := context.Background()

	// Test with a recent timestamp (last 30 days)
	since := time.Now().AddDate(0, 0, -30)
	t.Logf("Testing ListUpdatedSince with timestamp: %s", since.Format(time.RFC3339))

	servers, err := client.Servers.ListUpdatedSince(ctx, since)
	if err != nil {
		t.Fatalf("ListUpdatedSince returned error: %v", err)
	}

	t.Logf("Found %d servers updated since %s", len(servers), since.Format("2006-01-02"))

	// Verify all returned servers have valid update timestamps
	for i, server := range servers {
		if i < 5 { // Log first 5 for debugging
			t.Logf("Server %d: %s (v%s) - status: %s", i+1, server.Name, server.Version, server.Status)
			if server.Meta != nil && server.Meta.Official != nil {
				t.Logf("  ID: %s, Updated: %s", server.Meta.Official.ID, server.Meta.Official.UpdatedAt.Format(time.RFC3339))
			}
		}

		// Verify timestamp if available
		if server.Meta != nil && server.Meta.Official != nil && !server.Meta.Official.UpdatedAt.IsZero() {
			serverUpdatedAt := server.Meta.Official.UpdatedAt
			if serverUpdatedAt.Before(since) {
				t.Errorf("Server %s updated_at %s is before since timestamp %s",
					server.Meta.Official.ID, serverUpdatedAt.Format(time.RFC3339), since.Format(time.RFC3339))
			}
		}
	}

	// Test with a very recent timestamp (last 24 hours)
	recent := time.Now().AddDate(0, 0, -1)
	t.Logf("Testing ListUpdatedSince with recent timestamp: %s", recent.Format(time.RFC3339))

	recentServers, err := client.Servers.ListUpdatedSince(ctx, recent)
	if err != nil {
		t.Fatalf("ListUpdatedSince with recent timestamp returned error: %v", err)
	}

	t.Logf("Found %d servers updated in last 24 hours", len(recentServers))

	// The number of recent servers should be <= total servers updated in last 30 days
	if len(recentServers) > len(servers) {
		t.Errorf("Recent servers count (%d) should not exceed total servers count (%d)",
			len(recentServers), len(servers))
	}

	// Test with a very old timestamp (should return many servers)
	old := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	t.Logf("Testing ListUpdatedSince with old timestamp: %s", old.Format(time.RFC3339))

	oldServers, err := client.Servers.ListUpdatedSince(ctx, old)
	if err != nil {
		t.Fatalf("ListUpdatedSince with old timestamp returned error: %v", err)
	}

	t.Logf("Found %d servers updated since %s", len(oldServers), old.Format("2006-01-02"))

	// Should get significantly more servers with older timestamp
	if len(oldServers) < len(servers) {
		t.Errorf("Old timestamp should return more servers (%d) than recent timestamp (%d)",
			len(oldServers), len(servers))
	}

	// Test with future timestamp (should return empty)
	future := time.Now().AddDate(0, 0, 1)
	t.Logf("Testing ListUpdatedSince with future timestamp: %s", future.Format(time.RFC3339))

	futureServers, err := client.Servers.ListUpdatedSince(ctx, future)
	if err != nil {
		t.Fatalf("ListUpdatedSince with future timestamp returned error: %v", err)
	}

	if len(futureServers) > 0 {
		t.Errorf("Future timestamp should return 0 servers, got %d", len(futureServers))
	}

	t.Log("Successfully verified ListUpdatedSince with various timestamps")
}
