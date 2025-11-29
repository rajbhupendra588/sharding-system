package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sharding-system/pkg/manager"
	"github.com/sharding-system/pkg/monitoring"
	"go.uber.org/zap"
)

// PostgresStatsHandler handles PostgreSQL statistics API endpoints
type PostgresStatsHandler struct {
	statsCollector *monitoring.PostgresStatsCollector
	manager        *manager.Manager
	logger         *zap.Logger
}

// NewPostgresStatsHandler creates a new PostgreSQL stats handler
func NewPostgresStatsHandler(
	statsCollector *monitoring.PostgresStatsCollector,
	manager *manager.Manager,
	logger *zap.Logger,
) *PostgresStatsHandler {
	return &PostgresStatsHandler{
		statsCollector: statsCollector,
		manager:        manager,
		logger:         logger,
	}
}

// GetDatabaseStats returns PostgreSQL stats for a specific database
// @Summary Get PostgreSQL stats for a database
// @Description Returns detailed PostgreSQL statistics for a specific database
// @Tags postgres-stats
// @Accept json
// @Produce json
// @Param id path string true "Database ID"
// @Success 200 {object} monitoring.PostgresStats "PostgreSQL statistics"
// @Failure 404 {object} map[string]interface{} "Database not found or stats not available"
// @Router /api/v1/databases/{id}/stats [get]
func (h *PostgresStatsHandler) GetDatabaseStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	databaseID := vars["id"]

	stats, err := h.statsCollector.GetStats(databaseID)
	if err != nil {
		h.logger.Warn("failed to get database stats",
			zap.String("database_id", databaseID),
			zap.Error(err))
		http.Error(w, "stats not available for database", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetAllDatabaseStats returns PostgreSQL stats for all registered databases
// @Summary Get PostgreSQL stats for all databases
// @Description Returns detailed PostgreSQL statistics for all registered databases
// @Tags postgres-stats
// @Accept json
// @Produce json
// @Success 200 {object} map[string]monitoring.PostgresStats "PostgreSQL statistics map"
// @Router /api/v1/databases/stats [get]
func (h *PostgresStatsHandler) GetAllDatabaseStats(w http.ResponseWriter, r *http.Request) {
	allStats := h.statsCollector.GetAllStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allStats)
}

// GetShardStats returns PostgreSQL stats for a specific shard
// @Summary Get PostgreSQL stats for a shard
// @Description Returns detailed PostgreSQL statistics for a specific shard
// @Tags postgres-stats
// @Accept json
// @Produce json
// @Param id path string true "Shard ID"
// @Success 200 {object} monitoring.PostgresStats "PostgreSQL statistics"
// @Failure 404 {object} map[string]interface{} "Shard not found or stats not available"
// @Router /api/v1/shards/{id}/stats [get]
func (h *PostgresStatsHandler) GetShardStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shardID := vars["id"]

	// Verify shard exists
	_, err := h.manager.GetShard(shardID)
	if err != nil {
		http.Error(w, "shard not found", http.StatusNotFound)
		return
	}

	// Try to get stats using shard ID as database ID
	stats, err := h.statsCollector.GetStats(shardID)
	if err != nil {
		h.logger.Warn("failed to get shard stats",
			zap.String("shard_id", shardID),
			zap.Error(err))
		http.Error(w, "stats not available for shard", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// RegisterRoutes registers PostgreSQL stats API routes
func (h *PostgresStatsHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/databases/{id}/stats", h.GetDatabaseStats).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/databases/stats", h.GetAllDatabaseStats).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/shards/{id}/stats", h.GetShardStats).Methods("GET", "OPTIONS")
}

// endpointToDSN converts a PostgreSQL endpoint URL to DSN format
// Supports both postgresql:// and postgres:// URLs
func endpointToDSN(endpoint string) (string, error) {
	// If endpoint is already in DSN format (contains "host="), return as is
	if strings.Contains(endpoint, "host=") {
		return endpoint, nil
	}

	// Parse as URL
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}

	// Build DSN components
	parts := []string{}

	if parsedURL.Hostname() != "" {
		parts = append(parts, "host="+parsedURL.Hostname())
	}

	if parsedURL.Port() != "" {
		parts = append(parts, "port="+parsedURL.Port())
	}

	if parsedURL.User != nil {
		if username := parsedURL.User.Username(); username != "" {
			parts = append(parts, "user="+username)
		}
		if password, ok := parsedURL.User.Password(); ok && password != "" {
			parts = append(parts, "password="+password)
		}
	}

	// Database name from path (remove leading slash)
	dbname := strings.TrimPrefix(parsedURL.Path, "/")
	if dbname != "" {
		parts = append(parts, "dbname="+dbname)
	}

	// Add SSL mode and connection timeout
	parts = append(parts, "sslmode=prefer", "connect_timeout=10")

	return strings.Join(parts, " "), nil
}

