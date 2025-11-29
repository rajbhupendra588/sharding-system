package autoscale

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sharding-system/pkg/catalog"
	"github.com/sharding-system/pkg/manager"
	"github.com/sharding-system/pkg/models"
	"go.uber.org/zap"
)

// AutoSplitter automatically splits hot shards
type AutoSplitter struct {
	detector     *HotShardDetector
	manager      *manager.Manager
	catalog      catalog.Catalog
	logger       *zap.Logger
	enabled      bool
	mu           sync.RWMutex
	splitHistory map[string]time.Time // Track when shards were last split
	cooldown     time.Duration        // Minimum time between splits for same shard
}

// NewAutoSplitter creates a new auto-splitter
func NewAutoSplitter(
	detector *HotShardDetector,
	manager *manager.Manager,
	catalog catalog.Catalog,
	logger *zap.Logger,
) *AutoSplitter {
	return &AutoSplitter{
		detector:     detector,
		manager:      manager,
		catalog:      catalog,
		logger:       logger,
		enabled:      true,
		splitHistory: make(map[string]time.Time),
		cooldown:     30 * time.Minute, // 30 minute cooldown between splits
	}
}

// Start begins automatic splitting monitoring
func (s *AutoSplitter) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()

	s.logger.Info("auto-splitter started")

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("auto-splitter stopped")
			return
		case <-ticker.C:
			if s.IsEnabled() {
				s.checkAndSplit(ctx)
			}
		}
	}
}

// checkAndSplit checks for hot shards and splits them if needed
func (s *AutoSplitter) checkAndSplit(ctx context.Context) {
	hotShards := s.detector.GetHotShards()

	for _, shardID := range hotShards {
		// Check cooldown period
		if s.isInCooldown(shardID) {
			s.logger.Debug("shard in cooldown period, skipping split",
				zap.String("shard_id", shardID))
			continue
		}

		// Check if shard is already being split
		if s.isShardSplitting(shardID) {
			s.logger.Debug("shard already being split, skipping",
				zap.String("shard_id", shardID))
			continue
		}

		// Perform automatic split
		if err := s.splitShard(ctx, shardID); err != nil {
			s.logger.Error("failed to auto-split shard",
				zap.String("shard_id", shardID),
				zap.Error(err))
		}
	}
}

// splitShard automatically splits a hot shard
func (s *AutoSplitter) splitShard(ctx context.Context, shardID string) error {
	s.logger.Info("auto-splitting hot shard", zap.String("shard_id", shardID))

	// Get source shard
	sourceShard, err := s.catalog.GetShardByID(shardID)
	if err != nil {
		return fmt.Errorf("failed to get source shard: %w", err)
	}

	// Determine split strategy (split into 2 shards by default)
	targetShards := s.createTargetShards(sourceShard, 2)

	// Create split request
	splitReq := &models.SplitRequest{
		SourceShardID: shardID,
		TargetShards:  targetShards,
	}

	// Execute split
	job, err := s.manager.SplitShard(ctx, splitReq)
	if err != nil {
		return fmt.Errorf("failed to execute split: %w", err)
	}

	// Record split time
	s.mu.Lock()
	s.splitHistory[shardID] = time.Now()
	s.mu.Unlock()

	s.logger.Info("auto-split initiated",
		zap.String("shard_id", shardID),
		zap.String("job_id", job.ID),
		zap.Int("target_shards", len(targetShards)))

	return nil
}

// createTargetShards creates target shard requests for splitting
func (s *AutoSplitter) createTargetShards(sourceShard *models.Shard, count int) []models.CreateShardRequest {
	targetShards := make([]models.CreateShardRequest, 0, count)

	for i := 0; i < count; i++ {
		// Use default vnode count (will be set during shard creation)
		vnodeCount := 256 / count // Distribute vnodes across split shards

		targetShard := models.CreateShardRequest{
			Name:            fmt.Sprintf("%s-split-%d", sourceShard.Name, i+1),
			ClientAppID:     sourceShard.ClientAppID,
			PrimaryEndpoint: fmt.Sprintf("%s-split-%d", sourceShard.PrimaryEndpoint, i+1),
			Replicas:        sourceShard.Replicas, // Copy replica configuration
			VNodeCount:      vnodeCount,           // Distribute vnodes
		}
		targetShards = append(targetShards, targetShard)
	}

	return targetShards
}

// isInCooldown checks if a shard is in cooldown period
func (s *AutoSplitter) isInCooldown(shardID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lastSplit, ok := s.splitHistory[shardID]
	if !ok {
		return false
	}

	return time.Since(lastSplit) < s.cooldown
}

// isShardSplitting checks if a shard is currently being split
func (s *AutoSplitter) isShardSplitting(shardID string) bool {
	// Check if there's an active reshard job for this shard
	// This would require access to manager's job list
	// For now, return false (simplified)
	return false
}

// IsEnabled returns whether auto-splitting is enabled
func (s *AutoSplitter) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// Enable enables automatic splitting
func (s *AutoSplitter) Enable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = true
	s.logger.Info("auto-splitting enabled")
}

// Disable disables automatic splitting
func (s *AutoSplitter) Disable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = false
	s.logger.Info("auto-splitting disabled")
}

// SetCooldown sets the cooldown period between splits
func (s *AutoSplitter) SetCooldown(duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cooldown = duration
	s.logger.Info("cooldown period updated", zap.Duration("cooldown", duration))
}

