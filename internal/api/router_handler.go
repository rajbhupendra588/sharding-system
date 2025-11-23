package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sharding-system/internal/errors"
	"github.com/sharding-system/pkg/models"
	"github.com/sharding-system/pkg/router"
	"go.uber.org/zap"
)

// @title Sharding System Router API
// @version 1.0
// @description API for routing requests to shards based on shard keys
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@sharding-system.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /v1

// RouterHandler handles HTTP requests for the router
type RouterHandler struct {
	router *router.Router
	logger *zap.Logger
}

// NewRouterHandler creates a new router handler
func NewRouterHandler(r *router.Router, logger *zap.Logger) *RouterHandler {
	return &RouterHandler{
		router: r,
		logger: logger,
	}
}

// ExecuteQuery handles query execution requests
// @Summary Execute a query on a shard
// @Description Executes a SQL query on the shard determined by the shard key
// @Tags router
// @Accept json
// @Produce json
// @Param request body models.QueryRequest true "Query Request"
// @Success 200 {object} models.QueryResponse "Query executed successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /execute [post]
func (h *RouterHandler) ExecuteQuery(w http.ResponseWriter, r *http.Request) {
	var req models.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, errors.Wrap(err, http.StatusBadRequest, "invalid request body"))
		return
	}

	if req.ShardKey == "" {
		h.writeError(w, errors.New(http.StatusBadRequest, "shard_key is required"))
		return
	}

	if req.Query == "" {
		h.writeError(w, errors.New(http.StatusBadRequest, "query is required"))
		return
	}

	if req.Consistency == "" {
		req.Consistency = "strong"
	}

	resp, err := h.router.ExecuteQuery(r.Context(), &req)
	if err != nil {
		h.logger.Error("query execution failed", zap.Error(err))
		h.writeError(w, errors.Wrap(err, http.StatusInternalServerError, "query execution failed"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// GetShardForKey handles shard lookup requests
// @Summary Get shard ID for a key
// @Description Returns the shard ID that handles the given key
// @Tags router
// @Accept json
// @Produce json
// @Param key query string true "Shard key"
// @Success 200 {object} map[string]string "Shard ID"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /shard-for-key [get]
func (h *RouterHandler) GetShardForKey(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		h.writeError(w, errors.New(http.StatusBadRequest, "key parameter is required"))
		return
	}

	shardID, err := h.router.GetShardForKey(key)
	if err != nil {
		h.logger.Error("failed to get shard", zap.Error(err))
		h.writeError(w, errors.Wrap(err, http.StatusInternalServerError, "failed to get shard"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"shard_id": shardID}); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// writeError writes an error response in a standardized format
func (h *RouterHandler) writeError(w http.ResponseWriter, err *errors.Error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTPStatus())
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    err.Code,
			"message": err.Message,
		},
	})
}

// SetupRouterRoutes sets up router HTTP routes
func SetupRouterRoutes(router *mux.Router, handler *RouterHandler) {
	// Root route
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service": "sharding-router",
			"version": "1.0.0",
			"endpoints": []string{
				"POST /v1/execute",
				"GET /v1/shard-for-key?key=<key>",
				"GET /v1/health",
				"GET /health",
			},
		})
	}).Methods("GET", "OPTIONS")
	
	router.HandleFunc("/v1/execute", handler.ExecuteQuery).Methods("POST", "OPTIONS")
	router.HandleFunc("/v1/shard-for-key", handler.GetShardForKey).Methods("GET", "OPTIONS")
	
	// Health endpoint under /v1
	router.HandleFunc("/v1/health", func(w http.ResponseWriter, r *http.Request) {
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

