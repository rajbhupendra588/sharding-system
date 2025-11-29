package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
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
	// OAuth methods
	GetUserByOAuth(provider, oauthID string) (*security.User, error)
	GetUserByEmail(email string) (*security.User, error)
	CreateOrUpdateOAuthUser(oauthInfo *security.OAuthUserInfo) (*security.User, error)
}

// AuthHandler handles authentication requests
type AuthHandler struct {
	authManager *security.AuthManager
	userStore   UserStore
	oauthConfig *security.OAuthConfig
	logger      *zap.Logger
	frontendURL string // Frontend URL for OAuth redirects
}

// NewAuthHandler creates a new auth handler with database-backed user store
func NewAuthHandler(authManager *security.AuthManager, userStoreDSN string, baseURL string, logger *zap.Logger) (*AuthHandler, error) {
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

	oauthConfig := security.NewOAuthConfig(baseURL, logger)

	// Determine frontend URL (default to localhost:3000 for development)
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	return &AuthHandler{
		authManager: authManager,
		userStore:   userStore,
		oauthConfig: oauthConfig,
		logger:      logger,
		frontendURL: frontendURL,
	}, nil
}

// SetOAuthConfig sets OAuth configuration
func (h *AuthHandler) SetOAuthConfig(googleClientID, googleClientSecret, githubClientID, githubClientSecret, facebookClientID, facebookClientSecret string) {
	configuredCount := 0
	if googleClientID != "" && googleClientSecret != "" {
		h.oauthConfig.SetGoogleConfig(googleClientID, googleClientSecret)
		h.logger.Info("Google OAuth configured", zap.String("client_id", maskSecret(googleClientID)))
		configuredCount++
	}
	if githubClientID != "" && githubClientSecret != "" {
		h.oauthConfig.SetGitHubConfig(githubClientID, githubClientSecret)
		h.logger.Info("GitHub OAuth configured", zap.String("client_id", maskSecret(githubClientID)))
		configuredCount++
	}
	if facebookClientID != "" && facebookClientSecret != "" {
		h.oauthConfig.SetFacebookConfig(facebookClientID, facebookClientSecret)
		h.logger.Info("Facebook OAuth configured", zap.String("client_id", maskSecret(facebookClientID)))
		configuredCount++
	}

	if configuredCount == 0 {
		h.logger.Info("No OAuth providers configured - social login disabled. Set GOOGLE_OAUTH_CLIENT_ID, GITHUB_OAUTH_CLIENT_ID, or FACEBOOK_OAUTH_CLIENT_ID to enable.")
	} else {
		h.logger.Info("OAuth social login enabled", zap.Int("providers", configuredCount))
	}
}

// maskSecret masks sensitive parts of client ID for logging
func maskSecret(secret string) string {
	if len(secret) <= 8 {
		return "***"
	}
	return secret[:4] + "..." + secret[len(secret)-4:]
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

// writeJSONError writes a JSON error response
func (h *AuthHandler) writeJSONError(w http.ResponseWriter, code int, errorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    errorCode,
			"message": message,
		},
	})
}

