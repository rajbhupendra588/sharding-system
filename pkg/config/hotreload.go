package config

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ReloadCallback is called when configuration changes
type ReloadCallback func(old, new *Config) error

// HotReloader watches configuration files and reloads on changes
type HotReloader struct {
	logger        *zap.Logger
	configPath    string
	currentConfig *Config
	currentHash   string
	callbacks     []ReloadCallback
	mu            sync.RWMutex
	checkInterval time.Duration
	stopCh        chan struct{}
}

// HotReloaderConfig holds configuration for the hot reloader
type HotReloaderConfig struct {
	ConfigPath    string
	CheckInterval time.Duration
}

// NewHotReloader creates a new configuration hot reloader
func NewHotReloader(logger *zap.Logger, cfg HotReloaderConfig) (*HotReloader, error) {
	if cfg.CheckInterval == 0 {
		cfg.CheckInterval = 10 * time.Second
	}

	config, err := LoadConfig(cfg.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load initial config: %w", err)
	}

	hash, err := calculateConfigHash(cfg.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate config hash: %w", err)
	}

	return &HotReloader{
		logger:        logger,
		configPath:    cfg.ConfigPath,
		currentConfig: config,
		currentHash:   hash,
		callbacks:     make([]ReloadCallback, 0),
		checkInterval: cfg.CheckInterval,
		stopCh:        make(chan struct{}),
	}, nil
}

// OnReload registers a callback to be called when configuration changes
func (hr *HotReloader) OnReload(callback ReloadCallback) {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	hr.callbacks = append(hr.callbacks, callback)
}

// GetConfig returns the current configuration
func (hr *HotReloader) GetConfig() *Config {
	hr.mu.RLock()
	defer hr.mu.RUnlock()
	return hr.currentConfig
}

// Start starts watching for configuration changes
func (hr *HotReloader) Start(ctx context.Context) {
	ticker := time.NewTicker(hr.checkInterval)
	defer ticker.Stop()

	hr.logger.Info("config hot-reload started", zap.String("path", hr.configPath), zap.Duration("interval", hr.checkInterval))

	for {
		select {
		case <-ctx.Done():
			hr.logger.Info("config hot-reload stopped")
			return
		case <-hr.stopCh:
			hr.logger.Info("config hot-reload stopped")
			return
		case <-ticker.C:
			if err := hr.checkAndReload(); err != nil {
				hr.logger.Error("failed to check/reload config", zap.Error(err))
			}
		}
	}
}

// Stop stops the hot reloader
func (hr *HotReloader) Stop() {
	close(hr.stopCh)
}

func (hr *HotReloader) checkAndReload() error {
	newHash, err := calculateConfigHash(hr.configPath)
	if err != nil {
		return fmt.Errorf("failed to calculate config hash: %w", err)
	}

	hr.mu.RLock()
	currentHash := hr.currentHash
	hr.mu.RUnlock()

	if newHash == currentHash {
		return nil
	}

	hr.logger.Info("configuration change detected, reloading", zap.String("old_hash", currentHash), zap.String("new_hash", newHash))

	newConfig, err := LoadConfig(hr.configPath)
	if err != nil {
		return fmt.Errorf("failed to load new config: %w", err)
	}

	if err := hr.validateConfig(newConfig); err != nil {
		hr.logger.Warn("new configuration is invalid, not reloading", zap.Error(err))
		return fmt.Errorf("invalid config: %w", err)
	}

	hr.mu.Lock()
	oldConfig := hr.currentConfig
	hr.mu.Unlock()

	hr.mu.RLock()
	callbacks := hr.callbacks
	hr.mu.RUnlock()

	for _, callback := range callbacks {
		if err := callback(oldConfig, newConfig); err != nil {
			hr.logger.Error("reload callback failed", zap.Error(err))
		}
	}

	hr.mu.Lock()
	hr.currentConfig = newConfig
	hr.currentHash = newHash
	hr.mu.Unlock()

	hr.logger.Info("configuration reloaded successfully")
	return nil
}

func (hr *HotReloader) validateConfig(cfg *Config) error {
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}
	if cfg.Sharding.VNodeCount < 1 {
		return fmt.Errorf("invalid vnode count: %d", cfg.Sharding.VNodeCount)
	}
	if cfg.Sharding.MaxConnections < 1 {
		return fmt.Errorf("invalid max connections: %d", cfg.Sharding.MaxConnections)
	}
	return nil
}

// ForceReload forces a configuration reload
func (hr *HotReloader) ForceReload() error {
	return hr.checkAndReload()
}

func calculateConfigHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// ConfigWatcher provides Kubernetes ConfigMap/Secret watching capabilities
type ConfigWatcher struct {
	logger        *zap.Logger
	hotReloader   *HotReloader
	configMapName string
	secretName    string
	namespace     string
	mu            sync.RWMutex
	configMapData map[string]string
	secretData    map[string][]byte
}

// ConfigWatcherConfig holds configuration for the ConfigMap watcher
type ConfigWatcherConfig struct {
	ConfigMapName string
	SecretName    string
	Namespace     string
}

// NewConfigWatcher creates a new ConfigMap/Secret watcher
func NewConfigWatcher(logger *zap.Logger, hotReloader *HotReloader, cfg ConfigWatcherConfig) *ConfigWatcher {
	return &ConfigWatcher{
		logger:        logger,
		hotReloader:   hotReloader,
		configMapName: cfg.ConfigMapName,
		secretName:    cfg.SecretName,
		namespace:     cfg.Namespace,
		configMapData: make(map[string]string),
		secretData:    make(map[string][]byte),
	}
}

// UpdateConfigMap updates configuration from a ConfigMap
func (cw *ConfigWatcher) UpdateConfigMap(data map[string]string) error {
	cw.mu.Lock()
	cw.configMapData = data
	cw.mu.Unlock()

	if configJSON, ok := data["manager.json"]; ok {
		return cw.updateFromJSON(configJSON)
	}
	return nil
}

// UpdateSecret updates configuration from a Secret
func (cw *ConfigWatcher) UpdateSecret(data map[string][]byte) error {
	cw.mu.Lock()
	cw.secretData = data
	cw.mu.Unlock()
	cw.logger.Info("secret values updated")
	return nil
}

func (cw *ConfigWatcher) updateFromJSON(configJSON string) error {
	var cfg Config
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return fmt.Errorf("invalid config JSON: %w", err)
	}

	tmpPath := filepath.Join(os.TempDir(), "sharding-config-temp.json")
	if err := os.WriteFile(tmpPath, []byte(configJSON), 0644); err != nil {
		return fmt.Errorf("failed to write temp config: %w", err)
	}

	cw.logger.Info("configuration updated from ConfigMap")
	return cw.hotReloader.ForceReload()
}

// GetConfigMapData returns current ConfigMap data
func (cw *ConfigWatcher) GetConfigMapData() map[string]string {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	result := make(map[string]string)
	for k, v := range cw.configMapData {
		result[k] = v
	}
	return result
}

// DynamicConfig provides dynamic configuration values that can change at runtime
type DynamicConfig struct {
	mu     sync.RWMutex
	values map[string]interface{}
	logger *zap.Logger
}

// NewDynamicConfig creates a new dynamic configuration
func NewDynamicConfig(logger *zap.Logger) *DynamicConfig {
	return &DynamicConfig{values: make(map[string]interface{}), logger: logger}
}

// Set sets a dynamic configuration value
func (dc *DynamicConfig) Set(key string, value interface{}) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.values[key] = value
	dc.logger.Debug("dynamic config updated", zap.String("key", key))
}

// Get gets a dynamic configuration value
func (dc *DynamicConfig) Get(key string) (interface{}, bool) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	val, ok := dc.values[key]
	return val, ok
}

// GetString gets a string configuration value
func (dc *DynamicConfig) GetString(key, defaultValue string) string {
	val, ok := dc.Get(key)
	if !ok {
		return defaultValue
	}
	if str, ok := val.(string); ok {
		return str
	}
	return defaultValue
}

// GetInt gets an integer configuration value
func (dc *DynamicConfig) GetInt(key string, defaultValue int) int {
	val, ok := dc.Get(key)
	if !ok {
		return defaultValue
	}
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	}
	return defaultValue
}

// GetBool gets a boolean configuration value
func (dc *DynamicConfig) GetBool(key string, defaultValue bool) bool {
	val, ok := dc.Get(key)
	if !ok {
		return defaultValue
	}
	if b, ok := val.(bool); ok {
		return b
	}
	return defaultValue
}

// GetDuration gets a duration configuration value
func (dc *DynamicConfig) GetDuration(key string, defaultValue time.Duration) time.Duration {
	val, ok := dc.Get(key)
	if !ok {
		return defaultValue
	}
	switch v := val.(type) {
	case time.Duration:
		return v
	case string:
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultValue
}

// All returns all dynamic configuration values
func (dc *DynamicConfig) All() map[string]interface{} {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	result := make(map[string]interface{})
	for k, v := range dc.values {
		result[k] = v
	}
	return result
}

