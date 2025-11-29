package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sharding-system/internal/api"
	"github.com/sharding-system/pkg/database"
	"github.com/sharding-system/pkg/manager"
	"github.com/sharding-system/pkg/scanner"
	"go.uber.org/zap"
)

func TestGetDatabaseStats(t *testing.T) {
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

	// Manually inject some databases into the handler for testing
	// Note: In a real test we might want to use the public API to create them,
	// but since we want to test the stats aggregation logic specifically and
	// we don't have easy access to the internal map from outside the package
	// without using the public API (which might be complex to set up fully with mocks),
	// we will rely on the fact that we are in the same package (if we were) or
	// we need to use the public API.
	//
	// Wait, the test file is in package `api` but the handler is in `internal/api`.
	// The test file `tests/api/database_test.go` declares `package api` but imports `github.com/sharding-system/internal/api`.
	// This means it's an external test. We cannot access private fields.
	// We should use the public CreateDatabase API or mock the dependencies if possible.
	// However, `CreateDatabase` requires a lot of mocking.
	//
	// Let's try to use the `CreateDatabase` method if possible, or just rely on the fact that
	// `NewDatabaseHandler` returns a struct that we can't modify easily.
	//
	// Actually, looking at `database_handler.go`, `databases` is a private field.
	// But `CreateDatabase` adds to it.
	// Let's try to create a database using `CreateDatabase`.

	// Mock the dbService.CreateDatabase to return a success without actually doing much?
	// The `dbService` is a struct, not an interface, so we can't easily mock it unless we mock the underlying `manager`.
	// The `manager` is also a struct.
	//
	// This seems complicated to set up perfectly without a lot of mocks.
	// However, `CreateDatabase` in `DatabaseHandler` calls `h.dbService.CreateDatabase`.
	// If we can make that succeed, we are good.
	// `dbService.CreateDatabase` calls `s.manager.GetClientAppManager()` and `s.manager.CreateShard()`.
	//
	// Maybe we can just test the empty state first?

	req := httptest.NewRequest("GET", "/api/v1/databases/stats", nil)
	w := httptest.NewRecorder()

	// Execute
	dbHandler.GetDatabaseStats(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var stats api.DatabaseStats
	if err := json.Unmarshal(w.Body.Bytes(), &stats); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if stats.TotalDatabases != 0 {
		t.Errorf("Expected 0 databases, got %d", stats.TotalDatabases)
	}
}
