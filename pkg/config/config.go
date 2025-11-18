package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config holds the application configuration
type Config struct {
	Server     ServerConfig     `json:"server"`
	Metadata   MetadataConfig   `json:"metadata"`
	Sharding   ShardingConfig   `json:"sharding"`
	Security   SecurityConfig   `json:"security"`
	Observability ObservabilityConfig `json:"observability"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	ReadTimeout  time.Duration `json:"-"`
	WriteTimeout time.Duration `json:"-"`
	IdleTimeout  time.Duration `json:"-"`
	ReadTimeoutStr  string     `json:"read_timeout"`
	WriteTimeoutStr string     `json:"write_timeout"`
	IdleTimeoutStr  string     `json:"idle_timeout"`
}

// MetadataConfig holds metadata store configuration
type MetadataConfig struct {
	Type     string   `json:"type"` // "etcd" or "postgres"
	Endpoints []string `json:"endpoints"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Database string   `json:"database"`
	Timeout  time.Duration `json:"-"`
	TimeoutStr string `json:"timeout"`
}

// ShardingConfig holds sharding-specific configuration
type ShardingConfig struct {
	Strategy        string `json:"strategy"` // "hash" or "range"
	HashFunction    string `json:"hash_function"` // "murmur3" or "xxhash"
	VNodeCount      int    `json:"vnode_count"`
	ReplicaPolicy   string `json:"replica_policy"` // "primary" or "replica_ok"
	MaxConnections  int    `json:"max_connections"`
	ConnectionTTL   time.Duration `json:"-"`
	ConnectionTTLStr string `json:"connection_ttl"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	EnableTLS      bool   `json:"enable_tls"`
	TLSCertPath    string `json:"tls_cert_path"`
	TLSKeyPath     string `json:"tls_key_path"`
	EnableRBAC     bool   `json:"enable_rbac"`
	JWTSecret      string `json:"jwt_secret"`
	AuditLogPath   string `json:"audit_log_path"`
}

// ObservabilityConfig holds observability configuration
type ObservabilityConfig struct {
	MetricsPort    int    `json:"metrics_port"`
	EnableTracing  bool   `json:"enable_tracing"`
	TracingEndpoint string `json:"tracing_endpoint"`
	LogLevel       string `json:"log_level"`
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Parse duration strings
	if err := parseDurations(&config); err != nil {
		return nil, fmt.Errorf("failed to parse durations: %w", err)
	}

	// Set defaults
	setDefaults(&config)

	return &config, nil
}

// parseDurations parses duration strings into time.Duration
func parseDurations(c *Config) error {
	var err error

	// Parse server timeouts
	if c.Server.ReadTimeoutStr != "" {
		c.Server.ReadTimeout, err = time.ParseDuration(c.Server.ReadTimeoutStr)
		if err != nil {
			return fmt.Errorf("invalid read_timeout: %w", err)
		}
	}
	if c.Server.WriteTimeoutStr != "" {
		c.Server.WriteTimeout, err = time.ParseDuration(c.Server.WriteTimeoutStr)
		if err != nil {
			return fmt.Errorf("invalid write_timeout: %w", err)
		}
	}
	if c.Server.IdleTimeoutStr != "" {
		c.Server.IdleTimeout, err = time.ParseDuration(c.Server.IdleTimeoutStr)
		if err != nil {
			return fmt.Errorf("invalid idle_timeout: %w", err)
		}
	}

	// Parse metadata timeout
	if c.Metadata.TimeoutStr != "" {
		c.Metadata.Timeout, err = time.ParseDuration(c.Metadata.TimeoutStr)
		if err != nil {
			return fmt.Errorf("invalid metadata timeout: %w", err)
		}
	}

	// Parse sharding connection TTL
	if c.Sharding.ConnectionTTLStr != "" {
		c.Sharding.ConnectionTTL, err = time.ParseDuration(c.Sharding.ConnectionTTLStr)
		if err != nil {
			return fmt.Errorf("invalid connection_ttl: %w", err)
		}
	}

	return nil
}

func setDefaults(c *Config) {
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = 30 * time.Second
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = 30 * time.Second
	}
	if c.Server.IdleTimeout == 0 {
		c.Server.IdleTimeout = 120 * time.Second
	}
	if c.Sharding.Strategy == "" {
		c.Sharding.Strategy = "hash"
	}
	if c.Sharding.HashFunction == "" {
		c.Sharding.HashFunction = "murmur3"
	}
	if c.Sharding.VNodeCount == 0 {
		c.Sharding.VNodeCount = 256
	}
	if c.Sharding.MaxConnections == 0 {
		c.Sharding.MaxConnections = 100
	}
	if c.Sharding.ConnectionTTL == 0 {
		c.Sharding.ConnectionTTL = 5 * time.Minute
	}
	if c.Observability.MetricsPort == 0 {
		c.Observability.MetricsPort = 9090
	}
	if c.Observability.LogLevel == "" {
		c.Observability.LogLevel = "info"
	}
}

