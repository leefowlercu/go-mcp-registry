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
		fmt.Println("Usage: go run main.go <server-name>")
		fmt.Println("\nExample:")
		fmt.Println("  go run main.go ai.waystation/gmail")
		fmt.Println("\nTo see available servers, run:")
		fmt.Println("  go run ../list/main.go")
		os.Exit(1)
	}

	serverName := os.Args[1]

	// Create a client with default settings
	client := mcp.NewClient(nil)
	ctx := context.Background()

	// Get server by name (API v2 uses names, not IDs)
	fmt.Printf("Getting details for server: %s\n", serverName)
	server, resp, err := client.Servers.Get(ctx, serverName, nil)
	if err != nil {
		log.Fatal(err)
	}

	if server == nil {
		fmt.Printf("Server '%s' not found\n", serverName)
		fmt.Println("\nTip: Use 'go run ../list/main.go' to see available servers")
		os.Exit(1)
	}

	// Display server information
	fmt.Println("\nServer Details:")
	fmt.Printf("Name: %s\n", server.Name)
	fmt.Printf("Version: %s\n", server.Version)
	fmt.Printf("Description: %s\n", server.Description)

	if server.Repository.URL != "" {
		fmt.Printf("Repository: %s\n", server.Repository.URL)
	}

	// Note: Status field has been removed from ServerJSON in API v2
	// Status is now part of registry metadata in ServerResponse.Meta.Official

	// Show remotes (transport configurations)
	if len(server.Remotes) > 0 {
		fmt.Println("\nRemotes:")
		for _, remote := range server.Remotes {
			fmt.Printf("- Type: %s\n", remote.Type)
			if remote.URL != "" {
				fmt.Printf("  URL: %s\n", remote.URL)
			}
		}
	}

	// Show packages
	if len(server.Packages) > 0 {
		fmt.Printf("\nPackages: %d available\n", len(server.Packages))
		for i, pkg := range server.Packages {
			if i < 3 { // Show first 3 packages
				fmt.Printf("- Registry: %s\n", pkg.RegistryType)
				fmt.Printf("  Identifier: %s\n", pkg.Identifier)
				if pkg.Version != "" {
					fmt.Printf("  Version: %s\n", pkg.Version)
				}
				fmt.Println()
			}
		}
		if len(server.Packages) > 3 {
			fmt.Printf("... and %d more packages\n", len(server.Packages)-3)
		}
	}

	// Note: Registry metadata (ServerID, PublishedAt, UpdatedAt, IsLatest, Status)
	// has been moved from ServerJSON.Meta.Official to ServerResponse.Meta.Official in API v2.
	// Since Get() returns unwrapped ServerJSON, this metadata is not directly accessible here.
	// To access registry metadata, you would need to use List() which returns ServerResponse.

	// Show rate limit information
	if resp.Rate.Limit > 0 {
		fmt.Printf("\nRate Limit: %d/%d remaining\n", resp.Rate.Remaining, resp.Rate.Limit)
	}
}
