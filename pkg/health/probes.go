package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ProbeStatus represents the status of a health probe
type ProbeStatus string

const (
	ProbeStatusHealthy   ProbeStatus = "healthy"
	ProbeStatusUnhealthy ProbeStatus = "unhealthy"
	ProbeStatusDegraded  ProbeStatus = "degraded"
)

// ComponentHealth represents the health of a component
type ComponentHealth struct {
	Name      string      `json:"name"`
	Status    ProbeStatus `json:"status"`
	Message   string      `json:"message,omitempty"`
	LastCheck time.Time   `json:"last_check"`
	Details   interface{} `json:"details,omitempty"`
}

// HealthProbe defines the interface for health probes
type HealthProbe interface {
	Name() string
	Check(ctx context.Context) (*ComponentHealth, error)
}

// ProbeManager manages health probes for Kubernetes liveness and readiness
type ProbeManager struct {
	logger          *zap.Logger
	probes          map[string]HealthProbe
	livenessProbes  []string
	readinessProbes []string
	startupProbes   []string
	mu              sync.RWMutex
	componentHealth map[string]*ComponentHealth
	checkInterval   time.Duration
	startupComplete bool
	startupTimeout  time.Duration
	startedAt       time.Time
}

// ProbeManagerConfig holds configuration for the probe manager
type ProbeManagerConfig struct {
	CheckInterval  time.Duration
	StartupTimeout time.Duration
}

// NewProbeManager creates a new probe manager
func NewProbeManager(logger *zap.Logger, cfg ProbeManagerConfig) *ProbeManager {
	if cfg.CheckInterval == 0 {
		cfg.CheckInterval = 10 * time.Second
	}
	if cfg.StartupTimeout == 0 {
		cfg.StartupTimeout = 60 * time.Second
	}

	return &ProbeManager{
		logger:          logger,
		probes:          make(map[string]HealthProbe),
		livenessProbes:  make([]string, 0),
		readinessProbes: make([]string, 0),
		startupProbes:   make([]string, 0),
		componentHealth: make(map[string]*ComponentHealth),
		checkInterval:   cfg.CheckInterval,
		startupTimeout:  cfg.StartupTimeout,
		startedAt:       time.Now(),
	}
}

// RegisterProbe registers a health probe
func (pm *ProbeManager) RegisterProbe(probe HealthProbe, isLiveness, isReadiness, isStartup bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	name := probe.Name()
	pm.probes[name] = probe

	if isLiveness {
		pm.livenessProbes = append(pm.livenessProbes, name)
	}
	if isReadiness {
		pm.readinessProbes = append(pm.readinessProbes, name)
	}
	if isStartup {
		pm.startupProbes = append(pm.startupProbes, name)
	}

	pm.logger.Info("registered health probe", zap.String("name", name), zap.Bool("liveness", isLiveness), zap.Bool("readiness", isReadiness), zap.Bool("startup", isStartup))
}

// Start starts the probe manager
func (pm *ProbeManager) Start(ctx context.Context) {
	ticker := time.NewTicker(pm.checkInterval)
	defer ticker.Stop()

	pm.logger.Info("probe manager started", zap.Duration("interval", pm.checkInterval))
	pm.runAllProbes(ctx)

	for {
		select {
		case <-ctx.Done():
			pm.logger.Info("probe manager stopped")
			return
		case <-ticker.C:
			pm.runAllProbes(ctx)
		}
	}
}

func (pm *ProbeManager) runAllProbes(ctx context.Context) {
	pm.mu.RLock()
	probes := make([]HealthProbe, 0, len(pm.probes))
	for _, p := range pm.probes {
		probes = append(probes, p)
	}
	pm.mu.RUnlock()

	for _, probe := range probes {
		health, err := probe.Check(ctx)
		if err != nil {
			health = &ComponentHealth{Name: probe.Name(), Status: ProbeStatusUnhealthy, Message: err.Error(), LastCheck: time.Now()}
		}
		pm.mu.Lock()
		pm.componentHealth[probe.Name()] = health
		pm.mu.Unlock()
	}

	if !pm.startupComplete {
		pm.checkStartupComplete()
	}
}

