package monitoring

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// PrometheusCollector collects and exposes metrics for Prometheus
type PrometheusCollector struct {
	logger             *zap.Logger
	registry           *prometheus.Registry
	collectors         map[string]*ShardCollector
	mu                 sync.RWMutex
	collectionInterval time.Duration

	// Metrics
	shardQueryTotal     *prometheus.CounterVec
	shardQueryDuration  *prometheus.HistogramVec
	shardConnections    *prometheus.GaugeVec
	shardReplicationLag *prometheus.GaugeVec
	shardCPUUsage       *prometheus.GaugeVec
	shardMemoryUsage    *prometheus.GaugeVec
	shardDiskUsage      *prometheus.GaugeVec
	shardErrorRate      *prometheus.GaugeVec
	clusterHealth       *prometheus.GaugeVec
	routerLatency       *prometheus.HistogramVec
	routerThroughput    *prometheus.CounterVec
	catalogUpdates      prometheus.Counter
	failoverEvents      *prometheus.CounterVec
	reshardingProgress  *prometheus.GaugeVec
	
	// PostgreSQL statistics metrics
	postgresDatabaseSize      *prometheus.GaugeVec
	postgresTableCount        *prometheus.GaugeVec
	postgresTableRows         *prometheus.GaugeVec
	postgresIndexCount        *prometheus.GaugeVec
	postgresConnections       *prometheus.GaugeVec
	postgresMaxConnections    *prometheus.GaugeVec
	postgresCacheHitRatio     *prometheus.GaugeVec
	postgresDeadTuples        *prometheus.GaugeVec
	postgresDatabaseUptime     *prometheus.GaugeVec
}

// ShardCollector collects metrics for a specific shard
type ShardCollector struct {
	shardID     string
	dsn         string
	logger      *zap.Logger
	db          *sql.DB
	lastMetrics *ShardDetailedMetrics
	mu          sync.RWMutex
}

// ShardDetailedMetrics contains detailed metrics for a shard
type ShardDetailedMetrics struct {
	// Connection metrics
	ActiveConnections  int64
	IdleConnections    int64
	MaxConnections     int64
	WaitingConnections int64

	// Query metrics
	TotalQueries     int64
	QueriesPerSecond float64
	AvgQueryTime     float64
	SlowQueries      int64

	// Replication metrics
	ReplicationLag   float64
	ReplicationState string
	WALWritePosition int64
	WALFlushPosition int64

	// Resource metrics
	CPUUsage       float64
	MemoryUsage    float64
	DiskUsage      float64
	DiskReadBytes  int64
	DiskWriteBytes int64

	// Table metrics
	TableCount    int64
	TotalRows     int64
	DeadTuples    int64
	IndexHitRatio float64

	// Transaction metrics
	TransactionsCommit   int64
	TransactionsRollback int64
	Deadlocks            int64

	CollectedAt time.Time
}

// NewPrometheusCollector creates a new Prometheus collector
func NewPrometheusCollector(logger *zap.Logger, collectionInterval time.Duration) *PrometheusCollector {
	registry := prometheus.NewRegistry()
	registry.MustRegister(prometheus.NewGoCollector())
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	pc := &PrometheusCollector{
		logger:             logger,
		registry:           registry,
		collectors:         make(map[string]*ShardCollector),
		collectionInterval: collectionInterval,
	}

	// Initialize metrics
	pc.initMetrics()

	return pc
}

