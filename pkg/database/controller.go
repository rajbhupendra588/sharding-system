// Package database provides high-level database management
package database

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sharding-system/pkg/operator"
	"github.com/sharding-system/pkg/schema"
	"go.uber.org/zap"
)

// Database represents a fully managed sharded database
type Database struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	DisplayName      string                 `json:"display_name"`
	Description      string                 `json:"description,omitempty"`
	Status           string                 `json:"status"` // "creating", "ready", "scaling", "migrating", "failed", "deleting"
	Template         string                 `json:"template,omitempty"`
	ShardCount       int                    `json:"shard_count"`
	ShardKey         string                 `json:"shard_key"`
	ShardKeyType     string                 `json:"shard_key_type"` // "uuid", "integer", "string"
	Strategy         string                 `json:"strategy"`       // "hash", "range"
	ConnectionString string                 `json:"connection_string"`
	ProxyEndpoint    string                 `json:"proxy_endpoint"`
	Shards           []ShardStatus          `json:"shards"`
	SchemaVersion    int                    `json:"schema_version"`
	Tables           []TableInfo            `json:"tables,omitempty"`
	Config           DatabaseConfig         `json:"config"`
	Metrics          DatabaseMetrics        `json:"metrics"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	ReadyAt          *time.Time             `json:"ready_at,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// ShardStatus represents status of a single shard
type ShardStatus struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Host         string    `json:"host"`
	Port         int       `json:"port"`
	Status       string    `json:"status"`
	Size         int64     `json:"size_bytes"`
	RowCount     int64     `json:"row_count"`
	Connections  int       `json:"connections"`
	ReplicaLag   int64     `json:"replica_lag_ms"`
	LastHealthAt time.Time `json:"last_health_at"`
}

// TableInfo represents a table in the database
type TableInfo struct {
	Name        string   `json:"name"`
	ShardKey    string   `json:"shard_key,omitempty"`
	Columns     []Column `json:"columns"`
	Indexes     []string `json:"indexes,omitempty"`
	RowCount    int64    `json:"row_count"`
	SizeBytes   int64    `json:"size_bytes"`
	IsBroadcast bool     `json:"is_broadcast"` // Replicated to all shards
}

// Column represents a table column
type Column struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
	Default  string `json:"default,omitempty"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Resources   ResourceConfig `json:"resources"`
	Storage     StorageConfig  `json:"storage"`
	Replication ReplicaConfig  `json:"replication"`
	Backup      BackupConfig   `json:"backup"`
}

// ResourceConfig defines compute resources
type ResourceConfig struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

// StorageConfig defines storage settings
type StorageConfig struct {
	SizePerShard string `json:"size_per_shard"`
	StorageClass string `json:"storage_class"`
	TotalSize    string `json:"total_size"`
}

// ReplicaConfig defines replication settings
type ReplicaConfig struct {
	Enabled          bool `json:"enabled"`
	ReplicasPerShard int  `json:"replicas_per_shard"`
}

// BackupConfig defines backup settings
type BackupConfig struct {
	Enabled   bool   `json:"enabled"`
	Schedule  string `json:"schedule"` // Cron expression
	Retention int    `json:"retention_days"`
}

// DatabaseMetrics holds real-time metrics
type DatabaseMetrics struct {
	QueriesPerSecond  float64   `json:"queries_per_second"`
	AvgLatencyMs      float64   `json:"avg_latency_ms"`
	ConnectionsActive int       `json:"connections_active"`
	ConnectionsIdle   int       `json:"connections_idle"`
	StorageUsedBytes  int64     `json:"storage_used_bytes"`
	LastUpdated       time.Time `json:"last_updated"`
}

// CreateDatabaseRequest represents a request to create a new database
type CreateDatabaseRequest struct {
	Name           string                 `json:"name"`
	DisplayName    string                 `json:"display_name,omitempty"`
	Description    string                 `json:"description,omitempty"`
	Template       string                 `json:"template,omitempty"` // "starter", "production", "enterprise"
	ShardCount     int                    `json:"shard_count"`
	ShardKey       string                 `json:"shard_key"`
	ShardKeyType   string                 `json:"shard_key_type"`
	Strategy       string                 `json:"strategy"`
	Schema         string                 `json:"schema,omitempty"`          // Initial SQL schema
	SchemaTemplate string                 `json:"schema_template,omitempty"` // Pre-defined template
	Resources      *ResourceConfig        `json:"resources,omitempty"`
	Storage        *StorageConfig         `json:"storage,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// Controller manages sharded databases at a high level
