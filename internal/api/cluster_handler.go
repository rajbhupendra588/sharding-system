package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sharding-system/pkg/cluster"
	"go.uber.org/zap"
)

// ClusterHandler handles HTTP requests for cluster management
type ClusterHandler struct {
	clusterManager *cluster.ClusterManager
	logger         *zap.Logger
}

// NewClusterHandler creates a new cluster handler
func NewClusterHandler(clusterManager *cluster.ClusterManager, logger *zap.Logger) *ClusterHandler {
	return &ClusterHandler{
		clusterManager: clusterManager,
		logger:         logger,
	}
}

// RegisterClusterRequest represents a request to register a cluster
type RegisterClusterRequest struct {
	Name            string            `json:"name"`
	Type            string            `json:"type"` // "cloud" or "onprem"
	Provider        string            `json:"provider,omitempty"` // "aws", "gcp", "azure", "onprem"
	KubeconfigPath  string            `json:"kubeconfig_path,omitempty"`
	Kubeconfig      string            `json:"kubeconfig,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// RegisterCluster handles cluster registration requests
// @Summary Register a new Kubernetes cluster
// @Description Registers a new Kubernetes cluster (cloud or on-premise) for database scanning
// @Tags clusters
// @Accept json
// @Produce json
// @Param request body RegisterClusterRequest true "Cluster Configuration"
// @Success 201 {object} cluster.Cluster "Cluster registered successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /clusters [post]
func (h *ClusterHandler) RegisterCluster(w http.ResponseWriter, r *http.Request) {
	var req RegisterClusterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if req.Type == "" {
		req.Type = "onprem" // Default
	}

	cluster, err := h.clusterManager.RegisterCluster(r.Context(), req.Name, req.Type, req.Provider, req.KubeconfigPath, req.Kubeconfig, req.Metadata)
	if err != nil {
		h.logger.Error("failed to register cluster", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
// @Produce json
// @Success 200 {array} cluster.Cluster "List of clusters"
// @Router /clusters [get]
func (h *ClusterHandler) ListClusters(w http.ResponseWriter, r *http.Request) {
	clusters := h.clusterManager.ListClusters()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clusters)
}

// GetCluster handles cluster retrieval requests
// @Summary Get cluster by ID
// @Description Retrieves cluster information by cluster ID
// @Tags clusters
// @Produce json
// @Param id path string true "Cluster ID"
// @Success 200 {object} cluster.Cluster "Cluster information"
// @Failure 404 {object} map[string]interface{} "Cluster not found"
// @Router /clusters/{id} [get]
func (h *ClusterHandler) GetCluster(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterID := vars["id"]

	cluster, err := h.clusterManager.GetCluster(clusterID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cluster)
}

// DeleteCluster handles cluster deletion requests
// @Summary Delete a cluster
// @Description Removes a cluster from the system
// @Tags clusters
// @Param id path string true "Cluster ID"
// @Success 204 "Cluster deleted successfully"
// @Failure 404 {object} map[string]interface{} "Cluster not found"
// @Router /clusters/{id} [delete]
func (h *ClusterHandler) DeleteCluster(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterID := vars["id"]

	if err := h.clusterManager.DeleteCluster(clusterID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TestConnection handles cluster connection test requests
// @Summary Test connection to a cluster
// @Description Tests connectivity to a Kubernetes cluster
// @Tags clusters
// @Param id path string true "Cluster ID"
// @Success 200 {object} map[string]interface{} "Connection test result"
// @Failure 400 {object} map[string]interface{} "Connection test failed"
// @Router /clusters/{id}/test [post]
func (h *ClusterHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterID := vars["id"]

	if err := h.clusterManager.TestConnection(r.Context(), clusterID); err != nil {
		h.logger.Error("connection test failed", zap.String("cluster_id", clusterID), zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "failed",
			"error":  err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"message": "Connection test passed",
	})
}

// RefreshConnections handles refresh all connections requests
// @Summary Refresh connections to all clusters
// @Description Tests connectivity to all registered clusters
// @Tags clusters
// @Success 200 {object} map[string]interface{} "Refresh results"
// @Router /clusters/refresh [post]
func (h *ClusterHandler) RefreshConnections(w http.ResponseWriter, r *http.Request) {
	errors := h.clusterManager.RefreshAllConnections(r.Context())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"refreshed": len(h.clusterManager.ListClusters()) - len(errors),
		"errors":    errors,
	})
}

// RegisterRoutes registers cluster management routes
func (h *ClusterHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/clusters", h.ListClusters).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/clusters", h.RegisterCluster).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/clusters/{id}", h.GetCluster).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/clusters/{id}", h.DeleteCluster).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/api/v1/clusters/{id}/test", h.TestConnection).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/clusters/refresh", h.RefreshConnections).Methods("POST", "OPTIONS")
}

