package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sharding-system/pkg/cluster"
	"github.com/sharding-system/pkg/discovery"
	"github.com/sharding-system/pkg/scanner"
	"go.uber.org/zap"
)

// ScannerHandler handles HTTP requests for database scanning
type ScannerHandler struct {
	clusterManager *cluster.ClusterManager
	scanner        *scanner.LegacyDatabaseScanner
	logger         *zap.Logger
}

// NewScannerHandler creates a new scanner handler
func NewScannerHandler(clusterManager *cluster.ClusterManager, scanner *scanner.LegacyDatabaseScanner, logger *zap.Logger) *ScannerHandler {
	return &ScannerHandler{
		clusterManager: clusterManager,
		scanner:        scanner,
		logger:         logger,
	}
}

// ScanRequest represents a request to scan a database
type ScanRequest struct {
	ClusterID       string `json:"cluster_id"`
	DatabaseName    string `json:"database_name,omitempty"`
	DatabaseHost    string `json:"database_host"`
	DatabasePort    string `json:"database_port"`
	DatabaseUser    string `json:"database_user"`
	DatabasePassword string `json:"database_password"`
	DatabaseURL     string `json:"database_url,omitempty"`
}

// ScanDatabase handles database scanning requests
// @Summary Scan a database
// @Description Scans a database to extract schema information (tables, columns, indexes, etc.)
// @Tags scanning
// @Accept json
// @Produce json
// @Param request body ScanRequest true "Database Connection Details"
// @Success 200 {object} scanner.ScanResult "Scan result"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /scan [post]
func (h *ScannerHandler) ScanDatabase(w http.ResponseWriter, r *http.Request) {
	var req ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.ClusterID == "" {
		http.Error(w, "cluster_id is required", http.StatusBadRequest)
		return
	}

	// Get cluster info
	cluster, err := h.clusterManager.GetCluster(req.ClusterID)
	if err != nil {
		http.Error(w, "cluster not found: "+err.Error(), http.StatusNotFound)
		return
	}

	// Create discovered app from request
	app := &discovery.DiscoveredApp{
		DatabaseName: req.DatabaseName,
		DatabaseHost: req.DatabaseHost,
		DatabasePort: req.DatabasePort,
		DatabaseUser: req.DatabaseUser,
		DatabaseURL:  req.DatabaseURL,
	}

	// Perform scan
	result, err := h.scanner.ScanDatabase(r.Context(), app, req.ClusterID, cluster.Name, req.DatabasePassword)
	if err != nil {
		h.logger.Error("database scan failed", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(result) // Return partial result if available
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ScanClusterDatabasesRequest represents a request to scan all databases in a cluster
type ScanClusterDatabasesRequest struct {
	ClusterID        string `json:"cluster_id"`
	DatabasePassword string `json:"database_password,omitempty"` // Optional default password
}

// ScanClusterDatabases scans all discovered databases in a cluster
// @Summary Scan all databases in a cluster
// @Description Discovers and scans all databases found in a Kubernetes cluster
// @Tags scanning
// @Accept json
// @Produce json
// @Param request body ScanClusterDatabasesRequest true "Cluster Scan Configuration"
// @Success 200 {array} scanner.ScanResult "List of scan results"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /scan/cluster [post]
func (h *ScannerHandler) ScanClusterDatabases(w http.ResponseWriter, r *http.Request) {
	var req ScanClusterDatabasesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.ClusterID == "" {
		http.Error(w, "cluster_id is required", http.StatusBadRequest)
		return
	}

	// Get cluster client
	client, err := h.clusterManager.GetClient(req.ClusterID)
	if err != nil {
		http.Error(w, "failed to get cluster client: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create discovery service for this cluster
	cluster, err := h.clusterManager.GetCluster(req.ClusterID)
	if err != nil {
		http.Error(w, "cluster not found", http.StatusNotFound)
		return
	}

	// For now, we'll use a simplified discovery that works with the client
	// In a full implementation, we'd create a multi-cluster discovery service
	discoveryService, err := discovery.NewKubernetesDiscoveryFromClient(client, h.logger, []string{})
	if err != nil {
		http.Error(w, "failed to create discovery service: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Discover applications
	apps, err := discoveryService.DiscoverApplications(r.Context())
	if err != nil {
		http.Error(w, "failed to discover applications: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Scan each discovered database
	results := make([]*scanner.ScanResult, 0)
	for _, app := range apps {
		if app.DatabaseHost == "" && app.DatabaseURL == "" {
			continue // Skip apps without database info
		}

		result, err := h.scanner.ScanDatabase(r.Context(), &app, req.ClusterID, cluster.Name, req.DatabasePassword)
		if err != nil {
			h.logger.Warn("failed to scan database",
				zap.String("app", app.Name),
				zap.String("database", app.DatabaseName),
				zap.Error(err))
			// Include failed scans in results
			results = append(results, result)
			continue
		}

		results = append(results, result)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// RegisterRoutes registers scanner routes
func (h *ScannerHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/scan", h.ScanDatabase).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/scan/cluster", h.ScanClusterDatabases).Methods("POST", "OPTIONS")
}

