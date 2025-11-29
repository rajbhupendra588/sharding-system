// Package operator provides Kubernetes operator for automatic PostgreSQL provisioning
package operator

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ShardedDatabase represents a sharded database resource
type ShardedDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ShardedDatabaseSpec   `json:"spec"`
	Status ShardedDatabaseStatus `json:"status,omitempty"`
}

// ShardedDatabaseSpec defines the desired state
type ShardedDatabaseSpec struct {
	// Name of the database
	Name string `json:"name"`

	// Number of shards to create
	ShardCount int `json:"shardCount"`

	// Sharding strategy: "hash", "range", "directory"
	Strategy string `json:"strategy"`

	// Shard key column name
	ShardKey string `json:"shardKey"`

	// Resource configuration per shard
	Resources ShardResources `json:"resources"`

	// Storage configuration
	Storage StorageConfig `json:"storage"`

	// Replication configuration
	Replication ReplicationConfig `json:"replication,omitempty"`

	// Schema to apply on creation
	Schema string `json:"schema,omitempty"`
}

// ShardResources defines resource limits per shard
type ShardResources struct {
	CPU    string `json:"cpu"`    // e.g., "500m"
	Memory string `json:"memory"` // e.g., "1Gi"
}

// StorageConfig defines storage for each shard
type StorageConfig struct {
	Size         string `json:"size"`         // e.g., "10Gi"
	StorageClass string `json:"storageClass"` // e.g., "standard"
}

// ReplicationConfig defines replication settings
type ReplicationConfig struct {
	Enabled  bool `json:"enabled"`
	Replicas int  `json:"replicas"` // Number of read replicas per shard
}

// ShardedDatabaseStatus defines the observed state
type ShardedDatabaseStatus struct {
	Phase           string       `json:"phase"` // "Pending", "Creating", "Ready", "Failed"
	Shards          []ShardInfo  `json:"shards,omitempty"`
	ConnectionString string      `json:"connectionString,omitempty"`
	ProxyEndpoint   string       `json:"proxyEndpoint,omitempty"`
	CreatedAt       time.Time    `json:"createdAt,omitempty"`
	ReadyAt         *time.Time   `json:"readyAt,omitempty"`
	Message         string       `json:"message,omitempty"`
	SchemaVersion   int          `json:"schemaVersion"`
}

// ShardInfo contains information about a single shard
type ShardInfo struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	Database  string    `json:"database"`
	Status    string    `json:"status"` // "creating", "ready", "failed"
	PodName   string    `json:"podName"`
	PVCName   string    `json:"pvcName"`
	CreatedAt time.Time `json:"createdAt"`
}

// ShardedDatabaseList is a list of ShardedDatabase resources
type ShardedDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ShardedDatabase `json:"items"`
}

// DatabaseTemplate defines pre-configured database templates
type DatabaseTemplate struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	ShardCount  int               `json:"shardCount"`
	Resources   ShardResources    `json:"resources"`
	Storage     StorageConfig     `json:"storage"`
	Replication ReplicationConfig `json:"replication"`
}

// PredefinedTemplates provides ready-to-use configurations
var PredefinedTemplates = map[string]DatabaseTemplate{
	"starter": {
		Name:        "Starter",
		Description: "Perfect for development and small applications",
		ShardCount:  2,
		Resources:   ShardResources{CPU: "250m", Memory: "512Mi"},
		Storage:     StorageConfig{Size: "5Gi", StorageClass: "standard"},
		Replication: ReplicationConfig{Enabled: false, Replicas: 0},
	},
	"production": {
		Name:        "Production",
		Description: "Balanced configuration for production workloads",
		ShardCount:  4,
		Resources:   ShardResources{CPU: "1000m", Memory: "2Gi"},
		Storage:     StorageConfig{Size: "50Gi", StorageClass: "fast"},
		Replication: ReplicationConfig{Enabled: true, Replicas: 1},
	},
	"enterprise": {
		Name:        "Enterprise",
		Description: "High-performance configuration for large scale",
		ShardCount:  8,
		Resources:   ShardResources{CPU: "2000m", Memory: "4Gi"},
		Storage:     StorageConfig{Size: "100Gi", StorageClass: "fast"},
		Replication: ReplicationConfig{Enabled: true, Replicas: 2},
	},
}





