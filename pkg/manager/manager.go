package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sharding-system/pkg/catalog"
	"github.com/sharding-system/pkg/config"
	"github.com/sharding-system/pkg/hashing"
	"github.com/sharding-system/pkg/models"
	"github.com/sharding-system/pkg/pricing"
	"github.com/sharding-system/pkg/validation"
	"go.uber.org/zap"
)

// Manager manages shards and resharding operations
type Manager struct {
	catalog       catalog.Catalog
	logger        *zap.Logger
	jobs          map[string]*models.ReshardJob
	mu            sync.RWMutex
	resharder     Resharder
	pricingConfig config.PricingConfig
	clientAppMgr  *ClientAppManager
}

// Resharder handles data migration
type Resharder interface {
	Split(ctx context.Context, job *models.ReshardJob) error
	Merge(ctx context.Context, job *models.ReshardJob) error
}

// NewManager creates a new shard manager
func NewManager(catalog catalog.Catalog, logger *zap.Logger, resharder Resharder, pricingConfig config.PricingConfig) *Manager {
	return &Manager{
		catalog:       catalog,
		logger:        logger,
		jobs:          make(map[string]*models.ReshardJob),
		resharder:     resharder,
		pricingConfig: pricingConfig,
		clientAppMgr:  NewClientAppManager(catalog, logger),
	}
}

// GetClientAppManager returns the client application manager
func (m *Manager) GetClientAppManager() *ClientAppManager {
	return m.clientAppMgr
}

// GetPricingConfig returns the pricing configuration
func (m *Manager) GetPricingConfig() config.PricingConfig {
	return m.pricingConfig
}

// InitializeClientApps loads existing client applications from storage
// No auto-creation of default apps - only registered apps are shown
func (m *Manager) InitializeClientApps() error {
	// Client apps are loaded from etcd in NewClientAppManager
	// Nothing to do here - no auto-creation
	m.logger.Info("client app manager initialized")
	return nil
}

// CreateShard creates a new shard for a client application
func (m *Manager) CreateShard(ctx context.Context, req *models.CreateShardRequest) (*models.Shard, error) {
	// Validate client app ID is provided
	if req.ClientAppID == "" {
		return nil, fmt.Errorf("client_app_id is required - shards must belong to a client application")
	}

	// Verify client application exists
	clientAppMgr := m.GetClientAppManager()
	_, err := clientAppMgr.GetClientApp(req.ClientAppID)
	if err != nil {
		return nil, fmt.Errorf("client application not found: %s", req.ClientAppID)
	}

	// Check pricing limits (per client app)
	limits := pricing.GetLimits(m.pricingConfig.Tier)
	if limits.MaxShards != -1 {
		shards, err := m.ListShardsForClient(req.ClientAppID)
		if err != nil {
			return nil, fmt.Errorf("failed to list shards for limit check: %w", err)
		}
		if len(shards) >= limits.MaxShards {
			return nil, fmt.Errorf("shard limit reached for client application %s (max %d)", req.ClientAppID, limits.MaxShards)
		}
	}

	// Validate database connection before creating shard
	// Shards must have a valid database connection
	if !validation.HasDatabaseInfo(req.Host, fmt.Sprintf("%d", req.Port), req.Database, req.Username, req.Password, req.PrimaryEndpoint) {
		return nil, fmt.Errorf("insufficient database information: shards require a valid database connection (host, database, and either primary_endpoint or host+database)")
	}

	// Validate database connection
	if err := validation.ValidateDatabaseConnection(ctx, req.Host, fmt.Sprintf("%d", req.Port), req.Database, req.Username, req.Password, req.PrimaryEndpoint); err != nil {
		return nil, fmt.Errorf("database connection validation failed: %w. Shards cannot be created without a valid database connection", err)
	}

	// Determine status - but don't set to active if validation fails
	status := req.Status
	if status == "" {
		// Only set to active if we have a valid database connection (already validated above)
		status = "active"
	} else if status == "active" {
		// If explicitly setting to active, ensure database is still valid
		// (already validated above, but this is a safety check)
	}

	shard := &models.Shard{
		ID:              uuid.New().String(),
		Name:            req.Name,
		ClientAppID:     req.ClientAppID,
		PrimaryEndpoint: req.PrimaryEndpoint,
		Replicas:        req.Replicas,
		Status:          status,
		Version:         1,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		// Database connection details
		Host:     req.Host,
		Port:     req.Port,
		Database: req.Database,
		Username: req.Username,
		Password: req.Password,
		Weight:   req.Weight,
	}

	// Generate VNodes
	vnodeCount := req.VNodeCount
	if vnodeCount == 0 {
		vnodeCount = 256 // default
	}

	shard.VNodes = make([]models.VNode, vnodeCount)
	hashFunc := hashing.NewHashFunction("murmur3")
	for i := 0; i < vnodeCount; i++ {
		vnodeKey := shard.ID + "-vnode-" + fmt.Sprintf("%d", i)
		hash := hashFunc.Hash(vnodeKey)
		shard.VNodes[i] = models.VNode{
			ID:      uint64(i),
			ShardID: shard.ID,
			Hash:    hash,
		}
	}

	if err := m.catalog.CreateShard(shard); err != nil {
		return nil, fmt.Errorf("failed to create shard in catalog: %w", err)
	}

	m.logger.Info("created shard", zap.String("shard_id", shard.ID), zap.String("name", shard.Name))
	return shard, nil
}

