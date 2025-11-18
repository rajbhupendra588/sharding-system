package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Query metrics
	QueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "shard_query_duration_seconds",
			Help: "Duration of queries in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		[]string{"shard_id", "operation"},
	)

	QueryTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "shard_queries_total",
			Help: "Total number of queries",
		},
		[]string{"shard_id", "status"},
	)

	// Shard metrics
	ShardConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shard_connections_active",
			Help: "Number of active connections per shard",
		},
		[]string{"shard_id"},
	)

	ReplicationLag = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shard_replication_lag_seconds",
			Help: "Replication lag in seconds",
		},
		[]string{"shard_id", "replica"},
	)

	// Resharding metrics
	ReshardProgress = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "reshard_progress",
			Help: "Resharding progress (0.0 to 1.0)",
		},
		[]string{"job_id", "type"},
	)

	ReshardKeysMigrated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "reshard_keys_migrated_total",
			Help: "Total keys migrated during resharding",
		},
		[]string{"job_id"},
	)

	// Catalog metrics
	CatalogVersion = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "catalog_version",
			Help: "Current catalog version",
		},
	)

	CatalogUpdates = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "catalog_updates_total",
			Help: "Total catalog updates",
		},
	)
)

