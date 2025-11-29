package failover

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sharding-system/pkg/health"
	"github.com/sharding-system/pkg/manager"
	"go.uber.org/zap"
)

// FailoverController manages automatic failover operations
type FailoverController struct {
	manager      *manager.Manager
	healthCtrl   *health.Controller
	logger       *zap.Logger
	checkInterval time.Duration
	enabled      bool
	mu           sync.RWMutex
	running      bool
	stopCh       chan struct{}
	failoverHistory []*FailoverEvent
}

// FailoverEvent represents a failover event
type FailoverEvent struct {
	ID          string    `json:"id"`
	ShardID     string    `json:"shard_id"`
	OldPrimary  string    `json:"old_primary"`
	NewPrimary  string    `json:"new_primary"`
	Reason      string    `json:"reason"`
	Status      string    `json:"status"` // "success", "failed", "rolled_back"
	StartedAt   time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Error       string    `json:"error,omitempty"`
}

// NewFailoverController creates a new failover controller
func NewFailoverController(mgr *manager.Manager, healthCtrl *health.Controller, logger *zap.Logger, checkInterval time.Duration) *FailoverController {
	return &FailoverController{
		manager:        mgr,
		healthCtrl:     healthCtrl,
		logger:         logger,
		checkInterval:  checkInterval,
		enabled:        true,
		failoverHistory: make([]*FailoverEvent, 0),
		stopCh:         make(chan struct{}),
	}
}

// Start starts the failover monitoring loop
func (c *FailoverController) Start() {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return
	}
	c.running = true
	c.mu.Unlock()

	go c.monitorLoop()

	c.logger.Info("failover controller started",
		zap.Duration("check_interval", c.checkInterval))
}

// Stop stops the failover monitoring loop
func (c *FailoverController) Stop() {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return
	}
	c.running = false
	c.mu.Unlock()

	close(c.stopCh)

	c.logger.Info("failover controller stopped")
}

// Enable enables automatic failover
func (c *FailoverController) Enable() {
	c.mu.Lock()
	c.enabled = true
	c.mu.Unlock()
	c.logger.Info("automatic failover enabled")
}

// Disable disables automatic failover
func (c *FailoverController) Disable() {
	c.mu.Lock()
	c.enabled = false
	c.mu.Unlock()
	c.logger.Info("automatic failover disabled")
}

// IsEnabled returns whether automatic failover is enabled
func (c *FailoverController) IsEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.enabled
}

// monitorLoop continuously monitors shards for failures
func (c *FailoverController) monitorLoop() {
	ticker := time.NewTicker(c.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			if c.IsEnabled() {
				c.checkAndFailover(context.Background())
			}
		}
	}
}

// checkAndFailover checks all shards and performs failover if needed
func (c *FailoverController) checkAndFailover(ctx context.Context) {
	// Get all shards
	shards, err := c.manager.ListShards()
	if err != nil {
		c.logger.Error("failed to list shards for failover check", zap.Error(err))
		return
	}

	for _, shard := range shards {
		// Get shard health status
		healthStatus, err := c.healthCtrl.GetHealth(shard.ID)
		if err != nil {
			c.logger.Warn("failed to get shard health",
				zap.String("shard_id", shard.ID),
				zap.Error(err))
			continue
		}

		// Check if primary is down and we have healthy replicas
		if !healthStatus.PrimaryUp && len(healthStatus.ReplicasUp) > 0 {
			c.logger.Warn("primary shard is down, initiating failover",
				zap.String("shard_id", shard.ID),
				zap.Strings("available_replicas", healthStatus.ReplicasUp))

			// Select best replica (first available for now)
			bestReplica := healthStatus.ReplicasUp[0]
			
			// Perform failover
			if err := c.performFailover(ctx, shard.ID, shard.PrimaryEndpoint, bestReplica); err != nil {
				c.logger.Error("failover failed",
					zap.String("shard_id", shard.ID),
					zap.Error(err))
			}
		}
	}
}

