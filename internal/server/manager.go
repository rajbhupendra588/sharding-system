package server

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	managerSwagger "github.com/sharding-system/docs/swagger/manager"
	"github.com/sharding-system/internal/api"
	"github.com/sharding-system/internal/middleware"
	"github.com/sharding-system/pkg/config"
	"github.com/sharding-system/pkg/health"
	"github.com/sharding-system/pkg/manager"
	"github.com/sharding-system/pkg/security"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

// ManagerServer represents the manager HTTP server
type ManagerServer struct {
	server           *http.Server
	logger           *zap.Logger
	healthController *health.Controller
}

// NewManagerServer creates a new manager server instance
func NewManagerServer(
	cfg *config.Config,
	shardManager *manager.Manager,
	healthController *health.Controller,
	logger *zap.Logger,
) (*ManagerServer, error) {
	// Setup HTTP handlers
	managerHandler := api.NewManagerHandler(shardManager, logger)

	// Initialize auth manager
	// JWT_SECRET is required if RBAC is enabled, optional for development
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		if cfg.Security.EnableRBAC {
			logger.Fatal("JWT_SECRET environment variable is required when RBAC is enabled")
		}
		// Development mode - use a default (not secure for production!)
		jwtSecret = "development-secret-not-for-production-use-min-32-chars"
		logger.Warn("JWT_SECRET not set - using development secret. Set JWT_SECRET in production!")
	}
	if len(jwtSecret) < 32 {
		logger.Fatal("JWT_SECRET must be at least 32 characters for security")
	}
	authManager := security.NewAuthManager(jwtSecret)

	// Get user database DSN from config or environment
	userDSN := cfg.Security.UserDatabaseDSN
	if userDSN == "" {
		userDSN = os.Getenv("USER_DATABASE_DSN")
	}

	authHandler, err := api.NewAuthHandler(authManager, userDSN, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth handler: %w", err)
	}

	muxRouter := mux.NewRouter()

	// Apply middleware - CORS must be first to ensure headers are set
	muxRouter.Use(middleware.CORS)
	muxRouter.Use(middleware.Recovery(logger))
	muxRouter.Use(middleware.Logging(logger))

	// Request size limit (10MB default)
	muxRouter.Use(middleware.RequestSizeLimit(middleware.DefaultMaxRequestSize))

	// Content-Type validation for POST/PUT/PATCH requests
	muxRouter.Use(middleware.ContentTypeValidation([]string{"application/json"}))

	// Enable auth middleware if RBAC is enabled in config
	if cfg.Security.EnableRBAC {
		muxRouter.Use(middleware.AuthMiddleware(authManager))
		logger.Info("RBAC enabled - authentication required for protected endpoints")
	} else {
		logger.Warn("RBAC disabled - endpoints are not protected. Enable in production!")
	}

	// Setup routes
	api.SetupManagerRoutes(muxRouter, managerHandler)
	api.SetupAuthRoutes(muxRouter, authHandler)

	// Setup Swagger documentation
	muxRouter.HandleFunc("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		doc := managerSwagger.SwaggerInfomanager.ReadDoc()
		w.Write([]byte(doc))
	}).Methods("GET", "OPTIONS")

	muxRouter.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8081/swagger/doc.json"), // The url pointing to API definition
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	)).Methods("GET", "OPTIONS")

	// Setup metrics endpoint with CORS support
	// Prometheus metrics handler wrapped to ensure CORS headers are set
	muxRouter.Handle("/metrics", promhttp.Handler()).Methods("GET", "OPTIONS")

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      muxRouter,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &ManagerServer{
		server:           server,
		logger:           logger,
		healthController: healthController,
	}, nil
}

// Start starts the HTTP server
func (s *ManagerServer) Start() error {
	s.logger.Info("starting manager server", zap.String("address", s.server.Addr))
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed: %w", err)
	}
	return nil
}

// Shutdown gracefully shuts down the server
func (s *ManagerServer) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down manager server")
	return s.server.Shutdown(ctx)
}

// StartAsync starts the server in a goroutine
func (s *ManagerServer) StartAsync() {
	go func() {
		if err := s.Start(); err != nil {
			s.logger.Fatal("manager server failed", zap.Error(err))
		}
	}()
}

// Handler returns the HTTP handler for testing purposes
func (s *ManagerServer) Handler() http.Handler {
	return s.server.Handler
}