func (pm *ProbeManager) checkStartupComplete() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.startupComplete {
		return
	}
	if time.Since(pm.startedAt) > pm.startupTimeout {
		pm.startupComplete = true
		pm.logger.Warn("startup timeout reached, marking startup as complete")
		return
	}
	for _, name := range pm.startupProbes {
		health, ok := pm.componentHealth[name]
		if !ok || health.Status != ProbeStatusHealthy {
			return
		}
	}
	pm.startupComplete = true
	pm.logger.Info("all startup probes passed, startup complete")
}

// MarkStartupComplete manually marks startup as complete
func (pm *ProbeManager) MarkStartupComplete() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.startupComplete = true
}

// LivenessHandler returns an HTTP handler for the liveness probe
func (pm *ProbeManager) LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		pm.mu.RLock()
		probeNames := pm.livenessProbes
		pm.mu.RUnlock()
		result := pm.checkProbes(ctx, probeNames)
		pm.writeResponse(w, result)
	}
}

// ReadinessHandler returns an HTTP handler for the readiness probe
func (pm *ProbeManager) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		pm.mu.RLock()
		probeNames := pm.readinessProbes
		pm.mu.RUnlock()
		result := pm.checkProbes(ctx, probeNames)
		pm.writeResponse(w, result)
	}
}

// StartupHandler returns an HTTP handler for the startup probe
func (pm *ProbeManager) StartupHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pm.mu.RLock()
		startupComplete := pm.startupComplete
		pm.mu.RUnlock()

		if startupComplete {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "ready", "message": "startup complete"})
			return
		}

		ctx := r.Context()
		pm.mu.RLock()
		probeNames := pm.startupProbes
		pm.mu.RUnlock()
		result := pm.checkProbes(ctx, probeNames)
		pm.writeResponse(w, result)
	}
}

// HealthHandler returns an HTTP handler for detailed health status
func (pm *ProbeManager) HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pm.mu.RLock()
		components := make(map[string]*ComponentHealth)
		for k, v := range pm.componentHealth {
			components[k] = v
		}
		startupComplete := pm.startupComplete
		pm.mu.RUnlock()

		overallStatus := ProbeStatusHealthy
		for _, health := range components {
			if health.Status == ProbeStatusUnhealthy {
				overallStatus = ProbeStatusUnhealthy
				break
			}
			if health.Status == ProbeStatusDegraded {
				overallStatus = ProbeStatusDegraded
			}
		}

		response := map[string]interface{}{"status": overallStatus, "startup_complete": startupComplete, "components": components, "timestamp": time.Now()}
		w.Header().Set("Content-Type", "application/json")
		if overallStatus == ProbeStatusUnhealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		json.NewEncoder(w).Encode(response)
	}
}

func (pm *ProbeManager) checkProbes(ctx context.Context, probeNames []string) *ProbeResult {
	result := &ProbeResult{Status: ProbeStatusHealthy, Components: make([]*ComponentHealth, 0, len(probeNames)), Timestamp: time.Now()}

	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, name := range probeNames {
		health, ok := pm.componentHealth[name]
		if !ok {
			health = &ComponentHealth{Name: name, Status: ProbeStatusUnhealthy, Message: "probe has not run yet", LastCheck: time.Time{}}
		}
		result.Components = append(result.Components, health)
		if health.Status == ProbeStatusUnhealthy {
			result.Status = ProbeStatusUnhealthy
		} else if health.Status == ProbeStatusDegraded && result.Status == ProbeStatusHealthy {
			result.Status = ProbeStatusDegraded
		}
	}
	return result
}

