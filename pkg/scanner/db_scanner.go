package scanner

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/sharding-system/pkg/models"
	"go.uber.org/zap"
)

// DatabaseScanner scans databases discovered in Kubernetes clusters
type DatabaseScanner struct {
	logger *zap.Logger
}

// NewDatabaseScanner creates a new database scanner
func NewDatabaseScanner(logger *zap.Logger) *DatabaseScanner {
	return &DatabaseScanner{
		logger: logger,
	}
}

// ScanDatabase performs a deep scan of a database
func (ds *DatabaseScanner) ScanDatabase(ctx context.Context, dbInfo *models.ScannedDatabase, password string) (*models.DatabaseScanResults, error) {
	// Build DSN
	dsn := ds.buildDSN(dbInfo, password)

	// Connect to database
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Set connection timeouts
	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	results := &models.DatabaseScanResults{
		HealthStatus: "healthy",
		Metadata:     make(map[string]interface{}),
	}

	// Collect basic info
	if err := ds.collectBasicInfo(ctx, db, results); err != nil {
		ds.logger.Warn("failed to collect basic info", zap.Error(err))
		results.HealthStatus = "degraded"
	}

	// Collect table stats
	if err := ds.collectTableStats(ctx, db, results); err != nil {
		ds.logger.Warn("failed to collect table stats", zap.Error(err))
	}

	// Collect index stats
	if err := ds.collectIndexStats(ctx, db, results); err != nil {
		ds.logger.Warn("failed to collect index stats", zap.Error(err))
	}

	// Collect connection stats
	if err := ds.collectConnectionStats(ctx, db, results); err != nil {
		ds.logger.Warn("failed to collect connection stats", zap.Error(err))
	}

	// Collect replication info
	if err := ds.collectReplicationInfo(ctx, db, results); err != nil {
		ds.logger.Warn("failed to collect replication info", zap.Error(err))
	}

	return results, nil
}

// buildDSN builds a PostgreSQL DSN from database info
func (ds *DatabaseScanner) buildDSN(dbInfo *models.ScannedDatabase, password string) string {
	parts := []string{
		fmt.Sprintf("host=%s", dbInfo.Host),
		fmt.Sprintf("port=%d", dbInfo.Port),
		fmt.Sprintf("dbname=%s", dbInfo.Database),
	}

	if dbInfo.Username != "" {
		parts = append(parts, fmt.Sprintf("user=%s", dbInfo.Username))
	}
	if password != "" {
		parts = append(parts, fmt.Sprintf("password=%s", password))
	}

	parts = append(parts, "sslmode=prefer", "connect_timeout=10")

	return strings.Join(parts, " ")
}

// collectBasicInfo collects basic database information
func (ds *DatabaseScanner) collectBasicInfo(ctx context.Context, db *sql.DB, results *models.DatabaseScanResults) error {
	queries := []struct {
		name  string
		query string
		dest  interface{}
	}{
		{"version", "SELECT version()", &results.Version},
		{"size", "SELECT pg_database_size(current_database())", &results.Size},
		{"uptime", "SELECT EXTRACT(EPOCH FROM (now() - pg_postmaster_start_time()))::bigint", &results.Uptime},
	}

	for _, q := range queries {
		if err := db.QueryRowContext(ctx, q.query).Scan(q.dest); err != nil {
			ds.logger.Debug("failed to query", zap.String("query", q.name), zap.Error(err))
		}
	}

	return nil
}

// collectTableStats collects table statistics
func (ds *DatabaseScanner) collectTableStats(ctx context.Context, db *sql.DB, results *models.DatabaseScanResults) error {
	query := `
		SELECT 
			schemaname || '.' || tablename as name,
			n_live_tup as row_count,
			pg_total_relation_size(schemaname||'.'||tablename) as total_size,
			pg_relation_size(schemaname||'.'||tablename) as table_size,
			pg_indexes_size(schemaname||'.'||tablename) as index_size,
			last_vacuum,
			last_autovacuum,
			last_analyze,
			last_autoanalyze
		FROM pg_stat_user_tables
		ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	var tables []models.TableStat
	var tableNames []string

	for rows.Next() {
		var stat models.TableStat
		var lastVacuum, lastAutovacuum, lastAnalyze, lastAutoanalyze sql.NullTime

		err := rows.Scan(
			&stat.Name,
			&stat.RowCount,
			&stat.TotalSize,
			&stat.Size,
			&stat.IndexSize,
			&lastVacuum,
			&lastAutovacuum,
			&lastAnalyze,
			&lastAutoanalyze,
		)
		if err != nil {
			continue
		}

		// Use the most recent vacuum/analyze time
		if lastVacuum.Valid {
			stat.LastVacuum = &lastVacuum.Time
		} else if lastAutovacuum.Valid {
			stat.LastVacuum = &lastAutovacuum.Time
		}

		if lastAnalyze.Valid {
			stat.LastAnalyze = &lastAnalyze.Time
		} else if lastAutoanalyze.Valid {
			stat.LastAnalyze = &lastAutoanalyze.Time
		}

		tables = append(tables, stat)
		tableNames = append(tableNames, stat.Name)
	}

	results.TableStats = tables
	results.TableNames = tableNames
	results.TableCount = len(tables)

	return nil
}

// collectIndexStats collects index statistics
func (ds *DatabaseScanner) collectIndexStats(ctx context.Context, db *sql.DB, results *models.DatabaseScanResults) error {
	query := `
		SELECT 
			schemaname || '.' || indexrelname as name,
			schemaname || '.' || relname as table_name,
			pg_relation_size(indexrelid) as size,
			idx_scan as scans,
			idx_tup_read as tuples_read,
			idx_tup_fetch as tuples_fetched
		FROM pg_stat_user_indexes
		ORDER BY pg_relation_size(indexrelid) DESC
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	var indexes []models.IndexStat

	for rows.Next() {
		var stat models.IndexStat
		if err := rows.Scan(
			&stat.Name,
			&stat.TableName,
			&stat.Size,
			&stat.Scans,
			&stat.TuplesRead,
			&stat.TuplesFetched,
		); err != nil {
			continue
		}

		indexes = append(indexes, stat)
	}

	results.IndexStats = indexes
	results.IndexCount = len(indexes)

	return nil
}

// collectConnectionStats collects connection statistics
func (ds *DatabaseScanner) collectConnectionStats(ctx context.Context, db *sql.DB, results *models.DatabaseScanResults) error {
	query := `
		SELECT 
			count(*) FILTER (WHERE state = 'active') as active,
			(SELECT setting::int FROM pg_settings WHERE name = 'max_connections') as max_conn
		FROM pg_stat_activity
		WHERE backend_type = 'client backend'
	`

	return db.QueryRowContext(ctx, query).Scan(&results.ConnectionCount, &results.MaxConnections)
}

// collectReplicationInfo collects replication information
func (ds *DatabaseScanner) collectReplicationInfo(ctx context.Context, db *sql.DB, results *models.DatabaseScanResults) error {
	// Check if replica
	var isReplica bool
	if err := db.QueryRowContext(ctx, "SELECT pg_is_in_recovery()").Scan(&isReplica); err != nil {
		return err
	}

	results.IsReplica = isReplica

	if isReplica {
		// Get replication lag
		query := `
			SELECT EXTRACT(EPOCH FROM (now() - pg_last_xact_replay_timestamp()))
			WHERE pg_last_xact_replay_timestamp() IS NOT NULL
		`
		var lag sql.NullFloat64
		if err := db.QueryRowContext(ctx, query).Scan(&lag); err == nil && lag.Valid {
			results.ReplicationLag = lag.Float64
		}
	}

	return nil
}

