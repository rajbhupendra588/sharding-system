package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sharding-system/pkg/security"
	"go.uber.org/zap"
)

// UserStore interface for abstraction (MAANG standard)
type UserStore interface {
	GetUser(username string) (*security.User, error)
	Authenticate(username, password string) (*security.User, error)
	AddUser(user *security.User) error
	GetAdminCount() (int, error)
	IsSetupRequired() (bool, error)
}

// AuthHandler handles authentication requests
type AuthHandler struct {
	authManager *security.AuthManager
	userStore   UserStore
	logger      *zap.Logger
}

// NewAuthHandler creates a new auth handler with database-backed user store
func NewAuthHandler(authManager *security.AuthManager, userStoreDSN string, logger *zap.Logger) (*AuthHandler, error) {
	var userStore UserStore
	
	if userStoreDSN != "" {
		// Use database-backed store (MAANG production standard)
		dbStore, err := security.NewDBUserStore(userStoreDSN, logger)
		if err != nil {
			logger.Warn("failed to initialize database user store, falling back to in-memory", zap.Error(err))
			userStore = security.NewUserStore()
		} else {
			userStore = dbStore
			logger.Info("using database-backed user store", zap.String("dsn", maskDSN(userStoreDSN)))
		}
	} else {
		// Fallback to in-memory store for development
		userStore = security.NewUserStore()
		logger.Warn("using in-memory user store - not recommended for production")
	}

	return &AuthHandler{
		authManager: authManager,
		userStore:   userStore,
		logger:      logger,
	}, nil
}

// maskDSN masks sensitive parts of DSN for logging
func maskDSN(dsn string) string {
	parts := strings.Split(dsn, "@")
	if len(parts) > 1 {
		return "***@" + parts[len(parts)-1]
	}
	return "***"
}


// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token    string   `json:"token"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
}

// Login handles login requests (MAANG standard: rate limiting handled by DBUserStore)
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":{"code":"BAD_REQUEST","message":"Invalid request body"}}`, http.StatusBadRequest)
		return
	}

	// Validate input (MAANG standard: input validation)
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" {
		http.Error(w, `{"error":{"code":"BAD_REQUEST","message":"Username is required"}}`, http.StatusBadRequest)
		return
	}
	if len(req.Password) == 0 {
		http.Error(w, `{"error":{"code":"BAD_REQUEST","message":"Password is required"}}`, http.StatusBadRequest)
		return
	}

	// Authenticate user (rate limiting and account lockout handled by UserStore)
	startTime := time.Now()
	user, err := h.userStore.Authenticate(req.Username, req.Password)
	authDuration := time.Since(startTime)

	if err != nil {
		// Generic error message (MAANG standard: don't reveal if user exists)
		h.logger.Warn("authentication failed",
			zap.String("username", req.Username),
			zap.Error(err),
			zap.Duration("duration_ms", authDuration),
		)
		http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"Invalid credentials"}}`, http.StatusUnauthorized)
		return
	}

	// Generate token
	token, err := h.authManager.GenerateToken(user.Username, user.Roles)
	if err != nil {
		h.logger.Error("failed to generate token", zap.Error(err))
		http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"Failed to generate token"}}`, http.StatusInternalServerError)
		return
	}

	h.logger.Info("successful login",
		zap.String("username", user.Username),
		zap.Strings("roles", user.Roles),
		zap.Duration("duration_ms", authDuration),
	)

	response := LoginResponse{
		Token:    token,
		Username: user.Username,
		Roles:    user.Roles,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SetupRequest represents an initial admin setup request
type SetupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// SetupResponse represents a setup response
type SetupResponse struct {
	Message  string `json:"message"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

// Setup handles initial admin setup (only allowed when no users exist)
func (h *AuthHandler) Setup(w http.ResponseWriter, r *http.Request) {
	// Check if setup is required
	setupRequired, err := h.userStore.IsSetupRequired()
	if err != nil {
		h.logger.Error("failed to check setup status", zap.Error(err))
		http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"Failed to check setup status"}}`, http.StatusInternalServerError)
		return
	}

	if !setupRequired {
		http.Error(w, `{"error":{"code":"BAD_REQUEST","message":"System already initialized. Setup can only be performed when no users exist."}}`, http.StatusBadRequest)
		return
	}

	var req SetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":{"code":"BAD_REQUEST","message":"Invalid request body"}}`, http.StatusBadRequest)
		return
	}

	// Validate input
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" {
		http.Error(w, `{"error":{"code":"BAD_REQUEST","message":"Username is required"}}`, http.StatusBadRequest)
		return
	}
	if len(req.Password) < 8 {
		http.Error(w, `{"error":{"code":"BAD_REQUEST","message":"Password must be at least 8 characters"}}`, http.StatusBadRequest)
		return
	}

	// Validate password strength
	if err := security.ValidatePasswordStrength(req.Password); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":{"code":"BAD_REQUEST","message":"%s"}}`, err.Error()), http.StatusBadRequest)
		return
	}

	// Hash password
	passwordHash, err := security.HashPassword(req.Password)
	if err != nil {
		h.logger.Error("failed to hash password", zap.Error(err))
		http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"Failed to process password"}}`, http.StatusInternalServerError)
		return
	}

	// Create admin user
	adminUser := &security.User{
		Username:     req.Username,
		PasswordHash: passwordHash,
		Roles:        []string{"admin"},
		Active:       true,
	}

	if err := h.userStore.AddUser(adminUser); err != nil {
		h.logger.Error("failed to create admin user", zap.Error(err))
		http.Error(w, fmt.Sprintf(`{"error":{"code":"INTERNAL_ERROR","message":"%s"}}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// Generate token
	token, err := h.authManager.GenerateToken(adminUser.Username, adminUser.Roles)
	if err != nil {
		h.logger.Error("failed to generate token", zap.Error(err))
		http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"Failed to generate token"}}`, http.StatusInternalServerError)
		return
	}

	h.logger.Info("system setup completed", zap.String("username", adminUser.Username))

	response := SetupResponse{
		Message:  "System setup completed successfully",
		Username: adminUser.Username,
		Token:    token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// SetupAuthRoutes sets up authentication routes
func SetupAuthRoutes(router *mux.Router, handler *AuthHandler) {
	router.HandleFunc("/api/v1/auth/login", handler.Login).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/auth/setup", handler.Setup).Methods("POST", "OPTIONS")
}

