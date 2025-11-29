package monitoring

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// PostgresStatsCollector collects detailed PostgreSQL statistics
type PostgresStatsCollector struct {
	logger    *zap.Logger
	databases map[string]*DBConnection
	mu        sync.RWMutex
	interval  time.Duration
	stopCh    chan struct{}
}

// DBConnection represents a database connection for stats collection
type DBConnection struct {
	DSN         string
	DB          *sql.DB
	DatabaseID  string
	LastStats   *PostgresStats
	LastError   error
	LastCollect time.Time
}

// PostgresStats contains comprehensive PostgreSQL statistics
type PostgresStats struct {
	DatabaseID   string    `json:"database_id"`
	DatabaseName string    `json:"database_name"`
	Size         int64     `json:"size_bytes"`
	CollectedAt  time.Time `json:"collected_at"`
	Connections  ConnectionStats  `json:"connections"`
	Queries      QueryStats       `json:"queries"`
	Replication  ReplicationStats `json:"replication"`
	Tables       TableStats       `json:"tables"`
	Indexes      IndexStats       `json:"indexes"`
	Locks        LockStats        `json:"locks"`
	BGWriter     BGWriterStats    `json:"bg_writer"`
}

// ConnectionStats represents connection statistics
type ConnectionStats struct {
	Total          int                `json:"total"`
	Active         int                `json:"active"`
	Idle           int                `json:"idle"`
	IdleInTx       int                `json:"idle_in_transaction"`
	Waiting        int                `json:"waiting"`
	MaxConnections int                `json:"max_connections"`
	PercentUsed    float64            `json:"percent_used"`
	ByState        map[string]int     `json:"by_state"`
	ByApplication  map[string]int     `json:"by_application"`
}

// QueryStats represents query performance statistics
type QueryStats struct {
	TotalQueries     int64      `json:"total_queries"`
	QueriesPerSecond float64    `json:"queries_per_second"`
	AvgQueryTime     float64    `json:"avg_query_time_ms"`
	MaxQueryTime     float64    `json:"max_query_time_ms"`
	SlowQueries      int64      `json:"slow_queries"`
	CacheHitRatio    float64    `json:"cache_hit_ratio"`
	TopQueries       []TopQuery `json:"top_queries,omitempty"`
}

// TopQuery represents a frequently executed query
type TopQuery struct {
	Query           string  `json:"query"`
	Calls           int64   `json:"calls"`
	TotalTime       float64 `json:"total_time_ms"`
	MeanTime        float64 `json:"mean_time_ms"`
	Rows            int64   `json:"rows"`
	SharedBlocksHit int64   `json:"shared_blocks_hit"`
}

// ReplicationStats represents replication statistics
type ReplicationStats struct {
	IsReplica      bool          `json:"is_replica"`
	ReplicationLag float64       `json:"replication_lag_seconds"`
	ReplicaCount   int           `json:"replica_count"`
	WALPosition    string        `json:"wal_position"`
	ReplayLag      float64       `json:"replay_lag_bytes"`
	Replicas       []ReplicaInfo `json:"replicas,omitempty"`
}

// ReplicaInfo represents info about a replica
type ReplicaInfo struct {
	ClientAddr string `json:"client_addr"`
	State      string `json:"state"`
	SentLag    int64  `json:"sent_lag_bytes"`
	WriteLag   int64  `json:"write_lag_bytes"`
	FlushLag   int64  `json:"flush_lag_bytes"`
	ReplayLag  int64  `json:"replay_lag_bytes"`
	SyncState  string `json:"sync_state"`
}

// TableStats represents table statistics
type TableStats struct {
	TotalTables   int         `json:"total_tables"`
	TotalRows     int64       `json:"total_rows"`
	LiveTuples    int64       `json:"live_tuples"`
	DeadTuples    int64       `json:"dead_tuples"`
	SeqScans      int64       `json:"sequential_scans"`
	IndexScans    int64       `json:"index_scans"`
	SeqScanRatio  float64     `json:"seq_scan_ratio"`
	LargestTables []TableInfo `json:"largest_tables,omitempty"`
}

// TableInfo represents info about a specific table
type TableInfo struct {
	Schema    string `json:"schema"`
	TableName string `json:"table_name"`
	Rows      int64  `json:"rows"`
	Size      int64  `json:"size_bytes"`
	SeqScans  int64  `json:"seq_scans"`
	IdxScans  int64  `json:"idx_scans"`
}

// IndexStats represents index statistics
type IndexStats struct {
	TotalIndexes     int     `json:"total_indexes"`
	IndexSize        int64   `json:"index_size_bytes"`
	IndexHitRatio    float64 `json:"index_hit_ratio"`
	UnusedIndexes    int     `json:"unused_indexes"`
	DuplicateIndexes int     `json:"duplicate_indexes"`
}

