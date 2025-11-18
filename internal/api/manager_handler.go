package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sharding-system/pkg/manager"
	"github.com/sharding-system/pkg/models"
	"go.uber.org/zap"
)

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