// initMetrics initializes all Prometheus metrics
func (pc *PrometheusCollector) initMetrics() {
	pc.shardQueryTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sharding_shard_queries_total",
			Help: "Total number of queries executed on shards",
		},
		[]string{"shard_id", "database", "status"},
	)

	pc.shardQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "sharding_shard_query_duration_seconds",
			Help:    "Duration of queries in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
		},
		[]string{"shard_id", "database", "operation"},
	)

	pc.shardConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sharding_shard_connections",
			Help: "Number of connections per shard",
		},
		[]string{"shard_id", "database", "state"},
	)

	pc.shardReplicationLag = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sharding_shard_replication_lag_seconds",
			Help: "Replication lag in seconds",
		},
		[]string{"shard_id", "database", "replica"},
	)

	pc.shardCPUUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sharding_shard_cpu_usage_percent",
			Help: "CPU usage percentage",
		},
		[]string{"shard_id", "database"},
	)

	pc.shardMemoryUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sharding_shard_memory_usage_percent",
			Help: "Memory usage percentage",
		},
		[]string{"shard_id", "database"},
	)

	pc.shardDiskUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sharding_shard_disk_usage_percent",
			Help: "Disk usage percentage",
		},
		[]string{"shard_id", "database"},
	)

	pc.shardErrorRate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sharding_shard_error_rate",
			Help: "Error rate per shard",
		},
		[]string{"shard_id", "database"},
	)

	pc.clusterHealth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sharding_cluster_health",
			Help: "Cluster health status (1=healthy, 0=unhealthy)",
		},
		[]string{"component"},
	)

	pc.routerLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "sharding_router_latency_seconds",
			Help:    "Router request latency in seconds",
			Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0},
		},
		[]string{"method", "path", "status"},
	)

	pc.routerThroughput = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sharding_router_requests_total",
			Help: "Total router requests",
		},
		[]string{"method", "path", "status"},
	)

	pc.catalogUpdates = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sharding_catalog_updates_total",
			Help: "Total catalog updates",
		},
	)

	pc.failoverEvents = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sharding_failover_events_total",
			Help: "Total failover events",
		},
		[]string{"shard_id", "reason", "success"},
	)

	pc.reshardingProgress = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sharding_resharding_progress",
			Help: "Resharding progress (0.0 to 1.0)",
		},
		[]string{"job_id", "source_shard", "target_shard"},
	)

	// PostgreSQL statistics metrics
	pc.postgresDatabaseSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_database_size_bytes",
			Help: "PostgreSQL database size in bytes",
		},
		[]string{"cluster_id", "cluster_name", "namespace", "database_name", "database_host"},
	)

	pc.postgresTableCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_table_count",
			Help: "Number of tables in PostgreSQL database",
		},
		[]string{"cluster_id", "cluster_name", "namespace", "database_name", "database_host"},
	)

	pc.postgresTableRows = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_table_rows",
			Help: "Number of rows in a PostgreSQL table",
		},
		[]string{"cluster_id", "cluster_name", "namespace", "database_name", "database_host", "table_name"},
	)

	pc.postgresIndexCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_index_count",
			Help: "Number of indexes in PostgreSQL database",
		},
		[]string{"cluster_id", "cluster_name", "namespace", "database_name", "database_host"},
	)

	pc.postgresConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_connections",
			Help: "Current number of PostgreSQL connections",
		},
		[]string{"cluster_id", "cluster_name", "namespace", "database_name", "database_host", "state"},
	)

	pc.postgresMaxConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_max_connections",
			Help: "Maximum number of PostgreSQL connections",
		},
		[]string{"cluster_id", "cluster_name", "namespace", "database_name", "database_host"},
	)

	pc.postgresCacheHitRatio = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_cache_hit_ratio",
			Help: "PostgreSQL cache hit ratio (0.0 to 1.0)",
		},
		[]string{"cluster_id", "cluster_name", "namespace", "database_name", "database_host"},
	)

	pc.postgresDeadTuples = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_dead_tuples",
			Help: "Number of dead tuples in PostgreSQL database",
		},
		[]string{"cluster_id", "cluster_name", "namespace", "database_name", "database_host"},
	)

	pc.postgresDatabaseUptime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_database_uptime_seconds",
			Help: "PostgreSQL database uptime in seconds",
		},
		[]string{"cluster_id", "cluster_name", "namespace", "database_name", "database_host"},
	)

	// Register all metrics
	pc.registry.MustRegister(
		pc.shardQueryTotal,
		pc.shardQueryDuration,
		pc.shardConnections,
		pc.shardReplicationLag,
		pc.shardCPUUsage,
		pc.shardMemoryUsage,
		pc.shardDiskUsage,
		pc.shardErrorRate,
		pc.clusterHealth,
		pc.routerLatency,
		pc.routerThroughput,
		pc.catalogUpdates,
		pc.failoverEvents,
		pc.reshardingProgress,
		pc.postgresDatabaseSize,
		pc.postgresTableCount,
		pc.postgresTableRows,
		pc.postgresIndexCount,
		pc.postgresConnections,
		pc.postgresMaxConnections,
		pc.postgresCacheHitRatio,
		pc.postgresDeadTuples,
		pc.postgresDatabaseUptime,
	)
}