// LockStats represents lock statistics
type LockStats struct {
	Total       int            `json:"total"`
	Granted     int            `json:"granted"`
	Waiting     int            `json:"waiting"`
	Deadlocks   int64          `json:"deadlocks"`
	LocksByType map[string]int `json:"by_type"`
	LocksByMode map[string]int `json:"by_mode"`
}

// BGWriterStats represents background writer statistics
type BGWriterStats struct {
	CheckpointsRequired  int64 `json:"checkpoints_timed"`
	CheckpointsRequested int64 `json:"checkpoints_requested"`
	BuffersCheckpoint    int64 `json:"buffers_checkpoint"`
	BuffersClean         int64 `json:"buffers_clean"`
	MaxWrittenClean      int64 `json:"maxwritten_clean"`
	BuffersBackend       int64 `json:"buffers_backend"`
	BuffersBackendFsync  int64 `json:"buffers_backend_fsync"`
	BuffersAlloc         int64 `json:"buffers_alloc"`
}

// NewPostgresStatsCollector creates a new PostgreSQL stats collector
func NewPostgresStatsCollector(logger *zap.Logger, interval time.Duration) *PostgresStatsCollector {
	return &PostgresStatsCollector{
		logger:    logger,
		databases: make(map[string]*DBConnection),
		interval:  interval,
		stopCh:    make(chan struct{}),
	}
}

// RegisterDatabase registers a database for stats collection
func (psc *PostgresStatsCollector) RegisterDatabase(databaseID, dsn string) error {
	psc.mu.Lock()
	defer psc.mu.Unlock()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	psc.databases[databaseID] = &DBConnection{
		DSN:        dsn,
		DB:         db,
		DatabaseID: databaseID,
	}

	psc.logger.Info("registered database for stats collection", zap.String("database_id", databaseID))
	return nil
}

// UnregisterDatabase removes a database from stats collection
func (psc *PostgresStatsCollector) UnregisterDatabase(databaseID string) {
	psc.mu.Lock()
	defer psc.mu.Unlock()

	if conn, ok := psc.databases[databaseID]; ok {
		if conn.DB != nil {
			conn.DB.Close()
		}
		delete(psc.databases, databaseID)
		psc.logger.Info("unregistered database from stats collection", zap.String("database_id", databaseID))
	}
}

// Start starts the stats collection loop
func (psc *PostgresStatsCollector) Start(ctx context.Context) {
	ticker := time.NewTicker(psc.interval)
	defer ticker.Stop()

	psc.logger.Info("PostgreSQL stats collector started", zap.Duration("interval", psc.interval))

	for {
		select {
		case <-ctx.Done():
			psc.logger.Info("PostgreSQL stats collector stopped")
			return
		case <-psc.stopCh:
			psc.logger.Info("PostgreSQL stats collector stopped")
			return
		case <-ticker.C:
			psc.collectAll(ctx)
		}
	}
}

// Stop stops the stats collector
func (psc *PostgresStatsCollector) Stop() {
	close(psc.stopCh)

	psc.mu.Lock()
	defer psc.mu.Unlock()

	for _, conn := range psc.databases {
		if conn.DB != nil {
			conn.DB.Close()
		}
	}
}

// collectAll collects stats from all registered databases
func (psc *PostgresStatsCollector) collectAll(ctx context.Context) {
	psc.mu.RLock()
	databases := make([]*DBConnection, 0, len(psc.databases))
	for _, db := range psc.databases {
		databases = append(databases, db)
	}
	psc.mu.RUnlock()

	for _, dbConn := range databases {
		stats, err := psc.CollectStats(ctx, dbConn)
		if err != nil {
			psc.logger.Warn("failed to collect stats",
				zap.String("database_id", dbConn.DatabaseID),
				zap.Error(err))
			dbConn.LastError = err
			continue
		}

		dbConn.LastStats = stats
		dbConn.LastCollect = time.Now()
		dbConn.LastError = nil
	}
}

