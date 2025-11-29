package models

import (
	"time"
)

// Cluster represents a Kubernetes cluster configuration
type Cluster struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"` // "cloud" (aws, gcp, azure) or "onprem"
	Provider    string            `json:"provider,omitempty"` // "aws", "gcp", "azure", "onprem"
	Kubeconfig  string            `json:"kubeconfig,omitempty"` // Path to kubeconfig or base64 encoded
	Context     string            `json:"context,omitempty"`     // K8s context name
	Endpoint    string            `json:"endpoint,omitempty"`    // K8s API endpoint
	Credentials map[string]string `json:"credentials,omitempty"` // Provider-specific credentials
	Status      string            `json:"status"`                 // "active", "inactive", "error"
	LastScan    *time.Time        `json:"last_scan,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Metadata    map[string]string  `json:"metadata,omitempty"`
}

// CreateClusterRequest represents a request to register a new cluster
type CreateClusterRequest struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Provider    string            `json:"provider,omitempty"`
	Kubeconfig  string            `json:"kubeconfig,omitempty"`
	Context     string            `json:"context,omitempty"`
	Endpoint    string            `json:"endpoint,omitempty"`
	Credentials map[string]string `json:"credentials,omitempty"`
	Metadata    map[string]string  `json:"metadata,omitempty"`
}

// ScannedDatabase represents a database discovered and scanned from a cluster
type ScannedDatabase struct {
	ID              string            `json:"id"`
	ClusterID       string            `json:"cluster_id"`
	ClusterName     string            `json:"cluster_name"`
	Namespace       string            `json:"namespace"`
	AppName         string            `json:"app_name"`
	AppType         string            `json:"app_type"` // "deployment", "statefulset"
	DatabaseName    string            `json:"database_name"`
	DatabaseType    string            `json:"database_type"` // "postgresql", "mysql", etc.
	Host            string            `json:"host"`
	Port            int               `json:"port"`
	Database        string            `json:"database"`
	Username        string            `json:"username,omitempty"`
	Status          string            `json:"status"` // "discovered", "scanned", "error"
	ScanError       string            `json:"scan_error,omitempty"`
	ScanResults     *DatabaseScanResults `json:"scan_results,omitempty"`
	DiscoveredAt    time.Time         `json:"discovered_at"`
	LastScannedAt   *time.Time        `json:"last_scanned_at,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
	Annotations     map[string]string `json:"annotations,omitempty"`
}

// DatabaseScanResults contains detailed scan results for a database
type DatabaseScanResults struct {
	Version         string                 `json:"version,omitempty"`
	Size            int64                  `json:"size,omitempty"` // bytes
	TableCount      int                    `json:"table_count,omitempty"`
	TableNames      []string               `json:"table_names,omitempty"`
	IndexCount      int                    `json:"index_count,omitempty"`
	ConnectionCount int                    `json:"connection_count,omitempty"`
	MaxConnections  int                    `json:"max_connections,omitempty"`
	IsReplica       bool                   `json:"is_replica,omitempty"`
	ReplicationLag  float64                `json:"replication_lag,omitempty"`
	Uptime          int64                  `json:"uptime,omitempty"` // seconds
	TableStats      []TableStat            `json:"table_stats,omitempty"`
	IndexStats      []IndexStat            `json:"index_stats,omitempty"`
	HealthStatus    string                 `json:"health_status"` // "healthy", "degraded", "unhealthy"
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// TableStat contains statistics for a database table
type TableStat struct {
	Name            string `json:"name"`
	RowCount        int64  `json:"row_count"`
	Size            int64  `json:"size"` // bytes
	IndexSize       int64  `json:"index_size"` // bytes
	TotalSize       int64  `json:"total_size"` // bytes
	LastVacuum      *time.Time `json:"last_vacuum,omitempty"`
	LastAnalyze     *time.Time `json:"last_analyze,omitempty"`
}

// IndexStat contains statistics for a database index
type IndexStat struct {
	Name         string `json:"name"`
	TableName    string `json:"table_name"`
	Size         int64  `json:"size"` // bytes
	Scans        int64  `json:"scans"`
	TuplesRead   int64  `json:"tuples_read"`
	TuplesFetched int64 `json:"tuples_fetched"`
}

// ScanRequest represents a request to scan databases in clusters
type ScanRequest struct {
	ClusterIDs []string `json:"cluster_ids,omitempty"` // Empty means scan all clusters
	DeepScan   bool     `json:"deep_scan,omitempty"`   // Perform deep scan vs quick discovery
}

// ScanResult represents the result of a scan operation
type ScanResult struct {
	ID          string            `json:"id"`
	ClusterID   string            `json:"cluster_id"`
	Status      string            `json:"status"` // "running", "completed", "failed"
	DatabasesFound int            `json:"databases_found"`
	DatabasesScanned int          `json:"databases_scanned"`
	DatabasesFailed int           `json:"databases_failed"`
	StartedAt   time.Time         `json:"started_at"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
	Error       string            `json:"error,omitempty"`
	Results     []ScannedDatabase `json:"results,omitempty"`
}

