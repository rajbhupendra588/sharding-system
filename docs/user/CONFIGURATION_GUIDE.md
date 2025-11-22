# Configuration Guide

## Overview

This guide explains how to configure the Sharding System. The system uses JSON configuration files for both the Manager and Router services.

## Configuration Files

The system uses two main configuration files:

- **`configs/manager.json`**: Configuration for the Shard Manager service (control plane)
- **`configs/router.json`**: Configuration for the Shard Router service (data plane)

## Configuration Structure

### Manager Configuration (`configs/manager.json`)

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8081,
    "read_timeout": "30s",
    "write_timeout": "30s",
    "idle_timeout": "120s"
  },
  "metadata": {
    "type": "etcd",
    "endpoints": ["localhost:2389"],
    "timeout": "5s"
  },
  "sharding": {
    "strategy": "hash",
    "hash_function": "murmur3",
    "vnode_count": 256,
    "replica_policy": "replica_ok",
    "max_connections": 100,
    "connection_ttl": "5m"
  },
  "security": {
    "enable_tls": false,
    "enable_rbac": true,
    "audit_log_path": "/var/log/sharding/audit.log",
    "user_database_dsn": ""
  },
  "observability": {
    "metrics_port": 9091,
    "enable_tracing": false,
    "log_level": "info"
  }
}
```

### Router Configuration (`configs/router.json`)

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8080,
    "read_timeout": "30s",
    "write_timeout": "30s",
    "idle_timeout": "120s"
  },
  "metadata": {
    "type": "etcd",
    "endpoints": ["localhost:2389"],
    "timeout": "5s"
  },
  "sharding": {
    "strategy": "hash",
    "hash_function": "murmur3",
    "vnode_count": 256,
    "replica_policy": "replica_ok",
    "max_connections": 100,
    "connection_ttl": "5m"
  },
  "security": {
    "enable_tls": false,
    "enable_rbac": true
  },
  "observability": {
    "metrics_port": 9090,
    "enable_tracing": false,
    "log_level": "info"
  }
}
```

## Configuration Parameters

### Server Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `host` | string | `"0.0.0.0"` | Host address to bind to |
| `port` | integer | `8081` (Manager)<br/>`8080` (Router) | Port to listen on |
| `read_timeout` | duration | `"30s"` | HTTP read timeout |
| `write_timeout` | duration | `"30s"` | HTTP write timeout |
| `idle_timeout` | duration | `"120s"` | HTTP idle connection timeout |

**Duration Format:** Use Go duration format (e.g., `"30s"`, `"5m"`, `"1h"`)

### Metadata Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `type` | string | `"etcd"` | Metadata store type (`"etcd"` or `"postgresql"`) |
| `endpoints` | array | `["localhost:2389"]` | Metadata store endpoints |
| `timeout` | duration | `"5s"` | Connection timeout |

**Example for PostgreSQL:**
```json
{
  "metadata": {
    "type": "postgresql",
    "endpoints": ["postgresql://user:pass@localhost:5432/metadata"],
    "timeout": "5s"
  }
}
```

### Sharding Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `strategy` | string | `"hash"` | Sharding strategy (currently only `"hash"` supported) |
| `hash_function` | string | `"murmur3"` | Hash function to use (`"murmur3"`) |
| `vnode_count` | integer | `256` | Number of virtual nodes per shard |
| `replica_policy` | string | `"replica_ok"` | Replica read policy |
| `max_connections` | integer | `100` | Maximum connections per shard |
| `connection_ttl` | duration | `"5m"` | Connection time-to-live |

**Virtual Nodes:** Higher values provide better load balancing but use more memory. Recommended range: 128-512.

### Security Configuration

#### Manager Security Options

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `enable_tls` | boolean | `false` | Enable TLS/SSL encryption |
| `enable_rbac` | boolean | `true` | Enable Role-Based Access Control |
| `audit_log_path` | string | `"/var/log/sharding/audit.log"` | Path to audit log file |
| `user_database_dsn` | string | `""` | User database connection string (for RBAC) |

