package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sharding-system/pkg/branch"
	"go.uber.org/zap"
)

// BranchHandler handles database branching API endpoints
type BranchHandler struct {
	service *branch.BranchService
	logger  *zap.Logger
}

// NewBranchHandler creates a new branch handler
func NewBranchHandler(service *branch.BranchService, logger *zap.Logger) *BranchHandler {
	return &BranchHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers branch API routes
func (h *BranchHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/v1/databases/{dbName}/branches", h.ListBranches).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/databases/{dbName}/branches", h.CreateBranch).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/v1/branches/{branchID}", h.GetBranch).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/branches/{branchID}", h.DeleteBranch).Methods("DELETE", "OPTIONS")
	r.HandleFunc("/api/v1/branches/{branchID}/merge", h.MergeBranch).Methods("POST", "OPTIONS")
}

// CreateBranch creates a new branch from a database
// @Summary Create database branch
// @Description Creates a new development branch from a production database
// @Tags branches
// @Accept json
// @Produce json
// @Param dbName path string true "Database Name"
// @Param request body object true "Branch details" example({"name": "dev-branch"})
// @Success 201 {object} branch.Branch
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Internal server error"
// @Router /databases/{dbName}/branches [post]
func (h *BranchHandler) CreateBranch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbName := vars["dbName"]

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "branch name is required", http.StatusBadRequest)
		return
	}

	branch, err := h.service.CreateBranch(r.Context(), dbName, req.Name)
	if err != nil {
		h.logger.Error("failed to create branch", zap.String("db_name", dbName), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(branch)
}

// ListBranches lists all branches for a database
// @Summary List branches
// @Description Returns all branches for a database
// @Tags branches
// @Produce json
// @Param dbName path string true "Database Name"
// @Success 200 {array} branch.Branch
// @Router /databases/{dbName}/branches [get]
func (h *BranchHandler) ListBranches(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbName := vars["dbName"]

	branches := h.service.ListBranches(dbName)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(branches)
}

// GetBranch retrieves a branch by ID
// @Summary Get branch
// @Description Retrieves branch details by ID
// @Tags branches
// @Produce json
// @Param branchID path string true "Branch ID"
// @Success 200 {object} branch.Branch
// @Failure 404 {string} string "Branch not found"
// @Router /branches/{branchID} [get]
func (h *BranchHandler) GetBranch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	branchID := vars["branchID"]

	branch, err := h.service.GetBranch(branchID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(branch)
}

// DeleteBranch deletes a branch
// @Summary Delete branch
// @Description Deletes a database branch
// @Tags branches
// @Param branchID path string true "Branch ID"
// @Success 204 "Branch deleted successfully"
// @Failure 404 {string} string "Branch not found"
// @Failure 500 {string} string "Internal server error"
// @Router /branches/{branchID} [delete]
func (h *BranchHandler) DeleteBranch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	branchID := vars["branchID"]

	err := h.service.DeleteBranch(r.Context(), branchID)
	if err != nil {
		h.logger.Error("failed to delete branch", zap.String("branch_id", branchID), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// MergeBranch merges a branch into its parent database
// @Summary Merge branch
// @Description Merges schema changes from branch to parent database
// @Tags branches
// @Param branchID path string true "Branch ID"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "Bad request"
// @Failure 404 {string} string "Branch not found"
// @Failure 500 {string} string "Internal server error"
// @Router /branches/{branchID}/merge [post]
func (h *BranchHandler) MergeBranch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	branchID := vars["branchID"]

	err := h.service.MergeBranch(r.Context(), branchID)
	if err != nil {
		h.logger.Error("failed to merge branch", zap.String("branch_id", branchID), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "merge initiated"})
}