func (pm *ProbeManager) writeResponse(w http.ResponseWriter, result *ProbeResult) {
	w.Header().Set("Content-Type", "application/json")
	switch result.Status {
	case ProbeStatusHealthy, ProbeStatusDegraded:
		w.WriteHeader(http.StatusOK)
	case ProbeStatusUnhealthy:
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(result)
}

// ProbeResult represents the result of probe checks
type ProbeResult struct {
	Status     ProbeStatus        `json:"status"`
	Components []*ComponentHealth `json:"components"`
	Timestamp  time.Time          `json:"timestamp"`
}

// GetComponentHealth returns the health of a specific component
func (pm *ProbeManager) GetComponentHealth(name string) (*ComponentHealth, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	health, ok := pm.componentHealth[name]
	return health, ok
}

// DatabaseProbe implements health probe for database connectivity
type DatabaseProbe struct {
	name    string
	checkFn func(ctx context.Context) error
}

// NewDatabaseProbe creates a new database probe
func NewDatabaseProbe(name string, checkFn func(ctx context.Context) error) *DatabaseProbe {
	return &DatabaseProbe{name: name, checkFn: checkFn}
}

func (p *DatabaseProbe) Name() string { return p.name }

func (p *DatabaseProbe) Check(ctx context.Context) (*ComponentHealth, error) {
	health := &ComponentHealth{Name: p.name, LastCheck: time.Now()}
	if err := p.checkFn(ctx); err != nil {
		health.Status = ProbeStatusUnhealthy
		health.Message = err.Error()
		return health, nil
	}
	health.Status = ProbeStatusHealthy
	health.Message = "database is healthy"
	return health, nil
}

// CatalogProbe implements health probe for the catalog service
type CatalogProbe struct {
	name    string
	checkFn func(ctx context.Context) (bool, error)
}

// NewCatalogProbe creates a new catalog probe
func NewCatalogProbe(name string, checkFn func(ctx context.Context) (bool, error)) *CatalogProbe {
	return &CatalogProbe{name: name, checkFn: checkFn}
}

func (p *CatalogProbe) Name() string { return p.name }

func (p *CatalogProbe) Check(ctx context.Context) (*ComponentHealth, error) {
	health := &ComponentHealth{Name: p.name, LastCheck: time.Now()}
	healthy, err := p.checkFn(ctx)
	if err != nil {
		health.Status = ProbeStatusUnhealthy
		health.Message = err.Error()
		return health, nil
	}
	if healthy {
		health.Status = ProbeStatusHealthy
		health.Message = "catalog is healthy"
	} else {
		health.Status = ProbeStatusDegraded
		health.Message = "catalog is degraded"
	}
	return health, nil
}

// ExternalServiceProbe implements health probe for external service connectivity
type ExternalServiceProbe struct {
	name    string
	url     string
	timeout time.Duration
}

// NewExternalServiceProbe creates a new external service probe
func NewExternalServiceProbe(name, url string, timeout time.Duration) *ExternalServiceProbe {
	return &ExternalServiceProbe{name: name, url: url, timeout: timeout}
}

func (p *ExternalServiceProbe) Name() string { return p.name }

func (p *ExternalServiceProbe) Check(ctx context.Context) (*ComponentHealth, error) {
	health := &ComponentHealth{Name: p.name, LastCheck: time.Now()}
	client := &http.Client{Timeout: p.timeout}
	req, err := http.NewRequestWithContext(ctx, "GET", p.url, nil)
	if err != nil {
		health.Status = ProbeStatusUnhealthy
		health.Message = err.Error()
		return health, nil
	}
	resp, err := client.Do(req)
	if err != nil {
		health.Status = ProbeStatusUnhealthy
		health.Message = err.Error()
		return health, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		health.Status = ProbeStatusHealthy
		health.Message = "external service is reachable"
	} else {
		health.Status = ProbeStatusDegraded
		health.Message = "external service returned non-2xx status"
	}
	return health, nil
}

