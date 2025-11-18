package router

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/sharding-system/pkg/catalog"
	"github.com/sharding-system/pkg/models"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// Router routes queries to appropriate shards
type Router struct {
	catalog      catalog.Catalog
	logger       *zap.Logger
	connections  map[string]*sql.DB
	mu           sync.RWMutex
	maxConns     int
	connTTL      time.Duration
	replicaPolicy string
}

// NewRouter creates a new router instance
func NewRouter(catalog catalog.Catalog, logger *zap.Logger, maxConns int, connTTL time.Duration, replicaPolicy string) *Router {
	return &Router{
		catalog:      catalog,
		logger:       logger,
		connections:  make(map[string]*sql.DB),
		maxConns:     maxConns,
		connTTL:      connTTL,
		replicaPolicy: replicaPolicy,
	}
}

// ExecuteQuery executes a query on the appropriate shard
func (r *Router) ExecuteQuery(ctx context.Context, req *models.QueryRequest) (*models.QueryResponse, error) {
	start := time.Now()

	// Get shard for the key
	shard, err := r.catalog.GetShard(req.ShardKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get shard: %w", err)
	}

	// Select endpoint based on consistency requirement
	endpoint := shard.PrimaryEndpoint
	if req.Consistency == "eventual" && r.replicaPolicy == "replica_ok" && len(shard.Replicas) > 0 {
		// Use replica for read-only queries with eventual consistency
		endpoint = shard.Replicas[0]
	}

	// Get or create connection pool
	db, err := r.getConnection(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	// Execute query
	rows, err := db.QueryContext(ctx, req.Query, req.Params...)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	// Convert rows to response
	resultRows := make([]interface{}, 0)
	columns, _ := rows.Columns()
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				rowMap[col] = string(b)
			} else {
				rowMap[col] = val
			}
		}
		resultRows = append(resultRows, rowMap)
	}

	latency := time.Since(start)

	r.logger.Info("query executed",
		zap.String("shard_id", shard.ID),
		zap.String("endpoint", endpoint),
		zap.Duration("latency", latency),
		zap.Int("row_count", len(resultRows)),
	)

	return &models.QueryResponse{
		ShardID:   shard.ID,
		Rows:      resultRows,
		RowCount:  len(resultRows),
		LatencyMs: float64(latency.Nanoseconds()) / 1e6,
	}, nil
}

// GetShardForKey returns the shard ID for a given key
func (r *Router) GetShardForKey(key string) (string, error) {
	shard, err := r.catalog.GetShard(key)
	if err != nil {
		return "", err
	}
	return shard.ID, nil
}

// getConnection gets or creates a database connection pool
func (r *Router) getConnection(endpoint string) (*sql.DB, error) {
	r.mu.RLock()
	db, exists := r.connections[endpoint]
	r.mu.RUnlock()

	if exists {
		// Check if connection is still alive
		if err := db.Ping(); err == nil {
			return db, nil
		}
		// Connection is dead, remove it
		r.mu.Lock()
		delete(r.connections, endpoint)
		r.mu.Unlock()
	}

	// Create new connection
	r.mu.Lock()
	defer r.mu.Unlock()

	// Double check after acquiring write lock
	if db, exists := r.connections[endpoint]; exists {
		return db, nil
	}

	db, err := sql.Open("postgres", endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(r.maxConns)
	db.SetMaxIdleConns(r.maxConns / 2)
	db.SetConnMaxLifetime(r.connTTL)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	r.connections[endpoint] = db
	return db, nil
}

// Close closes all connections
func (r *Router) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for endpoint, db := range r.connections {
		if err := db.Close(); err != nil {
			r.logger.Error("failed to close connection", zap.String("endpoint", endpoint), zap.Error(err))
		}
	}

	return nil
}

