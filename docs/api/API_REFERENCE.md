# API Reference

This document provides a complete reference for the Sharding System REST API.

## Base URLs

- **Manager Service**: `http://localhost:8081` (default)
- **Router Service**: `http://localhost:8080` (default)

## Authentication

When RBAC is enabled, most endpoints require authentication via JWT tokens.

### Obtaining a Token

```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600
}
```

### Using the Token

Include the token in the `Authorization` header:

```
Authorization: Bearer <token>
```

## Manager Service API

The Manager Service provides endpoints for shard management, resharding operations, and system administration.

### Shard Management

#### List All Shards

```http
GET /api/v1/shards
Authorization: Bearer <token>
```

**Response:**
```json
[
  {
    "id": "shard-1",
    "name": "shard-1",
    "hash_range_start": 0,
    "hash_range_end": 18446744073709551615,
    "primary_endpoint": "postgresql://user:pass@localhost:5432/shard1",
    "replicas": [
      "postgresql://user:pass@localhost:5433/shard1-replica1"
    ],
    "status": "active",
    "version": 1,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "vnodes": [
      {
        "id": 1,
        "shard_id": "shard-1",
        "hash": 123456789
      }
    ]
  }
]
```

**Status Codes:**
- `200 OK`: Success
- `401 Unauthorized`: Authentication required
- `500 Internal Server Error`: Server error

#### Get Shard Details

```http
GET /api/v1/shards/{id}
Authorization: Bearer <token>
```

**Parameters:**
- `id` (path): Shard identifier

**Response:** Same as single shard object from List Shards

**Status Codes:**
- `200 OK`: Success
- `404 Not Found`: Shard not found
- `401 Unauthorized`: Authentication required

#### Create Shard

```http
POST /api/v1/shards
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "shard-2",
  "primary_endpoint": "postgresql://user:pass@localhost:5432/shard2",
  "replicas": [
    "postgresql://user:pass@localhost:5433/shard2-replica1"
  ],
  "vnode_count": 256
}
```

**Request Body:**
- `name` (string, required): Shard name
- `primary_endpoint` (string, required): Primary database connection string
- `replicas` (array of strings, optional): Read replica connection strings
- `vnode_count` (integer, optional): Number of virtual nodes (default: 256)

**Response:**
```json
{
  "id": "shard-2",
  "name": "shard-2",
  "hash_range_start": 0,
  "hash_range_end": 18446744073709551615,
  "primary_endpoint": "postgresql://user:pass@localhost:5432/shard2",
  "replicas": [
    "postgresql://user:pass@localhost:5433/shard2-replica1"
  ],
  "status": "active",
  "version": 1,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

**Status Codes:**
- `201 Created`: Shard created successfully
- `400 Bad Request`: Invalid request
- `401 Unauthorized`: Authentication required
- `500 Internal Server Error`: Server error

#### Delete Shard

```http
DELETE /api/v1/shards/{id}
Authorization: Bearer <token>
```

**Parameters:**
- `id` (path): Shard identifier

**Response:** No content

**Status Codes:**
- `204 No Content`: Shard deleted successfully
- `400 Bad Request`: Cannot delete shard (e.g., has data)
- `401 Unauthorized`: Authentication required
- `404 Not Found`: Shard not found

#### Promote Replica

```http
POST /api/v1/shards/{id}/promote
Authorization: Bearer <token>
Content-Type: application/json

{
  "replica_endpoint": "postgresql://user:pass@localhost:5433/shard1-replica1"
}
```

**Parameters:**
- `id` (path): Shard identifier

**Request Body:**
- `replica_endpoint` (string, required): Replica connection string to promote

**Response:**
```json
{
  "status": "promoted"
}
```

**Status Codes:**
- `200 OK`: Replica promoted successfully
- `400 Bad Request`: Invalid request
- `401 Unauthorized`: Authentication required
- `404 Not Found`: Shard not found

### Resharding Operations

#### Split Shard

```http
POST /api/v1/reshard/split
Authorization: Bearer <token>
Content-Type: application/json

{
  "source_shard_id": "shard-1",
  "target_shards": [
    {
      "name": "shard-1a",
      "primary_endpoint": "postgresql://user:pass@localhost:5432/shard1a",
      "replicas": [],
      "vnode_count": 128
    },
    {
      "name": "shard-1b",
      "primary_endpoint": "postgresql://user:pass@localhost:5432/shard1b",
      "replicas": [],
      "vnode_count": 128
    }
  ],
  "split_point": 9223372036854775807
}
```

**Request Body:**
- `source_shard_id` (string, required): ID of shard to split
- `target_shards` (array, required): Array of new shard configurations
- `split_point` (integer, optional): Explicit hash value to split at

**Response:**
```json
{
  "id": "job-123",
  "type": "split",
  "source_shards": ["shard-1"],
  "target_shards": ["shard-1a", "shard-1b"],
  "status": "pending",
  "progress": 0.0,
  "started_at": "2024-01-01T00:00:00Z",
  "keys_migrated": 0,
  "total_keys": 0
}
```

**Status Codes:**
- `202 Accepted`: Resharding job started
- `400 Bad Request`: Invalid request
- `401 Unauthorized`: Authentication required
- `500 Internal Server Error`: Server error

#### Merge Shards

```http
POST /api/v1/reshard/merge
Authorization: Bearer <token>
Content-Type: application/json

