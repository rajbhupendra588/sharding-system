package database

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sharding-system/pkg/manager"
	"github.com/sharding-system/pkg/models"
	"go.uber.org/zap"
)

// SimpleDatabase represents a sharded database with simplified management (Phase 1)
// This is a simplified version that wraps the existing Database type
type SimpleDatabase struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	DisplayName     string            `json:"display_name"`
	Description     string            `json:"description"`
	Template        string            `json:"template"`
	ShardKey        string            `json:"shard_key"`
	ClientAppID     string            `json:"client_app_id"`
	ShardIDs        []string          `json:"shard_ids"`
	Status          string            `json:"status"` // "creating", "ready", "failed", "scaling"
	ConnectionString string           `json:"connection_string"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// SimpleCreateDatabaseRequest represents a simplified request to create a database
type SimpleCreateDatabaseRequest struct {
	Name        string `json:"name"`                  // Required: Database name
	Template    string `json:"template,omitempty"`    // Optional: "starter", "production", "enterprise" (default: "starter")
	ShardKey    string `json:"shard_key,omitempty"`   // Optional: Auto-detected if not provided
	DisplayName string `json:"display_name,omitempty"` // Optional: Display name
	Description string `json:"description,omitempty"`  // Optional: Description
}

// DatabaseService provides simplified database management
type DatabaseService struct {
	manager    *manager.Manager
	logger     *zap.Logger
	routerHost string
	routerPort int
}

// NewDatabaseService creates a new database service
func NewDatabaseService(mgr *manager.Manager, logger *zap.Logger, routerHost string, routerPort int) *DatabaseService {
	return &DatabaseService{
		manager:    mgr,
		logger:     logger,
		routerHost: routerHost,
		routerPort: routerPort,
	}
}

// CreateDatabase creates a new sharded database with minimal configuration
func (s *DatabaseService) CreateDatabase(ctx context.Context, req SimpleCreateDatabaseRequest) (*SimpleDatabase, error) {
	// Validate request
	if req.Name == "" {
		return nil, fmt.Errorf("database name is required")
	}

	// Get template
	template := GetTemplate(req.Template)
	if req.Template == "" {
		req.Template = "starter"
	}

	// Auto-detect shard key if not provided
	shardKey := req.ShardKey
	if shardKey == "" {
		shardKey = "id" // Default shard key
	}

	// Create or get client application
	clientAppMgr := s.manager.GetClientAppManager()
	clientApp, err := s.getOrCreateClientApp(ctx, clientAppMgr, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create client app: %w", err)
	}

	// Create database record
	db := &SimpleDatabase{
		ID:              uuid.New().String(),
		Name:            req.Name,
		DisplayName:     req.DisplayName,
		Description:     req.Description,
		Template:        req.Template,
		ShardKey:        shardKey,
		ClientAppID:     clientApp.ID,
		ShardIDs:        make([]string, 0),
		Status:          "creating",
		ConnectionString: s.generateConnectionString(req.Name),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Metadata:        make(map[string]interface{}),
	}

	// Create shards asynchronously
	go s.createShards(ctx, db, template, shardKey)

	s.logger.Info("started creating database",
		zap.String("database_id", db.ID),
		zap.String("name", req.Name),
		zap.String("template", req.Template),
		zap.Int("shard_count", template.ShardCount))

	return db, nil
}

// getOrCreateClientApp gets existing client app or creates a new one
// NOTE: This function currently creates client apps without database info, which will fail validation.
// This service is for auto-provisioning databases, so it should:
// 1. Provision the database first (using operator)
// 2. Then register the client app with actual database endpoints
// 3. Then create shards with validated endpoints
// For now, this will fail validation - the service needs to be refactored to provision databases first.
func (s *DatabaseService) getOrCreateClientApp(ctx context.Context, clientAppMgr *manager.ClientAppManager, name string) (*manager.ClientAppInfo, error) {
	// Try to get existing app first (simplified - in production, use proper lookup)
	// For now, always create a new one
	// NOTE: This will fail with current validation - database must be provisioned first
	app, err := clientAppMgr.RegisterClientApp(ctx, name, fmt.Sprintf("Auto-created for database %s", name), "", "", "", "", "", "", "", "")
	if err != nil {
		return nil, err
	}
	return app, nil
}

// createShards creates all shards for the database
func (s *DatabaseService) createShards(ctx context.Context, db *SimpleDatabase, template DatabaseTemplate, shardKey string) {
	shardIDs := make([]string, 0, template.ShardCount)

	for i := 0; i < template.ShardCount; i++ {
		shardName := fmt.Sprintf("%s-shard-%d", db.Name, i+1)
		
		// Create shard request
		shardReq := &models.CreateShardRequest{
			Name:        shardName,
			ClientAppID: db.ClientAppID,
			// Note: In a real implementation, these would be auto-provisioned
			// For now, we require them to be provided or use mock endpoints
			PrimaryEndpoint: fmt.Sprintf("postgresql://localhost:5432/%s_shard_%d", db.Name, i+1),
			Replicas:        make([]string, template.ReplicasPerShard),
			VNodeCount:     template.VNodeCount,
		}

		// Add replica endpoints
		for j := 0; j < template.ReplicasPerShard; j++ {
			shardReq.Replicas[j] = fmt.Sprintf("postgresql://localhost:5433/%s_shard_%d_replica_%d", db.Name, i+1, j+1)
		}

		shard, err := s.manager.CreateShard(ctx, shardReq)
		if err != nil {
			s.logger.Error("failed to create shard",
				zap.String("database_id", db.ID),
				zap.String("shard_name", shardName),
				zap.Error(err))
			
			// Update database status to failed
			db.Status = "failed"
			return
		}

		shardIDs = append(shardIDs, shard.ID)
		s.logger.Info("created shard",
			zap.String("database_id", db.ID),
			zap.String("shard_id", shard.ID),
			zap.String("shard_name", shardName))
	}

	// Update database with shard IDs
	db.ShardIDs = shardIDs
	db.Status = "ready"
	db.UpdatedAt = time.Now()

	s.logger.Info("database creation completed",
		zap.String("database_id", db.ID),
		zap.String("name", db.Name),
		zap.Int("shard_count", len(shardIDs)))
}

// generateConnectionString generates a connection string for the database
func (s *DatabaseService) generateConnectionString(dbName string) string {
	// In production, this would use a load balancer or router endpoint
	if s.routerHost != "" {
		return fmt.Sprintf("postgresql://%s:%d/%s", s.routerHost, s.routerPort, dbName)
	}
	return fmt.Sprintf("postgresql://localhost:%d/%s", s.routerPort, dbName)
}

