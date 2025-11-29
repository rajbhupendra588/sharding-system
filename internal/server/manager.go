package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	managerSwagger "github.com/sharding-system/docs/swagger/manager"
	"github.com/sharding-system/internal/api"
	"github.com/sharding-system/internal/middleware"
	"github.com/sharding-system/pkg/autoscale"
	"github.com/sharding-system/pkg/backup"
	"github.com/sharding-system/pkg/branch"
	"github.com/sharding-system/pkg/catalog"
	"github.com/sharding-system/pkg/config"
	"github.com/sharding-system/pkg/database"
	"github.com/sharding-system/pkg/failover"
	"github.com/sharding-system/pkg/health"
	"github.com/sharding-system/pkg/manager"
	"github.com/sharding-system/pkg/models"
	"github.com/sharding-system/pkg/monitoring"
	"github.com/sharding-system/pkg/operator"
	"github.com/sharding-system/pkg/scanner"
	"github.com/sharding-system/pkg/schema"
	"github.com/sharding-system/pkg/security"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ManagerServer represents the manager HTTP server
type ManagerServer struct {
	server           *http.Server
	logger           *zap.Logger
	healthController *health.Controller
	backupService    *backup.BackupService
	failoverCtrl     *failover.FailoverController
	loadMonitor      *monitoring.LoadMonitor
	autoSplitter     *autoscale.AutoSplitter
	branchService    *branch.BranchService
	monitorCtx       context.Context
	monitorCancel    context.CancelFunc
	splitterCtx      context.Context
	splitterCancel   context.CancelFunc
}

// buildDSNFromShard builds a PostgreSQL DSN from shard connection details
func buildDSNFromShard(shard *models.Shard) string {
	// If PrimaryEndpoint is provided and is a full connection string, use it
	if shard.PrimaryEndpoint != "" {
		// Check if it's already a DSN format (starts with postgres:// or postgresql://)
		if len(shard.PrimaryEndpoint) > 10 && (shard.PrimaryEndpoint[:10] == "postgres://" || shard.PrimaryEndpoint[:14] == "postgresql://") {
			return shard.PrimaryEndpoint
		}
	}

	// Build DSN from individual connection details
	if shard.Host == "" || shard.Database == "" {
		return ""
	}

	port := shard.Port
	if port == 0 {
		port = 5432 // Default PostgreSQL port
	}

	dsn := fmt.Sprintf("host=%s port=%d dbname=%s", shard.Host, port, shard.Database)

	if shard.Username != "" {
		dsn += fmt.Sprintf(" user=%s", shard.Username)
	}

	if shard.Password != "" {
		dsn += fmt.Sprintf(" password=%s", shard.Password)
	}

	dsn += " sslmode=prefer connect_timeout=10"

	return dsn
}

// registerExistingShardsForMetrics registers all existing active shards with the Prometheus collector
func registerExistingShardsForMetrics(
	shardManager *manager.Manager,
	prometheusCollector *monitoring.PrometheusCollector,
	logger *zap.Logger,
) {
	shards, err := shardManager.ListShards()
	if err != nil {
		logger.Warn("failed to list shards for metrics registration", zap.Error(err))
		return
	}

	registeredCount := 0
	for _, shard := range shards {
		if shard.Status != "active" {
			continue
		}

		dsn := buildDSNFromShard(&shard)
		if dsn == "" {
			logger.Debug("skipping shard - no connection details available",
				zap.String("shard_id", shard.ID),
				zap.String("shard_name", shard.Name))
			continue
		}

		if err := prometheusCollector.RegisterShard(shard.ID, dsn); err != nil {
			logger.Warn("failed to register existing shard for metrics",
				zap.String("shard_id", shard.ID),
				zap.String("shard_name", shard.Name),
				zap.Error(err))
		} else {
			registeredCount++
			logger.Info("registered existing shard for metrics collection",
				zap.String("shard_id", shard.ID),
				zap.String("shard_name", shard.Name))
		}
	}

	logger.Info("completed registration of existing shards for metrics",
		zap.Int("total_shards", len(shards)),
		zap.Int("registered", registeredCount))
}

