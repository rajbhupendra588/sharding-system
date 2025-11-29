package scanner

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sharding-system/pkg/discovery"
	"go.uber.org/zap"
	_ "github.com/lib/pq" // PostgreSQL driver
	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

// ScanResult represents the result of scanning a database
type ScanResult struct {
	ID              string            `json:"id"`
	ClusterID       string            `json:"cluster_id"`
	ClusterName     string            `json:"cluster_name"`
	DatabaseName    string            `json:"database_name"`
	DatabaseHost    string            `json:"database_host"`
	DatabasePort    string            `json:"database_port"`
	DatabaseType    string            `json:"database_type"` // "postgresql", "mysql", etc.
	Status          string            `json:"status"`        // "success", "failed", "partial"
	Error           string            `json:"error,omitempty"`
	Tables          []TableInfo        `json:"tables"`
	Schemas         []SchemaInfo       `json:"schemas,omitempty"` // For PostgreSQL
	SizeBytes       int64              `json:"size_bytes"`
	TableCount      int                `json:"table_count"`
	TotalRowCount   int64              `json:"total_row_count"`
	ScannedAt       time.Time          `json:"scanned_at"`
	DurationMs      int64              `json:"duration_ms"`
	Metadata        map[string]string  `json:"metadata,omitempty"`
}

