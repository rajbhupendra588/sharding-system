package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sharding-system/internal/server"
	"github.com/sharding-system/pkg/catalog"
	"github.com/sharding-system/pkg/config"
	"github.com/sharding-system/pkg/health"
	"github.com/sharding-system/pkg/manager"
	"github.com/sharding-system/pkg/resharder"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/manager.json"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	// Initialize catalog
	cat, err := catalog.NewEtcdCatalog(cfg.Metadata.Endpoints, logger)
	if err != nil {
		logger.Fatal("failed to initialize catalog", zap.Error(err))
	}

	// Initialize resharder
	resharderInstance := resharder.NewResharder(cat, logger)

	// Initialize manager
	shardManager := manager.NewManager(cat, logger, resharderInstance)

	// Initialize health controller
	healthController := health.NewController(
		cat,
		logger,
		30*time.Second,  // check interval
		5*time.Second,   // replication lag threshold
	)

	// Start health monitoring
	healthCtx, healthCancel := context.WithCancel(context.Background())
	defer healthCancel()
	go healthController.Start(healthCtx)

	// Create and start server
	srv, err := server.NewManagerServer(cfg, shardManager, healthController, logger)
	if err != nil {
		logger.Fatal("failed to create server", zap.Error(err))
	}

	srv.StartAsync()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.WriteTimeout)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", zap.Error(err))
	}
}