// RegisterShard registers a shard for metrics collection
func (pc *PrometheusCollector) RegisterShard(shardID, dsn string) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	collector := &ShardCollector{
		shardID: shardID,
		dsn:     dsn,
		logger:  pc.logger.With(zap.String("shard_id", shardID)),
	}

	// Try to establish database connection
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		pc.logger.Warn("failed to connect to shard for metrics", zap.String("shard_id", shardID), zap.Error(err))
	} else {
		collector.db = db
		db.SetMaxOpenConns(2)
		db.SetMaxIdleConns(1)
	}

	pc.collectors[shardID] = collector
	pc.logger.Info("registered shard for metrics collection", zap.String("shard_id", shardID))

	return nil
}

// UnregisterShard removes a shard from metrics collection
func (pc *PrometheusCollector) UnregisterShard(shardID string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if collector, ok := pc.collectors[shardID]; ok {
		if collector.db != nil {
			collector.db.Close()
		}
		delete(pc.collectors, shardID)
	}
}

// Start starts the metrics collection loop
func (pc *PrometheusCollector) Start(ctx context.Context) {
	ticker := time.NewTicker(pc.collectionInterval)
	defer ticker.Stop()

	pc.logger.Info("Prometheus collector started", zap.Duration("interval", pc.collectionInterval))

	// Initial collection
	pc.collectAll(ctx)

	for {
		select {
		case <-ctx.Done():
			pc.logger.Info("Prometheus collector stopped")
			return
		case <-ticker.C:
			pc.collectAll(ctx)
		}
	}
}

// collectAll collects metrics from all registered shards
func (pc *PrometheusCollector) collectAll(ctx context.Context) {
	pc.mu.RLock()
	collectors := make([]*ShardCollector, 0, len(pc.collectors))
	for _, c := range pc.collectors {
		collectors = append(collectors, c)
	}
	pc.mu.RUnlock()

	for _, collector := range collectors {
		metrics, err := collector.Collect(ctx)
		if err != nil {
			pc.logger.Warn("failed to collect metrics", zap.String("shard_id", collector.shardID), zap.Error(err))
			continue
		}

		pc.updateMetrics(collector.shardID, "default", metrics)
	}
}

// updateMetrics updates Prometheus metrics with collected data
func (pc *PrometheusCollector) updateMetrics(shardID, database string, metrics *ShardDetailedMetrics) {
	pc.shardConnections.WithLabelValues(shardID, database, "active").Set(float64(metrics.ActiveConnections))
	pc.shardConnections.WithLabelValues(shardID, database, "idle").Set(float64(metrics.IdleConnections))
	pc.shardConnections.WithLabelValues(shardID, database, "waiting").Set(float64(metrics.WaitingConnections))

	pc.shardReplicationLag.WithLabelValues(shardID, database, "primary").Set(metrics.ReplicationLag)

	pc.shardCPUUsage.WithLabelValues(shardID, database).Set(metrics.CPUUsage)
	pc.shardMemoryUsage.WithLabelValues(shardID, database).Set(metrics.MemoryUsage)
	pc.shardDiskUsage.WithLabelValues(shardID, database).Set(metrics.DiskUsage)
}

// Collect collects metrics from a shard
func (sc *ShardCollector) Collect(ctx context.Context) (*ShardDetailedMetrics, error) {
	if sc.db == nil {
		return &ShardDetailedMetrics{CollectedAt: time.Now()}, nil
	}

	metrics := &ShardDetailedMetrics{
		CollectedAt: time.Now(),
	}

	// Collect connection stats
	if err := sc.collectConnectionStats(ctx, metrics); err != nil {
		sc.logger.Warn("failed to collect connection stats", zap.Error(err))
	}

	// Collect replication stats
	if err := sc.collectReplicationStats(ctx, metrics); err != nil {
		sc.logger.Warn("failed to collect replication stats", zap.Error(err))
	}

	// Collect database stats
	if err := sc.collectDatabaseStats(ctx, metrics); err != nil {
		sc.logger.Warn("failed to collect database stats", zap.Error(err))
	}

	// Collect table stats
	if err := sc.collectTableStats(ctx, metrics); err != nil {
		sc.logger.Warn("failed to collect table stats", zap.Error(err))
	}

	sc.mu.Lock()
	sc.lastMetrics = metrics
	sc.mu.Unlock()

	return metrics, nil
}

