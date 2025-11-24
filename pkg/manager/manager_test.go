package manager

import (
	"context"
	"errors"
	"testing"

	"github.com/sharding-system/pkg/config"
	"github.com/sharding-system/pkg/models"
	"go.uber.org/zap/zaptest"
)

// MockCatalog implements catalog.Catalog for testing
type MockCatalog struct {
	shards map[string]*models.Shard
}

func NewMockCatalog() *MockCatalog {
	return &MockCatalog{
		shards: make(map[string]*models.Shard),
	}
}

func (m *MockCatalog) GetShard(key string, clientAppID string) (*models.Shard, error) {
	for _, shard := range m.shards {
		return shard, nil
	}
	return nil, errors.New("no shard found")
}

func (m *MockCatalog) GetShardByID(shardID string) (*models.Shard, error) {
	shard, ok := m.shards[shardID]
	if !ok {
		return nil, errors.New("shard not found")
	}
	return shard, nil
}

func (m *MockCatalog) ListShards(clientAppID string) ([]models.Shard, error) {
	shards := make([]models.Shard, 0, len(m.shards))
	for _, shard := range m.shards {
		shards = append(shards, *shard)
	}
	return shards, nil
}

func (m *MockCatalog) CreateShard(shard *models.Shard) error {
	m.shards[shard.ID] = shard
	return nil
}

func (m *MockCatalog) UpdateShard(shard *models.Shard) error {
	m.shards[shard.ID] = shard
	return nil
}

func (m *MockCatalog) DeleteShard(shardID string) error {
	delete(m.shards, shardID)
	return nil
}

func (m *MockCatalog) GetCatalogVersion() (int64, error) {
	return 1, nil
}

func (m *MockCatalog) Watch(ctx context.Context) (<-chan *models.ShardCatalog, error) {
	ch := make(chan *models.ShardCatalog)
	return ch, nil
}

// MockResharder implements Resharder for testing
type MockResharder struct {
	splitError error
	mergeError error
}

func (m *MockResharder) Split(ctx context.Context, job *models.ReshardJob) error {
	job.Status = "completed"
	job.Progress = 1.0
	return m.splitError
}

func (m *MockResharder) Merge(ctx context.Context, job *models.ReshardJob) error {
	job.Status = "completed"
	job.Progress = 1.0
	return m.mergeError
}

func TestManager_CreateShard(t *testing.T) {
	logger := zaptest.NewLogger(t)
	catalog := NewMockCatalog()
	resharder := &MockResharder{}

	manager := NewManager(catalog, logger, resharder, config.PricingConfig{Tier: "pro"})

	// First, register a client application (required for shard creation)
	ctx := context.Background()
	clientApp, err := manager.GetClientAppManager().RegisterClientApp(ctx, "test-client-app", "Test app", "", "", "", "", "", "", "", "")
	if err != nil {
		t.Fatalf("Failed to register client app: %v", err)
	}

	req := &models.CreateShardRequest{
		Name:            "test-shard",
		ClientAppID:     clientApp.ID, // Use the registered client app ID
		PrimaryEndpoint: "postgres://localhost/test",
		Replicas:        []string{"postgres://localhost/replica"},
		VNodeCount:      10,
	}

	shard, err := manager.CreateShard(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if shard == nil {
		t.Fatal("Expected non-nil shard")
	}
	if shard.Name != "test-shard" {
		t.Errorf("Expected name=test-shard, got %s", shard.Name)
	}
	if shard.Status != "active" {
		t.Errorf("Expected status=active, got %s", shard.Status)
	}
	if len(shard.VNodes) != 10 {
		t.Errorf("Expected 10 vnodes, got %d", len(shard.VNodes))
	}
}

func TestManager_GetShard(t *testing.T) {
	logger := zaptest.NewLogger(t)
	catalog := NewMockCatalog()
	resharder := &MockResharder{}

	manager := NewManager(catalog, logger, resharder, config.PricingConfig{Tier: "pro"})

	shard := &models.Shard{
		ID:              "shard1",
		Name:            "test-shard",
		PrimaryEndpoint: "postgres://localhost/test",
		Status:          "active",
	}
	catalog.CreateShard(shard)

	got, err := manager.GetShard("shard1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if got.ID != "shard1" {
		t.Errorf("Expected shard1, got %s", got.ID)
	}
}

func TestManager_ListShards(t *testing.T) {
	logger := zaptest.NewLogger(t)
	catalog := NewMockCatalog()
	resharder := &MockResharder{}

	manager := NewManager(catalog, logger, resharder, config.PricingConfig{Tier: "pro"})

	shard1 := &models.Shard{ID: "shard1", Name: "shard1"}
	shard2 := &models.Shard{ID: "shard2", Name: "shard2"}
	catalog.CreateShard(shard1)
	catalog.CreateShard(shard2)

	shards, err := manager.ListShards()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(shards) != 2 {
		t.Errorf("Expected 2 shards, got %d", len(shards))
	}
}

func TestManager_DeleteShard_ActiveShard(t *testing.T) {
	logger := zaptest.NewLogger(t)
	catalog := NewMockCatalog()
	resharder := &MockResharder{}

	manager := NewManager(catalog, logger, resharder, config.PricingConfig{Tier: "pro"})

	shard := &models.Shard{
		ID:     "shard1",
		Status: "active",
	}
	catalog.CreateShard(shard)

	err := manager.DeleteShard("shard1")
	if err == nil {
		t.Error("Expected error when deleting active shard")
	}
}

func TestManager_DeleteShard_InactiveShard(t *testing.T) {
	logger := zaptest.NewLogger(t)
	catalog := NewMockCatalog()
	resharder := &MockResharder{}

	manager := NewManager(catalog, logger, resharder, config.PricingConfig{Tier: "pro"})

	shard := &models.Shard{
		ID:     "shard1",
		Status: "inactive",
	}
	catalog.CreateShard(shard)

	err := manager.DeleteShard("shard1")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestManager_GetReshardJob(t *testing.T) {
	logger := zaptest.NewLogger(t)
	catalog := NewMockCatalog()
	resharder := &MockResharder{}

	manager := NewManager(catalog, logger, resharder, config.PricingConfig{Tier: "pro"})

	job := &models.ReshardJob{
		ID:     "job1",
		Status: "pending",
	}
	manager.mu.Lock()
	manager.jobs["job1"] = job
	manager.mu.Unlock()

	got, err := manager.GetReshardJob("job1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if got.ID != "job1" {
		t.Errorf("Expected job1, got %s", got.ID)
	}
}

func TestManager_GetReshardJob_NotFound(t *testing.T) {
	logger := zaptest.NewLogger(t)
	catalog := NewMockCatalog()
	resharder := &MockResharder{}

	manager := NewManager(catalog, logger, resharder, config.PricingConfig{Tier: "pro"})

	_, err := manager.GetReshardJob("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent job")
	}
}
