package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sharding-system/pkg/backup"
	"go.uber.org/zap"
)

// BackupHandler handles backup management API endpoints
type BackupHandler struct {
	backupService *backup.BackupService
	logger        *zap.Logger
}

// NewBackupHandler creates a new backup handler
func NewBackupHandler(backupService *backup.BackupService, logger *zap.Logger) *BackupHandler {
	return &BackupHandler{
		backupService: backupService,
		logger:        logger,
	}
}

// CreateBackup handles backup creation requests
// @Summary Create a backup for a database
// @Description Creates a new backup for the specified database
// @Tags backups
// @Accept json
// @Produce json
// @Param id path string true "Database ID"
// @Param request body map[string]string true "Backup request (optional type: 'full' or 'incremental')"
// @Success 202 {object} backup.Backup "Backup created"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/databases/{id}/backups [post]
func (h *BackupHandler) CreateBackup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	databaseID := vars["id"]

	var req struct {
		Type string `json:"type"` // "full" or "incremental"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Default to full backup if no body provided
		req.Type = "full"
	}

	if req.Type == "" {
		req.Type = "full"
	}

	backup, err := h.backupService.CreateBackup(r.Context(), databaseID, req.Type)
	if err != nil {
		h.logger.Error("failed to create backup", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(backup)
}

// ListBackups handles backup listing requests
// @Summary List backups for a database
// @Description Returns a list of all backups for the specified database
// @Tags backups
// @Accept json
// @Produce json
// @Param id path string true "Database ID"
// @Success 200 {array} backup.Backup "List of backups"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/databases/{id}/backups [get]
func (h *BackupHandler) ListBackups(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	databaseID := vars["id"]

	backups, err := h.backupService.ListBackups(databaseID)
	if err != nil {
		h.logger.Error("failed to list backups", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(backups)
}

// GetBackup handles backup retrieval requests
// @Summary Get backup by ID
// @Description Retrieves backup information by backup ID
// @Tags backups
// @Accept json
// @Produce json
// @Param id path string true "Database ID"
// @Param backup_id path string true "Backup ID"
// @Success 200 {object} backup.Backup "Backup information"
// @Failure 404 {object} map[string]interface{} "Backup not found"
// @Router /api/v1/databases/{id}/backups/{backup_id} [get]
func (h *BackupHandler) GetBackup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	backupID := vars["backup_id"]

	backup, err := h.backupService.GetBackup(backupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(backup)
}

// RestoreBackup handles backup restore requests
// @Summary Restore database from backup
// @Description Restores a database from a backup
// @Tags backups
// @Accept json
// @Produce json
// @Param id path string true "Database ID"
// @Param backup_id path string true "Backup ID"
// @Param request body map[string]string true "Restore request (optional target_database_id)"
// @Success 202 {object} map[string]string "Restore started"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/databases/{id}/backups/{backup_id}/restore [post]
func (h *BackupHandler) RestoreBackup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	databaseID := vars["id"]
	backupID := vars["backup_id"]

	var req struct {
		TargetDatabaseID string `json:"target_database_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.TargetDatabaseID = databaseID // Default to same database
	}

	if req.TargetDatabaseID == "" {
		req.TargetDatabaseID = databaseID
	}

	if err := h.backupService.RestoreBackup(r.Context(), backupID, req.TargetDatabaseID); err != nil {
		h.logger.Error("failed to restore backup", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "restore_started",
		"message": "Backup restore has been initiated",
	})
}

// ScheduleBackup handles backup scheduling requests
// @Summary Schedule automatic backups
// @Description Schedules automatic backups for a database using cron syntax
// @Tags backups
// @Accept json
// @Produce json
// @Param id path string true "Database ID"
// @Param request body map[string]string true "Schedule request (schedule: cron expression)"
// @Success 200 {object} map[string]string "Backup scheduled"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Router /api/v1/databases/{id}/backups/schedule [post]
func (h *BackupHandler) ScheduleBackup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	databaseID := vars["id"]

	var req struct {
		Schedule string `json:"schedule"` // Cron expression, e.g., "0 2 * * *" for daily at 2 AM
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Schedule == "" {
		req.Schedule = "0 2 * * *" // Default: daily at 2 AM
	}

	if err := h.backupService.ScheduleBackup(databaseID, req.Schedule); err != nil {
		h.logger.Error("failed to schedule backup", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "scheduled",
		"schedule": req.Schedule,
		"message":  "Automatic backups have been scheduled",
	})
}

// SetupBackupRoutes sets up backup management routes
func SetupBackupRoutes(router *mux.Router, handler *BackupHandler) {
	router.HandleFunc("/api/v1/databases/{id}/backups", handler.CreateBackup).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/databases/{id}/backups", handler.ListBackups).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/databases/{id}/backups/{backup_id}", handler.GetBackup).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/v1/databases/{id}/backups/{backup_id}/restore", handler.RestoreBackup).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/databases/{id}/backups/schedule", handler.ScheduleBackup).Methods("POST", "OPTIONS")
}

