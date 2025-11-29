package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sharding-system/internal/api"
	"github.com/sharding-system/pkg/catalog"
	"github.com/sharding-system/pkg/config"
	"github.com/sharding-system/pkg/database"
	"github.com/sharding-system/pkg/manager"
	"github.com/sharding-system/pkg/models"
	"github.com/sharding-system/pkg/resharder"
	"github.com/sharding-system/pkg/scanner"
	"go.uber.org/zap"
)

func TestCreateDatabase(t *testing.T) {
	// Setup
	logger, _ := zap.NewDevelopment()
	catalog := setupMockCatalog(t)
	resharder := setupMockResharder(catalog)
	pricingConfig := setupMockPricingConfig()

	shardManager := manager.NewManager(catalog, logger, resharder, pricingConfig)
	dbService := database.NewDatabaseService(shardManager, logger, "localhost", 8080)
	clusterManager := scanner.NewClusterManager(logger)
	dbScanner := scanner.NewDatabaseScanner(logger)
	multiClusterScanner := scanner.NewMultiClusterScanner(clusterManager, dbScanner, logger)
	dbHandler := api.NewDatabaseHandler(dbService, clusterManager, multiClusterScanner, logger)

	// Test request
	reqBody := database.CreateDatabaseRequest{
		Name:     "test-db",
		Template: "starter",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/databases", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	dbHandler.CreateDatabase(w, req)

	// Assert
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var db database.Database
	if err := json.Unmarshal(w.Body.Bytes(), &db); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if db.Name != "test-db" {
		t.Errorf("Expected name 'test-db', got '%s'", db.Name)
	}

	if db.Template != "starter" {
		t.Errorf("Expected template 'starter', got '%s'", db.Template)
	}

	if db.Status != "creating" {
		t.Errorf("Expected status 'creating', got '%s'", db.Status)
	}
}

func TestListTemplates(t *testing.T) {
	// Setup
	logger, _ := zap.NewDevelopment()
	catalog := setupMockCatalog(t)
	resharder := setupMockResharder(catalog)
	pricingConfig := setupMockPricingConfig()

	shardManager := manager.NewManager(catalog, logger, resharder, pricingConfig)
	dbService := database.NewDatabaseService(shardManager, logger, "localhost", 8080)
	clusterManager := scanner.NewClusterManager(logger)
	dbScanner := scanner.NewDatabaseScanner(logger)
	multiClusterScanner := scanner.NewMultiClusterScanner(clusterManager, dbScanner, logger)
	dbHandler := api.NewDatabaseHandler(dbService, clusterManager, multiClusterScanner, logger)

	req := httptest.NewRequest("GET", "/api/v1/databases/templates", nil)
	w := httptest.NewRecorder()

	// Execute
	dbHandler.ListTemplates(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var templates []database.DatabaseTemplate
	if err := json.Unmarshal(w.Body.Bytes(), &templates); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(templates) == 0 {
		t.Error("Expected at least one template")
	}

	// Check for starter template
	foundStarter := false
	for _, tpl := range templates {
		if tpl.Name == "starter" {
			foundStarter = true
			if tpl.ShardCount != 2 {
				t.Errorf("Expected starter template to have 2 shards, got %d", tpl.ShardCount)
			}
			break
		}
	}

	if !foundStarter {
		t.Error("Expected to find 'starter' template")
	}
}

// Helper functions
func setupMockCatalog(t *testing.T) *MockCatalog {
	return &MockCatalog{}
}

func setupMockResharder(c catalog.Catalog) *resharder.Resharder {
	logger, _ := zap.NewDevelopment()
	return resharder.NewResharder(c, logger)
}

func setupMockPricingConfig() config.PricingConfig {
	return config.PricingConfig{Tier: "free"}
}

// MockCatalog implements catalog.Catalog for testing
type MockCatalog struct{}

func (m *MockCatalog) GetShard(key string, clientAppID string) (*models.Shard, error) {
	return nil, nil
}
func (m *MockCatalog) GetShardByID(shardID string) (*models.Shard, error) {
	return nil, nil
}
func (m *MockCatalog) ListShards(clientAppID string) ([]models.Shard, error) {
	return []models.Shard{}, nil
}
func (m *MockCatalog) CreateShard(shard *models.Shard) error {
	return nil
}
func (m *MockCatalog) UpdateShard(shard *models.Shard) error {
	return nil
}
func (m *MockCatalog) DeleteShard(shardID string) error {
	return nil
}
func (m *MockCatalog) GetCatalogVersion() (int64, error) {
	return 1, nil
}
func (m *MockCatalog) Watch(ctx context.Context) (<-chan *models.ShardCatalog, error) {
	return nil, nil
}
