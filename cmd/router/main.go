package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sharding-system/internal/server"
	"github.com/sharding-system/pkg/catalog"
	"github.com/sharding-system/pkg/config"
	"github.com/sharding-system/pkg/router"
	"go.uber.org/zap"
)

// @title Sharding System Router API
// @version 1.0
// @description API for routing requests to shards based on shard keys
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@sharding-system.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /v1
func main() {
	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/router.json"
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

	// Initialize router
	shardRouter := router.NewRouter(
		cat,
		logger,
		cfg.Sharding.MaxConnections,
		cfg.Sharding.ConnectionTTL,
		cfg.Sharding.ReplicaPolicy,
		cfg.Pricing,
	)
	defer shardRouter.Close()

	// Create and start server
	srv, err := server.NewRouterServer(cfg, shardRouter, logger)
	if err != nil {
		logger.Fatal("failed to create server", zap.Error(err))
	}

	srv.StartAsync()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.WriteTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error", zap.Error(err))
	}
}
