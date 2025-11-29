package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sharding-system/pkg/proxy"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	// Load configuration
	config := proxy.NewProxyConfig()
	
	configPath := os.Getenv("PROXY_CONFIG_PATH")
	if configPath != "" {
		if err := config.LoadFromFile(configPath); err != nil {
			logger.Warn("failed to load config file, using defaults", zap.Error(err))
		}
	}
	
	// Override from environment
	if addr := os.Getenv("PROXY_LISTEN_ADDR"); addr != "" {
		config.ListenAddr = addr
	}
	if addr := os.Getenv("PROXY_ADMIN_ADDR"); addr != "" {
		config.AdminAddr = addr
	}
	if url := os.Getenv("SHARDING_MANAGER_URL"); url != "" {
		config.ManagerURL = url
	}

	// Create and start proxy
	proxyServer := proxy.NewShardingProxy(config, logger)
	
	if err := proxyServer.Start(); err != nil {
		logger.Fatal("failed to start proxy", zap.Error(err))
	}

	// Print connection info
	fmt.Println("")
	fmt.Println("╔═══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║           SHARDING PROXY - ZERO CODE SHARDING                     ║")
	fmt.Println("╠═══════════════════════════════════════════════════════════════════╣")
	fmt.Printf("║  Database Proxy:  %s                                       ║\n", config.ListenAddr)
	fmt.Printf("║  Admin API:       %s                                       ║\n", config.AdminAddr)
	fmt.Printf("║  Manager URL:     %s                            ║\n", config.ManagerURL)
	fmt.Println("╠═══════════════════════════════════════════════════════════════════╣")
	fmt.Println("║                                                                   ║")
	fmt.Println("║  ZERO CODE SHARDING - Just change your connection string!        ║")
	fmt.Println("║                                                                   ║")
	fmt.Println("║  Before: jdbc:postgresql://db-server:5432/mydb                    ║")
	fmt.Printf("║  After:  jdbc:postgresql://localhost%s/mydb                  ║\n", config.ListenAddr)
	fmt.Println("║                                                                   ║")
	fmt.Println("║  Configure sharding rules via Admin API:                         ║")
	fmt.Printf("║  curl http://localhost%s/api/v1/rules                        ║\n", config.AdminAddr)
	fmt.Println("║                                                                   ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════╝")
	fmt.Println("")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	if err := proxyServer.Stop(); err != nil {
		logger.Error("error during shutdown", zap.Error(err))
	}
}

