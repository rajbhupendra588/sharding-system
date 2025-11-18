# API Endpoints Reference

## Router API (Port 8080)

### Root Endpoint
```bash
GET http://localhost:8080/
```
Returns service information and available endpoints.

### Health Check
```bash
GET http://localhost:8080/health
```
Returns: `OK`

### Get Shard for Key
```bash
GET http://localhost:8080/v1/shard-for-key?key=<your-key>
```
Returns the shard ID for a given key.

**Example:**
```bash
curl http://localhost:8080/v1/shard-for-key?key=user-123
```

### Execute Query
```bash
POST http://localhost:8080/v1/execute
Content-Type: application/json

{
  "shard_key": "user-123",
  "query": "SELECT * FROM users WHERE id = $1",
  "params": ["user-123"],
  "consistency": "strong"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/v1/execute \
  -H "Content-Type: application/json" \
  -d '{
    "shard_key": "test",
    "query": "SELECT version()",
    "params": [],
    "consistency": "strong"
  }'
```

## Manager API (Port 8081)

### Root Endpoint
```bash
GET http://localhost:8081/
```
Returns service information and available endpoints.

### Health Check
```bash
GET http://localhost:8081/health
```
Returns: `OK`

### List Shards
```bash
GET http://localhost:8081/api/v1/shards
```
Returns a list of all shards.

**Example:**
```bash
curl http://localhost:8081/api/v1/shards
```

### Create Shard
```bash
POST http://localhost:8081/api/v1/shards
Content-Type: application/json

{
  "name": "shard-01",
  "primary_endpoint": "postgres://user:pass@host:5432/db",
  "replicas": ["postgres://rep1:5432/db"],
  "vnode_count": 256
}
```

**Example:**
```bash
curl -X POST http://localhost:8081/api/v1/shards \
  -H "Content-Type: application/json" \
  -d '{
    "name": "shard-01",
    "primary_endpoint": "postgres://postgres:postgres@localhost:5432/shard1",
    "replicas": [],
    "vnode_count": 256
  }'
```

### Get Shard by ID
```bash
GET http://localhost:8081/api/v1/shards/{id}
```

### Delete Shard
```bash
DELETE http://localhost:8081/api/v1/shards/{id}
```

### Promote Replica
```bash
POST http://localhost:8081/api/v1/shards/{id}/promote
Content-Type: application/json

{
  "replica_endpoint": "postgres://rep1:5432/db"
}
```

### Split Shard
```bash
POST http://localhost:8081/api/v1/reshard/split
Content-Type: application/json

{
  "source_shard_id": "shard-01-id",
  "target_shards": [
    {
      "name": "shard-02",
      "primary_endpoint": "postgres://host2:5432/db",
      "replicas": [],
      "vnode_count": 256
    }
  ]
}
```

### Merge Shards
```bash
POST http://localhost:8081/api/v1/reshard/merge
Content-Type: application/json

{
  "source_shard_ids": ["shard-01", "shard-02"],
  "target_shard": {
    "name": "shard-merged",
    "primary_endpoint": "postgres://host:5432/db",
    "replicas": [],
    "vnode_count": 256
  }
}
```

### Get Reshard Job Status
```bash
GET http://localhost:8081/api/v1/reshard/jobs/{job-id}
```

## Quick Test Commands

```bash
# Test Router
curl http://localhost:8080/
curl http://localhost:8080/health
curl 'http://localhost:8080/v1/shard-for-key?key=test'

# Test Manager
curl http://localhost:8081/
curl http://localhost:8081/health
curl http://localhost:8081/api/v1/shards
```

