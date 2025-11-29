package monitoring

import (
	"context"
	"sync"
	"time"

	"github.com/sharding-system/pkg/catalog"
	"go.uber.org/zap"
)

// ShardMetrics represents real-time metrics for a shard
type ShardMetrics struct {
	ShardID         string    `json:"shard_id"`
	QueryRate       float64   `json:"query_rate"`        // Queries per second
	ConnectionCount int       `json:"connection_count"`  // Active connections
	CPUUsage        float64   `json:"cpu_usage"`         // CPU usage percentage (0-100)
	MemoryUsage     float64   `json:"memory_usage"`      // Memory usage percentage (0-100)
	StorageUsage    float64   `json:"storage_usage"`     // Storage usage percentage (0-100)
	AvgLatencyMs    float64   `json:"avg_latency_ms"`    // Average query latency in milliseconds
	ErrorRate       float64   `json:"error_rate"`        // Error rate percentage (0-100)
	Timestamp       time.Time `json:"timestamp"`          // When metrics were collected
}

// LoadMonitor monitors shard load metrics
type LoadMonitor struct {
	catalog    catalog.Catalog
	logger     *zap.Logger
	metrics    map[string]*ShardMetrics
	mu         sync.RWMutex
	interval   time.Duration
	stopCh     chan struct{}
	collectors map[string]MetricsCollector
}

// MetricsCollector collects metrics for a specific shard
type MetricsCollector interface {
	CollectMetrics(ctx context.Context, shardID string) (*ShardMetrics, error)
}

// NewLoadMonitor creates a new load monitor
func NewLoadMonitor(catalog catalog.Catalog, logger *zap.Logger, interval time.Duration) *LoadMonitor {
	return &LoadMonitor{
		catalog:    catalog,
		logger:     logger,
		metrics:    make(map[string]*ShardMetrics),
		interval:   interval,
		stopCh:     make(chan struct{}),
		collectors: make(map[string]MetricsCollector),
	}
}

// RegisterCollector registers a metrics collector for a shard
func (m *LoadMonitor) RegisterCollector(shardID string, collector MetricsCollector) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.collectors[shardID] = collector
}

// Start begins monitoring shards
func (m *LoadMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	m.logger.Info("load monitor started", zap.Duration("interval", m.interval))

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("load monitor stopped")
			return
		case <-m.stopCh:
			m.logger.Info("load monitor stopped")
			return
		case <-ticker.C:
			m.collectMetrics(ctx)
		}
	}
}

// Stop stops the load monitor
func (m *LoadMonitor) Stop() {
	close(m.stopCh)
}

// collectMetrics collects metrics for all shards
func (m *LoadMonitor) collectMetrics(ctx context.Context) {
	// List all shards (empty string = all client apps)
	shards, err := m.catalog.ListShards("")
	if err != nil {
		m.logger.Error("failed to list shards for monitoring", zap.Error(err))
		return
	}

	for _, shard := range shards {
		if shard.Status != "active" {
			continue // Skip non-active shards
		}

		metrics, err := m.collectShardMetrics(ctx, shard.ID)
		if err != nil {
			m.logger.Warn("failed to collect metrics for shard",
				zap.String("shard_id", shard.ID),
				zap.Error(err))
			continue
		}

		m.mu.Lock()
		m.metrics[shard.ID] = metrics
		m.mu.Unlock()
	}
}

// collectShardMetrics collects metrics for a specific shard
func (m *LoadMonitor) collectShardMetrics(ctx context.Context, shardID string) (*ShardMetrics, error) {
	m.mu.RLock()
	collector, hasCollector := m.collectors[shardID]
	m.mu.RUnlock()

	if hasCollector {
		return collector.CollectMetrics(ctx, shardID)
	}

	// Default collector - returns zero metrics when no collector is configured
	// In production, configure a MetricsCollector to query Prometheus, PostgreSQL stats, etc.
	return m.defaultCollector(ctx, shardID)
}

// defaultCollector provides default metrics collection
// Returns zero metrics when no collector is configured
// In production, configure a MetricsCollector to query actual metrics from Prometheus, PostgreSQL, etc.
func (m *LoadMonitor) defaultCollector(ctx context.Context, shardID string) (*ShardMetrics, error) {
	// Return zero metrics - actual metrics should be collected via MetricsCollector
	return &ShardMetrics{
		ShardID:         shardID,
		QueryRate:       0.0,
		ConnectionCount: 0,
		CPUUsage:        0.0,
		MemoryUsage:     0.0,
		StorageUsage:    0.0,
		AvgLatencyMs:    0.0,
		ErrorRate:       0.0,
		Timestamp:       time.Now(),
	}, nil
}

// GetMetrics returns current metrics for a shard
func (m *LoadMonitor) GetMetrics(shardID string) (*ShardMetrics, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	metrics, ok := m.metrics[shardID]
	return metrics, ok
}

// GetAllMetrics returns metrics for all shards
func (m *LoadMonitor) GetAllMetrics() map[string]*ShardMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*ShardMetrics, len(m.metrics))
	for k, v := range m.metrics {
		result[k] = v
	}
	return result
}