// NewManagerServer creates a new manager server instance
func NewManagerServer(
	cfg *config.Config,
	shardManager *manager.Manager,
	healthController *health.Controller,
	catalog catalog.Catalog,
	logger *zap.Logger,
) (*ManagerServer, error) {
	// Setup HTTP handlers
	managerHandler := api.NewManagerHandler(shardManager, logger)

	// Note: Stats collector will be set later after initialization

	// Initialize Prometheus collector for metrics (needed before setting up handlers)
	prometheusCollector := monitoring.NewPrometheusCollector(logger, 30*time.Second)
	prometheusCtx, prometheusCancel := context.WithCancel(context.Background())
	go prometheusCollector.Start(prometheusCtx)
	logger.Info("Prometheus collector started")
	_ = prometheusCancel // Will be used in shutdown

	// Set Prometheus collector on manager handler for shard registration
	managerHandler.SetPrometheusCollector(prometheusCollector)

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

	// Build base URL for OAuth callbacks
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		protocol := "http"
		if cfg.Security.EnableTLS {
			protocol = "https"
		}
		baseURL = fmt.Sprintf("%s://%s:%d", protocol, cfg.Server.Host, cfg.Server.Port)
		// For localhost, use localhost instead of 0.0.0.0
		if cfg.Server.Host == "0.0.0.0" {
			baseURL = fmt.Sprintf("%s://localhost:%d", protocol, cfg.Server.Port)
		}
	}

	authHandler, err := api.NewAuthHandler(authManager, userDSN, baseURL, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth handler: %w", err)
	}

	// Configure OAuth providers from environment variables
	googleClientID := os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET")
	githubClientID := os.Getenv("GITHUB_OAUTH_CLIENT_ID")
	githubClientSecret := os.Getenv("GITHUB_OAUTH_CLIENT_SECRET")
	facebookClientID := os.Getenv("FACEBOOK_OAUTH_CLIENT_ID")
	facebookClientSecret := os.Getenv("FACEBOOK_OAUTH_CLIENT_SECRET")

	authHandler.SetOAuthConfig(googleClientID, googleClientSecret, githubClientID, githubClientSecret, facebookClientID, facebookClientSecret)

	muxRouter := mux.NewRouter()

	// Setup auth routes first to avoid shadowing by protected router
	api.SetupAuthRoutes(muxRouter, authHandler)

	// Apply middleware - CORS must be first to ensure headers are set
	muxRouter.Use(middleware.CORS)
	muxRouter.Use(middleware.Recovery(logger))
	muxRouter.Use(middleware.Logging(logger))

	// Request size limit (10MB default)
	muxRouter.Use(middleware.RequestSizeLimit(middleware.DefaultMaxRequestSize))

	// Content-Type validation for POST/PUT/PATCH requests
	muxRouter.Use(middleware.ContentTypeValidation([]string{"application/json"}))

	// Enable auth middleware if RBAC is enabled in config
	var protectedRouter *mux.Router
	if cfg.Security.EnableRBAC {
		protectedRouter = muxRouter.PathPrefix("/").Subrouter()
		protectedRouter.Use(middleware.AuthMiddleware(authManager))
		logger.Info("RBAC enabled - authentication required for protected endpoints")
	} else {
		protectedRouter = muxRouter
		logger.Warn("RBAC disabled - endpoints are not protected. Enable in production!")
	}

	// Initialize multi-cluster scanner (needed for database discovery)
	clusterManager := scanner.NewClusterManager(logger)
	dbScanner := scanner.NewDatabaseScanner(logger)
	multiClusterScanner := scanner.NewMultiClusterScanner(clusterManager, dbScanner, logger)

	// Initialize database service (simplified database creation)
	dbService := database.NewDatabaseService(shardManager, logger, cfg.Server.Host, cfg.Server.Port)
	databaseHandler := api.NewDatabaseHandler(dbService, clusterManager, multiClusterScanner, logger)
	databaseHandler.SetManager(shardManager) // Set manager to access client apps

	// Initialize backup service
	backupStoragePath := os.Getenv("BACKUP_STORAGE_PATH")
	if backupStoragePath == "" {
		backupStoragePath = filepath.Join(os.TempDir(), "sharding-backups")
	}
	backupService := backup.NewBackupService(backupStoragePath, logger)
	backupService.Start()
	backupHandler := api.NewBackupHandler(backupService, logger)

	// Initialize failover controller
	failoverCtrl := failover.NewFailoverController(
		shardManager,
		healthController,
		logger,
		10*time.Second, // Check every 10 seconds
	)
	failoverCtrl.Start()
	failoverHandler := api.NewFailoverHandler(failoverCtrl, logger)

	// Initialize Phase 2 services: Load Monitoring
	loadMonitor := monitoring.NewLoadMonitor(catalog, logger, 10*time.Second)
	monitorCtx, monitorCancel := context.WithCancel(context.Background())
	go loadMonitor.Start(monitorCtx)
	logger.Info("load monitor started")

	// Initialize Phase 2 services: Hot Shard Detector
	thresholds := autoscale.DefaultThresholds()
	hotShardDetector := autoscale.NewHotShardDetector(loadMonitor, thresholds, logger)
	logger.Info("hot shard detector initialized")

	// Initialize Phase 2 services: Auto-Splitter
	autoSplitter := autoscale.NewAutoSplitter(hotShardDetector, shardManager, catalog, logger)
	splitterCtx, splitterCancel := context.WithCancel(context.Background())
	go autoSplitter.Start(splitterCtx)
	logger.Info("auto-splitter started")

	// Initialize Phase 2 services: Database Branching
	// Need operator and schema manager for branch service
	namespace := os.Getenv("KUBERNETES_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}
	op, err := operator.NewOperator(logger, namespace)
	if err != nil {
		logger.Warn("failed to initialize kubernetes operator, branch service will be limited", zap.Error(err))
		op = nil // Will need to handle nil operator
	}
	schemaManager := schema.NewManager(logger)
	dbController := database.NewController(logger, op, schemaManager, namespace)
	branchService := branch.NewBranchService(backupService, dbController, op, logger)
	logger.Info("branch service initialized")

	// Create API handlers for Phase 2
	autoscaleHandler := api.NewAutoscaleHandler(hotShardDetector, autoSplitter, logger)
	metricsHandler := api.NewMetricsHandler(loadMonitor, logger)
	branchHandler := api.NewBranchHandler(branchService, logger)

	// Initialize PostgreSQL stats collector
	postgresStatsCollector := monitoring.NewPostgresStatsCollector(logger, 30*time.Second)
	postgresStatsCtx, postgresStatsCancel := context.WithCancel(context.Background())
	go postgresStatsCollector.Start(postgresStatsCtx)
	logger.Info("PostgreSQL stats collector started")
	_ = postgresStatsCancel // Will be used in shutdown

	// Set stats collector on manager handler
	managerHandler.SetPostgresStatsCollector(postgresStatsCollector)

	// Register existing active shards with stats collector
	registerExistingShards(shardManager, postgresStatsCollector, logger)

	// Cluster scanner already initialized above, create handler
	clusterScannerHandler := api.NewClusterScannerHandler(clusterManager, multiClusterScanner, prometheusCollector, postgresStatsCollector, logger)

	// Auto-register current Kubernetes cluster and scan for databases
	go func() {
		// Wait a bit for the server to be ready
		time.Sleep(5 * time.Second)
		if err := autoRegisterAndScanCurrentCluster(clusterManager, multiClusterScanner, databaseHandler, logger); err != nil {
			logger.Warn("failed to auto-register current cluster", zap.Error(err))
		}
	}()

	// Setup routes
	api.SetupPublicRoutes(muxRouter, managerHandler)
	api.SetupProtectedRoutes(protectedRouter, managerHandler)

	// Setup Phase 1 routes (database, backup, failover)
	api.SetupDatabaseRoutes(muxRouter, databaseHandler)
	api.SetupBackupRoutes(muxRouter, backupHandler)
	api.SetupFailoverRoutes(muxRouter, failoverHandler)

	// Setup Phase 2 routes (autoscale, metrics, branches)
	autoscaleHandler.RegisterRoutes(protectedRouter)
	metricsHandler.RegisterRoutes(protectedRouter)
	branchHandler.RegisterRoutes(protectedRouter)

	// Setup multi-cluster scanner routes
	clusterScannerHandler.RegisterRoutes(protectedRouter)

	// Setup PostgreSQL stats routes
	postgresStatsHandler := api.NewPostgresStatsHandler(postgresStatsCollector, shardManager, logger)
	postgresStatsHandler.RegisterRoutes(protectedRouter)

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
	muxRouter.Handle("/metrics", prometheusCollector.Handler()).Methods("GET", "OPTIONS")

	// Register existing active shards for metrics collection on startup
	go func() {
		// Wait a bit for the server to be ready
		time.Sleep(2 * time.Second)
		registerExistingShardsForMetrics(shardManager, prometheusCollector, logger)
	}()

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
		backupService:    backupService,
		failoverCtrl:     failoverCtrl,
		loadMonitor:      loadMonitor,
		autoSplitter:     autoSplitter,
		branchService:    branchService,
		monitorCtx:       monitorCtx,
		monitorCancel:    monitorCancel,
		splitterCtx:      splitterCtx,
		splitterCancel:   splitterCancel,
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

	// Stop Phase 2 services
	if s.monitorCancel != nil {
		s.monitorCancel()
	}
	if s.splitterCancel != nil {
		s.splitterCancel()
	}
	if s.loadMonitor != nil {
		s.loadMonitor.Stop()
	}

	// Stop backup service
	if s.backupService != nil {
		s.backupService.Stop()
	}

	// Stop failover controller
	if s.failoverCtrl != nil {
		s.failoverCtrl.Stop()
	}

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

// autoRegisterAndScanCurrentCluster automatically registers the current Kubernetes cluster
// and scans it for databases
func autoRegisterAndScanCurrentCluster(
	clusterManager *scanner.ClusterManager,
	multiClusterScanner *scanner.MultiClusterScanner,
	databaseHandler *api.DatabaseHandler,
	logger *zap.Logger,
) error {
	var config *rest.Config
	var err error

	// Try to get in-cluster config first (when running inside K8s)
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig file (for local development or when manager runs outside K8s)
		logger.Info("not running in Kubernetes cluster, trying local kubeconfig")
		config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		if err != nil {
			logger.Warn("failed to get Kubernetes config from kubeconfig file, skipping auto-registration",
				zap.Error(err),
				zap.String("hint", "Make sure KUBECONFIG environment variable is set or ~/.kube/config exists"))
			return nil // Not an error, just no K8s config available
		}
		logger.Info("using local kubeconfig for cluster auto-registration")
	}

	// Test if we can connect
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Warn("failed to create Kubernetes client, skipping auto-registration", zap.Error(err))
		return nil // Don't fail startup if we can't connect
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		logger.Warn("failed to connect to Kubernetes cluster, skipping auto-registration", zap.Error(err))
		return nil // Don't fail startup if we can't connect
	}

	// Get cluster name from environment or use default
	clusterName := os.Getenv("KUBERNETES_CLUSTER_NAME")
	if clusterName == "" {
		clusterName = "local-cluster"
	}

	// Check if cluster already registered
	clusters := clusterManager.ListClusters()
	for _, cluster := range clusters {
		if cluster.Name == clusterName {
			logger.Info("cluster already registered", zap.String("name", clusterName))
			// Still scan it
			return scanCluster(cluster.ID, multiClusterScanner, databaseHandler, logger)
		}
	}

	// Determine cluster type based on how we connected
	clusterType := "onprem"
	if config.Host != "" && (config.Host != "https://kubernetes.default.svc" && config.Host != "https://kubernetes.default") {
		// If we're using a custom endpoint, it might be a cloud cluster
		// But for auto-registration, we'll default to onprem
		clusterType = "onprem"
	}

	// Register the cluster
	cluster := &models.Cluster{
		Name:      clusterName,
		Type:      clusterType,
		Provider:  "kubernetes",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata: map[string]string{
			"auto_registered": "true",
		},
	}

	if err := clusterManager.RegisterCluster(context.Background(), cluster); err != nil {
		return fmt.Errorf("failed to register cluster: %w", err)
	}

	logger.Info("auto-registered current Kubernetes cluster", zap.String("name", clusterName), zap.String("id", cluster.ID))

	// Scan the cluster for databases
	return scanCluster(cluster.ID, multiClusterScanner, databaseHandler, logger)
}