// collectConnectionStats collects connection statistics
func (sc *ShardCollector) collectConnectionStats(ctx context.Context, metrics *ShardDetailedMetrics) error {
	query := `
		SELECT 
			count(*) FILTER (WHERE state = 'active') as active,
			count(*) FILTER (WHERE state = 'idle') as idle,
			count(*) FILTER (WHERE wait_event IS NOT NULL) as waiting,
			(SELECT setting::int FROM pg_settings WHERE name = 'max_connections') as max_conn
		FROM pg_stat_activity
		WHERE backend_type = 'client backend'
	`

	row := sc.db.QueryRowContext(ctx, query)
	err := row.Scan(&metrics.ActiveConnections, &metrics.IdleConnections, &metrics.WaitingConnections, &metrics.MaxConnections)
	if err != nil {
		return fmt.Errorf("failed to query connection stats: %w", err)
	}

	return nil
}

// collectReplicationStats collects replication statistics
func (sc *ShardCollector) collectReplicationStats(ctx context.Context, metrics *ShardDetailedMetrics) error {
	// Check if this is a replica
	query := `SELECT pg_is_in_recovery()`
	var isReplica bool
	if err := sc.db.QueryRowContext(ctx, query).Scan(&isReplica); err != nil {
		return fmt.Errorf("failed to check replica status: %w", err)
	}

	if isReplica {
		// Get replication lag for replica
		lagQuery := `
			SELECT EXTRACT(EPOCH FROM (now() - pg_last_xact_replay_timestamp()))
			WHERE pg_last_xact_replay_timestamp() IS NOT NULL
		`
		var lag sql.NullFloat64
		if err := sc.db.QueryRowContext(ctx, lagQuery).Scan(&lag); err == nil && lag.Valid {
			metrics.ReplicationLag = lag.Float64
		}
		metrics.ReplicationState = "replica"
	} else {
		// Get replication info for primary
		replicationQuery := `
			SELECT 
				pg_current_wal_lsn() - sent_lsn as send_lag
			FROM pg_stat_replication
			LIMIT 1
		`
		var sendLag sql.NullInt64
		if err := sc.db.QueryRowContext(ctx, replicationQuery).Scan(&sendLag); err == nil && sendLag.Valid {
			metrics.ReplicationLag = float64(sendLag.Int64) / 1024.0 / 1024.0 // Convert to MB
		}
		metrics.ReplicationState = "primary"
	}

	return nil
}

// collectDatabaseStats collects database-level statistics
func (sc *ShardCollector) collectDatabaseStats(ctx context.Context, metrics *ShardDetailedMetrics) error {
	query := `
		SELECT 
			xact_commit,
			xact_rollback,
			deadlocks,
			blks_hit::float / NULLIF(blks_hit + blks_read, 0) as cache_hit_ratio
		FROM pg_stat_database
		WHERE datname = current_database()
	`

	var cacheHitRatio sql.NullFloat64
	row := sc.db.QueryRowContext(ctx, query)
	err := row.Scan(&metrics.TransactionsCommit, &metrics.TransactionsRollback, &metrics.Deadlocks, &cacheHitRatio)
	if err != nil {
		return fmt.Errorf("failed to query database stats: %w", err)
	}

	if cacheHitRatio.Valid {
		metrics.IndexHitRatio = cacheHitRatio.Float64 * 100
	}

	return nil
}

// collectTableStats collects table-level statistics
func (sc *ShardCollector) collectTableStats(ctx context.Context, metrics *ShardDetailedMetrics) error {
	query := `
		SELECT 
			count(*) as table_count,
			COALESCE(sum(n_live_tup), 0) as total_rows,
			COALESCE(sum(n_dead_tup), 0) as dead_tuples
		FROM pg_stat_user_tables
	`

	row := sc.db.QueryRowContext(ctx, query)
	err := row.Scan(&metrics.TableCount, &metrics.TotalRows, &metrics.DeadTuples)
	if err != nil {
		return fmt.Errorf("failed to query table stats: %w", err)
	}

	return nil
}

