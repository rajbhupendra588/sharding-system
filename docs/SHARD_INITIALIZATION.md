# Shard Initialization and Status Management

## Overview

This document explains how shards are created, loaded, and managed during application startup.

## Key Concepts

### 1. Shards Are NOT Auto-Created on Startup

**Important**: The sharding system does **NOT** automatically create shards when the application starts. Shards are **loaded from etcd** if they already exist.

### 2. Shard Creation Flow

Shards are created through the following mechanisms:

#### A. Via API (`POST /api/v1/shard`)
- User explicitly creates a shard through the REST API
- Status defaults to **"active"** if not specified

#### B. Via Database Creation (`POST /api/v1/databases`)
- When creating a sharded database, shards are automatically created
- Each shard gets status **"active"** by default

#### C. Via Kubernetes Operator
- When provisioning databases in Kubernetes, the operator creates shards
- Status is set to **"ready"** initially, then **"active"** when ready

## Startup Sequence

### Step 1: Catalog Initialization (`cmd/manager/main.go`)

```go
// Initialize catalog (connects to etcd)
cat, err := catalog.NewEtcdCatalog(cfg.Metadata.Endpoints, logger)
```

**What happens:**
- Connects to etcd at `localhost:2389`
- Calls `loadCatalog()` to load existing shards from etcd

### Step 2: Load Existing Shards (`pkg/catalog/catalog.go`)

```go
func (c *EtcdCatalog) loadCatalog() error {
    // Get all shards from etcd with prefix "/shards/"
    resp, err := c.client.Get(ctx, "/shards/", clientv3.WithPrefix())
    
    // For each shard in etcd:
    for _, kv := range resp.Kvs {
        var shard models.Shard
        json.Unmarshal(kv.Value, &shard)
        
        // Add to cache and hash ring
        c.cache[shard.ID] = &shard
        c.hashRing.addShard(&shard)
    }
}
```

**What happens:**
- Reads all shard data from etcd (stored under `/shards/{shard_id}`)
- Each shard retains its **existing status** (e.g., "active", "inactive", "migrating")
- Shards are loaded into memory cache and hash ring

### Step 3: Manager Initialization (`cmd/manager/main.go`)

```go
shardManager := manager.NewManager(cat, logger, resharderInstance, cfg.Pricing)
```

**What happens:**
- Manager is created with reference to the catalog
- Manager can now list/create/update shards

### Step 4: Register Active Shards for Monitoring (`internal/server/manager.go`)

```go
// Register existing active shards with Prometheus collector
registerExistingShardsForMetrics(shardManager, prometheusCollector, logger)

// Register existing active shards with PostgreSQL stats collector
registerExistingShards(shardManager, postgresStatsCollector, logger)
```

**What happens:**
- Lists all shards: `shardManager.ListShards()`
- Filters for shards with `status == "active"`
- Registers active shards with monitoring systems

## Shard Status Values

From `pkg/models/models.go`:

```go
type Shard struct {
    Status string `json:"status"` // "active", "migrating", "readonly", "inactive"
}
```

**Status Values:**
- **"active"**: Shard is operational and accepting traffic (default for new shards)
- **"migrating"**: Shard is being migrated (during resharding)
- **"readonly"**: Shard is read-only (e.g., during cutover)
- **"inactive"**: Shard is disabled/not in use

## Default Status Assignment

### When Creating a Shard (`pkg/manager/manager.go`)

```go
func (m *Manager) CreateShard(ctx context.Context, req *models.CreateShardRequest) (*models.Shard, error) {
    // Determine status
    status := req.Status
    if status == "" {
        status = "active"  // ← Default status is "active"
    }
    
    shard := &models.Shard{
        Status: status,
        // ... other fields
    }
}
```

**Key Point**: If no status is specified in the request, shards are created with **"active"** status by default.

## Why You See "Active" Shards on Startup

The shards you see with "Active" status on startup are:

1. **Previously Created Shards**: Shards that were created earlier (via API, database creation, etc.) and stored in etcd
2. **Persisted in etcd**: These shards persist across restarts because they're stored in etcd
3. **Loaded on Startup**: The catalog loads them from etcd and they retain their "active" status

## Example Flow

### First Time Startup (No Shards)
```
1. Manager starts
2. Catalog connects to etcd
3. loadCatalog() finds no shards in etcd
4. System starts with 0 shards
```

### Subsequent Startup (With Existing Shards)
```
1. Manager starts
2. Catalog connects to etcd
3. loadCatalog() finds 25 shards in etcd
4. All 25 shards loaded into memory
5. Active shards (e.g., 20) registered for monitoring
6. System ready with 25 shards visible
```

### Creating a New Shard
```
1. User calls POST /api/v1/shard
2. Manager.CreateShard() is called
3. Status defaults to "active" (if not specified)
4. Shard saved to etcd at /shards/{shard_id}
5. Shard added to hash ring
6. Shard now visible in system
```

## Verification

You can verify shards in etcd:

```bash
# List all shards in etcd
docker exec sharding-etcd etcdctl get /shards/ --prefix

# Check shard status via API
curl http://localhost:8081/api/v1/shards | jq '.[] | {id, name, status}'
```

## Summary

- **Shards are NOT auto-created on startup** - they're loaded from etcd
- **New shards default to "active" status** when created
- **Existing shards retain their status** when loaded from etcd
- **Only "active" shards** are registered for monitoring on startup
- **Shards persist in etcd** across application restarts

