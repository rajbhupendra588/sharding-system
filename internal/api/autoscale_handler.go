package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sharding-system/pkg/autoscale"
	"go.uber.org/zap"
)

// AutoscaleHandler handles auto-scaling API endpoints
type AutoscaleHandler struct {
	detector *autoscale.HotShardDetector
	splitter *autoscale.AutoSplitter
	logger   *zap.Logger
}

// NewAutoscaleHandler creates a new autoscale handler
func NewAutoscaleHandler(
	detector *autoscale.HotShardDetector,
	splitter *autoscale.AutoSplitter,
	logger *zap.Logger,
) *AutoscaleHandler {
	return &AutoscaleHandler{
		detector: detector,
		splitter: splitter,
		logger:   logger,
	}
}

// RegisterRoutes registers autoscale API routes
func (h *AutoscaleHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/v1/autoscale/status", h.GetStatus).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/autoscale/enable", h.Enable).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/v1/autoscale/disable", h.Disable).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/v1/autoscale/hot-shards", h.GetHotShards).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/autoscale/cold-shards", h.GetColdShards).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/autoscale/thresholds", h.GetThresholds).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/autoscale/thresholds", h.UpdateThresholds).Methods("PUT", "OPTIONS")
}

// GetStatus returns the current status of auto-scaling
// @Summary Get auto-scale status
// @Description Returns whether auto-scaling is enabled
// @Tags autoscale
// @Produce json
// @Success 200 {object} map[string]bool
// @Router /autoscale/status [get]
func (h *AutoscaleHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	enabled := h.splitter.IsEnabled()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"enabled": enabled})
}

// Enable enables automatic scaling
// @Summary Enable auto-scaling
// @Description Enables automatic shard splitting
// @Tags autoscale
// @Success 200 {object} map[string]string
// @Router /autoscale/enable [post]
func (h *AutoscaleHandler) Enable(w http.ResponseWriter, r *http.Request) {
	h.splitter.Enable()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "enabled"})
}

// Disable disables automatic scaling
// @Summary Disable auto-scaling
// @Description Disables automatic shard splitting
// @Tags autoscale
// @Success 200 {object} map[string]string
// @Router /autoscale/disable [post]
func (h *AutoscaleHandler) Disable(w http.ResponseWriter, r *http.Request) {
	h.splitter.Disable()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "disabled"})
}

// GetHotShards returns list of hot shards
// @Summary Get hot shards
// @Description Returns list of shards that exceed thresholds
// @Tags autoscale
// @Produce json
// @Success 200 {object} map[string][]string
// @Router /autoscale/hot-shards [get]
func (h *AutoscaleHandler) GetHotShards(w http.ResponseWriter, r *http.Request) {
	hotShards := h.detector.GetHotShards()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"shards": hotShards})
}

// GetColdShards returns list of cold shards
// @Summary Get cold shards
// @Description Returns list of shards that are underutilized
// @Tags autoscale
// @Produce json
// @Success 200 {object} map[string][]string
// @Router /autoscale/cold-shards [get]
func (h *AutoscaleHandler) GetColdShards(w http.ResponseWriter, r *http.Request) {
	coldShards := h.detector.GetColdShards()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"shards": coldShards})
}

// GetThresholds returns current detection thresholds
// @Summary Get thresholds
// @Description Returns current thresholds for hot/cold shard detection
// @Tags autoscale
// @Produce json
// @Success 200 {object} autoscale.Thresholds
// @Router /autoscale/thresholds [get]
func (h *AutoscaleHandler) GetThresholds(w http.ResponseWriter, r *http.Request) {
	thresholds := h.detector.GetThresholds()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(thresholds)
}

// UpdateThresholds updates detection thresholds
// @Summary Update thresholds
// @Description Updates thresholds for hot/cold shard detection
// @Tags autoscale
// @Accept json
// @Produce json
// @Param thresholds body autoscale.Thresholds true "Threshold configuration"
// @Success 200 {object} autoscale.Thresholds
// @Router /autoscale/thresholds [put]
func (h *AutoscaleHandler) UpdateThresholds(w http.ResponseWriter, r *http.Request) {
	var thresholds autoscale.Thresholds
	if err := json.NewDecoder(r.Body).Decode(&thresholds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.detector.UpdateThresholds(thresholds)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(thresholds)
}