// GetShard retrieves a shard by ID
func (m *Manager) GetShard(shardID string) (*models.Shard, error) {
	return m.catalog.GetShardByID(shardID)
}

// ListShards lists all shards (for admin/management purposes)
func (m *Manager) ListShards() ([]models.Shard, error) {
	return m.catalog.ListShards("")
}

// ListShardsForClient lists shards for a specific client application
func (m *Manager) ListShardsForClient(clientAppID string) ([]models.Shard, error) {
	return m.catalog.ListShards(clientAppID)
}

// DeleteShard deletes a shard
func (m *Manager) DeleteShard(shardID string) error {
	shard, err := m.catalog.GetShardByID(shardID)
	if err != nil {
		return err
	}

	if shard.Status == "active" {
		return fmt.Errorf("cannot delete active shard %s", shardID)
	}

	return m.catalog.DeleteShard(shardID)
}

// UpdateShardStatus updates the status of a shard
func (m *Manager) UpdateShardStatus(shardID string, status string) error {
	shard, err := m.catalog.GetShardByID(shardID)
	if err != nil {
		return err
	}

	// If setting status to "active", validate database connection first
	if status == "active" {
		// Build DSN from shard connection details
		dsn := ""
		if shard.PrimaryEndpoint != "" {
			dsn = shard.PrimaryEndpoint
		} else if shard.Host != "" && shard.Database != "" {
			port := shard.Port
			if port == 0 {
				port = 5432
			}
			dsn = fmt.Sprintf("host=%s port=%d dbname=%s", shard.Host, port, shard.Database)
			if shard.Username != "" {
				dsn += fmt.Sprintf(" user=%s", shard.Username)
			}
			if shard.Password != "" {
				dsn += fmt.Sprintf(" password=%s", shard.Password)
			}
			dsn += " sslmode=prefer connect_timeout=10"
		}

		if dsn == "" {
			return fmt.Errorf("cannot set shard status to active: shard has no valid database connection information")
		}

		// Validate database connection
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := validation.ValidateDatabaseConnection(ctx, shard.Host, fmt.Sprintf("%d", shard.Port), shard.Database, shard.Username, shard.Password, shard.PrimaryEndpoint); err != nil {
			return fmt.Errorf("cannot set shard status to active: database connection validation failed: %w", err)
		}
	}

	shard.Status = status
	shard.UpdatedAt = time.Now()

	if err := m.catalog.UpdateShard(shard); err != nil {
		return fmt.Errorf("failed to update shard status: %w", err)
	}

	m.logger.Info("updated shard status",
		zap.String("shard_id", shardID),
		zap.String("status", status),
	)

	return nil
}

// SplitShard starts a split operation
func (m *Manager) SplitShard(ctx context.Context, req *models.SplitRequest) (*models.ReshardJob, error) {
	sourceShard, err := m.catalog.GetShardByID(req.SourceShardID)
	if err != nil {
		return nil, fmt.Errorf("source shard not found: %w", err)
	}

	if sourceShard.Status != "active" {
		return nil, fmt.Errorf("source shard is not active: %s", sourceShard.Status)
	}

	// Create target shards
	targetShards := make([]*models.Shard, 0, len(req.TargetShards))
	for _, targetReq := range req.TargetShards {
		shard, err := m.CreateShard(ctx, &targetReq)
		if err != nil {
			return nil, fmt.Errorf("failed to create target shard: %w", err)
		}
		shard.Status = "migrating"
		m.catalog.UpdateShard(shard)
		targetShards = append(targetShards, shard)
	}

	// Create reshard job
	job := &models.ReshardJob{
		ID:           uuid.New().String(),
		Type:         "split",
		SourceShards: []string{req.SourceShardID},
		TargetShards: make([]string, 0, len(targetShards)),
		Status:       "pending",
		Progress:     0.0,
		StartedAt:    time.Now(),
		TotalKeys:    0, // Will be determined during migration
	}

	for _, shard := range targetShards {
		job.TargetShards = append(job.TargetShards, shard.ID)
	}

	m.mu.Lock()
	m.jobs[job.ID] = job
	m.mu.Unlock()

	// Start async resharding
	go m.executeReshard(ctx, job)

	m.logger.Info("started split operation", zap.String("job_id", job.ID), zap.String("source_shard", req.SourceShardID))
	return job, nil
}

