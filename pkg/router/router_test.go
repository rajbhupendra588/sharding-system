package router

import (
	"context"
	"errors"
	"testing"
	"time"

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

func (m *MockCatalog) GetShard(key string) (*models.Shard, error) {
	// Simple mock: return first shard for any key
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

func (m *MockCatalog) ListShards() ([]models.Shard, error) {
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

func TestRouter_GetShardForKey(t *testing.T) {
	logger := zaptest.NewLogger(t)
	catalog := NewMockCatalog()
	
	shard := &models.Shard{
		ID:              "shard1",
		Name:            "test-shard",
		PrimaryEndpoint: "postgres://localhost/test",
		Status:          "active",
	}
	catalog.CreateShard(shard)
	
	router := NewRouter(catalog, logger, 10, 5*time.Minute, "primary")
	
	shardID, err := router.GetShardForKey("test-key")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if shardID != "shard1" {
		t.Errorf("Expected shard1, got %s", shardID)
	}
}

func TestRouter_GetShardForKey_NoShard(t *testing.T) {
	logger := zaptest.NewLogger(t)
	catalog := NewMockCatalog()
	
	router := NewRouter(catalog, logger, 10, 5*time.Minute, "primary")
	
	_, err := router.GetShardForKey("test-key")
	if err == nil {
		t.Error("Expected error when no shard exists")
	}
}

func TestRouter_Close(t *testing.T) {
	logger := zaptest.NewLogger(t)
	catalog := NewMockCatalog()
	
	router := NewRouter(catalog, logger, 10, 5*time.Minute, "primary")
	
	// Close should not panic
	err := router.Close()
	if err != nil {
		t.Errorf("Expected no error on close, got %v", err)
	}
}

func TestRouter_NewRouter(t *testing.T) {
	logger := zaptest.NewLogger(t)
	catalog := NewMockCatalog()
	
	router := NewRouter(catalog, logger, 10, 5*time.Minute, "replica_ok")
	
	if router == nil {
		t.Fatal("Expected non-nil router")
	}
	if router.maxConns != 10 {
		t.Errorf("Expected maxConns=10, got %d", router.maxConns)
	}
	if router.connTTL != 5*time.Minute {
		t.Errorf("Expected connTTL=5m, got %v", router.connTTL)
	}
	if router.replicaPolicy != "replica_ok" {
		t.Errorf("Expected replicaPolicy=replica_ok, got %s", router.replicaPolicy)
	}
}

// Note: ExecuteQuery tests would require a real database connection
// or a more sophisticated mock. For unit tests, we focus on the routing logic.

