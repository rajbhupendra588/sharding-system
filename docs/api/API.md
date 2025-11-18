# Sharding System API Documentation

## Control Plane API (Manager)

Base URL: `http://localhost:8081/api/v1`

### Shards

#### Create Shard
```http
POST /shards
Content-Type: application/json

{
  "name": "shard-01",
  "primary_endpoint": "postgres://user:pass@host:5432/db",
  "replicas": ["postgres://rep1:5432/db"],
  "vnode_count": 256
}
```

#### List Shards
```http
GET /shards
```

#### Get Shard
```http
GET /shards/{id}
```

#### Delete Shard
```http
DELETE /shards/{id}
```

#### Promote Replica
```http
POST /shards/{id}/promote
Content-Type: application/json

{
  "replica_endpoint": "postgres://rep1:5432/db"
}
```

### Resharding

#### Split Shard
```http
POST /reshard/split
Content-Type: application/json

{
  "source_shard_id": "shard-01",
  "target_shards": [
    {
      "name": "shard-02",
      "primary_endpoint": "postgres://host2:5432/db",
      "replicas": [],
      "vnode_count": 256
    },
    {
      "name": "shard-03",
      "primary_endpoint": "postgres://host3:5432/db",
      "replicas": [],
      "vnode_count": 256
    }
  ]
}
```

#### Merge Shards
```http
POST /reshard/merge
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

#### Get Reshard Job Status
```http
GET /reshard/jobs/{id}
```

## Data Plane API (Router)

Base URL: `http://localhost:8080/v1`

### Execute Query
```http
POST /execute
Content-Type: application/json

{
  "shard_key": "user-123",
  "query": "SELECT * FROM users WHERE id = $1",
  "params": ["user-123"],
  "consistency": "strong"
}
```

### Get Shard for Key
```http
GET /shard-for-key?key=user-123
```

## Health Endpoints

Both services expose health endpoints:

```http
GET /health
```

## Metrics

Prometheus metrics are available at:

- Router: `http://localhost:9090/metrics`
- Manager: `http://localhost:9091/metrics`

