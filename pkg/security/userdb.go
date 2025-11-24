package security

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// DBUserStore manages users in PostgreSQL database (MAANG production standard)
type DBUserStore struct {
	db     *sql.DB
	logger *zap.Logger
	mu     sync.RWMutex
	cache  map[string]*User // In-memory cache with TTL
}

// NewDBUserStore creates a new database-backed user store
func NewDBUserStore(dsn string, logger *zap.Logger) (*DBUserStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool (MAANG standards)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &DBUserStore{
		db:     db,
		logger: logger,
		cache:  make(map[string]*User),
	}

	// Initialize schema
	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	// Load default users if table is empty
	if err := store.ensureDefaultUsers(); err != nil {
		logger.Warn("failed to ensure default users", zap.Error(err))
	}

	return store, nil
}

// initSchema creates the users table if it doesn't exist
func (s *DBUserStore) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		username VARCHAR(255) PRIMARY KEY,
		password_hash VARCHAR(255) NOT NULL,
		roles JSONB NOT NULL DEFAULT '[]'::jsonb,
		active BOOLEAN NOT NULL DEFAULT true,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_login_at TIMESTAMP,
		failed_login_attempts INTEGER NOT NULL DEFAULT 0,
		locked_until TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_users_active ON users(active) WHERE active = true;
	CREATE INDEX IF NOT EXISTS idx_users_locked ON users(locked_until) WHERE locked_until IS NOT NULL;
	`

	_, err := s.db.Exec(query)
	return err
}

// ensureDefaultUsers creates default users if table is empty
// NOTE: This is now skipped - users must be set up via the setup endpoint
func (s *DBUserStore) ensureDefaultUsers() error {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil // Users already exist
	}

	// Don't create default users - require setup via /api/v1/auth/setup
	s.logger.Info("no users found - system requires initial admin setup via /api/v1/auth/setup")
	return nil
}

// GetUser retrieves a user by username (with caching)
func (s *DBUserStore) GetUser(username string) (*User, error) {
	// Check cache first
	s.mu.RLock()
	if user, exists := s.cache[username]; exists {
		s.mu.RUnlock()
		return user, nil
	}
	s.mu.RUnlock()

	// Query database
	var passwordHash string
	var rolesJSON []byte
	var active bool
	var lockedUntil sql.NullTime

	err := s.db.QueryRow(
		"SELECT password_hash, roles, active, locked_until FROM users WHERE username = $1",
		username,
	).Scan(&passwordHash, &rolesJSON, &active, &lockedUntil)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Check if account is locked
	if lockedUntil.Valid && lockedUntil.Time.After(time.Now()) {
		return nil, errors.New("account is locked")
	}

	// Parse roles
	var roles []string
	if err := json.Unmarshal(rolesJSON, &roles); err != nil {
		s.logger.Warn("failed to parse roles", zap.String("username", username), zap.Error(err))
		roles = []string{}
	}

	user := &User{
		Username:     username,
		PasswordHash: passwordHash,
		Roles:        roles,
		Active:       active,
	}

	// Cache user
	s.mu.Lock()
	s.cache[username] = user
	s.mu.Unlock()

	if !active {
		return nil, errors.New("user is inactive")
	}

	return user, nil
}

// Authenticate verifies user credentials and updates login tracking
func (s *DBUserStore) Authenticate(username, password string) (*User, error) {
	user, err := s.GetUser(username)
	if err != nil {
		// Record failed attempt
		s.recordFailedAttempt(username)
		return nil, err
	}

	// Verify password
	if err := VerifyPassword(user.PasswordHash, password); err != nil {
		// Record failed attempt
		s.recordFailedAttempt(username)
		// Clear cache to force refresh
		s.mu.Lock()
		delete(s.cache, username)
		s.mu.Unlock()
		return nil, errors.New("invalid password")
	}

	// Clear failed attempts and update last login
	s.recordSuccessfulLogin(username)

	// Clear cache to force refresh
	s.mu.Lock()
	delete(s.cache, username)
	s.mu.Unlock()

	return user, nil
}

// recordFailedAttempt records a failed login attempt
func (s *DBUserStore) recordFailedAttempt(username string) {
	_, err := s.db.Exec(`
		UPDATE users 
		SET failed_login_attempts = failed_login_attempts + 1,
		    updated_at = CURRENT_TIMESTAMP
		WHERE username = $1
	`, username)
	if err != nil {
		s.logger.Warn("failed to record failed attempt", zap.String("username", username), zap.Error(err))
		return
	}

	// Check if account should be locked (5 failed attempts)
	var attempts int
	err = s.db.QueryRow("SELECT failed_login_attempts FROM users WHERE username = $1", username).Scan(&attempts)
	if err == nil && attempts >= 5 {
		lockedUntil := time.Now().Add(15 * time.Minute)
		_, err = s.db.Exec(`
			UPDATE users 
			SET locked_until = $1, updated_at = CURRENT_TIMESTAMP
			WHERE username = $2
		`, lockedUntil, username)
		if err != nil {
			s.logger.Warn("failed to lock account", zap.String("username", username), zap.Error(err))
		}
	}
}

// recordSuccessfulLogin clears failed attempts and updates last login
func (s *DBUserStore) recordSuccessfulLogin(username string) {
	_, err := s.db.Exec(`
		UPDATE users 
		SET failed_login_attempts = 0,
		    locked_until = NULL,
		    last_login_at = CURRENT_TIMESTAMP,
		    updated_at = CURRENT_TIMESTAMP
		WHERE username = $1
	`, username)
	if err != nil {
		s.logger.Warn("failed to record successful login", zap.String("username", username), zap.Error(err))
	}
}

// AddUser adds a new user
func (s *DBUserStore) AddUser(user *User) error {
	// Check if user is being added as admin
	isAdmin := false
	for _, role := range user.Roles {
		if role == "admin" {
			isAdmin = true
			break
		}
	}

	// Enforce maximum of 2 admins
	if isAdmin {
		adminCount, err := s.GetAdminCount()
		if err != nil {
			return fmt.Errorf("failed to check admin count: %w", err)
		}

		// Check if this user already exists and is an admin
		existingUser, err := s.GetUser(user.Username)
		isExistingAdmin := false
		if err == nil && existingUser != nil {
			for _, role := range existingUser.Roles {
				if role == "admin" {
					isExistingAdmin = true
					break
				}
			}
		}

		// If adding a new admin (not updating existing admin), check limit
		if !isExistingAdmin && adminCount >= 2 {
			return fmt.Errorf("maximum of 2 admin users allowed (current: %d)", adminCount)
		}
	}

	rolesJSON, err := json.Marshal(user.Roles)
	if err != nil {
		return fmt.Errorf("failed to marshal roles: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO users (username, password_hash, roles, active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (username) DO UPDATE SET
			password_hash = EXCLUDED.password_hash,
			roles = EXCLUDED.roles,
			active = EXCLUDED.active,
			updated_at = CURRENT_TIMESTAMP
	`, user.Username, user.PasswordHash, rolesJSON, user.Active)

	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Clear cache
	s.mu.Lock()
	delete(s.cache, user.Username)
	s.mu.Unlock()

	return nil
}

// GetAdminCount returns the number of active admin users
func (s *DBUserStore) GetAdminCount() (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) 
		FROM users 
		WHERE active = true 
		AND 'admin' = ANY(SELECT jsonb_array_elements_text(roles))
	`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count admins: %w", err)
	}
	return count, nil
}

// IsSetupRequired checks if the system needs initial setup (no users exist)
func (s *DBUserStore) IsSetupRequired() (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check setup status: %w", err)
	}
	return count == 0, nil
}

// Close closes the database connection
func (s *DBUserStore) Close() error {
	return s.db.Close()
}