// TableInfo represents information about a database table
type TableInfo struct {
	Name           string            `json:"name"`
	Schema         string            `json:"schema,omitempty"` // For PostgreSQL
	Type           string            `json:"type"`            // "table", "view", "materialized_view"
	Columns        []ColumnInfo      `json:"columns"`
	Indexes        []IndexInfo       `json:"indexes"`
	RowCount       int64             `json:"row_count"`
	SizeBytes      int64             `json:"size_bytes"`
	PrimaryKey     []string          `json:"primary_key,omitempty"`
	ForeignKeys    []ForeignKeyInfo  `json:"foreign_keys,omitempty"`
	Constraints    []ConstraintInfo  `json:"constraints,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// ColumnInfo represents information about a table column
type ColumnInfo struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Nullable     bool   `json:"nullable"`
	DefaultValue string `json:"default_value,omitempty"`
	IsPrimaryKey bool   `json:"is_primary_key"`
	IsUnique     bool   `json:"is_unique"`
	IsIndexed    bool   `json:"is_indexed"`
	MaxLength    int    `json:"max_length,omitempty"`
	Comment      string `json:"comment,omitempty"`
}

// IndexInfo represents information about a database index
type IndexInfo struct {
	Name        string   `json:"name"`
	Columns     []string `json:"columns"`
	IsUnique    bool     `json:"is_unique"`
	IsPrimary   bool     `json:"is_primary"`
	Type        string   `json:"type,omitempty"` // "btree", "hash", "gin", etc.
	SizeBytes   int64    `json:"size_bytes,omitempty"`
}

// ForeignKeyInfo represents a foreign key relationship
type ForeignKeyInfo struct {
	Name              string `json:"name"`
	Columns           []string `json:"columns"`
	ReferencedTable   string `json:"referenced_table"`
	ReferencedColumns []string `json:"referenced_columns"`
	OnDelete          string `json:"on_delete,omitempty"` // "CASCADE", "RESTRICT", etc.
	OnUpdate          string `json:"on_update,omitempty"`
}

// ConstraintInfo represents a table constraint
type ConstraintInfo struct {
	Name      string `json:"name"`
	Type      string `json:"type"` // "CHECK", "UNIQUE", etc.
	Definition string `json:"definition,omitempty"`
}

// SchemaInfo represents a database schema (PostgreSQL)
type SchemaInfo struct {
	Name        string `json:"name"`
	Owner       string `json:"owner,omitempty"`
	TableCount  int    `json:"table_count"`
	SizeBytes   int64  `json:"size_bytes,omitempty"`
}

// LegacyDatabaseScanner scans databases to extract schema information (legacy - use db_scanner.go instead)
// This is kept for backward compatibility but db_scanner.go should be used for new code
type LegacyDatabaseScanner struct {
	logger *zap.Logger
}

// NewLegacyDatabaseScanner creates a new legacy database scanner
func NewLegacyDatabaseScanner(logger *zap.Logger) *LegacyDatabaseScanner {
	return &LegacyDatabaseScanner{
		logger: logger,
	}
}

// ScanDatabase scans a discovered database and extracts schema information
func (ds *LegacyDatabaseScanner) ScanDatabase(ctx context.Context, app *discovery.DiscoveredApp, clusterID, clusterName string, password string) (*ScanResult, error) {
	startTime := time.Now()
	result := &ScanResult{
		ID:           uuid.New().String(),
		ClusterID:    clusterID,
		ClusterName:   clusterName,
		DatabaseName: app.DatabaseName,
		DatabaseHost: app.DatabaseHost,
		DatabasePort: app.DatabasePort,
		Status:       "failed",
		ScannedAt:    startTime,
		Tables:       make([]TableInfo, 0),
		Metadata:     make(map[string]string),
	}

	// Determine database type from URL or port
	dbType := ds.detectDatabaseType(app.DatabaseURL, app.DatabasePort)
	result.DatabaseType = dbType

	// Build connection string
	connStr, err := ds.buildConnectionString(app, password, dbType)
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	// Connect to database
	db, err := sql.Open(dbType, connStr)
	if err != nil {
		result.Error = fmt.Sprintf("failed to connect: %v", err)
		return result, err
	}
	defer db.Close()

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		result.Error = fmt.Sprintf("connection test failed: %v", err)
		return result, err
	}

	// Scan based on database type
	switch dbType {
	case "postgres":
		err = ds.scanPostgreSQL(ctx, db, app.DatabaseName, result)
	case "mysql":
		err = ds.scanMySQL(ctx, db, app.DatabaseName, result)
	default:
		err = fmt.Errorf("unsupported database type: %s", dbType)
	}

	if err != nil {
		result.Error = err.Error()
		result.Status = "partial"
		if len(result.Tables) == 0 {
			result.Status = "failed"
		}
	} else {
		result.Status = "success"
	}

	result.DurationMs = time.Since(startTime).Milliseconds()
	result.TableCount = len(result.Tables)

	// Calculate total row count
	for _, table := range result.Tables {
		result.TotalRowCount += table.RowCount
		result.SizeBytes += table.SizeBytes
	}

	ds.logger.Info("database scan completed",
		zap.String("scan_id", result.ID),
		zap.String("database", app.DatabaseName),
		zap.String("cluster", clusterName),
		zap.String("status", result.Status),
		zap.Int("table_count", result.TableCount),
		zap.Int64("duration_ms", result.DurationMs))

	return result, nil
}

// detectDatabaseType detects database type from URL or port
func (ds *LegacyDatabaseScanner) detectDatabaseType(url, port string) string {
	if url != "" {
		urlLower := strings.ToLower(url)
		if strings.Contains(urlLower, "postgres") {
			return "postgres"
		}
		if strings.Contains(urlLower, "mysql") {
			return "mysql"
		}
	}

	// Default port detection
	if port == "5432" || port == "" {
		return "postgres"
	}
	if port == "3306" {
		return "mysql"
	}

	// Default to PostgreSQL
	return "postgres"
}

// buildConnectionString builds a database connection string
func (ds *LegacyDatabaseScanner) buildConnectionString(app *discovery.DiscoveredApp, password, dbType string) (string, error) {
	if app.DatabaseURL != "" && !strings.HasPrefix(app.DatabaseURL, "secret:") {
		// If password is provided and URL doesn't have it, inject it
		if password != "" && !strings.Contains(app.DatabaseURL, "@") {
			// URL format: postgres://user@host:port/dbname
			// Need to inject password
			parts := strings.Split(app.DatabaseURL, "@")
			if len(parts) == 2 {
				return fmt.Sprintf("%s:%s@%s", parts[0], password, parts[1]), nil
			}
		}
		return app.DatabaseURL, nil
	}

	// Build from components
	if app.DatabaseHost == "" {
		return "", fmt.Errorf("database host is required")
	}

	port := app.DatabasePort
	if port == "" {
		if dbType == "postgres" {
			port = "5432"
		} else if dbType == "mysql" {
			port = "3306"
		}
	}

	user := app.DatabaseUser
	if user == "" {
		user = "postgres" // Default
	}

	if dbType == "postgres" {
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			app.DatabaseHost, port, user, password, app.DatabaseName), nil
	} else if dbType == "mysql" {
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			user, password, app.DatabaseHost, port, app.DatabaseName), nil
	}

	return "", fmt.Errorf("unsupported database type: %s", dbType)
}

// scanPostgreSQL scans a PostgreSQL database
func (ds *LegacyDatabaseScanner) scanPostgreSQL(ctx context.Context, db *sql.DB, dbName string, result *ScanResult) error {
	// Get database size
	var sizeBytes int64
	err := db.QueryRowContext(ctx, "SELECT pg_database_size($1)", dbName).Scan(&sizeBytes)
	if err == nil {
		result.SizeBytes = sizeBytes
	}

	// Get all schemas
	schemas, err := ds.getPostgreSQLSchemas(ctx, db)
	if err == nil {
		result.Schemas = schemas
	}

	// Get all tables from all schemas (excluding system schemas)
	query := `
		SELECT schemaname, tablename, 'table' as tabletype
		FROM pg_tables
		WHERE schemaname NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
		UNION ALL
		SELECT schemaname, viewname as tablename, 'view' as tabletype
		FROM pg_views
		WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
		ORDER BY schemaname, tablename
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var schema, tableName, tableType string
		if err := rows.Scan(&schema, &tableName, &tableType); err != nil {
			continue
		}

		tableInfo, err := ds.scanPostgreSQLTable(ctx, db, schema, tableName, tableType)
		if err != nil {
			ds.logger.Warn("failed to scan table",
				zap.String("schema", schema),
				zap.String("table", tableName),
				zap.Error(err))
			continue
		}

		result.Tables = append(result.Tables, *tableInfo)
	}

	return nil
}

