package validation

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// ValidateDatabaseConnection validates a PostgreSQL database connection
// Returns an error if the connection cannot be established or the database is not accessible
func ValidateDatabaseConnection(ctx context.Context, host, port, database, username, password, primaryEndpoint string) error {
	// If primaryEndpoint is provided and is a full connection string, use it
	var dsn string
	if primaryEndpoint != "" {
		// Check if it's already a DSN format (starts with postgres:// or postgresql://)
		if strings.HasPrefix(primaryEndpoint, "postgres://") || strings.HasPrefix(primaryEndpoint, "postgresql://") {
			dsn = primaryEndpoint
		}
	}

	// Build DSN from individual connection details if not using primaryEndpoint
	if dsn == "" {
		if host == "" || database == "" {
			return fmt.Errorf("database host and database name are required for validation")
		}

		portStr := port
		if portStr == "" {
			portStr = "5432" // Default PostgreSQL port
		}

		dsn = fmt.Sprintf("host=%s port=%s dbname=%s", host, portStr, database)
		
		if username != "" {
			dsn += fmt.Sprintf(" user=%s", username)
		}
		
		if password != "" {
			dsn += fmt.Sprintf(" password=%s", password)
		}
		
		dsn += " sslmode=prefer connect_timeout=10"
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Open database connection
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	// Test connection with ping
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	// Verify database exists by running a simple query
	var result int
	if err := db.QueryRowContext(ctx, "SELECT 1").Scan(&result); err != nil {
		return fmt.Errorf("database query failed: %w", err)
	}

	return nil
}

// HasDatabaseInfo checks if sufficient database connection information is provided
func HasDatabaseInfo(host, port, database, username, password, primaryEndpoint string) bool {
	// If primaryEndpoint is provided and looks like a connection string, consider it valid
	if primaryEndpoint != "" {
		if strings.HasPrefix(primaryEndpoint, "postgres://") || strings.HasPrefix(primaryEndpoint, "postgresql://") {
			return true
		}
	}

	// Otherwise, require at least host and database name
	return host != "" && database != ""
}

