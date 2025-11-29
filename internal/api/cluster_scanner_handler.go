package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sharding-system/pkg/models"
	"github.com/sharding-system/pkg/monitoring"
	"github.com/sharding-system/pkg/scanner"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/clientcmd"
)

// ClusterScannerHandler handles HTTP requests for multi-cluster database scanning
type ClusterScannerHandler struct {
	clusterManager         *scanner.ClusterManager
	multiClusterScanner    *scanner.MultiClusterScanner
	prometheusCollector    *monitoring.PrometheusCollector
	postgresStatsCollector *monitoring.PostgresStatsCollector
	logger                 *zap.Logger
}

// NewClusterScannerHandler creates a new cluster scanner handler
func NewClusterScannerHandler(
	clusterManager *scanner.ClusterManager,
	multiClusterScanner *scanner.MultiClusterScanner,
	prometheusCollector *monitoring.PrometheusCollector,
	postgresStatsCollector *monitoring.PostgresStatsCollector,
	logger *zap.Logger,
) *ClusterScannerHandler {
	return &ClusterScannerHandler{
		clusterManager:         clusterManager,
		multiClusterScanner:     multiClusterScanner,
		prometheusCollector:    prometheusCollector,
		postgresStatsCollector: postgresStatsCollector,
		logger:                 logger,
	}
}

// RegisterCluster handles cluster registration requests
// @Summary Register a new Kubernetes cluster for scanning
// @Description Registers a Kubernetes cluster (cloud or on-prem) for database scanning
// @Tags clusters
// @Accept json
// @Produce json
// @Param request body models.CreateClusterRequest true "Cluster Configuration"
// @Success 201 {object} models.Cluster "Cluster registered successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /clusters [post]
func (h *ClusterScannerHandler) RegisterCluster(w http.ResponseWriter, r *http.Request) {
	var req models.CreateClusterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	cluster := &models.Cluster{
		Name:        req.Name,
		Type:        req.Type,
		Provider:    req.Provider,
		Kubeconfig:  req.Kubeconfig,
		Context:     req.Context,
		Endpoint:    req.Endpoint,
		Credentials: req.Credentials,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    req.Metadata,
	}

	if err := h.clusterManager.RegisterCluster(r.Context(), cluster); err != nil {
		h.logger.Error("failed to register cluster", zap.Error(err))
		// Check if it's a duplicate cluster error
		if strings.Contains(err.Error(), "already registered") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cluster)
}

// ListClusters handles cluster listing requests
// @Summary List all registered clusters
// @Description Returns a list of all registered Kubernetes clusters
// @Tags clusters
// @Accept json
// @Produce json
// @Success 200 {array} models.Cluster "List of clusters"
// @Router /clusters [get]
func (h *ClusterScannerHandler) ListClusters(w http.ResponseWriter, r *http.Request) {
	clusters := h.clusterManager.ListClusters()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clusters)
}

// GetCluster handles cluster retrieval requests
// @Summary Get cluster by ID
// @Description Retrieves cluster information by cluster ID
// @Tags clusters
// @Accept json
// @Produce json
// @Param id path string true "Cluster ID"
// @Success 200 {object} models.Cluster "Cluster information"
// @Failure 404 {object} map[string]interface{} "Cluster not found"
// @Router /clusters/{id} [get]
func (h *ClusterScannerHandler) GetCluster(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterID := vars["id"]

	conn, err := h.clusterManager.GetCluster(clusterID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conn.Cluster)
}

// DeleteCluster handles cluster deletion requests
// @Summary Delete a cluster
// @Description Removes a cluster from the scanning system
// @Tags clusters
// @Accept json
// @Produce json
// @Param id path string true "Cluster ID"
// @Success 204 "Cluster deleted successfully"
// @Failure 404 {object} map[string]interface{} "Cluster not found"
// @Router /clusters/{id} [delete]
func (h *ClusterScannerHandler) DeleteCluster(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterID := vars["id"]

	h.clusterManager.UnregisterCluster(clusterID)
	w.WriteHeader(http.StatusNoContent)
}

