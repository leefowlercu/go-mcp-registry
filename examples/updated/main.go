package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	mcp "github.com/leefowlercu/go-mcp-registry/mcp"
)

func main() {
	// Create a client with default settings
	client := mcp.NewClient(nil)
	ctx := context.Background()

	// Parse command-line arguments for custom time period
	hours := 24
	if len(os.Args) >= 2 {
		var err error
		hours, err = strconv.Atoi(os.Args[1])
		if err != nil {
			fmt.Println("Usage: go run main.go [hours]")
			fmt.Println("\nExamples:")
			fmt.Println("  go run main.go       # Check last 24 hours (default)")
			fmt.Println("  go run main.go 1     # Check last 1 hour")
			fmt.Println("  go run main.go 168   # Check last 7 days (168 hours)")
			os.Exit(1)
		}
	}

	// Calculate the timestamp for the lookback period
	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	fmt.Printf("Fetching servers updated in the last %d hour(s)...\n", hours)
	fmt.Printf("Since: %s\n\n", since.Format(time.RFC3339))

	// List servers updated since the specified timestamp
	// This method automatically handles pagination to return all matching servers
	servers, resp, err := client.Servers.ListByUpdatedSince(ctx, since)
	if err != nil {
		log.Fatalf("Error listing updated servers: %v", err)
	}

	// Display results
	if len(servers) == 0 {
		fmt.Printf("No servers have been updated in the last %d hour(s).\n", hours)
	} else {
		fmt.Printf("Found %d server(s) updated in the last %d hour(s):\n\n", len(servers), hours)

		// Group by server name (since we may have multiple versions)
		serverMap := make(map[string][]string)
		for _, server := range servers {
			serverMap[server.Name] = append(serverMap[server.Name], server.Version)
		}

		// Display grouped results
		i := 1
		for name, versions := range serverMap {
			fmt.Printf("%d. %s\n", i, name)
			if len(versions) == 1 {
				fmt.Printf("   Version: %s\n", versions[0])
			} else {
				fmt.Printf("   Versions: %v\n", versions)
			}
			i++
		}
	}

	// Show rate limit information
	if resp.Rate.Limit > 0 {
		fmt.Printf("\nRate Limit: %d/%d remaining\n", resp.Rate.Remaining, resp.Rate.Limit)
	}

	// Usage tips
	fmt.Println("\n💡 Use Cases:")
	fmt.Println("• Monitor registry for new server releases")
	fmt.Println("• Track changes to server configurations")
	fmt.Println("• Build automated update notification systems")
	fmt.Println("• Audit server update frequency")

	fmt.Println("\n💡 Performance Note:")
	fmt.Println("ListByUpdatedSince automatically handles pagination to fetch")
	fmt.Println("all matching results, making it easy to track registry changes.")
}
