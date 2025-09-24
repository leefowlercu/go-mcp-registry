package main

import (
	"context"
	"fmt"
	"log"
	"os"

	mcp "github.com/leefowlercu/go-mcp-registry/mcp"
	registryv0 "github.com/modelcontextprotocol/registry/pkg/api/v0"
)

func main() {
	// Check if server identifier was provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <server-id-or-name>")
		fmt.Println("\nYou can use either:")
		fmt.Println("  1. Server ID (from list endpoint)")
		fmt.Println("  2. Server Name (e.g., 'ai.waystation/gmail')")
		fmt.Println("\nTo see available servers, run:")
		fmt.Println("  go run ../list/main.go")
		os.Exit(1)
	}

	serverIdentifier := os.Args[1]

	// Create a client with default settings
	client := mcp.NewClient(nil)
	ctx := context.Background()

	var server *registryv0.ServerJSON
	var resp *mcp.Response
	var err error

	// Try to get server by ID first (UUID format)
	fmt.Printf("Getting details for server: %s\n", serverIdentifier)
	server, resp, err = client.Servers.Get(ctx, serverIdentifier, nil)

	// If that fails with 404, try to get by name
	if err != nil && resp != nil && resp.StatusCode == 404 {
		fmt.Printf("Server ID not found, trying by name...\n")
		servers, _, err := client.Servers.ListByName(ctx, serverIdentifier)
		if err != nil {
			log.Fatal(err)
		}
		if len(servers) == 0 {
			fmt.Printf("Server with identifier '%s' not found\n", serverIdentifier)
			fmt.Println("\nTip: Use 'go run ../list/main.go' to see available servers")
			os.Exit(1)
		}

		// If multiple versions found, use the first one (could be enhanced to select latest)
		server = &servers[0]
		if len(servers) > 1 {
			fmt.Printf("Found %d versions of server '%s', using version %s\n", len(servers), serverIdentifier, server.Version)
		}

		// Create a mock response for consistency
		resp = &mcp.Response{}
	} else if err != nil {
		log.Fatal(err)
	}

	// Display server information
	fmt.Println("\nServer Details:")
	fmt.Printf("Name: %s\n", server.Name)
	fmt.Printf("Version: %s\n", server.Version)
	fmt.Printf("Description: %s\n", server.Description)

	if server.Repository.URL != "" {
		fmt.Printf("Repository: %s\n", server.Repository.URL)
	}

	// Show server status if available
	if server.Status != "" {
		fmt.Printf("Status: %s\n", server.Status)
	}

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

	// Show metadata if available
	if server.Meta != nil && server.Meta.Official != nil {
		fmt.Println("\nRegistry Metadata:")
		fmt.Printf("Server ID: %s\n", server.Meta.Official.ServerID)
		fmt.Printf("Published: %s\n", server.Meta.Official.PublishedAt.Format("2006-01-02 15:04:05"))
		if !server.Meta.Official.UpdatedAt.IsZero() {
			fmt.Printf("Updated: %s\n", server.Meta.Official.UpdatedAt.Format("2006-01-02 15:04:05"))
		}
		fmt.Printf("Latest: %t\n", server.Meta.Official.IsLatest)
	}

	// Show rate limit information
	if resp.Rate.Limit > 0 {
		fmt.Printf("\nRate Limit: %d/%d remaining\n", resp.Rate.Remaining, resp.Rate.Limit)
	}
}