// CollectStats collects comprehensive statistics from a database
func (psc *PostgresStatsCollector) CollectStats(ctx context.Context, dbConn *DBConnection) (*PostgresStats, error) {
	if dbConn.DB == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	stats := &PostgresStats{
		DatabaseID:  dbConn.DatabaseID,
		CollectedAt: time.Now(),
	}

	if err := psc.collectDatabaseInfo(ctx, dbConn.DB, stats); err != nil {
		psc.logger.Warn("failed to collect database info", zap.Error(err))
	}
	if err := psc.collectConnectionStats(ctx, dbConn.DB, stats); err != nil {
		psc.logger.Warn("failed to collect connection stats", zap.Error(err))
	}
	if err := psc.collectQueryStats(ctx, dbConn.DB, stats); err != nil {
		psc.logger.Warn("failed to collect query stats", zap.Error(err))
	}
	if err := psc.collectReplicationStats(ctx, dbConn.DB, stats); err != nil {
		psc.logger.Warn("failed to collect replication stats", zap.Error(err))
	}
	if err := psc.collectTableStats(ctx, dbConn.DB, stats); err != nil {
		psc.logger.Warn("failed to collect table stats", zap.Error(err))
	}
	if err := psc.collectIndexStats(ctx, dbConn.DB, stats); err != nil {
		psc.logger.Warn("failed to collect index stats", zap.Error(err))
	}
	if err := psc.collectLockStats(ctx, dbConn.DB, stats); err != nil {
		psc.logger.Warn("failed to collect lock stats", zap.Error(err))
	}
	if err := psc.collectBGWriterStats(ctx, dbConn.DB, stats); err != nil {
		psc.logger.Warn("failed to collect bgwriter stats", zap.Error(err))
	}

	return stats, nil
}

func (psc *PostgresStatsCollector) collectDatabaseInfo(ctx context.Context, db *sql.DB, stats *PostgresStats) error {
	query := `SELECT current_database(), pg_database_size(current_database())`
	return db.QueryRowContext(ctx, query).Scan(&stats.DatabaseName, &stats.Size)
}

func (psc *PostgresStatsCollector) collectConnectionStats(ctx context.Context, db *sql.DB, stats *PostgresStats) error {
	query := `SELECT state, count(*) FROM pg_stat_activity WHERE backend_type = 'client backend' GROUP BY state`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	stats.Connections.ByState = make(map[string]int)
	for rows.Next() {
		var state sql.NullString
		var count int
		if err := rows.Scan(&state, &count); err != nil {
			continue
		}
		stateStr := "unknown"
		if state.Valid {
			stateStr = state.String
		}
		stats.Connections.ByState[stateStr] = count
		stats.Connections.Total += count

		switch stateStr {
		case "active":
			stats.Connections.Active = count
		case "idle":
			stats.Connections.Idle = count
		case "idle in transaction":
			stats.Connections.IdleInTx = count
		}
	}

	maxQuery := `SELECT setting::int FROM pg_settings WHERE name = 'max_connections'`
	if err := db.QueryRowContext(ctx, maxQuery).Scan(&stats.Connections.MaxConnections); err == nil {
		if stats.Connections.MaxConnections > 0 {
			stats.Connections.PercentUsed = float64(stats.Connections.Total) / float64(stats.Connections.MaxConnections) * 100
		}
	}

	return nil
}

func (psc *PostgresStatsCollector) collectQueryStats(ctx context.Context, db *sql.DB, stats *PostgresStats) error {
	query := `SELECT xact_commit + xact_rollback as total_queries, blks_hit::float / NULLIF(blks_hit + blks_read, 0) as cache_hit_ratio FROM pg_stat_database WHERE datname = current_database()`
	var cacheHitRatio sql.NullFloat64
	if err := db.QueryRowContext(ctx, query).Scan(&stats.Queries.TotalQueries, &cacheHitRatio); err != nil {
		return err
	}
	if cacheHitRatio.Valid {
		stats.Queries.CacheHitRatio = cacheHitRatio.Float64 * 100
	}
	return nil
}

func (psc *PostgresStatsCollector) collectReplicationStats(ctx context.Context, db *sql.DB, stats *PostgresStats) error {
	var isReplica bool
	if err := db.QueryRowContext(ctx, "SELECT pg_is_in_recovery()").Scan(&isReplica); err != nil {
		return err
	}
	stats.Replication.IsReplica = isReplica

	if isReplica {
		lagQuery := `SELECT COALESCE(EXTRACT(EPOCH FROM (now() - pg_last_xact_replay_timestamp())), 0)`
		db.QueryRowContext(ctx, lagQuery).Scan(&stats.Replication.ReplicationLag)
	} else {
		replicaQuery := `SELECT client_addr, state, pg_wal_lsn_diff(sent_lsn, replay_lsn) as replay_lag, sync_state FROM pg_stat_replication`
		rows, err := db.QueryContext(ctx, replicaQuery)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var ri ReplicaInfo
				var clientAddr sql.NullString
				if err := rows.Scan(&clientAddr, &ri.State, &ri.ReplayLag, &ri.SyncState); err == nil {
					if clientAddr.Valid {
						ri.ClientAddr = clientAddr.String
					}
					stats.Replication.Replicas = append(stats.Replication.Replicas, ri)
					stats.Replication.ReplicaCount++
				}
			}
		}
	}
	return nil
}

