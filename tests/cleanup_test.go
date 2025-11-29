package tests

import (
	"context"
	"testing"
	"time"

	"github.com/sharding-system/pkg/catalog"
	"github.com/sharding-system/pkg/config"
	"github.com/sharding-system/pkg/manager"
	"github.com/sharding-system/pkg/models"
	"go.uber.org/zap"
)

// MockResharder implements manager.Resharder
type MockResharder struct{}

func (m *MockResharder) Split(ctx context.Context, job *models.ReshardJob) error { return nil }
func (m *MockResharder) Merge(ctx context.Context, job *models.ReshardJob) error { return nil }

func TestCleanupLogic(t *testing.T) {
	// Setup
	logger, _ := zap.NewDevelopment()
	// Use EtcdCatalog if available, or we can use a memory catalog if we implemented one.
	// Since we don't have a memory catalog in the codebase (based on file list), we'll try to use EtcdCatalog
	// pointing to the local Etcd instance we started.
	
	cfg := config.MetadataConfig{
		Type:      "etcd",
		Endpoints: []string{"localhost:2389"},
		Timeout:   5 * time.Second,
	}
	
	cat, err := catalog.NewEtcdCatalog(cfg.Endpoints, logger)
	if err != nil {
		t.Fatalf("Failed to create catalog: %v", err)
	}
	
	mgr := manager.NewManager(cat, logger, &MockResharder{}, config.PricingConfig{Tier: "free"})
	clientAppMgr := mgr.GetClientAppManager()
	
	ctx := context.Background()
	
	// 1. Register Client App
	appName := "TestCleanupApp"
	app, err := clientAppMgr.RegisterClientApp(ctx, appName, "Desc", "db", "host", "5432", "user", "pass", "prefix:", "ns", "cluster")
	if err != nil {
		// If it already exists from previous run, try to get it or delete it
		// For simplicity, we assume clean state or ignore error if it's "already exists"
		// But better to use unique name
		appName = "TestCleanupApp_" + time.Now().Format("20060102150405")
		app, err = clientAppMgr.RegisterClientApp(ctx, appName, "Desc", "db", "host", "5432", "user", "pass", "prefix:", "ns", "cluster")
		if err != nil {
			t.Fatalf("Failed to register client app: %v", err)
		}
	}
	t.Logf("Registered Client App: %s (%s)", app.Name, app.ID)
	
	// 2. Create Shard
	shardReq := &models.CreateShardRequest{
		Name:            "TestShard",
		ClientAppID:     app.ID,
		PrimaryEndpoint: "localhost:5432",
		VNodeCount:      10,
	}
	shard, err := mgr.CreateShard(ctx, shardReq)
	if err != nil {
		t.Fatalf("Failed to create shard: %v", err)
	}
	t.Logf("Created Shard: %s (%s) Status: %s", shard.Name, shard.ID, shard.Status)
	
	// 3. Attempt to Delete Active Shard (Should Fail)
	err = mgr.DeleteShard(shard.ID)
	if err == nil {
		t.Errorf("Expected DeleteShard to fail for active shard, but it succeeded")
	} else {
		t.Logf("DeleteShard failed as expected: %v", err)
	}
	
	// 4. Manually set Shard to Inactive (Simulating admin action)
	// We need to use catalog directly as Manager doesn't expose UpdateShard
	shard.Status = "inactive"
	err = cat.UpdateShard(shard)
	if err != nil {
		t.Fatalf("Failed to update shard status: %v", err)
	}
	t.Log("Updated shard status to inactive")
	
	// 5. Delete Shard (Should Succeed)
	err = mgr.DeleteShard(shard.ID)
	if err != nil {
		t.Errorf("Failed to delete inactive shard: %v", err)
	} else {
		t.Log("Successfully deleted inactive shard")
	}
	
	// 6. Verify Shard is gone
	_, err = mgr.GetShard(shard.ID)
	if err == nil {
		t.Errorf("Shard still exists after deletion")
	} else {
		t.Log("Verified shard is gone")
	}
	
	// 7. Delete Client App
	err = clientAppMgr.DeleteClientApp(app.ID)
	if err != nil {
		t.Errorf("Failed to delete client app: %v", err)
	} else {
		t.Log("Successfully deleted client app")
	}
	
	// 8. Verify Client App is gone
	_, err = clientAppMgr.GetClientApp(app.ID)
	if err == nil {
		t.Errorf("Client app still exists after deletion")
	} else {
		t.Log("Verified client app is gone")
	}
}
