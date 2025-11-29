// Package schema provides automated schema management across shards
package schema

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// Migration represents a schema migration
type Migration struct {
	ID          string    `json:"id"`
	Version     int       `json:"version"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	SQL         string    `json:"sql"`
	Checksum    string    `json:"checksum"`
	AppliedAt   time.Time `json:"applied_at,omitempty"`
	Duration    int64     `json:"duration_ms,omitempty"`
}

// MigrationStatus tracks migration status per shard
type MigrationStatus struct {
	ShardID    string    `json:"shard_id"`
	ShardName  string    `json:"shard_name"`
	Version    int       `json:"version"`
	Status     string    `json:"status"` // "pending", "applying", "applied", "failed"
	Error      string    `json:"error,omitempty"`
	AppliedAt  time.Time `json:"applied_at,omitempty"`
	DurationMs int64     `json:"duration_ms,omitempty"`
}

// ShardConnection holds connection info for a shard
type ShardConnection struct {
	ID       string
	Name     string
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

// Manager handles schema migrations across shards
type Manager struct {
	logger     *zap.Logger
	migrations map[int]*Migration // version -> migration
	mu         sync.RWMutex
}

// NewManager creates a new schema manager
func NewManager(logger *zap.Logger) *Manager {
	return &Manager{
		logger:     logger,
		migrations: make(map[int]*Migration),
	}
}

// RegisterMigration registers a new migration
func (m *Manager) RegisterMigration(version int, name, description, sqlContent string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.migrations[version]; exists {
		return fmt.Errorf("migration version %d already exists", version)
	}

	checksum := computeChecksum(sqlContent)

	m.migrations[version] = &Migration{
		ID:          fmt.Sprintf("migration_%d", version),
		Version:     version,
		Name:        name,
		Description: description,
		SQL:         sqlContent,
		Checksum:    checksum,
	}

	m.logger.Info("registered migration",
		zap.Int("version", version),
		zap.String("name", name))

	return nil
}

// ApplyMigrations applies pending migrations to all shards
func (m *Manager) ApplyMigrations(ctx context.Context, shards []ShardConnection) ([]MigrationStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var allStatus []MigrationStatus
	var mu sync.Mutex
	var wg sync.WaitGroup
	errors := make(chan error, len(shards))

	for _, shard := range shards {
		wg.Add(1)
		go func(s ShardConnection) {
			defer wg.Done()

			status, err := m.applyMigrationsToShard(ctx, s)
			if err != nil {
				errors <- fmt.Errorf("shard %s: %w", s.Name, err)
			}

			mu.Lock()
			allStatus = append(allStatus, status...)
			mu.Unlock()
		}(shard)
	}

	wg.Wait()
	close(errors)

	// Collect errors
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return allStatus, fmt.Errorf("migration errors: %v", errs)
	}

	return allStatus, nil
}

// applyMigrationsToShard applies migrations to a single shard
func (m *Manager) applyMigrationsToShard(ctx context.Context, shard ShardConnection) ([]MigrationStatus, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		shard.Host, shard.Port, shard.Username, shard.Password, shard.Database)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Create migrations tracking table
	if err := m.createMigrationsTable(ctx, db); err != nil {
		return nil, fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get current version
	currentVersion, err := m.getCurrentVersion(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to get current version: %w", err)
	}

	var statuses []MigrationStatus

	// Apply pending migrations in order
	for version := currentVersion + 1; ; version++ {
		migration, exists := m.migrations[version]
		if !exists {
			break // No more migrations
		}

		status := MigrationStatus{
			ShardID:   shard.ID,
			ShardName: shard.Name,
			Version:   version,
			Status:    "applying",
		}

		start := time.Now()

		// Execute migration in transaction
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			status.Status = "failed"
			status.Error = err.Error()
			statuses = append(statuses, status)
			return statuses, err
		}

		if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
			tx.Rollback()
			status.Status = "failed"
			status.Error = err.Error()
			statuses = append(statuses, status)
			return statuses, fmt.Errorf("migration %d failed: %w", version, err)
		}

		// Record migration
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO _schema_migrations (version, name, checksum, applied_at, duration_ms)
			VALUES ($1, $2, $3, $4, $5)
		`, version, migration.Name, migration.Checksum, time.Now(), time.Since(start).Milliseconds()); err != nil {
			tx.Rollback()
			status.Status = "failed"
			status.Error = err.Error()
			statuses = append(statuses, status)
			return statuses, err
		}

		if err := tx.Commit(); err != nil {
			status.Status = "failed"
			status.Error = err.Error()
			statuses = append(statuses, status)
			return statuses, err
		}

		status.Status = "applied"
		status.AppliedAt = time.Now()
		status.DurationMs = time.Since(start).Milliseconds()
		statuses = append(statuses, status)

		m.logger.Info("applied migration",
			zap.String("shard", shard.Name),
			zap.Int("version", version),
			zap.String("name", migration.Name),
			zap.Int64("duration_ms", status.DurationMs))
	}

	return statuses, nil
}

