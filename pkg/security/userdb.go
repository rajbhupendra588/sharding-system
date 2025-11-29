package security

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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
		password_hash VARCHAR(255),
		roles JSONB NOT NULL DEFAULT '[]'::jsonb,
		active BOOLEAN NOT NULL DEFAULT true,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_login_at TIMESTAMP,
		failed_login_attempts INTEGER NOT NULL DEFAULT 0,
		locked_until TIMESTAMP,
		oauth_provider VARCHAR(50),
		oauth_id VARCHAR(255),
		email VARCHAR(255)
	);

	CREATE INDEX IF NOT EXISTS idx_users_active ON users(active) WHERE active = true;
	CREATE INDEX IF NOT EXISTS idx_users_locked ON users(locked_until) WHERE locked_until IS NOT NULL;
	CREATE INDEX IF NOT EXISTS idx_users_oauth ON users(oauth_provider, oauth_id) WHERE oauth_provider IS NOT NULL;
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE email IS NOT NULL;
	
	-- Add OAuth columns if they don't exist (for existing databases)
	DO $$ 
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='oauth_provider') THEN
			ALTER TABLE users ADD COLUMN oauth_provider VARCHAR(50);
		END IF;
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='oauth_id') THEN
			ALTER TABLE users ADD COLUMN oauth_id VARCHAR(255);
		END IF;
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='email') THEN
			ALTER TABLE users ADD COLUMN email VARCHAR(255);
		END IF;
		IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname='idx_users_oauth') THEN
			CREATE INDEX idx_users_oauth ON users(oauth_provider, oauth_id) WHERE oauth_provider IS NOT NULL;
		END IF;
		IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname='idx_users_email') THEN
			CREATE INDEX idx_users_email ON users(email) WHERE email IS NOT NULL;
		END IF;
	END $$;
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
	var passwordHash sql.NullString
	var rolesJSON []byte
	var active bool
	var lockedUntil sql.NullTime
	var oauthProvider sql.NullString
	var oauthID sql.NullString
	var email sql.NullString

	err := s.db.QueryRow(
		"SELECT password_hash, roles, active, locked_until, oauth_provider, oauth_id, email FROM users WHERE username = $1",
		username,
	).Scan(&passwordHash, &rolesJSON, &active, &lockedUntil, &oauthProvider, &oauthID, &email)

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
		PasswordHash: passwordHash.String,
		Roles:        roles,
		Active:       active,
	}
	if oauthProvider.Valid {
		user.OAuthProvider = oauthProvider.String
	}
	if oauthID.Valid {
		user.OAuthID = oauthID.String
	}
	if email.Valid {
		user.Email = email.String
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
		INSERT INTO users (username, password_hash, roles, active, oauth_provider, oauth_id, email)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (username) DO UPDATE SET
			password_hash = COALESCE(EXCLUDED.password_hash, users.password_hash),
			roles = EXCLUDED.roles,
			active = EXCLUDED.active,
			oauth_provider = COALESCE(EXCLUDED.oauth_provider, users.oauth_provider),
			oauth_id = COALESCE(EXCLUDED.oauth_id, users.oauth_id),
			email = COALESCE(EXCLUDED.email, users.email),
			updated_at = CURRENT_TIMESTAMP
	`, user.Username, user.PasswordHash, rolesJSON, user.Active, user.OAuthProvider, user.OAuthID, user.Email)

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

// GetUserByOAuth retrieves a user by OAuth provider and ID
func (s *DBUserStore) GetUserByOAuth(provider, oauthID string) (*User, error) {
	var username string
	var passwordHash sql.NullString
	var rolesJSON []byte
	var active bool
	var email sql.NullString

	err := s.db.QueryRow(
		"SELECT username, password_hash, roles, active, email FROM users WHERE oauth_provider = $1 AND oauth_id = $2",
		provider, oauthID,
	).Scan(&username, &passwordHash, &rolesJSON, &active, &email)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if !active {
		return nil, errors.New("user is inactive")
	}

	var roles []string
	if err := json.Unmarshal(rolesJSON, &roles); err != nil {
		s.logger.Warn("failed to parse roles", zap.String("username", username), zap.Error(err))
		roles = []string{}
	}

	user := &User{
		Username:      username,
		PasswordHash:  passwordHash.String,
		Roles:         roles,
		Active:        active,
		OAuthProvider: provider,
		OAuthID:       oauthID,
	}
	if email.Valid {
		user.Email = email.String
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *DBUserStore) GetUserByEmail(email string) (*User, error) {
	var username string
	var passwordHash sql.NullString
	var rolesJSON []byte
	var active bool
	var oauthProvider sql.NullString
	var oauthID sql.NullString

	err := s.db.QueryRow(
		"SELECT username, password_hash, roles, active, oauth_provider, oauth_id FROM users WHERE email = $1",
		email,
	).Scan(&username, &passwordHash, &rolesJSON, &active, &oauthProvider, &oauthID)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if !active {
		return nil, errors.New("user is inactive")
	}

	var roles []string
	if err := json.Unmarshal(rolesJSON, &roles); err != nil {
		s.logger.Warn("failed to parse roles", zap.String("username", username), zap.Error(err))
		roles = []string{}
	}

	user := &User{
		Username:     username,
		PasswordHash: passwordHash.String,
		Roles:        roles,
		Active:       active,
		Email:        email,
	}
	if oauthProvider.Valid {
		user.OAuthProvider = oauthProvider.String
	}
	if oauthID.Valid {
		user.OAuthID = oauthID.String
	}

	return user, nil
}

// CreateOrUpdateOAuthUser creates or updates a user from OAuth info
func (s *DBUserStore) CreateOrUpdateOAuthUser(oauthInfo *OAuthUserInfo) (*User, error) {
	// Try to find existing user by OAuth provider/ID
	user, err := s.GetUserByOAuth(string(oauthInfo.Provider), oauthInfo.ID)
	if err == nil {
		// User exists, update email if needed
		if oauthInfo.Email != "" && user.Email != oauthInfo.Email {
			_, updateErr := s.db.Exec(
				"UPDATE users SET email = $1, updated_at = CURRENT_TIMESTAMP WHERE username = $2",
				oauthInfo.Email, user.Username,
			)
			if updateErr != nil {
				s.logger.Warn("failed to update email", zap.String("username", user.Username), zap.Error(updateErr))
			} else {
				user.Email = oauthInfo.Email
			}
		}
		// Update last login
		s.recordSuccessfulLogin(user.Username)
		return user, nil
	}

	// Try to find by email to link accounts
	if oauthInfo.Email != "" {
		existingUser, emailErr := s.GetUserByEmail(oauthInfo.Email)
		if emailErr == nil {
			// Link OAuth to existing account
			_, linkErr := s.db.Exec(
				"UPDATE users SET oauth_provider = $1, oauth_id = $2, updated_at = CURRENT_TIMESTAMP WHERE username = $3",
				string(oauthInfo.Provider), oauthInfo.ID, existingUser.Username,
			)
			if linkErr != nil {
				s.logger.Warn("failed to link OAuth account", zap.String("username", existingUser.Username), zap.Error(linkErr))
			} else {
				existingUser.OAuthProvider = string(oauthInfo.Provider)
				existingUser.OAuthID = oauthInfo.ID
				s.recordSuccessfulLogin(existingUser.Username)
				return existingUser, nil
			}
		}
	}

	// Create new user
	// Generate username from email or name
	username := oauthInfo.Email
	if username == "" {
		username = oauthInfo.Name
	}
	if username == "" {
		username = fmt.Sprintf("%s_%s", oauthInfo.Provider, oauthInfo.ID)
	}
	// Clean username (remove @ and special chars, make lowercase)
	username = strings.ToLower(strings.ReplaceAll(username, "@", "_"))
	username = strings.ReplaceAll(username, " ", "_")
	
	// Ensure username is unique
	baseUsername := username
	counter := 1
	for {
		_, err := s.GetUser(username)
		if err != nil {
			break // Username is available
		}
		username = fmt.Sprintf("%s_%d", baseUsername, counter)
		counter++
	}

	// Default role for new OAuth users
	roles := []string{"viewer"}
	
	newUser := &User{
		Username:      username,
		PasswordHash:  "", // No password for OAuth users
		Roles:         roles,
		Active:        true,
		OAuthProvider: string(oauthInfo.Provider),
		OAuthID:       oauthInfo.ID,
		Email:         oauthInfo.Email,
	}

	if err := s.AddUser(newUser); err != nil {
		return nil, fmt.Errorf("failed to create OAuth user: %w", err)
	}

	s.recordSuccessfulLogin(newUser.Username)
	return newUser, nil
}

// Close closes the database connection
func (s *DBUserStore) Close() error {
	return s.db.Close()
}