// ScanClusters handles cluster scanning requests
// @Summary Scan databases in clusters
// @Description Scans databases across specified Kubernetes clusters (or all clusters if none specified)
// @Tags clusters
// @Accept json
// @Produce json
// @Param request body models.ScanRequest true "Scan Configuration"
// @Success 200 {object} models.ScanResult "Scan results"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /clusters/scan [post]
func (h *ClusterScannerHandler) ScanClusters(w http.ResponseWriter, r *http.Request) {
	var req models.ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.multiClusterScanner.ScanClusters(r.Context(), &req)
	if err != nil {
		h.logger.Error("failed to scan clusters", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Register discovered databases for metrics collection
	h.registerDatabasesForMetrics(result.Results)

	// Store scan results (if database handler is available, update it)
	// This will be done through a shared storage mechanism in production

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetScanResults handles scan results retrieval
// @Summary Get scan results
// @Description Returns scan results for databases across clusters
// @Tags clusters
// @Accept json
// @Produce json
// @Param cluster_id query string false "Filter by cluster ID"
// @Success 200 {array} models.ScannedDatabase "List of scanned databases"
// @Router /clusters/scan/results [get]
func (h *ClusterScannerHandler) GetScanResults(w http.ResponseWriter, r *http.Request) {
	// For now, return empty - in production, you'd store scan results
	// and retrieve them from storage
	results := []models.ScannedDatabase{}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// registerDatabasesForMetrics registers discovered databases for metrics collection
func (h *ClusterScannerHandler) registerDatabasesForMetrics(databases []models.ScannedDatabase) {
	for _, db := range databases {
		if db.Host == "" || db.Database == "" {
			continue
		}

		// Build DSN for PostgreSQL
		dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s sslmode=prefer connect_timeout=10",
			db.Host, db.Port, db.Database, db.Username)

		// Register with Prometheus collector
		if h.prometheusCollector != nil {
			shardID := fmt.Sprintf("%s-%s-%s", db.ClusterID, db.Namespace, db.AppName)
			if err := h.prometheusCollector.RegisterShard(shardID, dsn); err != nil {
				h.logger.Warn("failed to register database with Prometheus collector",
					zap.String("database", db.DatabaseName),
					zap.Error(err))
			} else {
				h.logger.Info("registered database for Prometheus metrics",
					zap.String("database", db.DatabaseName),
					zap.String("shard_id", shardID))
			}
		}

		// Register with PostgreSQL stats collector
		if h.postgresStatsCollector != nil {
			databaseID := fmt.Sprintf("%s-%s-%s", db.ClusterID, db.Namespace, db.AppName)
			if err := h.postgresStatsCollector.RegisterDatabase(databaseID, dsn); err != nil {
				h.logger.Warn("failed to register database with PostgreSQL stats collector",
					zap.String("database", db.DatabaseName),
					zap.Error(err))
			} else {
				h.logger.Info("registered database for PostgreSQL stats collection",
					zap.String("database", db.DatabaseName),
					zap.String("database_id", databaseID))
			}
		}
	}
}

// DiscoverAvailableClusters discovers available Kubernetes clusters from kubeconfig
// @Summary Discover available clusters from kubeconfig
// @Description Lists all available Kubernetes contexts/clusters from kubeconfig file
// @Tags clusters
// @Accept json
// @Produce json
// @Success 200 {array} map[string]interface{} "List of available clusters"
// @Router /clusters/discover [get]
func (h *ClusterScannerHandler) DiscoverAvailableClusters(w http.ResponseWriter, r *http.Request) {
	// Get kubeconfig path
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		kubeconfigPath = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}

	// Check if kubeconfig exists
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{})
		return
	}

	// Load kubeconfig
	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		h.logger.Warn("failed to load kubeconfig", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{})
		return
	}

	// Get registered cluster names to check which ones are already registered
	registeredClusters := h.clusterManager.ListClusters()
	registeredNames := make(map[string]bool)
	registeredContexts := make(map[string]bool)
	for _, cluster := range registeredClusters {
		registeredNames[cluster.Name] = true
		// Also check by context name if available
		if cluster.Context != "" {
			registeredContexts[cluster.Context] = true
		}
	}

	// Extract contexts and clusters
	availableClusters := make([]map[string]interface{}, 0)
	currentContext := config.CurrentContext

	for contextName, context := range config.Contexts {
		if context == nil {
			continue
		}

		clusterName := context.Cluster
		clusterInfo := config.Clusters[clusterName]
		if clusterInfo == nil {
			continue
		}

		// Determine cluster type based on server URL
		clusterType := "onprem"
		provider := "kubernetes"
		serverURL := clusterInfo.Server
		if serverURL != "" {
			serverURLLower := strings.ToLower(serverURL)
			if strings.Contains(serverURLLower, "eks") || strings.Contains(serverURLLower, "amazonaws.com") {
				clusterType = "cloud"
				provider = "aws"
			} else if strings.Contains(serverURLLower, "gke") || strings.Contains(serverURLLower, "googleapis.com") {
				clusterType = "cloud"
				provider = "gcp"
			} else if strings.Contains(serverURLLower, "aks") || strings.Contains(serverURLLower, "azure.com") {
				clusterType = "cloud"
				provider = "azure"
			}
		}

		// Check if already registered by name, context name, or cluster name
		isRegistered := registeredNames[contextName] || registeredNames[clusterName] || registeredContexts[contextName]

		availableClusters = append(availableClusters, map[string]interface{}{
			"context_name":  contextName,
			"cluster_name":  clusterName,
			"server_url":    serverURL,
			"type":          clusterType,
			"provider":      provider,
			"is_current":    contextName == currentContext,
			"is_registered": isRegistered,
			"namespace":     context.Namespace,
			"user":          context.AuthInfo,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(availableClusters)
}


// RegisterRoutes registers cluster scanner routes
// Note: Specific routes must be registered before parameterized routes
func (h *ClusterScannerHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/clusters", h.ListClusters).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/clusters", h.RegisterCluster).Methods("POST", "OPTIONS")
	// Specific routes must come before parameterized routes
	router.HandleFunc("/api/v1/clusters/discover", h.DiscoverAvailableClusters).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/clusters/scan", h.ScanClusters).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/clusters/scan/results", h.GetScanResults).Methods("GET", "OPTIONS")
	// Parameterized routes come last
	router.HandleFunc("/api/v1/clusters/{id}", h.GetCluster).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/clusters/{id}", h.DeleteCluster).Methods("DELETE", "OPTIONS")
}

