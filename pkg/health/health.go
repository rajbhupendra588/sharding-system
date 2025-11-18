package health

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

// Controller monitors shard health and handles failover
type Controller struct {
	catalog      catalog.Catalog
	logger       *zap.Logger
	healthStatus map[string]*models.ShardHealth
	mu           sync.RWMutex
	checkInterval time.Duration
	replicationLagThreshold time.Duration
}

// NewController creates a new health controller
func NewController(catalog catalog.Catalog, logger *zap.Logger, checkInterval, lagThreshold time.Duration) *Controller {
	return &Controller{
		catalog:                catalog,
		logger:                 logger,
		healthStatus:           make(map[string]*models.ShardHealth),
		checkInterval:          checkInterval,
		replicationLagThreshold: lagThreshold,
	}
}

// Start starts the health monitoring loop
func (c *Controller) Start(ctx context.Context) {
	ticker := time.NewTicker(c.checkInterval)
	defer ticker.Stop()

	// Initial check
	c.checkAllShards(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.checkAllShards(ctx)
		}
	}
}

// checkAllShards checks health of all shards
func (c *Controller) checkAllShards(ctx context.Context) {
	shards, err := c.catalog.ListShards()
	if err != nil {
		c.logger.Error("failed to list shards for health check", zap.Error(err))
		return
	}

	for _, shard := range shards {
		c.checkShard(ctx, &shard)
	}
}

// checkShard checks the health of a single shard
func (c *Controller) checkShard(ctx context.Context, shard *models.Shard) {
	health := &models.ShardHealth{
		ShardID:   shard.ID,
		Status:    "healthy",
		LastCheck: time.Now(),
		PrimaryUp: false,
		ReplicasUp: make([]string, 0),
		ReplicasDown: make([]string, 0),
	}

	// Check primary
	if c.checkEndpoint(ctx, shard.PrimaryEndpoint) {
		health.PrimaryUp = true
	} else {
		health.Status = "unhealthy"
		c.logger.Warn("primary shard is down",
			zap.String("shard_id", shard.ID),
			zap.String("endpoint", shard.PrimaryEndpoint),
		)
	}

	// Check replicas
	for _, replicaEndpoint := range shard.Replicas {
		if c.checkEndpoint(ctx, replicaEndpoint) {
			health.ReplicasUp = append(health.ReplicasUp, replicaEndpoint)
		} else {
			health.ReplicasDown = append(health.ReplicasDown, replicaEndpoint)
			if health.Status == "healthy" {
				health.Status = "degraded"
			}
		}
	}

	// Check replication lag (simplified - in production use actual lag metrics)
	health.ReplicationLag = c.getReplicationLag(ctx, shard)
	if health.ReplicationLag > c.replicationLagThreshold {
		if health.Status == "healthy" {
			health.Status = "degraded"
		}
		c.logger.Warn("replication lag exceeds threshold",
			zap.String("shard_id", shard.ID),
			zap.Duration("lag", health.ReplicationLag),
		)
	}

	c.mu.Lock()
	c.healthStatus[shard.ID] = health
	c.mu.Unlock()
}

// checkEndpoint checks if an endpoint is reachable
func (c *Controller) checkEndpoint(ctx context.Context, endpoint string) bool {
	db, err := sql.Open("postgres", endpoint)
	if err != nil {
		return false
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return false
	}

	return true
}

// getReplicationLag gets replication lag for a shard
func (c *Controller) getReplicationLag(ctx context.Context, shard *models.Shard) time.Duration {
	// In production, this would query the database for actual replication lag
	// For PostgreSQL, you'd query pg_stat_replication
	// For now, return 0 (no lag) as a placeholder
	
	if len(shard.Replicas) == 0 {
		return 0
	}

	// Simplified: check if replicas are responding
	// In production, query: SELECT EXTRACT(EPOCH FROM (now() - pg_last_xact_replay_timestamp())) as lag
	return 0
}

// GetHealth returns health status for a shard
func (c *Controller) GetHealth(shardID string) (*models.ShardHealth, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	health, exists := c.healthStatus[shardID]
	if !exists {
		return nil, fmt.Errorf("health status not found for shard %s", shardID)
	}

	return health, nil
}

// GetAllHealth returns health status for all shards
func (c *Controller) GetAllHealth() map[string]*models.ShardHealth {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]*models.ShardHealth)
	for k, v := range c.healthStatus {
		result[k] = v
	}

	return result
}

// ShouldFailover determines if a failover should occur
func (c *Controller) ShouldFailover(shardID string) (bool, string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	health, exists := c.healthStatus[shardID]
	if !exists || health.Status == "healthy" {
		return false, ""
	}

	// If primary is down and we have healthy replicas, failover
	if !health.PrimaryUp && len(health.ReplicasUp) > 0 {
		return true, health.ReplicasUp[0]
	}

	return false, ""
}

