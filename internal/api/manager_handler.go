package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sharding-system/pkg/manager"
	"github.com/sharding-system/pkg/models"
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
	manager *manager.Manager
	logger  *zap.Logger
}

// NewManagerHandler creates a new manager handler
func NewManagerHandler(m *manager.Manager, logger *zap.Logger) *ManagerHandler {
	return &ManagerHandler{
		manager: m,
		logger:  logger,
	}
}

// CreateShard handles shard creation requests
// @Summary Create a new shard
// @Description Creates a new database shard with the specified configuration
// @Tags shards
// @Accept json
// @Produce json
// @Param request body models.CreateShardRequest true "Shard Configuration"
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

// SetupManagerRoutes sets up manager HTTP routes
func SetupManagerRoutes(router *mux.Router, handler *ManagerHandler) {
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
			},
		})
	}).Methods("GET", "OPTIONS")
	
	// Register routes directly on the parent router to ensure middleware is applied
	router.HandleFunc("/api/v1/shards", handler.CreateShard).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/shards", handler.ListShards).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/shards/{id}", handler.GetShard).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/shards/{id}", handler.DeleteShard).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/api/v1/shards/{id}/promote", handler.PromoteReplica).Methods("POST", "OPTIONS")
	
	router.HandleFunc("/api/v1/reshard/split", handler.SplitShard).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/reshard/merge", handler.MergeShards).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/reshard/jobs/{id}", handler.GetReshardJob).Methods("GET", "OPTIONS")
	
	// Health endpoint under /api/v1
	router.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "healthy",
			"version": "1.0.0",
		})
	}).Methods("GET", "OPTIONS")
	
	// Legacy health endpoint (keep for backward compatibility)
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET", "OPTIONS")
}