// createMigrationsTable creates the migrations tracking table
func (m *Manager) createMigrationsTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS _schema_migrations (
			version INTEGER PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			checksum VARCHAR(64) NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE NOT NULL,
			duration_ms BIGINT NOT NULL
		)
	`)
	return err
}

// getCurrentVersion gets the current schema version
func (m *Manager) getCurrentVersion(ctx context.Context, db *sql.DB) (int, error) {
	var version int
	err := db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version), 0) FROM _schema_migrations
	`).Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// GetMigrationHistory returns migration history for a shard
func (m *Manager) GetMigrationHistory(ctx context.Context, shard ShardConnection) ([]Migration, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		shard.Host, shard.Port, shard.Username, shard.Password, shard.Database)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, `
		SELECT version, name, checksum, applied_at, duration_ms 
		FROM _schema_migrations 
		ORDER BY version
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []Migration
	for rows.Next() {
		var mig Migration
		if err := rows.Scan(&mig.Version, &mig.Name, &mig.Checksum, &mig.AppliedAt, &mig.Duration); err != nil {
			return nil, err
		}
		mig.ID = fmt.Sprintf("migration_%d", mig.Version)
		history = append(history, mig)
	}

	return history, nil
}

// ValidateMigrations checks if all shards have consistent schema versions
func (m *Manager) ValidateMigrations(ctx context.Context, shards []ShardConnection) (map[string]int, error) {
	versions := make(map[string]int)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, shard := range shards {
		wg.Add(1)
		go func(s ShardConnection) {
			defer wg.Done()

			dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
				s.Host, s.Port, s.Username, s.Password, s.Database)

			db, err := sql.Open("postgres", dsn)
			if err != nil {
				mu.Lock()
				versions[s.Name] = -1 // Error indicator
				mu.Unlock()
				return
			}
			defer db.Close()

			version, err := m.getCurrentVersion(ctx, db)
			if err != nil {
				version = -1
			}

			mu.Lock()
			versions[s.Name] = version
			mu.Unlock()
		}(shard)
	}

	wg.Wait()
	return versions, nil
}

// GetPendingMigrations returns migrations not yet applied
func (m *Manager) GetPendingMigrations(currentVersion int) []*Migration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var pending []*Migration
	for version, migration := range m.migrations {
		if version > currentVersion {
			pending = append(pending, migration)
		}
	}
	return pending
}

// ListMigrations returns all registered migrations
func (m *Manager) ListMigrations() []*Migration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Migration, 0, len(m.migrations))
	for _, mig := range m.migrations {
		result = append(result, mig)
	}
	return result
}

func computeChecksum(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// SchemaTemplate represents a pre-defined schema template
type SchemaTemplate struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tables      []string `json:"tables"`
	SQL         string   `json:"sql"`
}

// PredefinedSchemas provides common schema templates
var PredefinedSchemas = map[string]SchemaTemplate{
	"users": {
		ID:          "users",
		Name:        "Users Schema",
		Description: "Standard users table with authentication fields",
		Tables:      []string{"users"},
		SQL: `
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    avatar_url TEXT,
    email_verified BOOLEAN DEFAULT FALSE,
    status VARCHAR(20) DEFAULT 'active',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_status ON users(status);
`,
	},
	"orders": {
		ID:          "orders",
		Name:        "E-commerce Orders",
		Description: "Orders and order items for e-commerce applications",
		Tables:      []string{"orders", "order_items"},
		SQL: `
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    order_number VARCHAR(50) NOT NULL UNIQUE,
    status VARCHAR(30) DEFAULT 'pending',
    total_amount DECIMAL(12, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    shipping_address JSONB,
    billing_address JSONB,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(12, 2) NOT NULL,
    total_price DECIMAL(12, 2) NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
`,
	},
	"products": {
		ID:          "products",
		Name:        "Product Catalog",
		Description: "Products with categories and inventory",
		Tables:      []string{"products", "categories", "inventory"},
		SQL: `
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    parent_id UUID REFERENCES categories(id),
    description TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category_id UUID REFERENCES categories(id),
    price DECIMAL(12, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    images JSONB DEFAULT '[]',
    attributes JSONB DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS inventory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    warehouse_id VARCHAR(50) NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 0,
    reserved INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(product_id, warehouse_id)
);

CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_products_sku ON products(sku);
CREATE INDEX idx_products_status ON products(status);
CREATE INDEX idx_inventory_product ON inventory(product_id);
`,
	},
	"analytics": {
		ID:          "analytics",
		Name:        "Analytics Events",
		Description: "Time-series events for analytics and tracking",
		Tables:      []string{"events", "sessions"},
		SQL: `
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    device_id VARCHAR(100),
    ip_address INET,
    user_agent TEXT,
    country VARCHAR(2),
    city VARCHAR(100),
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ended_at TIMESTAMP WITH TIME ZONE,
    page_views INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID REFERENCES sessions(id),
    user_id UUID,
    event_type VARCHAR(100) NOT NULL,
    event_name VARCHAR(255) NOT NULL,
    properties JSONB DEFAULT '{}',
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_started ON sessions(started_at);
CREATE INDEX idx_events_session ON events(session_id);
CREATE INDEX idx_events_user ON events(user_id);
CREATE INDEX idx_events_type ON events(event_type);
CREATE INDEX idx_events_timestamp ON events(timestamp);
`,
	},
}
