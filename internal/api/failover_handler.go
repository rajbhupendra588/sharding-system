package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sharding-system/pkg/failover"
	"go.uber.org/zap"
)

// FailoverHandler handles failover management API endpoints
type FailoverHandler struct {
	failoverCtrl *failover.FailoverController
	logger       *zap.Logger
}

// NewFailoverHandler creates a new failover handler
func NewFailoverHandler(failoverCtrl *failover.FailoverController, logger *zap.Logger) *FailoverHandler {
	return &FailoverHandler{
		failoverCtrl: failoverCtrl,
		logger:       logger,
	}
}

// GetFailoverStatus handles failover status requests
// @Summary Get failover status
// @Description Returns the current status of automatic failover
// @Tags failover
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Failover status"
// @Router /api/v1/failover/status [get]
func (h *FailoverHandler) GetFailoverStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"enabled": h.failoverCtrl.IsEnabled(),
		"status":  "active",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// EnableFailover handles enable failover requests
// @Summary Enable automatic failover
// @Description Enables automatic failover for all shards
// @Tags failover
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Failover enabled"
// @Router /api/v1/failover/enable [post]
func (h *FailoverHandler) EnableFailover(w http.ResponseWriter, r *http.Request) {
	h.failoverCtrl.Enable()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "enabled",
		"message": "Automatic failover has been enabled",
	})
}

// DisableFailover handles disable failover requests
// @Summary Disable automatic failover
// @Description Disables automatic failover for all shards
// @Tags failover
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Failover disabled"
// @Router /api/v1/failover/disable [post]
func (h *FailoverHandler) DisableFailover(w http.ResponseWriter, r *http.Request) {
	h.failoverCtrl.Disable()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "disabled",
		"message": "Automatic failover has been disabled",
	})
}

// GetFailoverHistory handles failover history requests
// @Summary Get failover history
// @Description Returns the history of failover events
// @Tags failover
// @Accept json
// @Produce json
// @Param shard_id query string false "Filter by shard ID"
// @Success 200 {array} failover.FailoverEvent "Failover history"
// @Router /api/v1/failover/history [get]
func (h *FailoverHandler) GetFailoverHistory(w http.ResponseWriter, r *http.Request) {
	shardID := r.URL.Query().Get("shard_id")

	var history []*failover.FailoverEvent
	if shardID != "" {
		history = h.failoverCtrl.GetFailoverHistoryForShard(shardID)
	} else {
		history = h.failoverCtrl.GetFailoverHistory()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// SetupFailoverRoutes sets up failover management routes
func SetupFailoverRoutes(router *mux.Router, handler *FailoverHandler) {
	router.HandleFunc("/api/v1/failover/status", handler.GetFailoverStatus).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/failover/enable", handler.EnableFailover).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/failover/disable", handler.DisableFailover).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/failover/history", handler.GetFailoverHistory).Methods("GET", "OPTIONS")
}

