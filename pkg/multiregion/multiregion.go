package multiregion

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// RegionConfig holds configuration for a region
type RegionConfig struct {
	Name                string        `json:"name"`
	Endpoint            string        `json:"endpoint"`
	Priority            int           `json:"priority"`
	Weight              int           `json:"weight"`
	HealthCheckPath     string        `json:"health_check_path"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	IsLocal             bool          `json:"is_local"`
	Metadata            map[string]string `json:"metadata,omitempty"`
}

// RegionStatus represents the current status of a region
type RegionStatus struct {
	Name        string        `json:"name"`
	Endpoint    string        `json:"endpoint"`
	IsHealthy   bool          `json:"is_healthy"`
	IsPrimary   bool          `json:"is_primary"`
	Latency     time.Duration `json:"latency"`
	LastCheck   time.Time     `json:"last_check"`
	ErrorCount  int           `json:"error_count"`
	Connections int           `json:"active_connections"`
}

// MultiRegionManager manages cross-region sharding coordination
type MultiRegionManager struct {
	logger          *zap.Logger
	localRegion     string
	regions         map[string]*Region
	mu              sync.RWMutex
	primaryRegion   string
	failoverEnabled bool
	client          *http.Client
	stopCh          chan struct{}
}

// Region represents a single region
type Region struct {
	config RegionConfig
	status RegionStatus
	mu     sync.RWMutex
}

// MultiRegionConfig holds configuration for multi-region support
type MultiRegionConfig struct {
	LocalRegion     string
	Regions         []RegionConfig
	FailoverEnabled bool
	SyncInterval    time.Duration
}

// NewMultiRegionManager creates a new multi-region manager
func NewMultiRegionManager(logger *zap.Logger, cfg MultiRegionConfig) (*MultiRegionManager, error) {
	mrm := &MultiRegionManager{
		logger:          logger,
		localRegion:     cfg.LocalRegion,
		regions:         make(map[string]*Region),
		failoverEnabled: cfg.FailoverEnabled,
		client:          &http.Client{Timeout: 10 * time.Second},
		stopCh:          make(chan struct{}),
	}

	for _, regionCfg := range cfg.Regions {
		region := &Region{
			config: regionCfg,
			status: RegionStatus{Name: regionCfg.Name, Endpoint: regionCfg.Endpoint, IsHealthy: true, IsPrimary: regionCfg.Priority == 1},
		}
		mrm.regions[regionCfg.Name] = region
		if regionCfg.Priority == 1 {
			mrm.primaryRegion = regionCfg.Name
		}
	}
	return mrm, nil
}

// Start starts the multi-region manager
func (m *MultiRegionManager) Start(ctx context.Context) {
	m.logger.Info("multi-region manager started", zap.String("local_region", m.localRegion), zap.Int("total_regions", len(m.regions)))
	go m.healthCheckLoop(ctx)
	go m.syncLoop(ctx)
}

// Stop stops the multi-region manager
func (m *MultiRegionManager) Stop() {
	close(m.stopCh)
}

func (m *MultiRegionManager) healthCheckLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.checkAllRegions(ctx)
		}
	}
}

func (m *MultiRegionManager) checkAllRegions(ctx context.Context) {
	m.mu.RLock()
	regions := make([]*Region, 0, len(m.regions))
	for _, r := range m.regions {
		regions = append(regions, r)
	}
	m.mu.RUnlock()

	var wg sync.WaitGroup
	for _, region := range regions {
		wg.Add(1)
		go func(r *Region) {
			defer wg.Done()
			m.checkRegion(ctx, r)
		}(region)
	}
	wg.Wait()

	if m.failoverEnabled {
		m.checkFailover()
	}
}

func (m *MultiRegionManager) checkRegion(ctx context.Context, region *Region) {
	start := time.Now()
	healthURL := region.config.Endpoint + region.config.HealthCheckPath
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		m.updateRegionStatus(region, false, 0, err)
		return
	}
	resp, err := m.client.Do(req)
	latency := time.Since(start)
	if err != nil {
		m.updateRegionStatus(region, false, latency, err)
		return
	}
	defer resp.Body.Close()
	healthy := resp.StatusCode >= 200 && resp.StatusCode < 300
	m.updateRegionStatus(region, healthy, latency, nil)
}

func (m *MultiRegionManager) updateRegionStatus(region *Region, healthy bool, latency time.Duration, err error) {
	region.mu.Lock()
	defer region.mu.Unlock()
	region.status.LastCheck = time.Now()
	region.status.Latency = latency
	if healthy {
		region.status.IsHealthy = true
		region.status.ErrorCount = 0
	} else {
		region.status.ErrorCount++
		if region.status.ErrorCount >= 3 {
			region.status.IsHealthy = false
			m.logger.Warn("region marked unhealthy", zap.String("region", region.config.Name), zap.Int("error_count", region.status.ErrorCount), zap.Error(err))
		}
	}
}

func (m *MultiRegionManager) checkFailover() {
	m.mu.Lock()
	defer m.mu.Unlock()

	primaryRegion := m.regions[m.primaryRegion]
	if primaryRegion == nil {
		return
	}

	primaryRegion.mu.RLock()
	primaryHealthy := primaryRegion.status.IsHealthy
	primaryRegion.mu.RUnlock()

	if primaryHealthy {
		return
	}

	var bestRegion *Region
	bestPriority := 999
	for _, region := range m.regions {
		region.mu.RLock()
		healthy := region.status.IsHealthy
		priority := region.config.Priority
		region.mu.RUnlock()

		if healthy && priority < bestPriority && region.config.Name != m.primaryRegion {
			bestRegion = region
			bestPriority = priority
		}
	}

	if bestRegion != nil {
		m.logger.Info("initiating failover", zap.String("from", m.primaryRegion), zap.String("to", bestRegion.config.Name))
		primaryRegion.mu.Lock()
		primaryRegion.status.IsPrimary = false
		primaryRegion.mu.Unlock()
		bestRegion.mu.Lock()
		bestRegion.status.IsPrimary = true
		bestRegion.mu.Unlock()
		m.primaryRegion = bestRegion.config.Name
	}
}

func (m *MultiRegionManager) syncLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.syncState(ctx)
		}
	}
}

func (m *MultiRegionManager) syncState(ctx context.Context) {
	m.mu.RLock()
	regions := make([]*Region, 0)
	for _, r := range m.regions {
		if !r.config.IsLocal && r.status.IsHealthy {
			regions = append(regions, r)
		}
	}
	m.mu.RUnlock()

	for _, region := range regions {
		if err := m.syncWithRegion(ctx, region); err != nil {
			m.logger.Warn("failed to sync with region", zap.String("region", region.config.Name), zap.Error(err))
		}
	}
}

func (m *MultiRegionManager) syncWithRegion(ctx context.Context, region *Region) error {
	return nil
}

// GetPrimaryRegion returns the current primary region
func (m *MultiRegionManager) GetPrimaryRegion() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.primaryRegion
}

// GetRegionStatus returns status of a specific region
func (m *MultiRegionManager) GetRegionStatus(name string) (*RegionStatus, error) {
	m.mu.RLock()
	region, ok := m.regions[name]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("region not found: %s", name)
	}
	region.mu.RLock()
	status := region.status
	region.mu.RUnlock()
	return &status, nil
}

// GetAllRegionStatuses returns status of all regions
func (m *MultiRegionManager) GetAllRegionStatuses() []RegionStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	statuses := make([]RegionStatus, 0, len(m.regions))
	for _, region := range m.regions {
		region.mu.RLock()
		statuses = append(statuses, region.status)
		region.mu.RUnlock()
	}
	return statuses
}

// RouteToRegion determines which region should handle a request
func (m *MultiRegionManager) RouteToRegion(key string) (*RegionConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	primary := m.regions[m.primaryRegion]
	if primary != nil {
		primary.mu.RLock()
		healthy := primary.status.IsHealthy
		primary.mu.RUnlock()
		if healthy {
			return &primary.config, nil
		}
	}

	for _, region := range m.regions {
		region.mu.RLock()
		healthy := region.status.IsHealthy
		region.mu.RUnlock()
		if healthy {
			return &region.config, nil
		}
	}
	return nil, fmt.Errorf("no healthy regions available")
}

// IsLocalPrimary returns true if the local region is the primary
func (m *MultiRegionManager) IsLocalPrimary() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.localRegion == m.primaryRegion
}

// CrossRegionReplicator handles data replication across regions
type CrossRegionReplicator struct {
	logger         *zap.Logger
	localRegion    string
	remoteRegions  map[string]*RemoteRegion
	mu             sync.RWMutex
	replicationLag map[string]time.Duration
}

// RemoteRegion represents a remote region for replication
type RemoteRegion struct {
	Name       string
	Endpoint   string
	LastSync   time.Time
	SyncStatus string
	Lag        time.Duration
}

// NewCrossRegionReplicator creates a new cross-region replicator
func NewCrossRegionReplicator(logger *zap.Logger, localRegion string) *CrossRegionReplicator {
	return &CrossRegionReplicator{
		logger:         logger,
		localRegion:    localRegion,
		remoteRegions:  make(map[string]*RemoteRegion),
		replicationLag: make(map[string]time.Duration),
	}
}

// AddRemoteRegion adds a remote region for replication
func (r *CrossRegionReplicator) AddRemoteRegion(name, endpoint string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.remoteRegions[name] = &RemoteRegion{Name: name, Endpoint: endpoint, SyncStatus: "pending"}
}

// Replicate replicates data to remote regions
func (r *CrossRegionReplicator) Replicate(ctx context.Context, data interface{}) error {
	r.mu.RLock()
	regions := make([]*RemoteRegion, 0)
	for _, region := range r.remoteRegions {
		regions = append(regions, region)
	}
	r.mu.RUnlock()

	var errs []error
	for _, region := range regions {
		if err := r.replicateToRegion(ctx, region, data); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", region.Name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("replication errors: %v", errs)
	}
	return nil
}

func (r *CrossRegionReplicator) replicateToRegion(ctx context.Context, region *RemoteRegion, data interface{}) error {
	start := time.Now()
	r.mu.Lock()
	region.LastSync = time.Now()
	region.SyncStatus = "synced"
	region.Lag = time.Since(start)
	r.replicationLag[region.Name] = region.Lag
	r.mu.Unlock()
	return nil
}

// GetReplicationLag returns replication lag for all regions
func (r *CrossRegionReplicator) GetReplicationLag() map[string]time.Duration {
	r.mu.RLock()
	defer r.mu.RUnlock()
	lag := make(map[string]time.Duration)
	for k, v := range r.replicationLag {
		lag[k] = v
	}
	return lag
}

