package models

import (
	"time"
)

// Shard represents a database shard
type Shard struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	HashRangeStart  uint64    `json:"hash_range_start"`
	HashRangeEnd    uint64    `json:"hash_range_end"`
	PrimaryEndpoint string    `json:"primary_endpoint"`
	Replicas        []string  `json:"replicas"`
	Status          string    `json:"status"` // "active", "migrating", "readonly", "inactive"
	Version         int64     `json:"version"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	VNodes          []VNode  `json:"vnodes,omitempty"`
}

// VNode represents a virtual node in consistent hashing
type VNode struct {
	ID       uint64 `json:"id"`
	ShardID  string `json:"shard_id"`
	Hash     uint64 `json:"hash"`
}

// ShardCatalog represents the complete shard mapping
type ShardCatalog struct {
	Version   int64   `json:"version"`
	Shards    []Shard `json:"shards"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ReshardJob represents a resharding operation
type ReshardJob struct {
	ID            string    `json:"id"`
	Type          string    `json:"type"` // "split" or "merge"
	SourceShards  []string  `json:"source_shards"`
	TargetShards  []string  `json:"target_shards"`
	Status        string    `json:"status"` // "pending", "precopy", "deltasync", "cutover", "completed", "failed"
	Progress      float64   `json:"progress"` // 0.0 to 1.0
	StartedAt     time.Time `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	ErrorMessage  string    `json:"error_message,omitempty"`
	KeysMigrated  int64     `json:"keys_migrated"`
	TotalKeys     int64     `json:"total_keys"`
}

// ShardHealth represents health status of a shard
type ShardHealth struct {
	ShardID        string        `json:"shard_id"`
	Status         string        `json:"status"` // "healthy", "degraded", "unhealthy"
	ReplicationLag time.Duration `json:"replication_lag"`
	LastCheck      time.Time     `json:"last_check"`
	PrimaryUp      bool          `json:"primary_up"`
	ReplicasUp     []string      `json:"replicas_up"`
	ReplicasDown   []string      `json:"replicas_down"`
}

// QueryRequest represents a query request
type QueryRequest struct {
	ShardKey    string                 `json:"shard_key"`
	Query       string                 `json:"query"`
	Params      []interface{}          `json:"params"`
	Consistency string                 `json:"consistency"` // "strong" or "eventual"
	Options     map[string]interface{} `json:"options,omitempty"`
}

// QueryResponse represents a query response
type QueryResponse struct {
	ShardID   string        `json:"shard_id"`
	Rows      []interface{} `json:"rows"`
	RowCount  int           `json:"row_count"`
	LatencyMs float64       `json:"latency_ms"`
}

// CreateShardRequest represents a request to create a shard
type CreateShardRequest struct {
	Name            string   `json:"name"`
	PrimaryEndpoint string   `json:"primary_endpoint"`
	Replicas        []string `json:"replicas"`
	VNodeCount      int      `json:"vnode_count"`
}

// SplitRequest represents a request to split a shard
type SplitRequest struct {
	SourceShardID string   `json:"source_shard_id"`
	TargetShards  []CreateShardRequest `json:"target_shards"`
	SplitPoint    uint64   `json:"split_point,omitempty"` // Optional explicit split point
}

// MergeRequest represents a request to merge shards
type MergeRequest struct {
	SourceShardIDs []string `json:"source_shard_ids"`
	TargetShard    CreateShardRequest `json:"target_shard"`
}