// scanCluster scans a cluster for databases and updates the database handler
func scanCluster(
	clusterID string,
	multiClusterScanner *scanner.MultiClusterScanner,
	databaseHandler *api.DatabaseHandler,
	logger *zap.Logger,
) error {
	logger.Info("scanning cluster for databases", zap.String("cluster_id", clusterID))

	req := &models.ScanRequest{
		ClusterIDs: []string{clusterID},
		DeepScan:   false, // Quick discovery scan
	}

	result, err := multiClusterScanner.ScanClusters(context.Background(), req)
	if err != nil {
		return fmt.Errorf("failed to scan cluster: %w", err)
	}

	// Update database handler with discovered databases
	databaseHandler.UpdateScanResults(result.Results)

	logger.Info("cluster scan completed",
		zap.String("cluster_id", clusterID),
		zap.Int("databases_found", result.DatabasesFound),
		zap.Int("databases_scanned", result.DatabasesScanned))

	return nil
}

// registerExistingShards registers all existing active shards with the PostgreSQL stats collector
func registerExistingShards(shardManager *manager.Manager, statsCollector *monitoring.PostgresStatsCollector, logger *zap.Logger) {
	shards, err := shardManager.ListShards()
	if err != nil {
		logger.Warn("failed to list shards for stats registration", zap.Error(err))
		return
	}

	registered := 0
	for _, shard := range shards {
		if shard.Status == "active" {
			dsn := buildDSNFromShard(&shard)
			if dsn != "" {
				if err := statsCollector.RegisterDatabase(shard.ID, dsn); err != nil {
					logger.Warn("failed to register existing shard with PostgreSQL stats collector",
						zap.String("shard_id", shard.ID),
						zap.Error(err))
				} else {
					registered++
					logger.Debug("registered existing shard for PostgreSQL stats collection",
						zap.String("shard_id", shard.ID))
				}
			}
		}
	}

	if registered > 0 {
		logger.Info("registered existing shards for PostgreSQL stats collection",
			zap.Int("count", registered))
	}
}
