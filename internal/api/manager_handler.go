package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sharding-system/pkg/discovery"
	"github.com/sharding-system/pkg/manager"
	"github.com/sharding-system/pkg/models"
	"github.com/sharding-system/pkg/monitoring"
	"github.com/sharding-system/pkg/pricing"
	"go.uber.org/zap"
)

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
	manager              *manager.Manager
	logger               *zap.Logger
	prometheusCollector  *monitoring.PrometheusCollector
	postgresStatsCollector *monitoring.PostgresStatsCollector
}

// NewManagerHandler creates a new manager handler
func NewManagerHandler(m *manager.Manager, logger *zap.Logger) *ManagerHandler {
	return &ManagerHandler{
		manager: m,
		logger:  logger,
	}
}

// SetPrometheusCollector sets the Prometheus collector for metrics registration
func (h *ManagerHandler) SetPrometheusCollector(pc *monitoring.PrometheusCollector) {
	h.prometheusCollector = pc
}

// SetPostgresStatsCollector sets the PostgreSQL stats collector for stats registration
func (h *ManagerHandler) SetPostgresStatsCollector(psc *monitoring.PostgresStatsCollector) {
	h.postgresStatsCollector = psc
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

	// Register shard with Prometheus collector for metrics if collector is available
	if h.prometheusCollector != nil && shard.Status == "active" {
		dsn := buildDSNFromShard(shard)
		if dsn != "" {
			if err := h.prometheusCollector.RegisterShard(shard.ID, dsn); err != nil {
				h.logger.Warn("failed to register shard for metrics collection",
					zap.String("shard_id", shard.ID),
					zap.Error(err))
			} else {
				h.logger.Info("registered shard for metrics collection",
					zap.String("shard_id", shard.ID))
			}
		}
	}

	// Register shard with PostgreSQL stats collector if collector is available
	if h.postgresStatsCollector != nil && shard.Status == "active" {
		dsn := buildDSNFromShard(shard)
		if dsn != "" {
			if err := h.postgresStatsCollector.RegisterDatabase(shard.ID, dsn); err != nil {
				h.logger.Warn("failed to register shard with PostgreSQL stats collector",
					zap.String("shard_id", shard.ID),
					zap.Error(err))
			} else {
				h.logger.Info("registered shard for PostgreSQL stats collection",
					zap.String("shard_id", shard.ID))
			}
		}
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
// @Description Returns a list of all shards in the system. Filter by client_app_id to get shards for a specific application.
// @Tags shards
// @Accept json
// @Produce json
// @Param client_app_id query string false "Filter by client application ID"
// @Success 200 {array} models.Shard "List of shards"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /shards [get]
func (h *ManagerHandler) ListShards(w http.ResponseWriter, r *http.Request) {
	// Check for client_app_id filter (used by Java client to fetch shard config)
	clientAppID := r.URL.Query().Get("client_app_id")
	
	var shards []models.Shard
	var err error
	
	if clientAppID != "" {
		// Filter shards by client app
		shards, err = h.manager.ListShardsForClient(clientAppID)
	} else {
		// Return all shards (admin view)
		shards, err = h.manager.ListShards()
	}
	
	if err != nil {
		h.logger.Error("failed to list shards", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

	// Unregister shard from metrics collection
	if h.prometheusCollector != nil {
		h.prometheusCollector.UnregisterShard(shardID)
		h.logger.Info("unregistered shard from metrics collection",
			zap.String("shard_id", shardID))
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

	// Register target shards for metrics collection
	if h.prometheusCollector != nil {
		for _, targetShardID := range job.TargetShards {
			shard, err := h.manager.GetShard(targetShardID)
			if err == nil {
				dsn := buildDSNFromShard(shard)
				if dsn != "" {
					if err := h.prometheusCollector.RegisterShard(targetShardID, dsn); err != nil {
						h.logger.Warn("failed to register target shard for metrics after split",
							zap.String("shard_id", targetShardID),
							zap.Error(err))
					} else {
						h.logger.Info("registered target shard for metrics after split",
							zap.String("shard_id", targetShardID))
					}
				}
			}
		}
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

	// Register target shard for metrics collection
	if h.prometheusCollector != nil && len(job.TargetShards) > 0 {
		targetShardID := job.TargetShards[0]
		shard, err := h.manager.GetShard(targetShardID)
		if err == nil {
			dsn := buildDSNFromShard(shard)
			if dsn != "" {
				if err := h.prometheusCollector.RegisterShard(targetShardID, dsn); err != nil {
					h.logger.Warn("failed to register target shard for metrics after merge",
						zap.String("shard_id", targetShardID),
						zap.Error(err))
				} else {
					h.logger.Info("registered target shard for metrics after merge",
						zap.String("shard_id", targetShardID))
				}
			}
		}
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

// UpdateShardStatus handles shard status update requests
// @Summary Update shard status
// @Description Updates the status of a shard (e.g., to inactive)
// @Tags shards
// @Accept json
// @Produce json
// @Param id path string true "Shard ID"
// @Param request body object true "Status Update Request" example({"status": "inactive"})
// @Success 200 {object} map[string]string "Status updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Router /shards/{id}/status [put]
func (h *ManagerHandler) UpdateShardStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shardID := vars["id"]

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Status != "active" && req.Status != "inactive" {
		http.Error(w, "invalid status: must be 'active' or 'inactive'", http.StatusBadRequest)
		return
	}

	// Get shard first to verify existence
	if _, err := h.manager.GetShard(shardID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := h.manager.UpdateShardStatus(shardID, req.Status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update metrics registration based on new status
	if h.prometheusCollector != nil {
		shard, err := h.manager.GetShard(shardID)
		if err == nil {
			if req.Status == "active" {
				// Register for metrics if becoming active
				dsn := buildDSNFromShard(shard)
				if dsn != "" {
					if err := h.prometheusCollector.RegisterShard(shardID, dsn); err != nil {
						h.logger.Warn("failed to register shard for metrics after status update",
							zap.String("shard_id", shardID),
							zap.Error(err))
					}
				}
			} else {
				// Unregister if becoming inactive
				h.prometheusCollector.UnregisterShard(shardID)
				h.logger.Info("unregistered shard from metrics collection",
					zap.String("shard_id", shardID))
			}
		}
	}

	// Update PostgreSQL stats registration based on new status
	if h.postgresStatsCollector != nil {
		shard, err := h.manager.GetShard(shardID)
		if err == nil {
			if req.Status == "active" {
				// Register for stats if becoming active
				dsn := buildDSNFromShard(shard)
				if dsn != "" {
					if err := h.postgresStatsCollector.RegisterDatabase(shardID, dsn); err != nil {
						h.logger.Warn("failed to register shard with PostgreSQL stats collector after status update",
							zap.String("shard_id", shardID),
							zap.Error(err))
					} else {
						h.logger.Info("registered shard for PostgreSQL stats collection after status update",
							zap.String("shard_id", shardID))
					}
				}
			} else {
				// Unregister if becoming inactive
				h.postgresStatsCollector.UnregisterDatabase(shardID)
				h.logger.Info("unregistered shard from PostgreSQL stats collection",
					zap.String("shard_id", shardID))
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
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
// @Description Returns a list of all client applications registered with the sharding system
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

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
	Name             string `json:"name"`
	Description      string `json:"description,omitempty"`
	DatabaseName     string `json:"database_name,omitempty"`     // Database name for which sharding needs to be created
	DatabaseHost     string `json:"database_host,omitempty"`     // Database host
	DatabasePort     string `json:"database_port,omitempty"`     // Database port
	DatabaseUser     string `json:"database_user,omitempty"`     // Database user
	DatabasePassword string `json:"database_password,omitempty"` // Database password
	KeyPrefix        string `json:"key_prefix,omitempty"`
	Namespace        string `json:"namespace,omitempty"`    // Kubernetes namespace
	ClusterName      string `json:"cluster_name,omitempty"` // Kubernetes cluster name
}

// CreateClientApp handles client application creation requests
// @Summary Create a new client application
// @Description Registers a new client application with the sharding system. Only registered apps are visible in the UI.
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

// DeleteClientApp handles client application deletion requests
// @Summary Delete a client application
// @Description De-registers a client application from the sharding system
// @Tags client-apps
// @Accept json
// @Produce json
// @Param id path string true "Client Application ID"
// @Success 204 "Client application deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Client application not found"
// @Router /client-apps/{id} [delete]
func (h *ManagerHandler) DeleteClientApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appID := vars["id"]

	clientAppMgr := h.manager.GetClientAppManager()
	if err := clientAppMgr.DeleteClientApp(appID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DiscoverClientApps handles client application discovery requests
// @Summary Discover applications from Kubernetes
// @Description Discovers applications running in Kubernetes clusters that can be registered as client applications
// @Tags client-apps
// @Accept json
// @Produce json
// @Success 200 {array} discovery.DiscoveredApp "List of discovered applications"
// @Failure 503 {object} map[string]interface{} "Kubernetes discovery not available"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /client-apps/discover [get]
func (h *ManagerHandler) DiscoverClientApps(w http.ResponseWriter, r *http.Request) {
	// Get list of registered client apps to check which ones are already registered
	clientAppMgr := h.manager.GetClientAppManager()
	registeredApps, err := clientAppMgr.ListClientApps()
	if err != nil {
		h.logger.Warn("failed to list registered apps for discovery", zap.Error(err))
		registeredApps = []*manager.ClientAppInfo{}
	}

	// Build list of registered app names for discovery service
	registeredNames := make([]string, 0, len(registeredApps))
	for _, app := range registeredApps {
		registeredNames = append(registeredNames, app.Name)
	}

	// Try to create Kubernetes discovery service
	var discoveryService discovery.DiscoveryService
	discoveryService, err = discovery.NewKubernetesDiscovery(h.logger, registeredNames)
	if err != nil {
		// Kubernetes not available - use mock discovery (returns empty list)
		h.logger.Info("Kubernetes discovery not available, using mock discovery", zap.Error(err))
		discoveryService = discovery.NewMockDiscovery(h.logger)
		discoveryService.UpdateRegisteredApps(registeredNames)
	}

	// Discover applications
	discoveredApps, err := discoveryService.DiscoverApplications(r.Context())
	if err != nil {
		h.logger.Error("failed to discover applications", zap.Error(err))
		// Return 503 Service Unavailable if discovery fails
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Kubernetes discovery not available",
			"message": err.Error(),
		})
		return
	}

	// Filter out applications without database information
	// Only applications with valid database connections should be discoverable
	filteredApps := make([]discovery.DiscoveredApp, 0)
	for _, app := range discoveredApps {
		// Skip apps without database info - they should not be discoverable for sharding
		if app.DatabaseHost == "" && app.DatabaseURL == "" {
			h.logger.Debug("filtering out discovered app without database information",
				zap.String("app_name", app.Name),
				zap.String("namespace", app.Namespace))
			continue
		}

		// Validate that we have sufficient database information
		hasDBInfo := app.DatabaseHost != "" && app.DatabaseName != ""
		if !hasDBInfo && app.DatabaseURL == "" {
			h.logger.Debug("filtering out discovered app with insufficient database information",
				zap.String("app_name", app.Name),
				zap.String("namespace", app.Namespace))
			continue
		}

		filteredApps = append(filteredApps, app)
	}

	// Ensure we always return an array (not null)
	if filteredApps == nil {
		filteredApps = []discovery.DiscoveredApp{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(filteredApps)
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
				"GET /api/v1/client-apps/discover",
			},
		})
	}).Methods("GET", "OPTIONS")

	router.HandleFunc("/api/v1/pricing", handler.GetPricing).Methods("GET", "OPTIONS")
	// Client apps endpoints - only registered apps are shown
	router.HandleFunc("/api/v1/client-apps", handler.ListClientApps).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/client-apps", handler.CreateClientApp).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/client-apps/discover", handler.DiscoverClientApps).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/client-apps/{id}", handler.GetClientApp).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/client-apps/{id}", handler.DeleteClientApp).Methods("DELETE", "OPTIONS")

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
	router.HandleFunc("/api/v1/shards/{id}/status", handler.UpdateShardStatus).Methods("PUT", "OPTIONS")

	router.HandleFunc("/api/v1/reshard/split", handler.SplitShard).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/reshard/merge", handler.MergeShards).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/reshard/jobs/{id}", handler.GetReshardJob).Methods("GET", "OPTIONS")
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
