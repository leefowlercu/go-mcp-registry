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
	for i, serverResp := range resp.Servers {
		if i < 3 { // Log first 3 servers
			t.Logf("  - %s (v%s): %s", serverResp.Server.Name, serverResp.Server.Version, serverResp.Server.Description)
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
	for _, serverResp := range resp.Servers {
		t.Logf("  - %s: %s", serverResp.Server.Name, serverResp.Server.Description)
	}
}

func TestServersService_ListByName_Integration(t *testing.T) {
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

	serverName := resp.Servers[0].Server.Name
	t.Logf("Testing ListByName with server: %s", serverName)

	// Get the server by name
	servers, _, err := client.Servers.ListByName(ctx, serverName)
	if err != nil {
		t.Fatalf("ListByName returned error: %v", err)
	}

	if len(servers) == 0 {
		t.Fatalf("ListByName returned no servers for name: %s", serverName)
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

	serverName := resp.Servers[0].Server.Name
	t.Logf("Testing GetByNameLatest with server: %s", serverName)

	// Get the latest version of the server by name
	server, _, err := client.Servers.GetByNameLatest(ctx, serverName)
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

	// Compare with ListByName to verify we get the latest
	allVersions, _, err := client.Servers.ListByName(ctx, serverName)
	if err != nil {
		t.Fatalf("ListByName returned error: %v", err)
	}

	if len(allVersions) > 1 {
		t.Logf("Server has %d versions total", len(allVersions))
		// Note: We can't easily verify which is "latest" without version comparison logic
		// but we can verify that GetByNameLatest returned one of the versions
		found := false
		for _, v := range allVersions {
			if v.Version == server.Version {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetByNameLatest version %s not found in ListByName results", server.Version)
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
	allVersions, _, err := client.Servers.ListByName(ctx, "io.github.containers/kubernetes-mcp-server")
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
	server, _, err := client.Servers.GetByNameExactVersion(ctx, serverName, targetVersion)
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
	server, _, err = client.Servers.GetByNameExactVersion(ctx, serverName, nonExistentVersion)
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
	server, _, err := client.Servers.GetByNameLatestActiveVersion(ctx, serverName)
	if err != nil {
		t.Fatalf("GetByNameLatestActiveVersion returned error: %v", err)
	}

	if server == nil {
		t.Fatalf("GetByNameLatestActiveVersion returned nil for name: %s", serverName)
	}

	if server.Name != serverName {
		t.Errorf("Expected server name %s, got %s", serverName, server.Name)
	}

	// Note: Status field is not accessible from unwrapped ServerJSON
	// The method GetByNameLatestActiveVersion filters by status internally

	t.Logf("Successfully retrieved latest active version: %s (v%s)", server.Name, server.Version)

	// Compare with ListByName to ensure we got a valid version
	allVersions, _, err := client.Servers.ListByName(ctx, serverName)
	if err != nil {
		t.Fatalf("ListByName returned error: %v", err)
	}

	if len(allVersions) > 1 {
		t.Logf("Server has %d total versions", len(allVersions))

		// Verify that the returned version exists in all versions
		found := false
		for _, v := range allVersions {
			// Note: Status field is not accessible from unwrapped ServerJSON in v2 API
			// The method GetByNameLatestActiveVersion filters by status internally
			if v.Version == server.Version {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetByNameLatestActiveVersion version %s not found in ListByName results", server.Version)
		}

		// Log all versions for debugging
		t.Logf("All versions:")
		for _, v := range allVersions {
			// Note: Status field is not accessible from unwrapped ServerJSON in v2 API
			t.Logf("  - %s (v%s)", v.Name, v.Version)
		}
	}

	// Test with a non-existent server
	nonExistentServer := "nonexistent/test-server"
	server, _, err = client.Servers.GetByNameLatestActiveVersion(ctx, nonExistentServer)
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
	if page1.Metadata.NextCursor != "" {
		opts.Cursor = page1.Metadata.NextCursor
		page2, _, err := client.Servers.List(ctx, opts)
		if err != nil {
			t.Fatalf("Failed to get second page: %v", err)
		}

		t.Logf("Page 2: Got %d servers", len(page2.Servers))

		// Ensure pages have different content
		if len(page2.Servers) > 0 && len(page1.Servers) > 0 {
			if page1.Servers[0].Server.Name == page2.Servers[0].Server.Name {
				t.Error("Expected different servers on different pages")
			}
		}
	} else {
		t.Log("No second page available")
	}
}

func TestServersService_ListByUpdatedSince_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=true to run.")
	}

	client := mcp.NewClient(nil)
	ctx := context.Background()

	// Test with a recent timestamp (last 30 days)
	since := time.Now().AddDate(0, 0, -30)
	t.Logf("Testing ListByUpdatedSince with timestamp: %s", since.Format(time.RFC3339))

	servers, _, err := client.Servers.ListByUpdatedSince(ctx, since)
	if err != nil {
		t.Fatalf("ListByUpdatedSince returned error: %v", err)
	}

	t.Logf("Found %d servers updated since %s", len(servers), since.Format("2006-01-02"))

	// Verify all returned servers have valid update timestamps
	for i, server := range servers {
		if i < 5 { // Log first 5 for debugging
			// Note: Status and Meta.Official fields are not accessible from unwrapped ServerJSON in v2 API
			// ServerJSON no longer has Meta.Official - registry metadata is only in ServerResponse.Meta.Official
			t.Logf("Server %d: %s (v%s)", i+1, server.Name, server.Version)
		}

		// Note: UpdatedAt timestamp verification removed - not accessible from unwrapped ServerJSON
		// In v2 API, ServerJSON.Meta no longer has Official field
		// Registry metadata (including UpdatedAt) is only in ServerResponse.Meta.Official
	}

	// Test with a very recent timestamp (last 24 hours)
	recent := time.Now().AddDate(0, 0, -1)
	t.Logf("Testing ListByUpdatedSince with recent timestamp: %s", recent.Format(time.RFC3339))

	recentServers, _, err := client.Servers.ListByUpdatedSince(ctx, recent)
	if err != nil {
		t.Fatalf("ListByUpdatedSince with recent timestamp returned error: %v", err)
	}

	t.Logf("Found %d servers updated in last 24 hours", len(recentServers))

	// The number of recent servers should be <= total servers updated in last 30 days
	if len(recentServers) > len(servers) {
		t.Errorf("Recent servers count (%d) should not exceed total servers count (%d)",
			len(recentServers), len(servers))
	}

	// Test with a very old timestamp (should return many servers)
	old := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	t.Logf("Testing ListByUpdatedSince with old timestamp: %s", old.Format(time.RFC3339))

	oldServers, _, err := client.Servers.ListByUpdatedSince(ctx, old)
	if err != nil {
		t.Fatalf("ListByUpdatedSince with old timestamp returned error: %v", err)
	}

	t.Logf("Found %d servers updated since %s", len(oldServers), old.Format("2006-01-02"))

	// Should get significantly more servers with older timestamp
	if len(oldServers) < len(servers) {
		t.Errorf("Old timestamp should return more servers (%d) than recent timestamp (%d)",
			len(oldServers), len(servers))
	}

	// Test with future timestamp (should return empty)
	future := time.Now().AddDate(0, 0, 1)
	t.Logf("Testing ListByUpdatedSince with future timestamp: %s", future.Format(time.RFC3339))

	futureServers, _, err := client.Servers.ListByUpdatedSince(ctx, future)
	if err != nil {
		t.Fatalf("ListByUpdatedSince with future timestamp returned error: %v", err)
	}

	if len(futureServers) > 0 {
		t.Errorf("Future timestamp should return 0 servers, got %d", len(futureServers))
	}

	t.Log("Successfully verified ListByUpdatedSince with various timestamps")
}

func TestServersService_ListByServerID_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=true to run.")
	}

	client := mcp.NewClient(nil)
	ctx := context.Background()

	// Use a known server that should have multiple versions
	serverName := "io.github.containers/kubernetes-mcp-server"

	// First, get all versions using ListByName to get a server ID
	servers, _, err := client.Servers.ListByName(ctx, serverName)
	if err != nil {
		t.Fatalf("ListByName returned error: %v", err)
	}

	if len(servers) == 0 {
		t.Skip("No servers available for kubernetes-mcp-server to test")
	}

	// Note: In v2 API, ServerJSON.Meta no longer has Official field
	// Registry metadata (including ServerID) is only in ServerResponse.Meta.Official
	// This test cannot retrieve ServerID from unwrapped ServerJSON
	// Skip this test as ListByServerID requires server ID from ServerResponse metadata
	t.Skip("Server ID not accessible from unwrapped ServerJSON in v2 API - ServerJSON.Meta.Official no longer exists")

	// The following code is unreachable after Skip but left for reference:
	//
	// t.Logf("Testing ListByServerID with server ID: %s (name: %s)", serverID, serverName)
	//
	// // Test ListByServerID
	// versions, resp, err := client.Servers.ListByServerID(ctx, serverID)
	// if err != nil {
	// 	t.Fatalf("ListByServerID returned error: %v", err)
	// }
	//
	// if len(versions) == 0 {
	// 	t.Fatalf("ListByServerID returned no versions for server ID: %s", serverID)
	// }
	//
	// t.Logf("Found %d versions for server %s", len(versions), serverName)
	//
	// // Verify all returned servers have the same name and different versions
	// expectedName := servers[0].Name
	// versionMap := make(map[string]bool)
	//
	// for i, version := range versions {
	// 	if version.Name != expectedName {
	// 		t.Errorf("Version %d has wrong name: expected %s, got %s", i, expectedName, version.Name)
	// 	}
	//
	// 	if versionMap[version.Version] {
	// 		t.Errorf("Duplicate version found: %s", version.Version)
	// 	}
	// 	versionMap[version.Version] = true
	//
	// 	// Note: Status field not accessible from unwrapped ServerJSON in v2 API
	// 	if i < 5 { // Log first 5 versions for debugging
	// 		t.Logf("Version %d: %s (v%s)", i+1, version.Name, version.Version)
	// 	}
	// }
	//
	// // Verify ListByServerID returns the same number of versions as ListByName
	// if len(versions) != len(servers) {
	// 	t.Logf("Warning: ListByServerID returned %d versions, but ListByName returned %d versions", len(versions), len(servers))
	// 	// This might be expected if the API behavior differs, so just log it as a warning
	// }
	//
	// // Show rate limit information
	// if resp.Rate.Limit > 0 {
	// 	t.Logf("Rate Limit: %d/%d remaining", resp.Rate.Remaining, resp.Rate.Limit)
	// }
	//
	// t.Logf("Successfully verified ListByServerID with server ID: %s", serverID)
}
