package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sharding-system/pkg/monitoring"
	"go.uber.org/zap"
)

// MetricsHandler handles load metrics API endpoints
type MetricsHandler struct {
	monitor *monitoring.LoadMonitor
	logger  *zap.Logger
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(monitor *monitoring.LoadMonitor, logger *zap.Logger) *MetricsHandler {
	return &MetricsHandler{
		monitor: monitor,
		logger:  logger,
	}
}

// RegisterRoutes registers metrics API routes
func (h *MetricsHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/v1/metrics/shard/{shardID}", h.GetShardMetrics).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/metrics/shard", h.GetAllMetrics).Methods("GET", "OPTIONS")
}

// GetShardMetrics returns metrics for a specific shard
// @Summary Get shard metrics
// @Description Returns current load metrics for a specific shard
// @Tags metrics
// @Produce json
// @Param shardID path string true "Shard ID"
// @Success 200 {object} monitoring.ShardMetrics
// @Failure 404 {string} string "Shard not found"
// @Router /metrics/shard/{shardID} [get]
func (h *MetricsHandler) GetShardMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shardID := vars["shardID"]

	metrics, ok := h.monitor.GetMetrics(shardID)
	if !ok {
		http.Error(w, "metrics not found for shard", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// GetAllMetrics returns metrics for all shards
// @Summary Get all metrics
// @Description Returns current load metrics for all shards
// @Tags metrics
// @Produce json
// @Success 200 {object} map[string]monitoring.ShardMetrics
// @Router /metrics/shard [get]
func (h *MetricsHandler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	allMetrics := h.monitor.GetAllMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allMetrics)
}