#### Router Security Options

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `enable_tls` | boolean | `false` | Enable TLS/SSL encryption |
| `enable_rbac` | boolean | `true` | Enable Role-Based Access Control |

**Environment Variables:**
- `JWT_SECRET`: Required when RBAC is enabled (minimum 32 characters)
- `USER_DATABASE_DSN`: User database connection string (alternative to config file)

### Observability Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `metrics_port` | integer | `9091` (Manager)<br/>`9090` (Router) | Prometheus metrics port |
| `enable_tracing` | boolean | `false` | Enable distributed tracing |
| `log_level` | string | `"info"` | Logging level (`"debug"`, `"info"`, `"warn"`, `"error"`) |

**Log Levels:**
- `debug`: Verbose logging for development
- `info`: Standard production logging
- `warn`: Warnings and errors only
- `error`: Errors only

## Environment Variables

Some configuration can be overridden via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `JWT_SECRET` | JWT secret for authentication | Required if RBAC enabled |
| `USER_DATABASE_DSN` | User database connection string | From config file |
| `CONFIG_PATH` | Path to configuration file | `configs/manager.json` or `configs/router.json` |

## Configuration Examples

### Development Configuration

```json
{
  "server": {
    "host": "localhost",
    "port": 8081
  },
  "metadata": {
    "type": "etcd",
    "endpoints": ["localhost:2389"]
  },
  "security": {
    "enable_rbac": false
  },
  "observability": {
    "log_level": "debug"
  }
}
```

### Production Configuration

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8081,
    "read_timeout": "60s",
    "write_timeout": "60s"
  },
  "metadata": {
    "type": "etcd",
    "endpoints": [
      "etcd-1.example.com:2379",
      "etcd-2.example.com:2379",
      "etcd-3.example.com:2379"
    ],
    "timeout": "10s"
  },
  "sharding": {
    "vnode_count": 512,
    "max_connections": 200
  },
  "security": {
    "enable_tls": true,
    "enable_rbac": true,
    "audit_log_path": "/var/log/sharding/audit.log"
  },
  "observability": {
    "metrics_port": 9091,
    "log_level": "info"
  }
}
```

## Configuration Validation

The system validates configuration on startup. Common validation errors:

- **Invalid port**: Port must be between 1-65535
- **Invalid timeout**: Duration must be parseable (e.g., `"30s"`)
- **Missing endpoints**: Metadata endpoints array cannot be empty
- **Invalid log level**: Must be one of: `debug`, `info`, `warn`, `error`

## Reloading Configuration

Currently, configuration changes require a service restart. Dynamic configuration reloading may be added in future versions.

## Best Practices

1. **Use Environment Variables for Secrets**: Never commit passwords or secrets to configuration files
2. **Separate Configs by Environment**: Use different config files for dev, staging, and production
3. **Enable RBAC in Production**: Always enable RBAC (`enable_rbac: true`) in production
4. **Set Appropriate Timeouts**: Adjust timeouts based on your network latency and query complexity
5. **Monitor Metrics Port**: Ensure metrics port is accessible for monitoring but not publicly exposed
6. **Use TLS in Production**: Enable TLS (`enable_tls: true`) for production deployments
7. **Configure Audit Logging**: Set appropriate audit log path and ensure log rotation

## Troubleshooting

### Configuration File Not Found

**Error:** `Failed to load configuration: open configs/manager.json: no such file or directory`

**Solution:** Ensure the configuration file exists at the specified path, or set `CONFIG_PATH` environment variable.

### Invalid Configuration

**Error:** `Invalid configuration: port must be between 1-65535`

**Solution:** Check the configuration file for invalid values. Review the parameter descriptions above.

### RBAC Enabled but JWT_SECRET Missing

**Error:** `JWT_SECRET environment variable is required when RBAC is enabled`

**Solution:** Set the `JWT_SECRET` environment variable with at least 32 characters, or disable RBAC for development.

