package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sharding-system/pkg/models"
)

// Client is the sharding client library for microservices
type Client struct {
	routerURL string
	httpClient *http.Client
}

// NewClient creates a new sharding client
func NewClient(routerURL string) *Client {
	return &Client{
		routerURL: routerURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetShardForKey returns the shard ID for a given key
func (c *Client) GetShardForKey(key string) (string, error) {
	url := fmt.Sprintf("%s/v1/shard-for-key?key=%s", c.routerURL, key)
	
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get shard: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		ShardID string `json:"shard_id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.ShardID, nil
}

// Query executes a query on the appropriate shard
func (c *Client) Query(shardKey string, query string, params []interface{}, consistency string) (*models.QueryResponse, error) {
	req := &models.QueryRequest{
		ShardKey:    shardKey,
		Query:       query,
		Params:      params,
		Consistency: consistency,
	}

	url := fmt.Sprintf("%s/v1/execute", c.routerURL)
	
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("query failed: %s", string(bodyBytes))
	}

	var result models.QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// QueryStrong executes a query with strong consistency (reads from primary)
func (c *Client) QueryStrong(shardKey string, query string, params ...interface{}) (*models.QueryResponse, error) {
	return c.Query(shardKey, query, params, "strong")
}

// QueryEventual executes a query with eventual consistency (can read from replica)
func (c *Client) QueryEventual(shardKey string, query string, params ...interface{}) (*models.QueryResponse, error) {
	return c.Query(shardKey, query, params, "eventual")
}