// performFailover performs the actual failover operation
func (c *FailoverController) performFailover(ctx context.Context, shardID string, oldPrimary string, newPrimary string) error {
	event := &FailoverEvent{
		ID:         fmt.Sprintf("failover-%d", time.Now().Unix()),
		ShardID:    shardID,
		OldPrimary: oldPrimary,
		NewPrimary: newPrimary,
		Reason:     "primary_unavailable",
		Status:     "in_progress",
		StartedAt:  time.Now(),
	}

	c.mu.Lock()
	c.failoverHistory = append(c.failoverHistory, event)
	c.mu.Unlock()

	c.logger.Info("performing failover",
		zap.String("event_id", event.ID),
		zap.String("shard_id", shardID),
		zap.String("old_primary", oldPrimary),
		zap.String("new_primary", newPrimary))

	// Promote replica to primary
	if err := c.manager.PromoteReplica(shardID, newPrimary); err != nil {
		event.Status = "failed"
		event.Error = err.Error()
		now := time.Now()
		event.CompletedAt = &now

		c.logger.Error("failover failed",
			zap.String("event_id", event.ID),
			zap.Error(err))

		return fmt.Errorf("failed to promote replica: %w", err)
	}

	// Verify failover success
	if err := c.verifyFailover(ctx, shardID, newPrimary); err != nil {
		// Rollback if verification fails
		c.logger.Warn("failover verification failed, attempting rollback",
			zap.String("event_id", event.ID),
			zap.Error(err))

		if rollbackErr := c.rollbackFailover(ctx, shardID, oldPrimary, newPrimary); rollbackErr != nil {
			c.logger.Error("rollback failed",
				zap.String("event_id", event.ID),
				zap.Error(rollbackErr))
		}

		event.Status = "rolled_back"
		event.Error = err.Error()
		now := time.Now()
		event.CompletedAt = &now

		return fmt.Errorf("failover verification failed: %w", err)
	}

	// Success
	event.Status = "success"
	now := time.Now()
	event.CompletedAt = &now

	c.logger.Info("failover completed successfully",
		zap.String("event_id", event.ID),
		zap.String("shard_id", shardID),
		zap.String("new_primary", newPrimary))

	return nil
}

// verifyFailover verifies that failover was successful
func (c *FailoverController) verifyFailover(ctx context.Context, shardID string, newPrimary string) error {
	// Wait a bit for the system to stabilize
	time.Sleep(2 * time.Second)

	// Check shard health again
	healthStatus, err := c.healthCtrl.GetHealth(shardID)
	if err != nil {
		return fmt.Errorf("failed to get shard health: %w", err)
	}

	// Verify new primary is up
	if !healthStatus.PrimaryUp {
		return fmt.Errorf("new primary is not up")
	}

	// Verify new primary matches expected
	shard, err := c.manager.GetShard(shardID)
	if err != nil {
		return fmt.Errorf("failed to get shard: %w", err)
	}

	if shard.PrimaryEndpoint != newPrimary {
		return fmt.Errorf("primary endpoint mismatch: expected %s, got %s", newPrimary, shard.PrimaryEndpoint)
	}

	return nil
}

// rollbackFailover attempts to rollback a failed failover
func (c *FailoverController) rollbackFailover(ctx context.Context, shardID string, oldPrimary string, newPrimary string) error {
	c.logger.Warn("rolling back failover",
		zap.String("shard_id", shardID),
		zap.String("old_primary", oldPrimary),
		zap.String("new_primary", newPrimary))

	// Try to promote old primary back
	// Note: This is a simplified rollback. In production, you'd need more sophisticated logic
	if err := c.manager.PromoteReplica(shardID, oldPrimary); err != nil {
		return fmt.Errorf("failed to rollback: %w", err)
	}

	return nil
}

// GetFailoverHistory returns failover history
func (c *FailoverController) GetFailoverHistory() []*FailoverEvent {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy
	history := make([]*FailoverEvent, len(c.failoverHistory))
	copy(history, c.failoverHistory)
	return history
}

// GetFailoverHistoryForShard returns failover history for a specific shard
func (c *FailoverController) GetFailoverHistoryForShard(shardID string) []*FailoverEvent {
	c.mu.RLock()
	defer c.mu.RUnlock()

	history := make([]*FailoverEvent, 0)
	for _, event := range c.failoverHistory {
		if event.ShardID == shardID {
			history = append(history, event)
		}
	}

	return history
}