{
  "source_shard_ids": ["shard-1a", "shard-1b"],
  "target_shard": {
    "name": "shard-1",
    "primary_endpoint": "postgresql://user:pass@localhost:5432/shard1",
    "replicas": [],
    "vnode_count": 256
  }
}
```

**Request Body:**
- `source_shard_ids` (array, required): IDs of shards to merge
- `target_shard` (object, required): Configuration for merged shard

**Response:** Same format as Split Shard, with `type: "merge"`

**Status Codes:**
- `202 Accepted`: Resharding job started
- `400 Bad Request`: Invalid request
- `401 Unauthorized`: Authentication required
- `500 Internal Server Error`: Server error

#### Get Resharding Job Status

```http
GET /api/v1/reshard/jobs/{id}
Authorization: Bearer <token>
```

**Parameters:**
- `id` (path): Job identifier

**Response:**
```json
{
  "id": "job-123",
  "type": "split",
  "source_shards": ["shard-1"],
  "target_shards": ["shard-1a", "shard-1b"],
  "status": "precopy",
  "progress": 0.45,
  "started_at": "2024-01-01T00:00:00Z",
  "completed_at": null,
  "error_message": "",
  "keys_migrated": 45000,
  "total_keys": 100000
}
```

**Job Status Values:**
- `pending`: Job queued, not started
- `precopy`: Copying historical data
- `deltasync`: Synchronizing incremental changes
- `cutover`: Switching traffic to new shards
- `completed`: Job finished successfully
- `failed`: Job failed (check `error_message`)

**Status Codes:**
- `200 OK`: Success
- `401 Unauthorized`: Authentication required
- `404 Not Found`: Job not found

### Health and Status

#### Health Check

```http
GET /api/v1/health
```

**Response:**
```json
{
  "status": "healthy",
  "version": "1.0.0"
}
```

**Status Codes:**
- `200 OK`: Service is healthy

#### Legacy Health Check

```http
GET /health
```

**Response:** `OK` (plain text)

## Router Service API

The Router Service provides endpoints for query execution and shard lookup.

### Query Execution

#### Execute Query

```http
POST /v1/execute
Content-Type: application/json

{
  "shard_key": "user-123",
  "query": "SELECT * FROM users WHERE id = $1",
  "params": ["user-123"],
  "consistency": "strong"
}
```

**Request Body:**
- `shard_key` (string, required): Key used to determine target shard
- `query` (string, required): SQL query to execute
- `params` (array, optional): Query parameters (for parameterized queries)
- `consistency` (string, optional): `"strong"` or `"eventual"` (default: `"strong"`)
- `options` (object, optional): Additional query options

**Response:**
```json
{
  "shard_id": "shard-1",
  "rows": [
    {
      "id": "user-123",
      "name": "John Doe",
      "email": "john@example.com"
    }
  ],
  "row_count": 1,
  "latency_ms": 12.5
}
```

**Consistency Levels:**
- `strong`: Reads from primary database (latest data, higher latency)
- `eventual`: Reads from replica (may be slightly stale, lower latency)

**Status Codes:**
- `200 OK`: Query executed successfully
- `400 Bad Request`: Invalid request (missing shard_key or query)
- `500 Internal Server Error`: Query execution failed

**Example with cURL:**
```bash
curl -X POST http://localhost:8080/v1/execute \
  -H "Content-Type: application/json" \
  -d '{
    "shard_key": "user-123",
    "query": "SELECT * FROM users WHERE id = $1",
    "params": ["user-123"],
    "consistency": "strong"
  }'
```

### Shard Lookup

#### Get Shard for Key

```http
GET /v1/shard-for-key?key={key}
```

**Query Parameters:**
- `key` (string, required): Shard key to look up

**Response:**
```json
{
  "shard_id": "shard-1"
}
```

**Status Codes:**
- `200 OK`: Success
- `400 Bad Request`: Missing key parameter
- `500 Internal Server Error`: Lookup failed

**Example:**
```bash
curl "http://localhost:8080/v1/shard-for-key?key=user-123"
```

### Health and Status

#### Health Check

```http
GET /v1/health
```

**Response:**
```json
{
  "status": "healthy",
  "version": "1.0.0"
}
```

#### Legacy Health Check

```http
GET /health
```

**Response:** `OK` (plain text)

#### Service Information

```http
GET /
```

**Response:**
```json
{
  "service": "sharding-router",
  "version": "1.0.0",
  "endpoints": [
    "POST /v1/execute",
    "GET /v1/shard-for-key?key=<key>",
    "GET /v1/health",
    "GET /health"
  ]
}
```

## Error Responses

All endpoints return errors in a consistent format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message"
  }
}
```

**Common Error Codes:**
- `INVALID_REQUEST`: Request validation failed
- `SHARD_NOT_FOUND`: Shard does not exist
- `AUTHENTICATION_REQUIRED`: Authentication token missing or invalid
- `AUTHORIZATION_FAILED`: User lacks required permissions
- `INTERNAL_ERROR`: Server error
- `SHARD_UNAVAILABLE`: Shard is not available (down or migrating)

## Rate Limiting

Currently, the API does not enforce rate limiting. This may be added in future versions.

## CORS

CORS (Cross-Origin Resource Sharing) is enabled by default. Configure allowed origins in the server configuration.

## Request Size Limits

- Maximum request body size: 10MB (configurable)
- Content-Type validation: Only `application/json` accepted for POST/PUT/PATCH requests

## Pagination

List endpoints (e.g., `GET /api/v1/shards`) currently return all results. Pagination may be added in future versions.

## Versioning

The API uses URL-based versioning:
- Current version: `v1`
- Manager endpoints: `/api/v1/*`
- Router endpoints: `/v1/*`

Future versions will use `/api/v2/*` and `/v2/*` respectively.

