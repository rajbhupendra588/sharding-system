package proxy

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// ShardingRule defines how a table should be sharded
type ShardingRule struct {
	Table       string `json:"table"`
	ShardKey    string `json:"shard_key"`     // Column to shard by (e.g., "user_id")
	Strategy    string `json:"strategy"`      // "hash", "range", "broadcast"
	Description string `json:"description"`
}

// ClientAppConfig holds sharding configuration for a client application
type ClientAppConfig struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Database      string         `json:"database"`       // Database name
	ShardingRules []ShardingRule `json:"sharding_rules"` // Table-level sharding rules
	DefaultShard  string         `json:"default_shard"`  // Default shard for unsharded tables
}

// ProxyConfig holds the proxy server configuration
type ProxyConfig struct {
	ListenAddr    string                      `json:"listen_addr"`    // e.g., ":5432"
	AdminAddr     string                      `json:"admin_addr"`     // e.g., ":8082"
	ManagerURL    string                      `json:"manager_url"`    // Sharding manager URL
	ClientApps    map[string]*ClientAppConfig `json:"client_apps"`    // App configs by database name
	mu            sync.RWMutex
}

// NewProxyConfig creates a new proxy configuration
func NewProxyConfig() *ProxyConfig {
	return &ProxyConfig{
		ListenAddr: ":5432",
		AdminAddr:  ":8082",
		ManagerURL: "http://localhost:8081",
		ClientApps: make(map[string]*ClientAppConfig),
	}
}

// LoadFromFile loads configuration from a JSON file
func (c *ProxyConfig) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}
	
	return nil
}

// SaveToFile saves configuration to a JSON file
func (c *ProxyConfig) SaveToFile(path string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(path, data, 0644)
}

// GetAppConfig returns the configuration for a database
func (c *ProxyConfig) GetAppConfig(database string) *ClientAppConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ClientApps[database]
}

// SetAppConfig sets the configuration for a database
func (c *ProxyConfig) SetAppConfig(database string, config *ClientAppConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ClientApps[database] = config
}

// GetShardingRule returns the sharding rule for a table
func (c *ClientAppConfig) GetShardingRule(table string) *ShardingRule {
	for i := range c.ShardingRules {
		if c.ShardingRules[i].Table == table {
			return &c.ShardingRules[i]
		}
	}
	return nil
}

// AddShardingRule adds or updates a sharding rule
func (c *ClientAppConfig) AddShardingRule(rule ShardingRule) {
	for i := range c.ShardingRules {
		if c.ShardingRules[i].Table == rule.Table {
			c.ShardingRules[i] = rule
			return
		}
	}
	c.ShardingRules = append(c.ShardingRules, rule)
}

// RemoveShardingRule removes a sharding rule
func (c *ClientAppConfig) RemoveShardingRule(table string) {
	rules := make([]ShardingRule, 0, len(c.ShardingRules))
	for _, rule := range c.ShardingRules {
		if rule.Table != table {
			rules = append(rules, rule)
		}
	}
	c.ShardingRules = rules
}

