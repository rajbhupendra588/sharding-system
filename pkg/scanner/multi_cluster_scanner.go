package scanner

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sharding-system/pkg/discovery"
	"github.com/sharding-system/pkg/models"
	"go.uber.org/zap"
)

// MultiClusterScanner scans databases across multiple Kubernetes clusters
type MultiClusterScanner struct {
	clusterManager *ClusterManager
	dbScanner      *DatabaseScanner
	k8sDiscovery   map[string]*discovery.KubernetesDiscovery
	logger         *zap.Logger
	mu             sync.RWMutex
}

// NewMultiClusterScanner creates a new multi-cluster scanner
func NewMultiClusterScanner(clusterManager *ClusterManager, dbScanner *DatabaseScanner, logger *zap.Logger) *MultiClusterScanner {
	return &MultiClusterScanner{
		clusterManager: clusterManager,
		dbScanner:      dbScanner,
		k8sDiscovery:   make(map[string]*discovery.KubernetesDiscovery),
		logger:         logger,
	}
}

// ScanClusters scans databases in the specified clusters
func (mcs *MultiClusterScanner) ScanClusters(ctx context.Context, request *models.ScanRequest) (*models.ScanResult, error) {
	scanResult := &models.ScanResult{
		ID:        generateClusterID(),
		Status:    "running",
		StartedAt: time.Now(),
		Results:   make([]models.ScannedDatabase, 0),
	}

	// Get clusters to scan
	clusters := mcs.getClustersToScan(request.ClusterIDs)

	if len(clusters) == 0 {
		return nil, fmt.Errorf("no clusters found to scan")
	}

	// Scan each cluster concurrently
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, cluster := range clusters {
		wg.Add(1)
		go func(clusterID string) {
			defer wg.Done()

			databases, err := mcs.scanCluster(ctx, clusterID, request.DeepScan)
			if err != nil {
				mcs.logger.Error("failed to scan cluster", zap.String("cluster_id", clusterID), zap.Error(err))
				mu.Lock()
				scanResult.DatabasesFailed += len(databases)
				mu.Unlock()
				return
			}

			mu.Lock()
			scanResult.DatabasesFound += len(databases)
			scanResult.DatabasesScanned += len(databases)
			scanResult.Results = append(scanResult.Results, databases...)
			mu.Unlock()
		}(cluster.ID)
	}

	wg.Wait()

	scanResult.Status = "completed"
	now := time.Now()
	scanResult.CompletedAt = &now

	// Update cluster last scan time
	for _, cluster := range clusters {
		conn, err := mcs.clusterManager.GetCluster(cluster.ID)
		if err == nil {
			now := time.Now()
			conn.Cluster.LastScan = &now
		}
	}

	return scanResult, nil
}

// scanCluster scans databases in a single cluster
func (mcs *MultiClusterScanner) scanCluster(ctx context.Context, clusterID string, deepScan bool) ([]models.ScannedDatabase, error) {
	conn, err := mcs.clusterManager.GetCluster(clusterID)
	if err != nil {
		return nil, err
	}

	// Get or create discovery for this cluster
	discovery, err := mcs.getOrCreateDiscovery(clusterID, conn)
	if err != nil {
		return nil, err
	}

	// Discover applications
	discoveredApps, err := discovery.DiscoverApplications(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover applications: %w", err)
	}

	// Convert to scanned databases
	databases := make([]models.ScannedDatabase, 0)
	for _, app := range discoveredApps {
		// Skip apps without database info - they should not appear in scanning
		if app.DatabaseHost == "" && app.DatabaseURL == "" {
			mcs.logger.Debug("skipping application without database information",
				zap.String("app_name", app.Name),
				zap.String("namespace", app.Namespace))
			continue
		}

		// Validate that we have sufficient database information
		hasDBInfo := app.DatabaseHost != "" && app.DatabaseName != ""
		if !hasDBInfo && app.DatabaseURL == "" {
			mcs.logger.Debug("skipping application with insufficient database information",
				zap.String("app_name", app.Name),
				zap.String("namespace", app.Namespace))
			continue
		}

		db := mcs.convertToScannedDatabase(clusterID, conn.Cluster.Name, &app)

		// Perform deep scan if requested
		if deepScan && app.DatabaseHost != "" {
			scanResults, err := mcs.performDeepScan(ctx, &app)
			if err != nil {
				db.Status = "error"
				db.ScanError = err.Error()
				mcs.logger.Warn("deep scan failed", zap.String("database", db.DatabaseName), zap.Error(err))
			} else {
				db.ScanResults = scanResults
				db.Status = "scanned"
				now := time.Now()
				db.LastScannedAt = &now
			}
		}

		databases = append(databases, *db)
	}

	return databases, nil
}

