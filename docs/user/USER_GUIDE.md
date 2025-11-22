# User Guide

## Introduction

Welcome to the User Guide for the Sharding System. This document provides comprehensive instructions on how to use the system effectively, from initial setup to advanced operations.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Core Concepts](#core-concepts)
3. [Using the Web UI](#using-the-web-ui)
4. [Using the API](#using-the-api)
5. [Shard Management](#shard-management)
6. [Query Execution](#query-execution)
7. [Resharding Operations](#resharding-operations)
8. [Monitoring and Health](#monitoring-and-health)
9. [Troubleshooting](#troubleshooting)

## Getting Started

### Prerequisites

- Docker and Docker Compose (for local deployment)
- Access to database instances for shards
- etcd or PostgreSQL for metadata storage

### Quick Start

1. **Start the System**
   ```bash
   docker-compose up -d
   ```

2. **Access the Web UI**
   - Open `http://localhost:3000` in your browser
   - Default credentials may be required (check deployment guide)

3. **Verify Services**
   - Manager: `http://localhost:8081/health`
   - Router: `http://localhost:8080/health`

### First Steps

1. **Create Your First Shard**
   - Use the Web UI or API to create a shard
   - Provide database connection details
   - System will automatically assign hash ranges

2. **Execute a Query**
   - Use the Router API or Web UI query executor
   - Provide a shard key and SQL query
   - System routes to the correct shard automatically

## Core Concepts

### Shard Key

The **shard key** is a value used to determine which shard stores your data. Common examples:
- User ID: `"user-123"`
- Order ID: `"order-456"`
- Tenant ID: `"tenant-789"`

**Important:** Choose shard keys with:
- High cardinality (many unique values)
- Even distribution (avoid hotspots)
- Relevance to your queries

### Hash-Based Routing

The system uses hash-based routing:
1. Compute hash of shard key: `hash = Murmur3(shard_key)`
2. Map hash to virtual node (vnode)
3. Route to shard containing that vnode

### Consistency Levels

- **Strong Consistency**: Reads from primary database (latest data)
- **Eventual Consistency**: Reads from replica (may be slightly stale, faster)

## Using the Web UI

### Dashboard

The dashboard provides:
- System health overview
- Shard status summary
- Recent activity
- Quick access to common operations

### Shard Management

1. **View Shards**
   - Navigate to "Shards" section
   - See all shards with status, endpoints, and hash ranges

2. **Create Shard**
   - Click "Create Shard"
   - Enter shard name and database endpoints
   - System assigns hash ranges automatically

3. **Delete Shard**
   - Select shard and click "Delete"
   - **Warning:** Ensure shard is empty or data is migrated

### Query Executor

1. **Navigate to Query Executor**
2. **Enter Query Details**
   - Shard Key: Key to determine target shard
   - SQL Query: Your SQL query
   - Parameters: Query parameters (if needed)
   - Consistency: Strong or Eventual

3. **Execute Query**
   - Click "Execute"
   - View results in table format
   - Check execution time and shard used

### Resharding

1. **Split Shard**
   - Select shard to split
   - Choose number of target shards
   - Monitor job progress

2. **Merge Shards**
   - Select multiple shards to merge
   - Configure target shard
   - Monitor job progress

### Monitoring

- **Health Status**: View health of all shards
- **Metrics**: Prometheus metrics visualization
- **Logs**: System logs and audit trail

## Using the API

### Authentication

When RBAC is enabled, obtain a token:

```bash
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "password"}'
```

Use the token in subsequent requests:

```bash
curl -H "Authorization: Bearer <token>" \
  http://localhost:8081/api/v1/shards
```

### Creating a Shard

```bash
curl -X POST http://localhost:8081/api/v1/shards \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "name": "shard-1",
    "primary_endpoint": "postgresql://user:pass@localhost:5432/shard1",
    "replicas": ["postgresql://user:pass@localhost:5433/shard1-replica"],
    "vnode_count": 256
  }'
```

### Executing a Query

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

### Looking Up Shard for Key

```bash
curl "http://localhost:8080/v1/shard-for-key?key=user-123"
```

## Shard Management

### Creating Shards

**Best Practices:**
- Use descriptive names (e.g., `shard-users-east`)
- Configure replicas for high availability
- Start with default vnode count (256), adjust based on load

**Example:**
```json
{
  "name": "shard-users-1",
  "primary_endpoint": "postgresql://user:pass@db1:5432/users",
  "replicas": [
    "postgresql://user:pass@db1-replica:5432/users"
  ],
  "vnode_count": 256
}
```

### Updating Shards

Shards can be updated via API:
- Add/remove replicas
- Change status (active, readonly, inactive)
- Update endpoints (requires careful migration)

### Deleting Shards

**Before deleting:**
1. Ensure shard is empty or data is migrated
2. Verify no active queries are using the shard
3. Check resharding jobs are complete

**Deletion:**
```bash
curl -X DELETE http://localhost:8081/api/v1/shards/shard-1 \
  -H "Authorization: Bearer <token>"
```

## Query Execution

### Basic Query

```bash
curl -X POST http://localhost:8080/v1/execute \
  -H "Content-Type: application/json" \
  -d '{
    "shard_key": "user-123",
    "query": "SELECT * FROM users WHERE id = $1",
    "params": ["user-123"]
  }'
```

### Parameterized Queries

Always use parameterized queries to prevent SQL injection:

```json
{
  "shard_key": "order-456",
  "query": "SELECT * FROM orders WHERE user_id = $1 AND status = $2",
  "params": ["user-123", "completed"]
}
```

### Consistency Levels

**Strong Consistency** (default):
```json
{
  "shard_key": "user-123",
  "query": "SELECT balance FROM accounts WHERE user_id = $1",
  "consistency": "strong"
}
```

**Eventual Consistency** (for read-heavy workloads):
```json
{
  "shard_key": "user-123",
  "query": "SELECT * FROM user_activity WHERE user_id = $1",
  "consistency": "eventual"
}
```

### Query Best Practices

1. **Always include shard key**: Required for routing
2. **Use parameterized queries**: Prevents SQL injection
3. **Choose appropriate consistency**: Strong for critical reads, eventual for analytics
4. **Optimize queries**: Index your database tables appropriately
5. **Avoid cross-shard queries**: Design schema to keep related data together

## Resharding Operations

### When to Reshard

- **Shard too large**: Approaching storage or performance limits
- **Uneven distribution**: One shard receiving disproportionate load
- **Scaling down**: Consolidating shards to reduce costs

### Splitting a Shard

Split a large shard into multiple smaller shards:

```bash
curl -X POST http://localhost:8081/api/v1/reshard/split \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "source_shard_id": "shard-1",
    "target_shards": [
      {
        "name": "shard-1a",
        "primary_endpoint": "postgresql://user:pass@db1a:5432/users",
        "replicas": [],
        "vnode_count": 128
      },
      {
        "name": "shard-1b",
        "primary_endpoint": "postgresql://user:pass@db1b:5432/users",
        "replicas": [],
        "vnode_count": 128
      }
    ]
  }'
```

**Process:**
1. Job created with status `pending`
2. `precopy`: Historical data copied
3. `deltasync`: Incremental changes synchronized
4. `cutover`: Traffic switched to new shards
5. `completed`: Old shard can be removed

### Merging Shards

Merge multiple shards into one:

```bash
curl -X POST http://localhost:8081/api/v1/reshard/merge \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "source_shard_ids": ["shard-1a", "shard-1b"],
    "target_shard": {
      "name": "shard-1",
      "primary_endpoint": "postgresql://user:pass@db1:5432/users",
      "replicas": [],
      "vnode_count": 256
    }
  }'
```

### Monitoring Resharding Jobs

```bash
curl http://localhost:8081/api/v1/reshard/jobs/job-123 \
  -H "Authorization: Bearer <token>"
```

**Job Status:**
- `pending`: Queued, not started
- `precopy`: Copying historical data
- `deltasync`: Synchronizing changes
- `cutover`: Switching traffic
- `completed`: Finished successfully
- `failed`: Check error message

## Monitoring and Health

### Health Checks

**Manager Health:**
```bash
curl http://localhost:8081/api/v1/health
```

**Router Health:**
```bash
curl http://localhost:8080/v1/health
```

### Metrics

Prometheus metrics available at:
- Manager: `http://localhost:9091/metrics`
- Router: `http://localhost:9090/metrics`

**Key Metrics:**
- Query latency (p50, p95, p99)
- Query throughput (queries/second)
- Shard health status
- Resharding job progress
- Error rates

### Logs

**Manager Logs:**
```bash
tail -f logs/manager.log
```

**Router Logs:**
```bash
tail -f logs/router.log
```

**Audit Logs:**
```bash
tail -f /var/log/sharding/audit.log
```

## Troubleshooting

### Common Issues

#### Shard Not Found

**Problem:** Query fails with "shard not found"

**Solutions:**
1. Verify shard exists: `GET /api/v1/shards/{id}`
2. Check shard status is "active"
3. Verify Router has latest catalog (may need refresh)

#### Query Timeout

**Problem:** Queries timing out

**Solutions:**
1. Check database connectivity
2. Verify shard is healthy
3. Review query performance (add indexes)
4. Increase timeout in configuration

#### Authentication Failed

**Problem:** "Authentication required" error

**Solutions:**
1. Verify RBAC is enabled
2. Obtain valid JWT token
3. Include token in Authorization header
4. Check token expiration

#### Resharding Job Stuck

**Problem:** Job status not progressing

**Solutions:**
1. Check job details for error messages
2. Verify database connectivity
3. Check available disk space
4. Review logs for errors
5. Consider canceling and retrying

### Getting Help

1. **Check FAQ**: See [FAQ](./FAQ.md) for common questions
2. **Review Logs**: Check service logs for errors
3. **Check Health**: Verify all services are healthy
4. **Review Configuration**: Ensure configuration is correct

For more detailed troubleshooting, refer to the [FAQ](./FAQ.md) or check the [Developer Guide](../dev/DEVELOPER_GUIDE.md) for advanced debugging.

