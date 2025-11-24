package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sharding-system/pkg/discovery"
	"github.com/sharding-system/pkg/manager"
	"github.com/sharding-system/pkg/models"
	"github.com/sharding-system/pkg/pricing"
	"go.uber.org/zap"
)

// Alias for discovery.DiscoveredApp to use in API docs
type DiscoveredApp = discovery.DiscoveredApp

// @title Sharding System Manager API
// @version 1.0
// @description API for managing shards and cluster state
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@sharding-system.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8081
// @BasePath /api/v1

// ManagerHandler handles HTTP requests for the manager
type ManagerHandler struct {
	manager   *manager.Manager
	logger    *zap.Logger
	discovery discovery.DiscoveryService
}

// NewManagerHandler creates a new manager handler
func NewManagerHandler(m *manager.Manager, logger *zap.Logger, discovery discovery.DiscoveryService) *ManagerHandler {
	return &ManagerHandler{
		manager:   m,
		logger:    logger,
		discovery: discovery,
	}
}

// CreateShard handles shard creation requests
// @Summary Create a new shard for a client application
// @Description Creates a new database shard with the specified configuration. Shards must belong to a client application.
// @Tags shards
// @Accept json
// @Produce json
// @Param request body models.CreateShardRequest true "Shard Configuration (must include client_app_id)"
// @Success 201 {object} models.Shard "Shard created successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /shards [post]
func (h *ManagerHandler) CreateShard(w http.ResponseWriter, r *http.Request) {
	var req models.CreateShardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.ClientAppID == "" {
		http.Error(w, "client_app_id is required - shards must belong to a client application", http.StatusBadRequest)
		return
	}

	shard, err := h.manager.CreateShard(r.Context(), &req)
	if err != nil {
		h.logger.Error("failed to create shard", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(shard)
}

// GetShard handles shard retrieval requests
// @Summary Get shard by ID
// @Description Retrieves shard information by shard ID
// @Tags shards
// @Accept json
// @Produce json
// @Param id path string true "Shard ID"
// @Success 200 {object} models.Shard "Shard information"
// @Failure 404 {object} map[string]interface{} "Shard not found"
// @Router /shards/{id} [get]
func (h *ManagerHandler) GetShard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shardID := vars["id"]

	shard, err := h.manager.GetShard(shardID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shard)
}

// ListShards handles shard listing requests
// @Summary List all shards
// @Description Returns a list of all shards in the system
// @Tags shards
// @Accept json
// @Produce json
// @Success 200 {array} models.Shard "List of shards"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /shards [get]
func (h *ManagerHandler) ListShards(w http.ResponseWriter, r *http.Request) {
	shards, err := h.manager.ListShards()
	if err != nil {
		h.logger.Error("failed to list shards", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update client apps with shard information when listing shards
	// This ensures client apps show which shards they're using
	if len(shards) > 0 {
		clientAppMgr := h.manager.GetClientAppManager()
		apps, _ := clientAppMgr.ListClientApps()

		// If no client apps exist but we have shards, create default one
		if len(apps) == 0 {
			defaultApp, err := clientAppMgr.RegisterClientApp(
				r.Context(),
				"Default Client Application",
				fmt.Sprintf("Auto-created to represent existing shard usage (%d active shards)", len(shards)),
				"", // database_name - empty for default
				"", // database_host - empty for default
				"", // database_port - empty for default
				"", // database_user - empty for default
				"", // database_password - empty for default
				"", // key_prefix - empty for default
				"", // namespace - empty for default
				"", // cluster_name - empty for default
			)
			if err == nil && defaultApp != nil {
				// Associate all shards with default app
				for _, shard := range shards {
					clientAppMgr.TrackRequest("", shard.ID)
				}
			}
		} else {
			// Associate shards with existing apps (especially default app)
			for _, shard := range shards {
				clientAppMgr.TrackRequest("", shard.ID)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shards)
}

// DeleteShard handles shard deletion requests
// @Summary Delete a shard
// @Description Deletes a shard by ID
// @Tags shards
// @Accept json
// @Produce json
// @Param id path string true "Shard ID"
// @Success 204 "Shard deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Router /shards/{id} [delete]
func (h *ManagerHandler) DeleteShard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shardID := vars["id"]

	if err := h.manager.DeleteShard(shardID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SplitShard handles split operation requests
// @Summary Split a shard
// @Description Splits a shard into multiple target shards
// @Tags resharding
// @Accept json
// @Produce json
// @Param request body models.SplitRequest true "Split Request"
// @Success 202 {object} models.ReshardJob "Split job started"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /reshard/split [post]
func (h *ManagerHandler) SplitShard(w http.ResponseWriter, r *http.Request) {
	var req models.SplitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	job, err := h.manager.SplitShard(r.Context(), &req)
	if err != nil {
		h.logger.Error("failed to start split", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(job)
}

// MergeShards handles merge operation requests
// @Summary Merge shards
// @Description Merges multiple source shards into a target shard
// @Tags resharding
// @Accept json
// @Produce json
// @Param request body models.MergeRequest true "Merge Request"
// @Success 202 {object} models.ReshardJob "Merge job started"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /reshard/merge [post]
func (h *ManagerHandler) MergeShards(w http.ResponseWriter, r *http.Request) {
	var req models.MergeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	job, err := h.manager.MergeShards(r.Context(), &req)
	if err != nil {
		h.logger.Error("failed to start merge", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(job)
}

// GetReshardJob handles reshard job status requests
// @Summary Get reshard job status
// @Description Retrieves the status of a resharding job by job ID
// @Tags resharding
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} models.ReshardJob "Job status"
// @Failure 404 {object} map[string]interface{} "Job not found"
// @Router /reshard/jobs/{id} [get]
func (h *ManagerHandler) GetReshardJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]

	job, err := h.manager.GetReshardJob(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// PromoteReplica handles replica promotion requests
// @Summary Promote a replica to primary
// @Description Promotes a replica to become the primary shard
// @Tags shards
// @Accept json
// @Produce json
// @Param id path string true "Shard ID"
// @Param request body object true "Replica Promotion Request" example({"replica_endpoint": "postgresql://replica:5432/db"})
// @Success 200 {object} map[string]string "Replica promoted successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Router /shards/{id}/promote [post]
func (h *ManagerHandler) PromoteReplica(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shardID := vars["id"]

	var req struct {
		ReplicaEndpoint string `json:"replica_endpoint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.manager.PromoteReplica(shardID, req.ReplicaEndpoint); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "promoted"})
}

// GetPricing handles pricing info requests
// @Summary Get pricing plan
// @Description Retrieves the current pricing plan and limits
// @Tags pricing
// @Accept json
// @Produce json
// @Success 200 {object} pricing.Limits "Pricing limits"
// @Router /pricing [get]
func (h *ManagerHandler) GetPricing(w http.ResponseWriter, r *http.Request) {
	config := h.manager.GetPricingConfig()
	limits := pricing.GetLimits(config.Tier)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(limits)
}

// ClientAppInfo represents client application information (exported for Swagger)
type ClientAppInfo = manager.ClientAppInfo

// ListClientApps handles client application listing requests
// @Summary List all client applications
// @Description Returns a list of all client applications using the sharding system
// @Tags client-apps
// @Accept json
// @Produce json
// @Success 200 {array} ClientAppInfo "List of client applications"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /client-apps [get]
func (h *ManagerHandler) ListClientApps(w http.ResponseWriter, r *http.Request) {
	clientAppMgr := h.manager.GetClientAppManager()
	apps, err := clientAppMgr.ListClientApps()
	if err != nil {
		h.logger.Error("failed to list client apps", zap.Error(err))
		// Return empty array instead of error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	// If no apps, return empty array
	if apps == nil {
		apps = []*ClientAppInfo{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apps)
}

// GetClientApp handles client application retrieval requests
// @Summary Get client application by ID
// @Description Retrieves client application information by ID
// @Tags client-apps
// @Accept json
// @Produce json
// @Param id path string true "Client Application ID"
// @Success 200 {object} ClientAppInfo "Client application information"
// @Failure 404 {object} map[string]interface{} "Client application not found"
// @Router /client-apps/{id} [get]
func (h *ManagerHandler) GetClientApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appID := vars["id"]

	clientAppMgr := h.manager.GetClientAppManager()
	app, err := clientAppMgr.GetClientApp(appID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(app)
}

// CreateClientAppRequest represents a request to create a client application
type CreateClientAppRequest struct {
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	DatabaseName   string `json:"database_name,omitempty"`   // Database name for which sharding needs to be created
	DatabaseHost   string `json:"database_host,omitempty"`    // Database host
	DatabasePort   string `json:"database_port,omitempty"`    // Database port
	DatabaseUser   string `json:"database_user,omitempty"`   // Database user
	DatabasePassword string `json:"database_password,omitempty"` // Database password
	KeyPrefix      string `json:"key_prefix,omitempty"`
	Namespace      string `json:"namespace,omitempty"`        // Kubernetes namespace
	ClusterName    string `json:"cluster_name,omitempty"`     // Kubernetes cluster name
}

// CreateClientApp handles client application creation requests
// @Summary Create a new client application
// @Description Registers a new client application with the sharding system
// @Tags client-apps
// @Accept json
// @Produce json
// @Param request body CreateClientAppRequest true "Client Application Configuration"
// @Success 201 {object} ClientAppInfo "Client application created successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /client-apps [post]
func (h *ManagerHandler) CreateClientApp(w http.ResponseWriter, r *http.Request) {
	var req CreateClientAppRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	clientAppMgr := h.manager.GetClientAppManager()
	app, err := clientAppMgr.RegisterClientApp(r.Context(), req.Name, req.Description, req.DatabaseName, req.DatabaseHost, req.DatabasePort, req.DatabaseUser, req.DatabasePassword, req.KeyPrefix, req.Namespace, req.ClusterName)
	if err != nil {
		h.logger.Error("failed to create client app", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(app)
}

// DiscoverApplications handles Kubernetes application discovery requests
// @Summary Discover applications in Kubernetes namespaces
// @Description Automatically discovers applications and their databases from Kubernetes namespaces
// @Tags client-apps
// @Accept json
// @Produce json
// @Success 200 {array} discovery.DiscoveredApp "List of discovered applications"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /client-apps/discover [get]
func (h *ManagerHandler) DiscoverApplications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.discovery == nil {
		// Return empty array instead of error when discovery is not available
		w.WriteHeader(http.StatusOK)
		// Explicitly create empty slice to ensure JSON encoding produces [] not null
		emptyArray := make([]discovery.DiscoveredApp, 0)
		json.NewEncoder(w).Encode(emptyArray)
		return
	}

	// Update registered apps list
	clientAppMgr := h.manager.GetClientAppManager()
	apps, _ := clientAppMgr.ListClientApps()
	registeredNames := make([]string, 0, len(apps))
	for _, app := range apps {
		registeredNames = append(registeredNames, app.Name)
	}
	h.discovery.UpdateRegisteredApps(registeredNames)

	// Discover applications
	discoveredApps, err := h.discovery.DiscoverApplications(r.Context())
	if err != nil {
		h.logger.Error("failed to discover applications", zap.Error(err))
		// Return empty array on error instead of 500, so UI can handle gracefully
		w.WriteHeader(http.StatusOK)
		emptyArray := make([]discovery.DiscoveredApp, 0)
		json.NewEncoder(w).Encode(emptyArray)
		return
	}

	// Ensure we always return an array, never null
	if discoveredApps == nil {
		discoveredApps = make([]discovery.DiscoveredApp, 0)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(discoveredApps)
}

// SetupPublicRoutes sets up public manager HTTP routes
func SetupPublicRoutes(router *mux.Router, handler *ManagerHandler) {
	// Root route
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service": "sharding-manager",
			"version": "1.0.0",
			"endpoints": []string{
				"GET /api/v1/shards",
				"POST /api/v1/shards",
				"GET /api/v1/shards/{id}",
				"POST /api/v1/reshard/split",
				"POST /api/v1/reshard/merge",
				"GET /api/v1/reshard/jobs/{id}",
				"GET /api/v1/health",
				"GET /health",
				"GET /api/v1/pricing",
				"GET /api/v1/client-apps",
			},
		})
	}).Methods("GET", "OPTIONS")

	router.HandleFunc("/api/v1/pricing", handler.GetPricing).Methods("GET", "OPTIONS")
	// Client apps endpoint - public for now (can be made protected later)
	router.HandleFunc("/api/v1/client-apps", handler.ListClientApps).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/client-apps", handler.CreateClientApp).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/client-apps/discover", handler.DiscoverApplications).Methods("GET", "OPTIONS")

	// Health endpoint under /api/v1
	router.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "healthy",
			"version": "1.0.0",
		})
	}).Methods("GET", "OPTIONS")

	// Legacy health endpoint (keep for backward compatibility)
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET", "OPTIONS")
}

// SetupProtectedRoutes sets up protected manager HTTP routes
func SetupProtectedRoutes(router *mux.Router, handler *ManagerHandler) {
	router.HandleFunc("/api/v1/shards", handler.CreateShard).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/shards", handler.ListShards).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/shards/{id}", handler.GetShard).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/shards/{id}", handler.DeleteShard).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/api/v1/shards/{id}/promote", handler.PromoteReplica).Methods("POST", "OPTIONS")

	router.HandleFunc("/api/v1/reshard/split", handler.SplitShard).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/reshard/merge", handler.MergeShards).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/reshard/jobs/{id}", handler.GetReshardJob).Methods("GET", "OPTIONS")

	// Note: Client application routes are in SetupPublicRoutes for now
	// Move to SetupProtectedRoutes if you want to require authentication
}
