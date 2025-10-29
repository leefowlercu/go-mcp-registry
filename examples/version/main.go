package main

import (
	"context"
	"fmt"
	"log"
	"os"

	mcp "github.com/leefowlercu/go-mcp-registry/mcp"
)

func main() {
	// Check if server name was provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <server-name> [version]")
		fmt.Println("\nThis example demonstrates version-specific retrieval using GetByNameExactVersion.")
		fmt.Println("If version is not provided, it lists all available versions first.")
		fmt.Println("\nExamples:")
		fmt.Println("  go run main.go ai.waystation/gmail")
		fmt.Println("  go run main.go ai.waystation/gmail 1.0.0")
		os.Exit(1)
	}

	serverName := os.Args[1]
	var targetVersion string
	if len(os.Args) >= 3 {
		targetVersion = os.Args[2]
	}

	// Create a client with default settings
	client := mcp.NewClient(nil)
	ctx := context.Background()

	// First, list all available versions for the server
	fmt.Printf("Fetching available versions for: %s\n", serverName)
	versions, _, err := client.Servers.ListVersionsByName(ctx, serverName)
	if err != nil {
		log.Fatalf("Error listing versions: %v", err)
	}

	if len(versions) == 0 {
		fmt.Printf("No versions found for server: %s\n", serverName)
		os.Exit(1)
	}

	fmt.Printf("\nFound %d version(s):\n", len(versions))
	for i, version := range versions {
		fmt.Printf("%d. Version %s\n", i+1, version.Version)
	}

	// If no version was specified, use the first one (typically the latest)
	if targetVersion == "" {
		targetVersion = versions[0].Version
		fmt.Printf("\nNo version specified, using: %s\n", targetVersion)
	}

	// Now fetch the specific version using GetByNameExactVersion
	// This method uses a dedicated API endpoint and is more performant
	// than client-side filtering
	fmt.Printf("\nFetching version %s using GetByNameExactVersion...\n", targetVersion)
	server, resp, err := client.Servers.GetByNameExactVersion(ctx, serverName, targetVersion)
	if err != nil {
		log.Fatalf("Error getting version: %v", err)
	}

	if server == nil {
		fmt.Printf("Version %s not found for server %s\n", targetVersion, serverName)
		os.Exit(1)
	}

	// Display server details
	fmt.Println("\n=== Server Details ===")
	fmt.Printf("Name: %s\n", server.Name)
	fmt.Printf("Version: %s\n", server.Version)
	fmt.Printf("Description: %s\n", server.Description)

	if server.Repository.URL != "" {
		fmt.Printf("Repository: %s\n", server.Repository.URL)
	}

	// Show remotes
	if len(server.Remotes) > 0 {
		fmt.Println("\nRemotes:")
		for _, remote := range server.Remotes {
			fmt.Printf("  - %s", remote.Type)
			if remote.URL != "" {
				fmt.Printf(": %s", remote.URL)
			}
			fmt.Println()
		}
	}

	// Show rate limit information
	if resp.Rate.Limit > 0 {
		fmt.Printf("\nRate Limit: %d/%d remaining\n", resp.Rate.Remaining, resp.Rate.Limit)
	}

	// Performance note
	fmt.Println("\n💡 Performance Note:")
	fmt.Println("GetByNameExactVersion uses the dedicated endpoint:")
	fmt.Printf("  GET /v0.1/servers/%s/versions/%s\n", serverName, targetVersion)
	fmt.Println("This is more efficient than client-side filtering for version lookups.")
}