// Login handles login requests (MAANG standard: rate limiting handled by DBUserStore)
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSONError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	// Validate input (MAANG standard: input validation)
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" {
		h.writeJSONError(w, http.StatusBadRequest, "BAD_REQUEST", "Username is required")
		return
	}
	if len(req.Password) == 0 {
		h.writeJSONError(w, http.StatusBadRequest, "BAD_REQUEST", "Password is required")
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
		h.writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid credentials")
		return
	}

	// Generate token
	token, err := h.authManager.GenerateToken(user.Username, user.Roles)
	if err != nil {
		h.logger.Error("failed to generate token", zap.Error(err))
		h.writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate token")
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
		h.writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check setup status")
		return
	}

	if !setupRequired {
		h.writeJSONError(w, http.StatusBadRequest, "BAD_REQUEST", "System already initialized. Setup can only be performed when no users exist.")
		return
	}

	var req SetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSONError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	// Validate input
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" {
		h.writeJSONError(w, http.StatusBadRequest, "BAD_REQUEST", "Username is required")
		return
	}
	if len(req.Password) < 8 {
		h.writeJSONError(w, http.StatusBadRequest, "BAD_REQUEST", "Password must be at least 8 characters")
		return
	}

	// Validate password strength
	if err := security.ValidatePasswordStrength(req.Password); err != nil {
		h.writeJSONError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	// Hash password
	passwordHash, err := security.HashPassword(req.Password)
	if err != nil {
		h.logger.Error("failed to hash password", zap.Error(err))
		h.writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process password")
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
		h.writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	// Generate token
	token, err := h.authManager.GenerateToken(adminUser.Username, adminUser.Roles)
	if err != nil {
		h.logger.Error("failed to generate token", zap.Error(err))
		h.writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate token")
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

// OAuthInitiate initiates OAuth flow by redirecting to provider
func (h *AuthHandler) OAuthInitiate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	providerStr := vars["provider"]
	provider := security.OAuthProvider(providerStr)

	if !h.oauthConfig.IsProviderEnabled(provider) {
		h.writeJSONError(w, http.StatusBadRequest, "BAD_REQUEST", "OAuth provider is not configured")
		return
	}

	// Generate state token
	state, err := security.GenerateState()
	if err != nil {
		h.logger.Error("failed to generate state", zap.Error(err))
		h.writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to initiate OAuth")
		return
	}

	// Store state in session/cookie (for production, use secure session storage)
	// For now, we'll include it in the redirect URL as a query param
	authURL, err := h.oauthConfig.GetAuthURL(provider, state)
	if err != nil {
		h.logger.Error("failed to get auth URL", zap.Error(err), zap.String("provider", providerStr))
		h.writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to initiate OAuth")
		return
	}

	// Set state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600, // 10 minutes
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to OAuth provider
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// OAuthCallback handles OAuth callback from provider
func (h *AuthHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	providerStr := vars["provider"]
	provider := security.OAuthProvider(providerStr)

	// Verify state
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		h.writeJSONError(w, http.StatusBadRequest, "BAD_REQUEST", "Missing state cookie")
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" || state != stateCookie.Value {
		h.writeJSONError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid state parameter")
		return
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		h.writeJSONError(w, http.StatusBadRequest, "BAD_REQUEST", "Missing authorization code")
		return
	}

	// Exchange code for token
	token, err := h.oauthConfig.ExchangeCode(provider, code)
	if err != nil {
		h.logger.Error("failed to exchange code", zap.Error(err), zap.String("provider", providerStr))
		h.writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to authenticate with OAuth provider")
		return
	}

	// Get user info from provider
	oauthInfo, err := h.oauthConfig.GetUserInfo(provider, token)
	if err != nil {
		h.logger.Error("failed to get user info", zap.Error(err), zap.String("provider", providerStr))
		h.writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve user information")
		return
	}

	// Get or create user
	// Check if userStore implements CreateOrUpdateOAuthUser method
	// Both DBUserStore and in-memory UserStore now support OAuth
	user, err := h.userStore.CreateOrUpdateOAuthUser(oauthInfo)
	if err != nil {
		h.logger.Error("failed to create/update OAuth user", zap.Error(err))
		h.writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create user account")
		return
	}

	// Generate JWT token
	jwtToken, err := h.authManager.GenerateToken(user.Username, user.Roles)
	if err != nil {
		h.logger.Error("failed to generate token", zap.Error(err))
		h.writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate token")
		return
	}

	h.logger.Info("OAuth login successful",
		zap.String("username", user.Username),
		zap.String("provider", providerStr),
		zap.Strings("roles", user.Roles),
	)

	// Redirect to frontend with token
	// In production, you might want to use a more secure method
	redirectURI := r.URL.Query().Get("redirect_uri")
	
	// Decode the redirect_uri if it's URL encoded
	if redirectURI != "" {
		if decoded, err := url.QueryUnescape(redirectURI); err == nil {
			redirectURI = decoded
		}
	}
	
	// Use provided redirect_uri, or default to frontend URL + /login
	if redirectURI == "" {
		redirectURI = fmt.Sprintf("%s/login", h.frontendURL)
	}

	// Redirect with token in URL fragment (more secure than query param)
	// Fragment is not sent to server, so it's more secure
	redirectURL := fmt.Sprintf("%s#token=%s&username=%s", redirectURI, jwtToken, user.Username)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// GetOAuthProviders returns list of enabled OAuth providers
func (h *AuthHandler) GetOAuthProviders(w http.ResponseWriter, r *http.Request) {
	providers := h.oauthConfig.GetEnabledProviders()
	providerStrs := make([]string, len(providers))
	for i, p := range providers {
		providerStrs[i] = string(p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"providers": providerStrs,
	})
}

// SetupAuthRoutes sets up authentication routes
func SetupAuthRoutes(router *mux.Router, handler *AuthHandler) {
	router.HandleFunc("/api/v1/auth/login", handler.Login).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/auth/setup", handler.Setup).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/auth/oauth/providers", handler.GetOAuthProviders).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/auth/oauth/{provider}", handler.OAuthInitiate).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/auth/oauth/{provider}/callback", handler.OAuthCallback).Methods("GET", "OPTIONS")
}
