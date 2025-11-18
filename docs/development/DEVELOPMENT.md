# Development Guide

## Prerequisites

- Go 1.21 or later
- Docker and Docker Compose
- etcd (or use Docker Compose)

## Local Development

### 1. Start Dependencies

```bash
docker-compose up -d etcd
```

### 2. Run Router

```bash
make run-router
# or
go run ./cmd/router
```

### 3. Run Manager

```bash
make run-manager
# or
go run ./cmd/manager
```

## Building

```bash
make build
```

This creates binaries in `bin/`:
- `bin/router` - Shard router service
- `bin/manager` - Shard manager service

## Docker Development

### Build Images

```bash
make docker-build
```

### Start All Services

```bash
make docker-up
```

### View Logs

```bash
make docker-logs
```

### Stop Services

```bash
make docker-down
```

## Testing

```bash
make test
```

## Project Structure

```
sharding-system/
├── cmd/              # Main applications
│   ├── router/       # Router service entry point
│   └── manager/      # Manager service entry point
├── pkg/              # Reusable packages
│   ├── catalog/      # Metadata catalog
│   ├── router/       # Query routing
│   ├── manager/      # Shard management
│   ├── resharder/    # Data migration
│   ├── health/       # Health monitoring
│   ├── client/       # Client library
│   ├── hashing/      # Consistent hashing
│   ├── models/       # Data models
│   ├── config/       # Configuration
│   ├── security/     # Security (auth, RBAC, audit)
│   └── observability/# Metrics and tracing
├── internal/         # Internal packages
│   └── api/          # HTTP handlers
├── configs/          # Configuration files
└── docs/             # Documentation
```

## Configuration

Configuration files are in `configs/`:
- `router.json` - Router configuration
- `manager.json` - Manager configuration

Configuration can be overridden via environment variables:
- `CONFIG_PATH` - Path to config file
- `METADATA_ENDPOINTS` - Comma-separated etcd endpoints

## Adding a New Shard

1. Create PostgreSQL instance
2. Call manager API to create shard:
```bash
curl -X POST http://localhost:8081/api/v1/shards \
  -H "Content-Type: application/json" \
  -d '{
    "name": "shard-01",
    "primary_endpoint": "postgres://user:pass@host:5432/db",
    "replicas": [],
    "vnode_count": 256
  }'
```

## Splitting a Shard

```bash
curl -X POST http://localhost:8081/api/v1/reshard/split \
  -H "Content-Type: application/json" \
  -d '{
    "source_shard_id": "shard-01",
    "target_shards": [...]
  }'
```

## Monitoring

- Router metrics: http://localhost:9090/metrics
- Manager metrics: http://localhost:9091/metrics

## Troubleshooting

### etcd Connection Issues
- Ensure etcd is running: `docker ps | grep etcd`
- Check etcd logs: `docker logs sharding-etcd`

### Router Can't Find Shards
- Verify shards exist: `curl http://localhost:8081/api/v1/shards`
- Check catalog version: Router should auto-refresh on catalog changes

### Migration Fails
- Check reshard job status: `curl http://localhost:8081/api/v1/reshard/jobs/{job_id}`
- Verify target shards are accessible
- Check logs for detailed error messages

