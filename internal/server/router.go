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
	"github.com/sharding-system/pkg/router"
	"go.uber.org/zap"
)

// RouterServer represents the router HTTP server
type RouterServer struct {
	server *http.Server
	logger *zap.Logger
}

// NewRouterServer creates a new router server instance
func NewRouterServer(
	cfg *config.Config,
	shardRouter *router.Router,
	logger *zap.Logger,
) (*RouterServer, error) {
	// Setup HTTP handlers
	routerHandler := api.NewRouterHandler(shardRouter, logger)
	muxRouter := mux.NewRouter()

	// Apply middleware - CORS must be first to ensure headers are set
	muxRouter.Use(middleware.CORS)
	muxRouter.Use(middleware.Recovery(logger))
	muxRouter.Use(middleware.Logging(logger))

	// Setup routes
	api.SetupRouterRoutes(muxRouter, routerHandler)

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

	return &RouterServer{
		server: server,
		logger: logger,
	}, nil
}

// Start starts the HTTP server
func (s *RouterServer) Start() error {
	s.logger.Info("starting router server", zap.String("address", s.server.Addr))
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed: %w", err)
	}
	return nil
}

// Shutdown gracefully shuts down the server
func (s *RouterServer) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down router server")
	return s.server.Shutdown(ctx)
}

// StartAsync starts the server in a goroutine
func (s *RouterServer) StartAsync() {
	go func() {
		if err := s.Start(); err != nil {
			s.logger.Fatal("router server failed", zap.Error(err))
		}
	}()
}