type Controller struct {
	logger        *zap.Logger
	operator      *operator.Operator
	schemaManager *schema.Manager
	databases     map[string]*Database
	mu            sync.RWMutex
	namespace     string

	// Event callbacks
	onDatabaseReady  func(*Database)
	onDatabaseFailed func(*Database, error)
}

// NewController creates a new database controller
func NewController(logger *zap.Logger, op *operator.Operator, sm *schema.Manager, namespace string) *Controller {
	return &Controller{
		logger:        logger,
		operator:      op,
		schemaManager: sm,
		databases:     make(map[string]*Database),
		namespace:     namespace,
	}
}

// SetOnDatabaseReady sets callback for when database becomes ready
func (c *Controller) SetOnDatabaseReady(callback func(*Database)) {
	c.onDatabaseReady = callback
}

// CreateDatabase creates a new sharded database with one API call
func (c *Controller) CreateDatabase(ctx context.Context, req CreateDatabaseRequest) (*Database, error) {
	// Validate request
	if req.Name == "" {
		return nil, fmt.Errorf("database name is required")
	}
	if req.ShardCount < 1 {
		req.ShardCount = 2 // Default
	}
	if req.ShardKey == "" {
		req.ShardKey = "id" // Default
	}
	if req.Strategy == "" {
		req.Strategy = "hash" // Default
	}
	if req.ShardKeyType == "" {
		req.ShardKeyType = "uuid" // Default
	}

	c.mu.Lock()
	if _, exists := c.databases[req.Name]; exists {
		c.mu.Unlock()
		return nil, fmt.Errorf("database %s already exists", req.Name)
	}

	// Apply template defaults
	template := operator.PredefinedTemplates["starter"]
	if req.Template != "" {
		if t, ok := operator.PredefinedTemplates[req.Template]; ok {
			template = t
		}
	}

	// Override with custom settings
	resources := template.Resources
	if req.Resources != nil {
		resources = operator.ShardResources{
			CPU:    req.Resources.CPU,
			Memory: req.Resources.Memory,
		}
	}

	storage := template.Storage
	if req.Storage != nil {
		storage = operator.StorageConfig{
			Size:         req.Storage.SizePerShard,
			StorageClass: req.Storage.StorageClass,
		}
	}

	// Determine initial schema
	initialSchema := req.Schema
	if initialSchema == "" && req.SchemaTemplate != "" {
		if schemaTempl, ok := schema.PredefinedSchemas[req.SchemaTemplate]; ok {
			initialSchema = schemaTempl.SQL
		}
	}

	displayName := req.DisplayName
	if displayName == "" {
		displayName = req.Name
	}

	// Create database record
	db := &Database{
		ID:            uuid.New().String(),
		Name:          req.Name,
		DisplayName:   displayName,
		Description:   req.Description,
		Status:        "creating",
		Template:      req.Template,
		ShardCount:    req.ShardCount,
		ShardKey:      req.ShardKey,
		ShardKeyType:  req.ShardKeyType,
		Strategy:      req.Strategy,
		SchemaVersion: 0,
		Config: DatabaseConfig{
			Resources: ResourceConfig{
				CPU:    resources.CPU,
				Memory: resources.Memory,
			},
			Storage: StorageConfig{
				SizePerShard: storage.Size,
				StorageClass: storage.StorageClass,
			},
			Replication: ReplicaConfig{
				Enabled:          template.Replication.Enabled,
				ReplicasPerShard: template.Replication.Replicas,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  req.Metadata,
	}

	c.databases[req.Name] = db
	c.mu.Unlock()

	// Provision infrastructure asynchronously
	go c.provisionDatabase(ctx, db, resources, storage, initialSchema)

	c.logger.Info("started creating database",
		zap.String("name", req.Name),
		zap.Int("shardCount", req.ShardCount),
		zap.String("template", req.Template))

	return db, nil
}

// provisionDatabase handles the async provisioning
func (c *Controller) provisionDatabase(ctx context.Context, db *Database, resources operator.ShardResources, storage operator.StorageConfig, initialSchema string) {
	// Create the sharded database via operator
	spec := operator.ShardedDatabaseSpec{
		Name:       db.Name,
		ShardCount: db.ShardCount,
		Strategy:   db.Strategy,
		ShardKey:   db.ShardKey,
		Resources:  resources,
		Storage:    storage,
		Schema:     initialSchema,
	}

	// Set up callback to track shard creation
	c.operator.SetOnShardReady(func(dbName string, shard operator.ShardInfo) {
		c.mu.Lock()
		defer c.mu.Unlock()

		if database, ok := c.databases[dbName]; ok {
			database.Shards = append(database.Shards, ShardStatus{
				ID:     shard.ID,
				Name:   shard.Name,
				Host:   shard.Host,
				Port:   shard.Port,
				Status: shard.Status,
			})
			database.UpdatedAt = time.Now()
		}
	})

	shardedDB, err := c.operator.CreateShardedDatabase(ctx, spec)
	if err != nil {
		c.mu.Lock()
		db.Status = "failed"
		db.UpdatedAt = time.Now()
		c.mu.Unlock()

		c.logger.Error("failed to create sharded database",
			zap.String("name", db.Name),
			zap.Error(err))

		if c.onDatabaseFailed != nil {
			c.onDatabaseFailed(db, err)
		}
		return
	}

	// Wait for operator to complete
	for {
		time.Sleep(5 * time.Second)

		opDB, exists := c.operator.GetDatabase(db.Name)
		if !exists {
			continue
		}

		if opDB.Status.Phase == "Ready" {
			c.mu.Lock()
			db.Status = "ready"
			db.ConnectionString = opDB.Status.ConnectionString
			db.ProxyEndpoint = opDB.Status.ProxyEndpoint
			now := time.Now()
			db.ReadyAt = &now
			db.UpdatedAt = now

			// Update shard info
			db.Shards = make([]ShardStatus, 0, len(opDB.Status.Shards))
			for _, s := range opDB.Status.Shards {
				db.Shards = append(db.Shards, ShardStatus{
					ID:     s.ID,
					Name:   s.Name,
					Host:   s.Host,
					Port:   s.Port,
					Status: s.Status,
				})
			}
			c.mu.Unlock()

			c.logger.Info("database ready",
				zap.String("name", db.Name),
				zap.String("connectionString", db.ConnectionString))

			if c.onDatabaseReady != nil {
				c.onDatabaseReady(db)
			}
			return
		}

		if opDB.Status.Phase == "Failed" {
			c.mu.Lock()
			db.Status = "failed"
			db.UpdatedAt = time.Now()
			c.mu.Unlock()

			if c.onDatabaseFailed != nil {
				c.onDatabaseFailed(db, fmt.Errorf("database provisioning failed: %s", shardedDB.Status.Message))
			}
			return
		}
	}
}

// GetDatabase retrieves a database by name
func (c *Controller) GetDatabase(name string) (*Database, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	db, exists := c.databases[name]
	return db, exists
}

// ListDatabases returns all databases
func (c *Controller) ListDatabases() []*Database {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]*Database, 0, len(c.databases))
	for _, db := range c.databases {
		result = append(result, db)
	}
	return result
}