func (psc *PostgresStatsCollector) collectTableStats(ctx context.Context, db *sql.DB, stats *PostgresStats) error {
	query := `SELECT count(*), COALESCE(sum(n_live_tup), 0), COALESCE(sum(n_dead_tup), 0), COALESCE(sum(seq_scan), 0), COALESCE(sum(idx_scan), 0) FROM pg_stat_user_tables`
	if err := db.QueryRowContext(ctx, query).Scan(&stats.Tables.TotalTables, &stats.Tables.LiveTuples, &stats.Tables.DeadTuples, &stats.Tables.SeqScans, &stats.Tables.IndexScans); err != nil {
		return err
	}
	stats.Tables.TotalRows = stats.Tables.LiveTuples
	if stats.Tables.SeqScans+stats.Tables.IndexScans > 0 {
		stats.Tables.SeqScanRatio = float64(stats.Tables.SeqScans) / float64(stats.Tables.SeqScans+stats.Tables.IndexScans) * 100
	}
	return nil
}

func (psc *PostgresStatsCollector) collectIndexStats(ctx context.Context, db *sql.DB, stats *PostgresStats) error {
	query := `SELECT count(*), COALESCE(sum(pg_relation_size(indexrelid)), 0) FROM pg_stat_user_indexes`
	if err := db.QueryRowContext(ctx, query).Scan(&stats.Indexes.TotalIndexes, &stats.Indexes.IndexSize); err != nil {
		return err
	}
	hitQuery := `SELECT sum(idx_blks_hit)::float / NULLIF(sum(idx_blks_hit + idx_blks_read), 0) FROM pg_statio_user_indexes`
	var hitRatio sql.NullFloat64
	if err := db.QueryRowContext(ctx, hitQuery).Scan(&hitRatio); err == nil && hitRatio.Valid {
		stats.Indexes.IndexHitRatio = hitRatio.Float64 * 100
	}
	unusedQuery := `SELECT count(*) FROM pg_stat_user_indexes WHERE idx_scan = 0`
	db.QueryRowContext(ctx, unusedQuery).Scan(&stats.Indexes.UnusedIndexes)
	return nil
}

func (psc *PostgresStatsCollector) collectLockStats(ctx context.Context, db *sql.DB, stats *PostgresStats) error {
	query := `SELECT count(*), count(*) FILTER (WHERE granted), count(*) FILTER (WHERE NOT granted) FROM pg_locks`
	if err := db.QueryRowContext(ctx, query).Scan(&stats.Locks.Total, &stats.Locks.Granted, &stats.Locks.Waiting); err != nil {
		return err
	}
	deadlockQuery := `SELECT deadlocks FROM pg_stat_database WHERE datname = current_database()`
	db.QueryRowContext(ctx, deadlockQuery).Scan(&stats.Locks.Deadlocks)
	return nil
}

func (psc *PostgresStatsCollector) collectBGWriterStats(ctx context.Context, db *sql.DB, stats *PostgresStats) error {
	query := `SELECT checkpoints_timed, checkpoints_req, buffers_checkpoint, buffers_clean, maxwritten_clean, buffers_backend, buffers_backend_fsync, buffers_alloc FROM pg_stat_bgwriter`
	return db.QueryRowContext(ctx, query).Scan(&stats.BGWriter.CheckpointsRequired, &stats.BGWriter.CheckpointsRequested, &stats.BGWriter.BuffersCheckpoint, &stats.BGWriter.BuffersClean, &stats.BGWriter.MaxWrittenClean, &stats.BGWriter.BuffersBackend, &stats.BGWriter.BuffersBackendFsync, &stats.BGWriter.BuffersAlloc)
}

// GetStats returns the latest stats for a database
func (psc *PostgresStatsCollector) GetStats(databaseID string) (*PostgresStats, error) {
	psc.mu.RLock()
	defer psc.mu.RUnlock()

	dbConn, ok := psc.databases[databaseID]
	if !ok {
		return nil, fmt.Errorf("database not registered: %s", databaseID)
	}
	if dbConn.LastStats == nil {
		return nil, fmt.Errorf("no stats available yet for database: %s", databaseID)
	}
	return dbConn.LastStats, nil
}

// GetAllStats returns stats for all registered databases
func (psc *PostgresStatsCollector) GetAllStats() map[string]*PostgresStats {
	psc.mu.RLock()
	defer psc.mu.RUnlock()

	result := make(map[string]*PostgresStats)
	for id, dbConn := range psc.databases {
		if dbConn.LastStats != nil {
			result[id] = dbConn.LastStats
		}
	}
	return result
}

