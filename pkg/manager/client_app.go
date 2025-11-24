package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sharding-system/pkg/catalog"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

// ClientAppManager manages client applications
type ClientAppManager struct {
	catalog    catalog.Catalog
	logger     *zap.Logger
	mu         sync.RWMutex
	clientApps map[string]*ClientAppInfo
	etcdClient *clientv3.Client // optional etcd client for persistence
}

// ClientAppInfo tracks information about a client application
type ClientAppInfo struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description,omitempty"`
	DatabaseName     string    `json:"database_name,omitempty"`     // Database name for sharding
	DatabaseHost     string    `json:"database_host,omitempty"`     // Database host
	DatabasePort     string    `json:"database_port,omitempty"`     // Database port
	DatabaseUser     string    `json:"database_user,omitempty"`     // Database user
	DatabasePassword string    `json:"database_password,omitempty"` // Database password
	Namespace        string    `json:"namespace,omitempty"`         // Kubernetes namespace
	ClusterName      string    `json:"cluster_name,omitempty"`      // Kubernetes cluster name
	Status           string    `json:"status"`                      // "active", "inactive"
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	LastSeen         time.Time `json:"last_seen"`
	// Track which shards this client uses
	ShardIDs []string `json:"shard_ids"`
	// Track request patterns
	RequestCount int64 `json:"request_count"`
	// Client identifier pattern (e.g., "app1:", "app2:")
	KeyPrefix string `json:"key_prefix,omitempty"`
}

// NewClientAppManager creates a new client application manager
func NewClientAppManager(catalogInst catalog.Catalog, logger *zap.Logger) *ClientAppManager {
	mgr := &ClientAppManager{
		catalog:    catalogInst,
		logger:     logger,
		clientApps: make(map[string]*ClientAppInfo),
	}
	// If using EtcdCatalog, capture the etcd client for persistence
	if etcdCat, ok := catalogInst.(*catalog.EtcdCatalog); ok {
		mgr.etcdClient = etcdCat.GetEtcdClient()
		// Load existing client apps from etcd
		if err := mgr.loadClientApps(); err != nil {
			logger.Error("failed to load client apps from etcd", zap.Error(err))
		}
	}
	return mgr
}

// RegisterClientApp registers a new client application
func (m *ClientAppManager) RegisterClientApp(ctx context.Context, name, description, databaseName, databaseHost, databasePort, databaseUser, databasePassword, keyPrefix, namespace, clusterName string) (*ClientAppInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if client with same name or key prefix exists
	for _, app := range m.clientApps {
		if app.Name == name {
			return nil, fmt.Errorf("client application with name '%s' already exists", name)
		}
		if keyPrefix != "" && app.KeyPrefix == keyPrefix {
			return nil, fmt.Errorf("client application with key prefix '%s' already exists", keyPrefix)
		}
	}

	app := &ClientAppInfo{
		ID:               uuid.New().String(),
		Name:             name,
		Description:      description,
		DatabaseName:     databaseName,
		DatabaseHost:     databaseHost,
		DatabasePort:     databasePort,
		DatabaseUser:     databaseUser,
		DatabasePassword: databasePassword,
		Namespace:        namespace,
		ClusterName:      clusterName,
		Status:           "active",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		LastSeen:         time.Now(),
		ShardIDs:         []string{},
		RequestCount:     0,
		KeyPrefix:        keyPrefix,
	}

	m.clientApps[app.ID] = app
	// Persist to etcd if possible
	if m.etcdClient != nil {
		if err := m.persistClientApp(app); err != nil {
			m.logger.Error("failed to persist client app to etcd", zap.Error(err))
		}
	}
	m.logger.Info("registered client application", zap.String("id", app.ID), zap.String("name", app.Name))

	return app, nil
}

// TrackRequest tracks a request from a client application
func (m *ClientAppManager) TrackRequest(shardKey string, shardID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var clientApp *ClientAppInfo

	// Try to identify client from shard key pattern
	if shardKey != "" {
		clientApp = m.identifyClientFromKey(shardKey)
	}

	// If no client found and we have a default app, use it
	if clientApp == nil {
		// Find default client app (the one without a key prefix)
		for _, app := range m.clientApps {
			if app.KeyPrefix == "" {
				clientApp = app
				break
			}
		}
	}

	// If still no client app, skip tracking
	if clientApp == nil {
		return
	}

	// Update client app info
	clientApp.LastSeen = time.Now()
	if shardKey != "" {
		clientApp.RequestCount++
	}
	clientApp.UpdatedAt = time.Now()

	// Track shard usage
	shardFound := false
	for _, sid := range clientApp.ShardIDs {
		if sid == shardID {
			shardFound = true
			break
		}
	}
	if !shardFound && shardID != "" {
		clientApp.ShardIDs = append(clientApp.ShardIDs, shardID)
	}
}