// DeleteDatabase deletes a database and all its resources
func (c *Controller) DeleteDatabase(ctx context.Context, name string) error {
	c.mu.Lock()
	db, exists := c.databases[name]
	if !exists {
		c.mu.Unlock()
		return fmt.Errorf("database %s not found", name)
	}

	db.Status = "deleting"
	db.UpdatedAt = time.Now()
	c.mu.Unlock()

	// Delete via operator
	if err := c.operator.DeleteDatabase(ctx, name); err != nil {
		return fmt.Errorf("failed to delete database: %w", err)
	}

	c.mu.Lock()
	delete(c.databases, name)
	c.mu.Unlock()

	c.logger.Info("deleted database", zap.String("name", name))
	return nil
}

// ScaleDatabase changes the number of shards
func (c *Controller) ScaleDatabase(ctx context.Context, name string, newShardCount int) error {
	c.mu.Lock()
	db, exists := c.databases[name]
	if !exists {
		c.mu.Unlock()
		return fmt.Errorf("database %s not found", name)
	}

	if db.Status != "ready" {
		c.mu.Unlock()
		return fmt.Errorf("database must be ready to scale (current: %s)", db.Status)
	}

	db.Status = "scaling"
	db.UpdatedAt = time.Now()
	c.mu.Unlock()

	// Scale via operator
	if err := c.operator.ScaleShards(ctx, name, newShardCount); err != nil {
		c.mu.Lock()
		db.Status = "ready" // Revert status
		c.mu.Unlock()
		return fmt.Errorf("failed to scale: %w", err)
	}

	c.mu.Lock()
	db.ShardCount = newShardCount
	db.Status = "ready"
	db.UpdatedAt = time.Now()
	c.mu.Unlock()

	c.logger.Info("scaled database",
		zap.String("name", name),
		zap.Int("newShardCount", newShardCount))

	return nil
}