// Handler returns the HTTP handler for Prometheus metrics
func (pc *PrometheusCollector) Handler() http.Handler {
	return promhttp.HandlerFor(pc.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// RecordQuery records a query execution
func (pc *PrometheusCollector) RecordQuery(shardID, database, operation, status string, duration time.Duration) {
	pc.shardQueryTotal.WithLabelValues(shardID, database, status).Inc()
	pc.shardQueryDuration.WithLabelValues(shardID, database, operation).Observe(duration.Seconds())
}

// RecordRouterRequest records a router request
func (pc *PrometheusCollector) RecordRouterRequest(method, path, status string, duration time.Duration) {
	pc.routerThroughput.WithLabelValues(method, path, status).Inc()
	pc.routerLatency.WithLabelValues(method, path, status).Observe(duration.Seconds())
}

// RecordFailover records a failover event
func (pc *PrometheusCollector) RecordFailover(shardID, reason string, success bool) {
	successStr := "false"
	if success {
		successStr = "true"
	}
	pc.failoverEvents.WithLabelValues(shardID, reason, successStr).Inc()
}

// RecordCatalogUpdate records a catalog update
func (pc *PrometheusCollector) RecordCatalogUpdate() {
	pc.catalogUpdates.Inc()
}

// SetClusterHealth sets the cluster health status
func (pc *PrometheusCollector) SetClusterHealth(component string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	pc.clusterHealth.WithLabelValues(component).Set(value)
}

// SetReshardingProgress sets the resharding progress
func (pc *PrometheusCollector) SetReshardingProgress(jobID, sourceShard, targetShard string, progress float64) {
	pc.reshardingProgress.WithLabelValues(jobID, sourceShard, targetShard).Set(progress)
}

// RecordPostgresStats records PostgreSQL statistics from scanned databases
func (pc *PrometheusCollector) RecordPostgresStats(clusterID, clusterName, namespace, databaseName, databaseHost string, stats *ShardDetailedMetrics) {
	labels := []string{clusterID, clusterName, namespace, databaseName, databaseHost}
	
	// Database size (if available)
	if stats.TableCount > 0 {
		pc.postgresTableCount.WithLabelValues(labels...).Set(float64(stats.TableCount))
	}
	
	// Total rows
	if stats.TotalRows > 0 {
		pc.postgresTableRows.WithLabelValues(append(labels, "total")...).Set(float64(stats.TotalRows))
	}
	
	// Dead tuples
	if stats.DeadTuples > 0 {
		pc.postgresDeadTuples.WithLabelValues(labels...).Set(float64(stats.DeadTuples))
	}
	
	// Connections
	pc.postgresConnections.WithLabelValues(append(labels, "active")...).Set(float64(stats.ActiveConnections))
	pc.postgresConnections.WithLabelValues(append(labels, "idle")...).Set(float64(stats.IdleConnections))
	pc.postgresConnections.WithLabelValues(append(labels, "waiting")...).Set(float64(stats.WaitingConnections))
	
	// Max connections
	if stats.MaxConnections > 0 {
		pc.postgresMaxConnections.WithLabelValues(labels...).Set(float64(stats.MaxConnections))
	}
	
	// Cache hit ratio
	if stats.IndexHitRatio > 0 {
		pc.postgresCacheHitRatio.WithLabelValues(labels...).Set(stats.IndexHitRatio / 100.0) // Convert from percentage to ratio
	}
}

// RecordPostgresTableStats records table-level PostgreSQL statistics
func (pc *PrometheusCollector) RecordPostgresTableStats(clusterID, clusterName, namespace, databaseName, databaseHost, tableName string, rowCount int64) {
	labels := []string{clusterID, clusterName, namespace, databaseName, databaseHost, tableName}
	pc.postgresTableRows.WithLabelValues(labels...).Set(float64(rowCount))
}

// GetShardMetrics returns the latest metrics for a shard
func (pc *PrometheusCollector) GetShardMetrics(shardID string) (*ShardDetailedMetrics, bool) {
	pc.mu.RLock()
	collector, ok := pc.collectors[shardID]
	pc.mu.RUnlock()

	if !ok {
		return nil, false
	}

	collector.mu.RLock()
	defer collector.mu.RUnlock()
	return collector.lastMetrics, collector.lastMetrics != nil
}

