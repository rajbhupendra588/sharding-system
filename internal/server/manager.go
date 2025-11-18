package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sharding-system/internal/api"
	"github.com/sharding-system/internal/middleware"
	"github.com/sharding-system/pkg/config"
	"github.com/sharding-system/pkg/health"
	"github.com/sharding-system/pkg/manager"
	"go.uber.org/zap"
)

// ManagerServer represents the manager HTTP server
type ManagerServer struct {
	server          *http.Server
	logger          *zap.Logger
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
	muxRouter := mux.NewRouter()

	// Apply middleware - CORS must be first to ensure headers are set
	muxRouter.Use(middleware.CORS)
	muxRouter.Use(middleware.Recovery(logger))
	muxRouter.Use(middleware.Logging(logger))

	// Setup routes
	api.SetupManagerRoutes(muxRouter, managerHandler)

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

