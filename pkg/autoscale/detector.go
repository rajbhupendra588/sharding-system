package autoscale

import (
	"sync"

	"github.com/sharding-system/pkg/monitoring"
	"go.uber.org/zap"
)

// Thresholds define when a shard is considered "hot" or "cold"
type Thresholds struct {
	MaxQueryRate    float64 // Maximum queries per second before splitting
	MaxCPUUsage     float64 // Maximum CPU usage percentage (0-100)
	MaxMemoryUsage  float64 // Maximum memory usage percentage (0-100)
	MaxStorageUsage float64 // Maximum storage usage percentage (0-100)
	MaxConnections  int     // Maximum connection count
	MaxLatencyMs    float64 // Maximum average latency in milliseconds
	MinQueryRate    float64 // Minimum queries per second for merge consideration
	MinCPUUsage     float64 // Minimum CPU usage for merge consideration
	MinStorageUsage float64 // Minimum storage usage for merge consideration
}

// DefaultThresholds returns default threshold values
func DefaultThresholds() Thresholds {
	return Thresholds{
		MaxQueryRate:    10000.0, // 10k queries/sec
		MaxCPUUsage:     80.0,    // 80% CPU
		MaxMemoryUsage:  80.0,    // 80% memory
		MaxStorageUsage: 80.0,    // 80% storage
		MaxConnections:  1000,    // 1000 connections
		MaxLatencyMs:    100.0,   // 100ms latency
		MinQueryRate:    100.0,   // Below 100 queries/sec for merge
		MinCPUUsage:     20.0,    // Below 20% CPU for merge
		MinStorageUsage: 30.0,    // Below 30% storage for merge
	}
}

// HotShardDetector detects shards that need to be split
type HotShardDetector struct {
	monitor    *monitoring.LoadMonitor
	thresholds Thresholds
	logger     *zap.Logger
	mu         sync.RWMutex
	history    map[string][]*monitoring.ShardMetrics // Track metrics history
}

// NewHotShardDetector creates a new hot shard detector
func NewHotShardDetector(monitor *monitoring.LoadMonitor, thresholds Thresholds, logger *zap.Logger) *HotShardDetector {
	return &HotShardDetector{
		monitor:    monitor,
		thresholds: thresholds,
		logger:     logger,
		history:    make(map[string][]*monitoring.ShardMetrics),
	}
}

// IsHotShard determines if a shard is "hot" and needs splitting
func (d *HotShardDetector) IsHotShard(shardID string) bool {
	metrics, ok := d.monitor.GetMetrics(shardID)
	if !ok {
		return false
	}

	// Add to history
	d.mu.Lock()
	if d.history[shardID] == nil {
		d.history[shardID] = make([]*monitoring.ShardMetrics, 0, 10)
	}
	d.history[shardID] = append(d.history[shardID], metrics)
	// Keep only last 10 measurements
	if len(d.history[shardID]) > 10 {
		d.history[shardID] = d.history[shardID][len(d.history[shardID])-10:]
	}
	d.mu.Unlock()

	// Check if any threshold is exceeded
	isHot := metrics.QueryRate > d.thresholds.MaxQueryRate ||
		metrics.CPUUsage > d.thresholds.MaxCPUUsage ||
		metrics.MemoryUsage > d.thresholds.MaxMemoryUsage ||
		metrics.StorageUsage > d.thresholds.MaxStorageUsage ||
		metrics.ConnectionCount > d.thresholds.MaxConnections ||
		metrics.AvgLatencyMs > d.thresholds.MaxLatencyMs

	if isHot {
		d.logger.Warn("hot shard detected",
			zap.String("shard_id", shardID),
			zap.Float64("query_rate", metrics.QueryRate),
			zap.Float64("cpu_usage", metrics.CPUUsage),
			zap.Float64("memory_usage", metrics.MemoryUsage),
			zap.Float64("storage_usage", metrics.StorageUsage),
			zap.Int("connections", metrics.ConnectionCount),
			zap.Float64("latency_ms", metrics.AvgLatencyMs))
	}

	return isHot
}

// IsColdShard determines if a shard is "cold" and can be merged
func (d *HotShardDetector) IsColdShard(shardID string) bool {
	metrics, ok := d.monitor.GetMetrics(shardID)
	if !ok {
		return false
	}

	// Check if all metrics are below minimum thresholds
	isCold := metrics.QueryRate < d.thresholds.MinQueryRate &&
		metrics.CPUUsage < d.thresholds.MinCPUUsage &&
		metrics.StorageUsage < d.thresholds.MinStorageUsage

	if isCold {
		d.logger.Info("cold shard detected",
			zap.String("shard_id", shardID),
			zap.Float64("query_rate", metrics.QueryRate),
			zap.Float64("cpu_usage", metrics.CPUUsage),
			zap.Float64("storage_usage", metrics.StorageUsage))
	}

	return isCold
}

// GetHotShards returns all shards that are currently hot
func (d *HotShardDetector) GetHotShards() []string {
	allMetrics := d.monitor.GetAllMetrics()
	hotShards := make([]string, 0)

	for shardID := range allMetrics {
		if d.IsHotShard(shardID) {
			hotShards = append(hotShards, shardID)
		}
	}

	return hotShards
}

// GetColdShards returns all shards that are currently cold
func (d *HotShardDetector) GetColdShards() []string {
	allMetrics := d.monitor.GetAllMetrics()
	coldShards := make([]string, 0)

	for shardID := range allMetrics {
		if d.IsColdShard(shardID) {
			coldShards = append(coldShards, shardID)
		}
	}

	return coldShards
}

// GetMetricsHistory returns metrics history for a shard
func (d *HotShardDetector) GetMetricsHistory(shardID string) []*monitoring.ShardMetrics {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.history[shardID]
}

// GetThresholds returns current thresholds
func (d *HotShardDetector) GetThresholds() Thresholds {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.thresholds
}

// UpdateThresholds updates the detection thresholds
func (d *HotShardDetector) UpdateThresholds(thresholds Thresholds) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.thresholds = thresholds
	d.logger.Info("thresholds updated",
		zap.Float64("max_query_rate", thresholds.MaxQueryRate),
		zap.Float64("max_cpu", thresholds.MaxCPUUsage),
		zap.Float64("max_storage", thresholds.MaxStorageUsage))
}

