package main

import (
	"fmt"
	"log"

	"github.com/sharding-system/pkg/client"
)

func main() {
	// Create a client pointing to the router
	client := client.NewClient("http://localhost:8080")

	// Get shard for a key
	shardID, err := client.GetShardForKey("user-123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Shard ID for 'user-123': %s\n", shardID)

	// Execute a query with strong consistency (reads from primary)
	result, err := client.QueryStrong(
		"user-123",
		"SELECT * FROM users WHERE id = $1",
		"user-123",
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Query returned %d rows\n", result.RowCount)
	fmt.Printf("Latency: %.2f ms\n", result.LatencyMs)

	// Execute a query with eventual consistency (can read from replica)
	result, err = client.QueryEventual(
		"user-123",
		"SELECT COUNT(*) FROM users",
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Count query returned %d rows\n", result.RowCount)
}