// ApplySchema applies a schema migration to all shards
func (c *Controller) ApplySchema(ctx context.Context, name string, sql string) error {
	c.mu.RLock()
	db, exists := c.databases[name]
	if !exists {
		c.mu.RUnlock()
		return fmt.Errorf("database %s not found", name)
	}

	if db.Status != "ready" {
		c.mu.RUnlock()
		return fmt.Errorf("database must be ready for schema changes")
	}
	c.mu.RUnlock()

	// Get shard connections
	shards := make([]schema.ShardConnection, 0, len(db.Shards))
	for _, s := range db.Shards {
		shards = append(shards, schema.ShardConnection{
			ID:       s.ID,
			Name:     s.Name,
			Host:     s.Host,
			Port:     s.Port,
			Database: db.Name,
			Username: "sharding_admin",
			Password: "", // Get from secrets
		})
	}

	// Register and apply migration
	c.mu.Lock()
	newVersion := db.SchemaVersion + 1
	c.mu.Unlock()

	if err := c.schemaManager.RegisterMigration(newVersion, fmt.Sprintf("migration_%d", newVersion), "", sql); err != nil {
		return err
	}

	statuses, err := c.schemaManager.ApplyMigrations(ctx, shards)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	// Check all applied
	for _, status := range statuses {
		if status.Status == "failed" {
			return fmt.Errorf("migration failed on shard %s: %s", status.ShardName, status.Error)
		}
	}

	c.mu.Lock()
	db.SchemaVersion = newVersion
	db.UpdatedAt = time.Now()
	c.mu.Unlock()

	c.logger.Info("applied schema migration",
		zap.String("database", name),
		zap.Int("version", newVersion))

	return nil
}

// GetConnectionInfo returns connection details for a database
func (c *Controller) GetConnectionInfo(name string) (*ConnectionInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	db, exists := c.databases[name]
	if !exists {
		return nil, fmt.Errorf("database %s not found", name)
	}

	return &ConnectionInfo{
		DatabaseName:     db.Name,
		ProxyHost:        fmt.Sprintf("sharding-proxy.%s.svc.cluster.local", c.namespace),
		ProxyPort:        6432,
		Username:         "sharding_admin",
		ConnectionString: db.ConnectionString,
		DirectShards:     db.Shards,
	}, nil
}

// ConnectionInfo contains all connection details
type ConnectionInfo struct {
	DatabaseName     string        `json:"database_name"`
	ProxyHost        string        `json:"proxy_host"`
	ProxyPort        int           `json:"proxy_port"`
	Username         string        `json:"username"`
	ConnectionString string        `json:"connection_string"`
	DirectShards     []ShardStatus `json:"direct_shards,omitempty"`
}

// ExportConfig exports database configuration as JSON
func (c *Controller) ExportConfig(name string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	db, exists := c.databases[name]
	if !exists {
		return "", fmt.Errorf("database %s not found", name)
	}

	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}



