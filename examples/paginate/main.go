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

	// Example 1: Manual pagination
	fmt.Println("Example 1: Manual pagination")
	fmt.Println("=============================")

	var allServers []string
	opts := &mcp.ServerListOptions{
		ListOptions: mcp.ListOptions{
			Limit: 5, // Small page size for demonstration
		},
	}

	pageNum := 1
	for {
		fmt.Printf("\nFetching page %d...\n", pageNum)
		resp, _, err := client.Servers.List(context.Background(), opts)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Got %d servers on page %d\n", len(resp.Servers), pageNum)
		for _, serverResponse := range resp.Servers {
			allServers = append(allServers, serverResponse.Server.Name)
			fmt.Printf("  - %s\n", serverResponse.Server.Name)
		}

		// Check if there are more pages
		if resp.Metadata.NextCursor == "" {
			fmt.Println("\nNo more pages available")
			break
		}

		// Update cursor for next page
		opts.Cursor = resp.Metadata.NextCursor
		pageNum++

		// For demonstration, limit to 3 pages
		if pageNum > 3 {
			fmt.Println("\nStopping after 3 pages for demonstration")
			break
		}
	}

	fmt.Printf("\nTotal servers collected: %d\n", len(allServers))

	// Example 2: Using ListAll helper method
	fmt.Println("\n\nExample 2: Using ListAll helper method")
	fmt.Println("=======================================")

	searchOpts := &mcp.ServerListOptions{
		Search: "mcp", // Search for MCP-related servers
		ListOptions: mcp.ListOptions{
			Limit: 10,
		},
	}

	fmt.Println("Fetching all MCP-related servers...")
	servers, _, err := client.Servers.ListAll(context.Background(), searchOpts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nFound %d total MCP-related servers:\n", len(servers))
	for i, server := range servers {
		fmt.Printf("%d. %s (v%s)\n", i+1, server.Name, server.Version)
		if i >= 9 { // Show first 10
			if len(servers) > 10 {
				fmt.Printf("... and %d more\n", len(servers)-10)
			}
			break
		}
	}
}
