package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sharding-system/internal/server"
	"github.com/sharding-system/pkg/catalog"
	"github.com/sharding-system/pkg/config"
	"github.com/sharding-system/pkg/manager"
	"github.com/sharding-system/pkg/models"
	"github.com/sharding-system/pkg/resharder"
	"github.com/sharding-system/pkg/router"
	"go.uber.org/zap"
)

// setupTestServers creates test instances of router and manager servers
func setupTestServers(t *testing.T) (*server.RouterServer, *server.ManagerServer, func()) {
	logger, _ := zap.NewDevelopment()

	// Create mock catalog
	cat, err := catalog.NewEtcdCatalog([]string{"localhost:2379"}, logger)
	if err != nil {
		t.Fatalf("Failed to create catalog: %v", err)
	}

	// Create router
	shardRouter := router.NewRouter(cat, logger, 10, 5*time.Minute, "replica_ok", config.PricingConfig{Tier: "pro"})

	// Create resharder
	resharderInstance := resharder.NewResharder(cat, logger)

	// Create manager
	shardManager := manager.NewManager(cat, logger, resharderInstance, config.PricingConfig{Tier: "pro"})

	// Create config
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:         "localhost",
			Port:         0, // Use random port
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
	}

	// Create router server
	routerSrv, err := server.NewRouterServer(cfg, shardRouter, logger)
	if err != nil {
		t.Fatalf("Failed to create router server: %v", err)
	}

	// Create manager server
	managerSrv, err := server.NewManagerServer(cfg, shardManager, nil, logger)
	if err != nil {
		t.Fatalf("Failed to create manager server: %v", err)
	}

	cleanup := func() {
		shardRouter.Close()
	}

	return routerSrv, managerSrv, cleanup
}

// TestRouterHealthEndpoint tests the router health endpoint
func TestRouterHealthEndpoint(t *testing.T) {
	routerSrv, _, cleanup := setupTestServers(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/v1/health", nil)
	w := httptest.NewRecorder()

	// Get handler from server
	routerSrv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if status, ok := response["status"].(string); !ok || status != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", response["status"])
	}
}

// TestManagerHealthEndpoint tests the manager health endpoint
func TestManagerHealthEndpoint(t *testing.T) {
	_, managerSrv, cleanup := setupTestServers(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()

	managerSrv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if status, ok := response["status"].(string); !ok || status != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", response["status"])
	}
}

// TestCreateShard tests shard creation
func TestCreateShard(t *testing.T) {
	_, managerSrv, cleanup := setupTestServers(t)
	defer cleanup()

	shardReq := models.CreateShardRequest{
		Name:            "test-shard-1",
		PrimaryEndpoint: "localhost:5432",
		Replicas:        []string{"localhost:5433"},
		VNodeCount:      256,
	}

	body, _ := json.Marshal(shardReq)
	req := httptest.NewRequest("POST", "/api/v1/shards", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	managerSrv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var shard models.Shard
	if err := json.Unmarshal(w.Body.Bytes(), &shard); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if shard.ID == "" {
		t.Error("Expected shard ID to be set")
	}

	if shard.Name != shardReq.Name {
		t.Errorf("Expected shard name %s, got %s", shardReq.Name, shard.Name)
	}
}

// TestListShards tests shard listing
func TestListShards(t *testing.T) {
	_, managerSrv, cleanup := setupTestServers(t)
	defer cleanup()

	// Create a shard first
	shardReq := models.CreateShardRequest{
		Name:            "test-shard-list",
		PrimaryEndpoint: "localhost:5432",
		Replicas:        []string{},
		VNodeCount:      256,
	}

	body, _ := json.Marshal(shardReq)
	createReq := httptest.NewRequest("POST", "/api/v1/shards", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	managerSrv.Handler().ServeHTTP(createW, createReq)

	// List shards
	req := httptest.NewRequest("GET", "/api/v1/shards", nil)
	w := httptest.NewRecorder()

	managerSrv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var shards []models.Shard
	if err := json.Unmarshal(w.Body.Bytes(), &shards); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(shards) == 0 {
		t.Error("Expected at least one shard, got 0")
	}
}

// TestGetShardForKey tests shard lookup by key
func TestGetShardForKey(t *testing.T) {
	routerSrv, managerSrv, cleanup := setupTestServers(t)
	defer cleanup()

	// Create a shard first
	shardReq := models.CreateShardRequest{
		Name:            "test-shard-key",
		PrimaryEndpoint: "localhost:5432",
		Replicas:        []string{},
		VNodeCount:      256,
	}

	body, _ := json.Marshal(shardReq)
	createReq := httptest.NewRequest("POST", "/api/v1/shards", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	managerSrv.Handler().ServeHTTP(createW, createReq)

	// Get shard for key
	req := httptest.NewRequest("GET", "/v1/shard-for-key?key=test-key-123", nil)
	w := httptest.NewRecorder()

	routerSrv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if shardID, ok := response["shard_id"]; !ok || shardID == "" {
		t.Error("Expected shard_id in response")
	}
}

// TestCORSHeaders tests CORS headers are present
func TestCORSHeaders(t *testing.T) {
	routerSrv, _, cleanup := setupTestServers(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/v1/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	routerSrv.Handler().ServeHTTP(w, req)

	corsHeaders := []string{
		"Access-Control-Allow-Origin",
		"Access-Control-Allow-Methods",
		"Access-Control-Allow-Headers",
	}

	for _, header := range corsHeaders {
		if w.Header().Get(header) == "" {
			t.Errorf("Missing CORS header: %s", header)
		}
	}
}

// TestMetricsEndpoint tests metrics endpoint
func TestMetricsEndpoint(t *testing.T) {
	routerSrv, managerSrv, cleanup := setupTestServers(t)
	defer cleanup()

	// Test router metrics
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	routerSrv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for router metrics, got %d", w.Code)
	}

	// Test manager metrics
	req2 := httptest.NewRequest("GET", "/metrics", nil)
	w2 := httptest.NewRecorder()
	managerSrv.Handler().ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected status 200 for manager metrics, got %d", w2.Code)
	}
}

// TestErrorHandling tests error responses
func TestErrorHandling(t *testing.T) {
	routerSrv, _, cleanup := setupTestServers(t)
	defer cleanup()

	// Test invalid query request
	queryReq := models.QueryRequest{
		ShardKey: "", // Missing required field
		Query:    "SELECT 1",
	}

	body, _ := json.Marshal(queryReq)
	req := httptest.NewRequest("POST", "/v1/execute", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	routerSrv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d. Body: %s", w.Code, w.Body.String())
	}

	var errorResp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &errorResp); err != nil {
		t.Fatalf("Failed to parse error response: %v", err)
	}

	if _, ok := errorResp["error"]; !ok {
		t.Error("Expected error field in error response")
	}
}
