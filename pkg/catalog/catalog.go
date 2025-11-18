package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sharding-system/pkg/hashing"
	"github.com/sharding-system/pkg/models"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

// Catalog manages shard metadata and routing information
type Catalog interface {
	GetShard(key string) (*models.Shard, error)
	GetShardByID(shardID string) (*models.Shard, error)
	ListShards() ([]models.Shard, error)
	CreateShard(shard *models.Shard) error
	UpdateShard(shard *models.Shard) error
	DeleteShard(shardID string) error
	GetCatalogVersion() (int64, error)
	Watch(ctx context.Context) (<-chan *models.ShardCatalog, error)
}

// EtcdCatalog implements Catalog using etcd
type EtcdCatalog struct {
	client     *clientv3.Client
	logger     *zap.Logger
	hashRing   *ConsistentHashRing
	mu         sync.RWMutex
	cache      map[string]*models.Shard
	version    int64
	watchChan  chan *models.ShardCatalog
}

// ConsistentHashRing wraps the hashing logic with catalog integration
type ConsistentHashRing struct {
	hashFunc *hashing.ConsistentHash
	shards   map[string]*models.Shard
	mu       sync.RWMutex
}

// NewEtcdCatalog creates a new etcd-based catalog
func NewEtcdCatalog(endpoints []string, logger *zap.Logger) (*EtcdCatalog, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	catalog := &EtcdCatalog{
		client:    client,
		logger:    logger,
		hashRing:  &ConsistentHashRing{shards: make(map[string]*models.Shard)},
		cache:     make(map[string]*models.Shard),
		watchChan: make(chan *models.ShardCatalog, 10),
	}

	// Load initial catalog
	if err := catalog.loadCatalog(); err != nil {
		logger.Warn("failed to load initial catalog", zap.Error(err))
	}

	return catalog, nil
}

// GetShard returns the shard for a given key
func (c *EtcdCatalog) GetShard(key string) (*models.Shard, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	c.hashRing.mu.RLock()
	defer c.hashRing.mu.RUnlock()

	shardID := c.hashRing.hashFunc.GetShard(key)
	if shardID == "" {
		return nil, fmt.Errorf("no shard found for key: %s", key)
	}

	shard, exists := c.cache[shardID]
	if !exists {
		return nil, fmt.Errorf("shard %s not found in cache", shardID)
	}

	return shard, nil
}

// GetShardByID returns a shard by its ID
func (c *EtcdCatalog) GetShardByID(shardID string) (*models.Shard, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	shard, exists := c.cache[shardID]
	if !exists {
		return nil, fmt.Errorf("shard %s not found", shardID)
	}

	return shard, nil
}

// ListShards returns all shards
func (c *EtcdCatalog) ListShards() ([]models.Shard, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	shards := make([]models.Shard, 0, len(c.cache))
	for _, shard := range c.cache {
		shards = append(shards, *shard)
	}

	return shards, nil
}

// CreateShard creates a new shard
func (c *EtcdCatalog) CreateShard(shard *models.Shard) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	shardData, err := json.Marshal(shard)
	if err != nil {
		return fmt.Errorf("failed to marshal shard: %w", err)
	}

	key := fmt.Sprintf("/shards/%s", shard.ID)
	
	// Use transaction to ensure atomicity
	txn := c.client.Txn(ctx)
	txn.If(clientv3.Compare(clientv3.Version(key), "=", 0)).
		Then(clientv3.OpPut(key, string(shardData))).
		Else(clientv3.OpGet(key))

	resp, err := txn.Commit()
	if err != nil {
		return fmt.Errorf("failed to create shard in etcd: %w", err)
	}

	if !resp.Succeeded {
		return fmt.Errorf("shard %s already exists", shard.ID)
	}

	// Update local cache and hash ring
	c.cache[shard.ID] = shard
	c.hashRing.addShard(shard)
	c.version++

	c.logger.Info("created shard", zap.String("shard_id", shard.ID))
	return nil
}

// UpdateShard updates an existing shard
func (c *EtcdCatalog) UpdateShard(shard *models.Shard) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	shard.UpdatedAt = time.Now()
	shard.Version++
	shardData, err := json.Marshal(shard)
	if err != nil {
		return fmt.Errorf("failed to marshal shard: %w", err)
	}

	key := fmt.Sprintf("/shards/%s", shard.ID)
	_, err = c.client.Put(ctx, key, string(shardData))
	if err != nil {
		return fmt.Errorf("failed to update shard in etcd: %w", err)
	}

	// Update local cache
	c.cache[shard.ID] = shard
	c.version++

	c.logger.Info("updated shard", zap.String("shard_id", shard.ID))
	return nil
}

// DeleteShard deletes a shard
func (c *EtcdCatalog) DeleteShard(shardID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("/shards/%s", shardID)
	_, err := c.client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete shard from etcd: %w", err)
	}

	// Remove from cache and hash ring
	delete(c.cache, shardID)
	c.hashRing.removeShard(shardID)
	c.version++

	c.logger.Info("deleted shard", zap.String("shard_id", shardID))
	return nil
}

// GetCatalogVersion returns the current catalog version
func (c *EtcdCatalog) GetCatalogVersion() (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.version, nil
}

// Watch watches for catalog changes
func (c *EtcdCatalog) Watch(ctx context.Context) (<-chan *models.ShardCatalog, error) {
	watchChan := make(chan *models.ShardCatalog, 10)

	go func() {
		defer close(watchChan)
		
		watchResp := c.client.Watch(ctx, "/shards/", clientv3.WithPrefix())
		for watchResp := range watchResp {
			for _, ev := range watchResp.Events {
				if ev.Type == clientv3.EventTypePut || ev.Type == clientv3.EventTypeDelete {
					// Reload catalog on change
					if err := c.loadCatalog(); err != nil {
						c.logger.Error("failed to reload catalog", zap.Error(err))
						continue
					}

					shards, _ := c.ListShards()
					catalog := &models.ShardCatalog{
						Version:   c.version,
						Shards:    shards,
						UpdatedAt: time.Now(),
					}
					select {
					case watchChan <- catalog:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return watchChan, nil
}

// loadCatalog loads the catalog from etcd
func (c *EtcdCatalog) loadCatalog() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.Get(ctx, "/shards/", clientv3.WithPrefix())
	if err != nil {
		return fmt.Errorf("failed to get shards from etcd: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*models.Shard)
	c.hashRing.mu.Lock()
	c.hashRing.shards = make(map[string]*models.Shard)
	c.hashRing.mu.Unlock()

	for _, kv := range resp.Kvs {
		var shard models.Shard
		if err := json.Unmarshal(kv.Value, &shard); err != nil {
			c.logger.Warn("failed to unmarshal shard", zap.Error(err))
			continue
		}

		c.cache[shard.ID] = &shard
		c.hashRing.addShard(&shard)
	}

	c.version = resp.Header.Revision
	return nil
}

// addShard adds a shard to the hash ring
func (r *ConsistentHashRing) addShard(shard *models.Shard) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.hashFunc == nil {
		r.hashFunc = hashing.NewConsistentHash(hashing.NewHashFunction("murmur3"))
	}

	vnodeCount := len(shard.VNodes)
	if vnodeCount == 0 {
		vnodeCount = 256 // default
	}

	r.hashFunc.AddShard(shard.ID, vnodeCount)
	r.shards[shard.ID] = shard
}

// removeShard removes a shard from the hash ring
func (r *ConsistentHashRing) removeShard(shardID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.hashFunc != nil {
		r.hashFunc.RemoveShard(shardID)
	}
	delete(r.shards, shardID)
}

