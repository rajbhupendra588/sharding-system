# Quick Start Guide

## Prerequisites

- Docker and Docker Compose installed
- Go 1.21+ (for local development)

## Quick Start with Docker Compose

1. **Start all services:**
```bash
docker-compose up -d
```

This starts:
- etcd (metadata store)
- Router service (port 8080)
- Manager service (port 8081)
- Two PostgreSQL shards (ports 5432, 5433)

2. **Verify services are running:**
```bash
docker-compose ps
```

3. **Check logs:**
```bash
docker-compose logs -f router
docker-compose logs -f manager
```

## Create Your First Shard

```bash
curl -X POST http://localhost:8081/api/v1/shards \
  -H "Content-Type: application/json" \
  -d '{
    "name": "shard-01",
    "primary_endpoint": "postgres://postgres:postgres@postgres-shard1:5432/shard1",
    "replicas": [],
    "vnode_count": 256
  }'
```

## Query a Shard

```bash
curl -X POST http://localhost:8080/v1/execute \
  -H "Content-Type: application/json" \
  -d '{
    "shard_key": "user-123",
    "query": "SELECT version()",
    "params": [],
    "consistency": "strong"
  }'
```

## Get Shard for a Key

```bash
curl http://localhost:8080/v1/shard-for-key?key=user-123
```

## List All Shards

```bash
curl http://localhost:8081/api/v1/shards
```

## Split a Shard

First, create target shards, then:

```bash
curl -X POST http://localhost:8081/api/v1/reshard/split \
  -H "Content-Type: application/json" \
  -d '{
    "source_shard_id": "shard-01-id",
    "target_shards": [
      {
        "name": "shard-02",
        "primary_endpoint": "postgres://postgres:postgres@postgres-shard2:5432/shard2",
        "replicas": [],
        "vnode_count": 256
      }
    ]
  }'
```

## Monitor Metrics

- Router metrics: http://localhost:9090/metrics
- Manager metrics: http://localhost:9091/metrics

## Using the Client Library

See `examples/client_example.go` for a complete example.

```go
import "github.com/sharding-system/pkg/client"

client := client.NewClient("http://localhost:8080")
shardID, err := client.GetShardForKey("user-123")
result, err := client.QueryStrong("user-123", "SELECT * FROM users WHERE id = $1", "user-123")
```

## Stop Services

```bash
docker-compose down
```

## Local Development

1. **Start etcd:**
```bash
docker-compose up -d etcd
```

2. **Run router:**
```bash
go run ./cmd/router
```

3. **Run manager (in another terminal):**
```bash
go run ./cmd/manager
```

## Next Steps

- Read [API Documentation](docs/API.md) for detailed API reference
- Read [Development Guide](docs/DEVELOPMENT.md) for development setup
- Check configuration files in `configs/` directory