// identifyClientFromKey tries to identify client from shard key pattern
func (m *ClientAppManager) identifyClientFromKey(shardKey string) *ClientAppInfo {
	// Look for key prefix pattern (e.g., "app1:", "app2:")
	for _, app := range m.clientApps {
		if app.KeyPrefix != "" && strings.HasPrefix(shardKey, app.KeyPrefix) {
			return app
		}
	}

	// Try to extract from common patterns: "appX:", "clientX:", etc.
	parts := strings.Split(shardKey, ":")
	if len(parts) > 0 {
		prefix := parts[0] + ":"
		for _, app := range m.clientApps {
			if app.KeyPrefix == prefix {
				return app
			}
		}
	}

	return nil
}

// ListClientApps returns all registered client applications
func (m *ClientAppManager) ListClientApps() ([]*ClientAppInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	apps := make([]*ClientAppInfo, 0, len(m.clientApps))
	for _, app := range m.clientApps {
		// Create a copy to avoid race conditions
		appCopy := *app
		// Copy slice to avoid sharing
		appCopy.ShardIDs = make([]string, len(app.ShardIDs))
		copy(appCopy.ShardIDs, app.ShardIDs)
		apps = append(apps, &appCopy)
	}

	return apps, nil
}

// GetClientApp returns a client application by ID
func (m *ClientAppManager) GetClientApp(id string) (*ClientAppInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	app, exists := m.clientApps[id]
	if !exists {
		return nil, fmt.Errorf("client application not found: %s", id)
	}

	return app, nil
}

// UpdateClientAppStatus updates the status of a client application
func (m *ClientAppManager) UpdateClientAppStatus(id string, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	app, exists := m.clientApps[id]
	if !exists {
		return fmt.Errorf("client application not found: %s", id)
	}

	app.Status = status
	app.UpdatedAt = time.Now()
	m.logger.Info("updated client app status", zap.String("id", id), zap.String("status", status))

	return nil
}

// DeleteClientApp removes a client application
func (m *ClientAppManager) DeleteClientApp(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.clientApps[id]; !exists {
		return fmt.Errorf("client application not found: %s", id)
	}

	delete(m.clientApps, id)
	// Remove from etcd if persisted
	if m.etcdClient != nil {
		key := fmt.Sprintf("/client_apps/%s", id)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := m.etcdClient.Delete(ctx, key); err != nil {
			m.logger.Error("failed to delete client app from etcd", zap.Error(err))
		}
	}
	m.logger.Info("deleted client application", zap.String("id", id))

	return nil
}

// DiscoverClientApps analyzes shard keys to discover client applications
func (m *ClientAppManager) DiscoverClientApps(shardKeys []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	prefixMap := make(map[string]int)

	// Count occurrences of each prefix
	for _, key := range shardKeys {
		parts := strings.Split(key, ":")
		if len(parts) > 0 {
			prefix := parts[0] + ":"
			prefixMap[prefix]++
		}
	}

	// Auto-register clients with significant usage
	for prefix, count := range prefixMap {
		if count >= 10 { // Threshold for auto-discovery
			// Check if already registered
			found := false
			for _, app := range m.clientApps {
				if app.KeyPrefix == prefix {
					found = true
					break
				}
			}

			if !found {
				app := &ClientAppInfo{
					ID:               uuid.New().String(),
					Name:             fmt.Sprintf("Client-%s", strings.TrimSuffix(prefix, ":")),
					Description:      fmt.Sprintf("Auto-discovered client with prefix '%s'", prefix),
					DatabaseName:     "",
					DatabaseHost:     "",
					DatabasePort:     "",
					DatabaseUser:     "",
					DatabasePassword: "",
					Status:           "active",
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
					LastSeen:         time.Now(),
					ShardIDs:         []string{},
					RequestCount:     int64(count),
					KeyPrefix:        prefix,
				}
				m.clientApps[app.ID] = app
				if m.etcdClient != nil {
					if err := m.persistClientApp(app); err != nil {
						m.logger.Error("failed to persist auto-discovered client app", zap.Error(err))
					}
				}
				m.logger.Info("auto-discovered client application", zap.String("prefix", prefix), zap.Int("requests", count))
			}
		}
	}
}

// persistClientApp stores a client app in etcd
func (m *ClientAppManager) persistClientApp(app *ClientAppInfo) error {
	if m.etcdClient == nil {
		return nil // nothing to persist
	}
	data, err := json.Marshal(app)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("/client_apps/%s", app.ID)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = m.etcdClient.Put(ctx, key, string(data))
	return err
}

// loadClientApps loads all client apps from etcd into memory
func (m *ClientAppManager) loadClientApps() error {
	if m.etcdClient == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := m.etcdClient.Get(ctx, "/client_apps/", clientv3.WithPrefix())
	if err != nil {
		return err
	}
	for _, kv := range resp.Kvs {
		var app ClientAppInfo
		if err := json.Unmarshal(kv.Value, &app); err != nil {
			m.logger.Error("failed to unmarshal client app from etcd", zap.Error(err))
			continue
		}
		m.clientApps[app.ID] = &app
	}
	return nil
}
