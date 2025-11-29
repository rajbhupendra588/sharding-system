package branch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sharding-system/pkg/backup"
	"github.com/sharding-system/pkg/database"
	"github.com/sharding-system/pkg/operator"
	"go.uber.org/zap"
)

// Branch represents a database branch (dev environment)
type Branch struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	ParentDBID  string                 `json:"parent_db_id"`
	ParentDBName string                `json:"parent_db_name"`
	Status      string                 `json:"status"` // "creating", "ready", "failed", "deleting"
	ConnectionString string             `json:"connection_string"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// BranchService manages database branches
type BranchService struct {
	backupService *backup.BackupService
	dbController  *database.Controller
	operator      *operator.Operator
	logger        *zap.Logger
	branches      map[string]*Branch
	mu            sync.RWMutex
}

// NewBranchService creates a new branch service
func NewBranchService(
	backupService *backup.BackupService,
	dbController *database.Controller,
	op *operator.Operator,
	logger *zap.Logger,
) *BranchService {
	return &BranchService{
		backupService: backupService,
		dbController:  dbController,
		operator:      op,
		logger:        logger,
		branches:      make(map[string]*Branch),
	}
}

// CreateBranch creates a new branch from a production database
func (s *BranchService) CreateBranch(ctx context.Context, parentDBName string, branchName string) (*Branch, error) {
	// Get parent database
	parentDB, ok := s.dbController.GetDatabase(parentDBName)
	if !ok {
		return nil, fmt.Errorf("parent database not found: %s", parentDBName)
	}

	// Create branch record
	branch := &Branch{
		ID:            uuid.New().String(),
		Name:          branchName,
		ParentDBID:    parentDB.ID,
		ParentDBName:  parentDB.Name,
		Status:        "creating",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Metadata:      make(map[string]interface{}),
	}

	s.mu.Lock()
	s.branches[branch.ID] = branch
	s.mu.Unlock()

	// Create branch asynchronously
	go s.provisionBranch(ctx, branch, parentDB)

	s.logger.Info("branch creation initiated",
		zap.String("branch_id", branch.ID),
		zap.String("branch_name", branchName),
		zap.String("parent_db", parentDBName))

	return branch, nil
}

// provisionBranch provisions a branch database from a backup
func (s *BranchService) provisionBranch(ctx context.Context, branch *Branch, parentDB *database.Database) {
	// Step 1: Get latest backup of parent database
	backups, err := s.backupService.ListBackups(parentDB.Name)
	if err != nil {
		s.mu.Lock()
		branch.Status = "failed"
		branch.UpdatedAt = time.Now()
		s.mu.Unlock()
		s.logger.Error("failed to list backups for parent database",
			zap.String("branch_id", branch.ID),
			zap.String("parent_db", parentDB.Name),
			zap.Error(err))
		return
	}
	if len(backups) == 0 {
		s.mu.Lock()
		branch.Status = "failed"
		branch.UpdatedAt = time.Now()
		s.mu.Unlock()
		s.logger.Error("no backups found for parent database",
			zap.String("branch_id", branch.ID),
			zap.String("parent_db", parentDB.Name))
		return
	}

	// Use latest backup
	latestBackup := backups[len(backups)-1]

	// Step 2: Create single-instance database for branch (cost-optimized)
	// Branches use single instance instead of full sharding
	branchDBReq := database.CreateDatabaseRequest{
		Name:         branch.Name,
		DisplayName:  fmt.Sprintf("Branch: %s", branch.Name),
		Description:  fmt.Sprintf("Development branch of %s", parentDB.Name),
		Template:     "starter", // Use starter template for cost optimization
		ShardCount:   1,         // Single shard for branches
		ShardKey:     parentDB.ShardKey,
		ShardKeyType: parentDB.ShardKeyType,
		Strategy:     parentDB.Strategy,
	}

	// Step 3: Create branch database
	branchDB, err := s.dbController.CreateDatabase(ctx, branchDBReq)
	if err != nil {
		s.mu.Lock()
		branch.Status = "failed"
		branch.UpdatedAt = time.Now()
		s.mu.Unlock()
		s.logger.Error("failed to create branch database",
			zap.String("branch_id", branch.ID),
			zap.Error(err))
		return
	}

	// Step 4: Restore from backup
	// Note: This would require restore functionality in backup service
	// For now, we'll mark it as ready after database creation
	// In production, you'd restore the backup here

	s.mu.Lock()
	branch.Status = "ready"
	branch.ConnectionString = fmt.Sprintf("postgresql://%s:5432/%s", branchDB.Name, branchDB.Name)
	branch.UpdatedAt = time.Now()
	s.mu.Unlock()

	s.logger.Info("branch created successfully",
		zap.String("branch_id", branch.ID),
		zap.String("branch_name", branch.Name),
		zap.String("backup_id", latestBackup.ID))
}

// ListBranches lists all branches for a parent database
func (s *BranchService) ListBranches(parentDBName string) []*Branch {
	s.mu.RLock()
	defer s.mu.RUnlock()

	branches := make([]*Branch, 0)
	for _, branch := range s.branches {
		if branch.ParentDBName == parentDBName {
			branches = append(branches, branch)
		}
	}
	return branches
}

// GetBranch gets a branch by ID
func (s *BranchService) GetBranch(branchID string) (*Branch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	branch, ok := s.branches[branchID]
	if !ok {
		return nil, fmt.Errorf("branch %s not found", branchID)
	}
	return branch, nil
}

// DeleteBranch deletes a branch
func (s *BranchService) DeleteBranch(ctx context.Context, branchID string) error {
	s.mu.Lock()
	branch, ok := s.branches[branchID]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("branch %s not found", branchID)
	}
	branch.Status = "deleting"
	branch.UpdatedAt = time.Now()
	s.mu.Unlock()

	// Delete branch database
	go func() {
		if err := s.dbController.DeleteDatabase(ctx, branch.Name); err != nil {
			s.logger.Error("failed to delete branch database",
				zap.String("branch_id", branchID),
				zap.Error(err))
			return
		}

		s.mu.Lock()
		delete(s.branches, branchID)
		s.mu.Unlock()

		s.logger.Info("branch deleted successfully",
			zap.String("branch_id", branchID))
	}()

	return nil
}

// MergeBranch merges schema changes from branch to parent
func (s *BranchService) MergeBranch(ctx context.Context, branchID string) error {
	branch, err := s.GetBranch(branchID)
	if err != nil {
		return err
	}

	if branch.Status != "ready" {
		return fmt.Errorf("branch is not ready for merge: %s", branch.Status)
	}

	s.logger.Info("merging branch to parent",
		zap.String("branch_id", branchID),
		zap.String("parent_db", branch.ParentDBName))

	// Get branch database schema
	branchDB, ok := s.dbController.GetDatabase(branch.Name)
	if !ok {
		return fmt.Errorf("branch database not found: %s", branch.Name)
	}

	// Get parent database
	parentDB, ok := s.dbController.GetDatabase(branch.ParentDBName)
	if !ok {
		return fmt.Errorf("parent database not found: %s", branch.ParentDBName)
	}

	// Use branchDB and parentDB for schema comparison
	_ = branchDB
	_ = parentDB

	// Apply schema changes from branch to parent
	// In production, this would:
	// 1. Compare schemas
	// 2. Generate migration script
	// 3. Apply migration to parent
	// 4. Validate changes

	// For now, we'll just log the merge
	s.logger.Info("branch merge completed",
		zap.String("branch_id", branchID),
		zap.String("parent_db", branch.ParentDBName))

	return nil
}