// MergeShards starts a merge operation
func (m *Manager) MergeShards(ctx context.Context, req *models.MergeRequest) (*models.ReshardJob, error) {
	// Validate source shards
	for _, shardID := range req.SourceShardIDs {
		shard, err := m.catalog.GetShardByID(shardID)
		if err != nil {
			return nil, fmt.Errorf("source shard not found: %s", shardID)
		}
		if shard.Status != "active" {
			return nil, fmt.Errorf("source shard is not active: %s", shardID)
		}
	}

	// Create target shard
	targetShard, err := m.CreateShard(ctx, &req.TargetShard)
	if err != nil {
		return nil, fmt.Errorf("failed to create target shard: %w", err)
	}
	targetShard.Status = "migrating"
	m.catalog.UpdateShard(targetShard)

	// Create reshard job
	job := &models.ReshardJob{
		ID:           uuid.New().String(),
		Type:         "merge",
		SourceShards: req.SourceShardIDs,
		TargetShards: []string{targetShard.ID},
		Status:       "pending",
		Progress:     0.0,
		StartedAt:    time.Now(),
		TotalKeys:    0,
	}

	m.mu.Lock()
	m.jobs[job.ID] = job
	m.mu.Unlock()

	// Start async resharding
	go m.executeReshard(ctx, job)

	m.logger.Info("started merge operation", zap.String("job_id", job.ID))
	return job, nil
}

// GetReshardJob retrieves a reshard job by ID
func (m *Manager) GetReshardJob(jobID string) (*models.ReshardJob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	job, exists := m.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	return job, nil
}

// executeReshard executes a resharding operation
func (m *Manager) executeReshard(ctx context.Context, job *models.ReshardJob) {
	m.mu.Lock()
	job.Status = "precopy"
	m.mu.Unlock()

	var err error
	if job.Type == "split" {
		err = m.resharder.Split(ctx, job)
	} else if job.Type == "merge" {
		err = m.resharder.Merge(ctx, job)
	} else {
		err = fmt.Errorf("unknown reshard type: %s", job.Type)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if err != nil {
		job.Status = "failed"
		job.ErrorMessage = err.Error()
		m.logger.Error("reshard failed", zap.String("job_id", job.ID), zap.Error(err))
	} else {
		job.Status = "completed"
		now := time.Now()
		job.CompletedAt = &now
		m.logger.Info("reshard completed", zap.String("job_id", job.ID))
	}
}

// PromoteReplica promotes a replica to primary
func (m *Manager) PromoteReplica(shardID string, replicaEndpoint string) error {
	shard, err := m.catalog.GetShardByID(shardID)
	if err != nil {
		return err
	}

	// Verify replica exists
	found := false
	for _, rep := range shard.Replicas {
		if rep == replicaEndpoint {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("replica endpoint not found: %s", replicaEndpoint)
	}

	// Update shard: old primary becomes replica, new primary is promoted
	oldPrimary := shard.PrimaryEndpoint
	shard.Replicas = append(shard.Replicas, oldPrimary)
	shard.PrimaryEndpoint = replicaEndpoint

	// Remove promoted replica from replicas list
	newReplicas := make([]string, 0)
	for _, rep := range shard.Replicas {
		if rep != replicaEndpoint {
			newReplicas = append(newReplicas, rep)
		}
	}
	shard.Replicas = newReplicas

	if err := m.catalog.UpdateShard(shard); err != nil {
		return fmt.Errorf("failed to update catalog: %w", err)
	}

	m.logger.Info("promoted replica to primary",
		zap.String("shard_id", shardID),
		zap.String("new_primary", replicaEndpoint),
	)

	return nil
}