// getOrCreateDiscovery gets or creates a Kubernetes discovery instance for a cluster
func (mcs *MultiClusterScanner) getOrCreateDiscovery(clusterID string, conn *ClusterConnection) (*discovery.KubernetesDiscovery, error) {
	mcs.mu.RLock()
	disc, ok := mcs.k8sDiscovery[clusterID]
	mcs.mu.RUnlock()

	if ok {
		return disc, nil
	}

	// Create new discovery instance using the cluster's k8s client
	disc, err := discovery.NewKubernetesDiscoveryFromClient(conn.Client, mcs.logger, []string{})
	if err != nil {
		return nil, err
	}

	mcs.mu.Lock()
	mcs.k8sDiscovery[clusterID] = disc
	mcs.mu.Unlock()

	return disc, nil
}

// convertToScannedDatabase converts a discovered app to a scanned database
func (mcs *MultiClusterScanner) convertToScannedDatabase(clusterID, clusterName string, app *discovery.DiscoveredApp) *models.ScannedDatabase {
	db := &models.ScannedDatabase{
		ID:           generateClusterID(),
		ClusterID:    clusterID,
		ClusterName:  clusterName,
		Namespace:    app.Namespace,
		AppName:      app.Name,
		AppType:      app.Type,
		DatabaseName: app.DatabaseName,
		DatabaseType: "postgresql", // Default, could be detected from URL
		Status:       "discovered",
		DiscoveredAt: time.Now(),
		Labels:       app.Labels,
		Annotations:  app.Annotations,
	}

	// Parse database connection info
	if app.DatabaseURL != "" {
		mcs.parseDatabaseURL(app.DatabaseURL, db)
	} else {
		db.Host = app.DatabaseHost
		if app.DatabasePort != "" {
			fmt.Sscanf(app.DatabasePort, "%d", &db.Port)
		} else {
			db.Port = 5432 // Default PostgreSQL port
		}
		db.Database = app.DatabaseName
		db.Username = app.DatabaseUser
	}

	return db
}

// parseDatabaseURL parses a database URL and populates the scanned database
func (mcs *MultiClusterScanner) parseDatabaseURL(url string, db *models.ScannedDatabase) {
	// Simple parsing - in production, use proper URL parsing
	// postgres://user:pass@host:port/dbname
	if strings.HasPrefix(url, "postgres://") || strings.HasPrefix(url, "postgresql://") {
		db.DatabaseType = "postgresql"
		// Extract components (simplified)
		parts := strings.Split(url, "@")
		if len(parts) == 2 {
			hostPort := strings.Split(parts[1], "/")
			if len(hostPort) >= 2 {
				db.Database = hostPort[1]
				hostPortParts := strings.Split(hostPort[0], ":")
				if len(hostPortParts) >= 2 {
					db.Host = hostPortParts[0]
					fmt.Sscanf(hostPortParts[1], "%d", &db.Port)
				}
			}
		}
	}
}

// performDeepScan performs a deep scan of a database
func (mcs *MultiClusterScanner) performDeepScan(ctx context.Context, app *discovery.DiscoveredApp) (*models.DatabaseScanResults, error) {
	// Convert to scanned database first
	db := &models.ScannedDatabase{
		Host:     app.DatabaseHost,
		Database: app.DatabaseName,
		Username: app.DatabaseUser,
	}

	if app.DatabasePort != "" {
		fmt.Sscanf(app.DatabasePort, "%d", &db.Port)
	} else {
		db.Port = 5432
	}

	// Try to get password from secrets (in production, use proper secret management)
	password := "" // Would need to fetch from K8s secrets

	return mcs.dbScanner.ScanDatabase(ctx, db, password)
}

// getClustersToScan returns clusters to scan based on the request
func (mcs *MultiClusterScanner) getClustersToScan(clusterIDs []string) []*models.Cluster {
	allClusters := mcs.clusterManager.ListClusters()

	if len(clusterIDs) == 0 {
		return allClusters
	}

	// Filter by requested IDs
	clusterMap := make(map[string]*models.Cluster)
	for _, cluster := range allClusters {
		clusterMap[cluster.ID] = cluster
	}

	result := make([]*models.Cluster, 0)
	for _, id := range clusterIDs {
		if cluster, ok := clusterMap[id]; ok {
			result = append(result, cluster)
		}
	}

	return result
}

