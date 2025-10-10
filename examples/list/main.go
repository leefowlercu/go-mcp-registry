package main

import (
	"context"
	"fmt"
	"log"

	mcp "github.com/leefowlercu/go-mcp-registry/mcp"
)

func main() {
	// Create a client with default settings
	client := mcp.NewClient(nil)

	// List servers with default options
	fmt.Println("Listing servers...")
	opts := &mcp.ServerListOptions{
		ListOptions: mcp.ListOptions{
			Limit: 20, // Get 20 servers per page
		},
	}

	resp, _, err := client.Servers.List(context.Background(), opts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d servers\n", len(resp.Servers))
	for _, serverResponse := range resp.Servers {
		fmt.Printf("- %s (v%s): %s\n", serverResponse.Server.Name, serverResponse.Server.Version, serverResponse.Server.Description)
	}

	// Check if there are more pages
	if resp.Metadata.NextCursor != "" {
		fmt.Printf("\nMore results available with cursor: %s\n", resp.Metadata.NextCursor)
	}

	// Example with search filter
	fmt.Println("\nSearching for GitHub-related servers...")
	searchOpts := &mcp.ServerListOptions{
		Search: "github",
		ListOptions: mcp.ListOptions{
			Limit: 10,
		},
	}

	searchResp, _, err := client.Servers.List(context.Background(), searchOpts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d servers matching 'github'\n", len(searchResp.Servers))
	for _, serverResponse := range searchResp.Servers {
		fmt.Printf("- %s: %s\n", serverResponse.Server.Name, serverResponse.Server.Description)
	}
}