// getPostgreSQLSchemas gets all schemas in the database
func (ds *LegacyDatabaseScanner) getPostgreSQLSchemas(ctx context.Context, db *sql.DB) ([]SchemaInfo, error) {
	query := `
		SELECT schema_name, schema_owner
		FROM information_schema.schemata
		WHERE schema_name NOT IN ('pg_catalog', 'information_schema', 'pg_toast', 'pg_temp_1', 'pg_toast_temp_1')
		ORDER BY schema_name
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schemas := make([]SchemaInfo, 0)
	for rows.Next() {
		var schema SchemaInfo
		if err := rows.Scan(&schema.Name, &schema.Owner); err != nil {
			continue
		}
		schemas = append(schemas, schema)
	}

	return schemas, nil
}

// scanPostgreSQLTable scans a single PostgreSQL table
func (ds *LegacyDatabaseScanner) scanPostgreSQLTable(ctx context.Context, db *sql.DB, schema, tableName, tableType string) (*TableInfo, error) {
	table := &TableInfo{
		Name:     tableName,
		Schema:   schema,
		Type:     tableType,
		Columns:  make([]ColumnInfo, 0),
		Indexes:  make([]IndexInfo, 0),
		Metadata: make(map[string]string),
	}

	fullTableName := fmt.Sprintf("%s.%s", schema, tableName)
	if schema == "public" {
		fullTableName = tableName
	}

	// Get columns
	columnQuery := `
		SELECT 
			column_name,
			data_type,
			is_nullable,
			column_default,
			character_maximum_length,
			col_description((SELECT oid FROM pg_class WHERE relname = $1 AND relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = $2)), ordinal_position)
		FROM information_schema.columns
		WHERE table_schema = $2 AND table_name = $1
		ORDER BY ordinal_position
	`

	rows, err := db.QueryContext(ctx, columnQuery, tableName, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	columnMap := make(map[string]*ColumnInfo)
	for rows.Next() {
		var col ColumnInfo
		var nullable, maxLength sql.NullString
		var comment sql.NullString

		if err := rows.Scan(&col.Name, &col.Type, &nullable, &col.DefaultValue, &maxLength, &comment); err != nil {
			continue
		}

		col.Nullable = nullable.String == "YES"
		if maxLength.Valid {
			// Parse max length if it's a number
			fmt.Sscanf(maxLength.String, "%d", &col.MaxLength)
		}
		if comment.Valid {
			col.Comment = comment.String
		}

		columnMap[col.Name] = &col
		table.Columns = append(table.Columns, col)
	}

	// Get primary key
	pkQuery := `
		SELECT a.attname
		FROM pg_index i
		JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
		WHERE i.indrelid = $1::regclass AND i.indisprimary
		ORDER BY a.attnum
	`

	pkRows, err := db.QueryContext(ctx, pkQuery, fullTableName)
	if err == nil {
		defer pkRows.Close()
		for pkRows.Next() {
			var colName string
			if err := pkRows.Scan(&colName); err == nil {
				if col, exists := columnMap[colName]; exists {
					col.IsPrimaryKey = true
					table.PrimaryKey = append(table.PrimaryKey, colName)
				}
			}
		}
	}

	// Get indexes
	indexQuery := `
		SELECT
			i.relname as index_name,
			array_agg(a.attname ORDER BY array_position(i.indkey, a.attnum)) as columns,
			ix.indisunique,
			ix.indisprimary,
			am.amname as index_type
		FROM pg_index ix
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_class t ON t.oid = ix.indrelid
		JOIN pg_am am ON am.oid = i.relam
		LEFT JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
		WHERE t.relname = $1 AND t.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = $2)
		GROUP BY i.relname, ix.indisunique, ix.indisprimary, am.amname
	`

	idxRows, err := db.QueryContext(ctx, indexQuery, tableName, schema)
	if err == nil {
		defer idxRows.Close()
		for idxRows.Next() {
			var idx IndexInfo
			var columns string
			if err := idxRows.Scan(&idx.Name, &columns, &idx.IsUnique, &idx.IsPrimary, &idx.Type); err == nil {
				// Parse PostgreSQL array format
				columns = strings.Trim(columns, "{}")
				if columns != "" {
					idx.Columns = strings.Split(columns, ",")
				}
				table.Indexes = append(table.Indexes, idx)
			}
		}
	}

	// Get row count and size
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", fullTableName)
	err = db.QueryRowContext(ctx, countQuery).Scan(&table.RowCount)
	if err != nil {
		table.RowCount = -1 // Unknown
	}

	sizeQuery := `
		SELECT pg_total_relation_size($1::regclass)
	`
	err = db.QueryRowContext(ctx, sizeQuery, fullTableName).Scan(&table.SizeBytes)
	if err != nil {
		table.SizeBytes = -1
	}

	// Get foreign keys
	fkQuery := `
		SELECT
			tc.constraint_name,
			kcu.column_name,
			ccu.table_name AS foreign_table_name,
			ccu.column_name AS foreign_column_name,
			rc.delete_rule,
			rc.update_rule
		FROM information_schema.table_constraints AS tc
		JOIN information_schema.key_column_usage AS kcu
			ON tc.constraint_name = kcu.constraint_name
		JOIN information_schema.constraint_column_usage AS ccu
			ON ccu.constraint_name = tc.constraint_name
		JOIN information_schema.referential_constraints AS rc
			ON rc.constraint_name = tc.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY'
			AND tc.table_schema = $2
			AND tc.table_name = $1
	`

	fkRows, err := db.QueryContext(ctx, fkQuery, tableName, schema)
	if err == nil {
		defer fkRows.Close()
		fkMap := make(map[string]*ForeignKeyInfo)
		for fkRows.Next() {
			var fkName, colName, refTable, refCol, onDelete, onUpdate string
			if err := fkRows.Scan(&fkName, &colName, &refTable, &refCol, &onDelete, &onUpdate); err == nil {
				if fk, exists := fkMap[fkName]; exists {
					fk.Columns = append(fk.Columns, colName)
					fk.ReferencedColumns = append(fk.ReferencedColumns, refCol)
				} else {
					fkMap[fkName] = &ForeignKeyInfo{
						Name:              fkName,
						Columns:           []string{colName},
						ReferencedTable:   refTable,
						ReferencedColumns: []string{refCol},
						OnDelete:          onDelete,
						OnUpdate:          onUpdate,
					}
				}
			}
		}
		for _, fk := range fkMap {
			table.ForeignKeys = append(table.ForeignKeys, *fk)
		}
	}

	return table, nil
}

// scanMySQL scans a MySQL database
func (ds *LegacyDatabaseScanner) scanMySQL(ctx context.Context, db *sql.DB, dbName string, result *ScanResult) error {
	// Get database size
	var sizeBytes int64
	sizeQuery := `
		SELECT SUM(data_length + index_length) as size
		FROM information_schema.tables
		WHERE table_schema = ?
	`
	err := db.QueryRowContext(ctx, sizeQuery, dbName).Scan(&sizeBytes)
	if err == nil {
		result.SizeBytes = sizeBytes
	}

	// Get all tables
	query := `
		SELECT table_name, table_type
		FROM information_schema.tables
		WHERE table_schema = ?
		ORDER BY table_name
	`

	rows, err := db.QueryContext(ctx, query, dbName)
	if err != nil {
		return fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName, tableType string
		if err := rows.Scan(&tableName, &tableType); err != nil {
			continue
		}

		tableInfo, err := ds.scanMySQLTable(ctx, db, dbName, tableName, tableType)
		if err != nil {
			ds.logger.Warn("failed to scan table",
				zap.String("table", tableName),
				zap.Error(err))
			continue
		}

		result.Tables = append(result.Tables, *tableInfo)
	}

	return nil
}

// scanMySQLTable scans a single MySQL table
func (ds *LegacyDatabaseScanner) scanMySQLTable(ctx context.Context, db *sql.DB, dbName, tableName, tableType string) (*TableInfo, error) {
	table := &TableInfo{
		Name:     tableName,
		Type:     strings.ToLower(tableType),
		Columns:  make([]ColumnInfo, 0),
		Indexes:  make([]IndexInfo, 0),
		Metadata: make(map[string]string),
	}

	// Get columns
	columnQuery := `
		SELECT 
			column_name,
			data_type,
			is_nullable,
			column_default,
			character_maximum_length,
			column_comment,
			column_key,
			extra
		FROM information_schema.columns
		WHERE table_schema = ? AND table_name = ?
		ORDER BY ordinal_position
	`

	rows, err := db.QueryContext(ctx, columnQuery, dbName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	columnMap := make(map[string]*ColumnInfo)
	for rows.Next() {
		var col ColumnInfo
		var nullable, maxLength, comment, columnKey, extra sql.NullString

		if err := rows.Scan(&col.Name, &col.Type, &nullable, &col.DefaultValue, &maxLength, &comment, &columnKey, &extra); err != nil {
			continue
		}

		col.Nullable = nullable.String == "YES"
		if maxLength.Valid {
			fmt.Sscanf(maxLength.String, "%d", &col.MaxLength)
		}
		if comment.Valid {
			col.Comment = comment.String
		}
		if columnKey.String == "PRI" {
			col.IsPrimaryKey = true
			table.PrimaryKey = append(table.PrimaryKey, col.Name)
		}
		if columnKey.String == "UNI" {
			col.IsUnique = true
		}

		columnMap[col.Name] = &col
		table.Columns = append(table.Columns, col)
	}

	// Get indexes
	indexQuery := `
		SELECT
			index_name,
			GROUP_CONCAT(column_name ORDER BY seq_in_index) as columns,
			non_unique = 0 as is_unique,
			index_type
		FROM information_schema.statistics
		WHERE table_schema = ? AND table_name = ?
		GROUP BY index_name, non_unique, index_type
	`

	idxRows, err := db.QueryContext(ctx, indexQuery, dbName, tableName)
	if err == nil {
		defer idxRows.Close()
		for idxRows.Next() {
			var idx IndexInfo
			var columns string
			var isUnique int
			if err := idxRows.Scan(&idx.Name, &columns, &isUnique, &idx.Type); err == nil {
				idx.IsUnique = isUnique == 1
				idx.IsPrimary = idx.Name == "PRIMARY"
				if columns != "" {
					idx.Columns = strings.Split(columns, ",")
				}
				table.Indexes = append(table.Indexes, idx)
			}
		}
	}

	// Get row count and size
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tableName)
	err = db.QueryRowContext(ctx, countQuery).Scan(&table.RowCount)
	if err != nil {
		table.RowCount = -1
	}

	sizeQuery := `
		SELECT data_length + index_length
		FROM information_schema.tables
		WHERE table_schema = ? AND table_name = ?
	`
	err = db.QueryRowContext(ctx, sizeQuery, dbName, tableName).Scan(&table.SizeBytes)
	if err != nil {
		table.SizeBytes = -1
	}

	// Get foreign keys
	fkQuery := `
		SELECT
			constraint_name,
			column_name,
			referenced_table_name,
			referenced_column_name,
			delete_rule,
			update_rule
		FROM information_schema.key_column_usage
		WHERE table_schema = ? 
			AND table_name = ?
			AND referenced_table_name IS NOT NULL
	`

	fkRows, err := db.QueryContext(ctx, fkQuery, dbName, tableName)
	if err == nil {
		defer fkRows.Close()
		fkMap := make(map[string]*ForeignKeyInfo)
		for fkRows.Next() {
			var fkName, colName, refTable, refCol, onDelete, onUpdate string
			if err := fkRows.Scan(&fkName, &colName, &refTable, &refCol, &onDelete, &onUpdate); err == nil {
				if fk, exists := fkMap[fkName]; exists {
					fk.Columns = append(fk.Columns, colName)
					fk.ReferencedColumns = append(fk.ReferencedColumns, refCol)
				} else {
					fkMap[fkName] = &ForeignKeyInfo{
						Name:              fkName,
						Columns:           []string{colName},
						ReferencedTable:   refTable,
						ReferencedColumns: []string{refCol},
						OnDelete:          onDelete,
						OnUpdate:          onUpdate,
					}
				}
			}
		}
		for _, fk := range fkMap {
			table.ForeignKeys = append(table.ForeignKeys, *fk)
		}
	}

	return table, nil
}

