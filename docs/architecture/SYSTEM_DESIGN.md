# System Design Document
## Standalone DB Sharding Microservice

## Table of Contents
1. [System Overview](#system-overview)
2. [Architecture](#architecture)
3. [Component Details](#component-details)
4. [Data Flow](#data-flow)
5. [Sharding Strategy](#sharding-strategy)
6. [Resharding Process](#resharding-process)
7. [High Availability & Failover](#high-availability--failover)
8. [Security Architecture](#security-architecture)
9. [Observability](#observability)
10. [Scalability & Performance](#scalability--performance)
11. [Deployment Architecture](#deployment-architecture)
12. [API Design](#api-design)
13. [Data Models](#data-models)
14. [Failure Scenarios & Recovery](#failure-scenarios--recovery)
15. [Operational Considerations](#operational-considerations)

---

## System Overview

### Purpose
A production-ready, self-contained database sharding service that provides transparent routing, online resharding, replication management, health monitoring, and comprehensive observability. The system enables horizontal scaling of databases by distributing data across multiple shards while maintaining a unified interface for applications.

### Key Requirements
- **Transparent Sharding**: Applications interact with a single logical database
- **Online Resharding**: Add/remove shards without downtime
- **High Availability**: Automatic failover and replica promotion
- **Consistency**: Strong consistency per shard, eventual consistency across shards
- **Security**: Authentication, authorization, and audit logging
- **Observability**: Comprehensive metrics, logging, and tracing
- **Performance**: Low-latency routing, connection pooling, read/write splitting

### Design Principles
1. **Separation of Concerns**: Control plane (management) separated from data plane (routing)
2. **Stateless Components**: Router and Manager are stateless for horizontal scaling
3. **Eventual Consistency**: Cross-shard operations use eventual consistency model
4. **Fail-Fast**: Quick detection and recovery from failures
5. **Operational Simplicity**: Clear APIs and comprehensive monitoring

---

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Client Applications                      │
│                    (Microservices using client-lib)              │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ HTTP/gRPC
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Shard Router (Data Plane)                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │   Router 1   │  │   Router 2   │  │   Router N   │         │
│  │  (Stateless) │  │  (Stateless) │  │  (Stateless) │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
│         │                  │                  │                  │
│         └──────────────────┴──────────────────┘                 │
│                            │                                     │
│                            ▼                                     │
│              ┌─────────────────────────┐                        │
│              │   Metadata Catalog      │                        │
│              │   (etcd/PostgreSQL)     │                        │
│              └─────────────────────────┘                        │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ Catalog Updates
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Shard Manager (Control Plane)                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │  Manager 1   │  │  Manager 2   │  │  Manager N   │         │
│  │  (Stateless) │  │  (Stateless) │  │  (Stateless) │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
│         │                  │                  │                  │
│         └──────────────────┴──────────────────┘                 │
│                            │                                     │
│                            ▼                                     │
│              ┌─────────────────────────┐                        │
│              │   Metadata Catalog      │                        │
│              │   (etcd/PostgreSQL)     │                        │
│              └─────────────────────────┘                        │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ Management Operations
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Supporting Services                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │  Re-sharder  │  │    Health    │  │   Security   │         │
│  │              │  │  Controller  │  │   Service    │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Database Shards                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │   Shard 1    │  │   Shard 2    │  │   Shard N    │         │
│  │  Primary     │  │  Primary     │  │  Primary     │         │
│  │  Replica 1   │  │  Replica 1   │  │  Replica 1   │         │
│  │  Replica 2   │  │  Replica 2   │  │  Replica 2   │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└─────────────────────────────────────────────────────────────────┘
```

### Component Layers

1. **Client Layer**: Applications using the client library
2. **Data Plane Layer**: Shard Router (stateless, horizontally scalable)
3. **Control Plane Layer**: Shard Manager (stateless, horizontally scalable)
4. **Metadata Layer**: Catalog store (etcd or PostgreSQL)
5. **Data Layer**: Database shards (PostgreSQL with replication)

---

## Component Details

### 1. Shard Router (Data Plane)

**Purpose**: Routes queries to appropriate shards based on shard key.

**Responsibilities**:
- Key-to-shard mapping using consistent hashing
- Connection pooling to shard databases
- Read/write routing (primary vs replica)
- Query execution and result aggregation
- Catalog cache management
- Request/response metrics

**Key Features**:
- **Stateless Design**: Can be horizontally scaled
- **Connection Pooling**: Per-shard connection pools with configurable limits
- **Read/Write Splitting**: Routes reads to replicas, writes to primary
- **Catalog Caching**: Caches shard mappings with TTL-based invalidation
- **Consistency Levels**: Supports strong and eventual consistency for reads

**Ports**:
- `8080`: HTTP API (data plane)
- `9090`: Prometheus metrics

**Performance Characteristics**:
- P99 latency: < 10ms (routing overhead)
- Throughput: 10,000+ QPS per router instance
- Connection pool: 100 connections per shard (configurable)

### 2. Shard Manager (Control Plane)

**Purpose**: Manages shard lifecycle and resharding operations.

**Responsibilities**:
- Shard CRUD operations
- Resharding orchestration (split/merge)
- Replica promotion
- Catalog management
- Health monitoring coordination
- Audit logging

**Key Features**:
- **Stateless Design**: Can be horizontally scaled with coordination via catalog
- **Resharding Orchestration**: Coordinates multi-phase resharding operations
- **Catalog Versioning**: Ensures consistent catalog updates
- **RBAC**: Role-based access control for operations
- **Audit Logging**: All operations logged for compliance

**Ports**:
- `8081`: HTTP API (control plane)
- `9091`: Prometheus metrics

**API Categories**:
- Shard Management: Create, Read, Update, Delete shards
- Resharding: Split and merge operations
- Health: Replica promotion, health status
- Catalog: Version management, watch/notify

### 3. Metadata Catalog

**Purpose**: Source of truth for shard mappings and routing rules.

**Implementation Options**:
- **etcd**: Distributed key-value store (default)
- **PostgreSQL**: Relational database (production alternative)

**Stored Data**:
- Shard metadata (ID, name, endpoints, hash ranges, status)
- Virtual node mappings (vnode → shard mapping)
- Catalog version (for consistency)
- Resharding job status
- Health status

**Key Features**:
- **Versioning**: Catalog version incremented on each update
- **Watch/Notify**: Routers watch for catalog changes
- **Consistency**: Strong consistency guarantees
- **High Availability**: etcd cluster or PostgreSQL HA setup

**Data Structure**:
```
/catalog/version -> int64
/catalog/shard/{shard_id} -> Shard JSON
/catalog/vnodes/{vnode_id} -> VNode JSON
/catalog/reshard/jobs/{job_id} -> ReshardJob JSON
```

### 4. Re-sharder

**Purpose**: Handles data migration during resharding operations.

**Process Phases**:

1. **Pre-copy Phase**:
   - Bulk copy existing data from source to target shards
   - Uses consistent hashing to determine target shard
   - Batched operations for efficiency
   - Progress tracking (keys migrated / total keys)

2. **Delta Sync Phase**:
   - Capture changes during pre-copy
   - Uses Change Data Capture (CDC) or WAL streaming
   - Replay changes to target shards
   - Continues until cutover

3. **Cutover Phase**:
   - Update catalog to route new requests to target shards
   - Drain in-flight requests to source shards
   - Final delta sync
   - Mark source shards as read-only

4. **Validation Phase**:
   - Verify data consistency between source and target
   - Check row counts, checksums
   - Report any discrepancies

**Features**:
- **Resumable**: Can resume from last checkpoint on failure
- **Progress Tracking**: Real-time progress updates
- **Rollback Support**: Can rollback on failure
- **Minimal Downtime**: Cutover typically < 1 second

### 5. Health Controller

**Purpose**: Monitors shard health and handles failover.

**Monitoring**:
- Primary/replica availability (ping checks)
- Replication lag (PostgreSQL replication lag queries)
- Connection pool health
- Query latency

**Failover Process**:
1. Detect primary failure (health check timeout)
2. Verify replication lag is acceptable
3. Promote replica to primary (via Manager API)
4. Update catalog with new primary endpoint
5. Routers automatically refresh catalog and route to new primary

**Health States**:
- **Healthy**: Primary and replicas up, lag < threshold
- **Degraded**: Primary up, some replicas down
- **Unhealthy**: Primary down or lag > threshold

**Configuration**:
- Health check interval: 5 seconds (default)
- Replication lag threshold: 10 seconds (default)
- Failure detection: 3 consecutive failures

### 6. Client Library

**Purpose**: Lightweight library for microservices to interact with sharding system.

**Features**:
- Shard key resolution (compute shard ID for a key)
- Query execution helpers
- Consistency level selection
- Connection management
- Retry logic with exponential backoff

**Usage Pattern**:
```go
client := shardclient.NewClient("http://router:8080")
shardID, err := client.GetShardForKey("user-123")
result, err := client.Query("user-123", "SELECT * FROM users WHERE id = $1", "user-123")
```

---

## Data Flow

### Query Flow (Read/Write)

```
1. Client Application
   │
   │ HTTP POST /v1/execute
   │ { shard_key: "user-123", query: "...", params: [...] }
   │
   ▼
2. Shard Router
   │
   │ a. Extract shard_key from request
   │ b. Compute hash(shard_key) → hash_value
   │ c. Lookup shard in catalog cache (or fetch from catalog)
   │ d. Determine target shard using consistent hashing
   │ e. Check query type (read vs write)
   │
   ├─ Write Query → Route to Primary
   └─ Read Query → Route to Replica (or primary if strong consistency)
   │
   ▼
3. Database Shard (Primary/Replica)
   │
   │ Execute query
   │ Return results
   │
   ▼
4. Shard Router
   │
   │ Aggregate results (if multi-shard query)
   │ Add metrics (latency, shard_id)
   │
   ▼
5. Client Application
   │
   │ Receive response
   │ { shard_id: "shard-01", rows: [...], latency_ms: 5.2 }
```

### Resharding Flow (Split)

```
1. Admin/Operator
   │
   │ POST /api/v1/reshard/split
   │ { source_shard_id: "shard-01", target_shards: [...] }
   │
   ▼
2. Shard Manager
   │
   │ a. Validate request
   │ b. Create reshard job (status: "pending")
   │ c. Create target shards in catalog
   │ d. Trigger re-sharder
   │
   ▼
3. Re-sharder
   │
   │ Phase 1: Pre-copy
   │ ├─ Fetch all keys from source shard
   │ ├─ For each key:
   │ │   ├─ Compute hash(key) → target_shard
   │ │   └─ Copy row to target shard
   │ └─ Update progress (keys_migrated / total_keys)
   │
   │ Phase 2: Delta Sync
   │ ├─ Start CDC/WAL streaming from source
   │ ├─ For each change:
   │ │   ├─ Compute hash(key) → target_shard
   │ │   └─ Apply change to target shard
   │ └─ Continue until cutover
   │
   │ Phase 3: Cutover
   │ ├─ Notify Manager to update catalog
   │ ├─ Manager updates catalog version
   │ ├─ Routers detect catalog change (watch)
   │ ├─ Routers refresh cache
   │ └─ New requests route to target shards
   │
   │ Phase 4: Validation
   │ ├─ Compare row counts
   │ ├─ Compare checksums
   │ └─ Report results
   │
   ▼
4. Shard Manager
   │
   │ Update job status: "completed"
   │ Mark source shard as "inactive" (optional)
   │
   ▼
5. Admin/Operator
   │
   │ GET /api/v1/reshard/jobs/{id}
   │ { status: "completed", progress: 1.0 }
```

### Failover Flow

```
1. Health Controller
   │
   │ Health check detects primary failure
   │ (3 consecutive failures)
   │
   ▼
2. Health Controller
   │
   │ Check replication lag on replicas
   │ Select replica with lowest lag
   │
   ▼
3. Shard Manager
   │
   │ POST /api/v1/shards/{id}/promote
   │ { replica_endpoint: "postgres://replica1:5432/db" }
   │
   ▼
4. Shard Manager
   │
   │ Update catalog:
   │ - Set replica as new primary
   │ - Update primary_endpoint
   │ - Increment catalog version
   │
   ▼
5. Routers (via watch)
   │
   │ Detect catalog version change
   │ Refresh catalog cache
   │ Update connection pools
   │
   ▼
6. New Requests
   │
   │ Route to new primary
   │ System continues operating
```

---

## Sharding Strategy

### Consistent Hashing

**Algorithm**: Consistent hashing with virtual nodes

**Hash Function**:
- **Default**: Murmur3 (fast, good distribution)
- **Alternative**: xxHash (faster, slightly less uniform)

**Virtual Nodes**:
- Default: 256 virtual nodes per shard
- Each virtual node maps to a hash value in the ring
- Provides better load distribution

**Hash Ring**:
```
Hash Space: [0, 2^64 - 1]

Shard 1: [vnode1, vnode2, ..., vnode256]
Shard 2: [vnode1, vnode2, ..., vnode256]
Shard 3: [vnode1, vnode2, ..., vnode256]
...
```

**Key Routing**:
1. Compute `hash(shard_key)` → `hash_value`
2. Find first virtual node with `vnode.hash >= hash_value` (clockwise)
3. Route to shard owning that virtual node

**Benefits**:
- Minimal data movement when adding/removing shards (~1/N where N = number of shards)
- Even distribution with virtual nodes
- Deterministic routing (same key → same shard)

### Shard Key Selection

**Best Practices**:
- Use high-cardinality keys (e.g., user_id, order_id)
- Avoid hot keys (keys with disproportionate traffic)
- Consider access patterns (co-locate related data)

**Key Types**:
- **Single Key**: `user-123` → routes to one shard
- **Composite Key**: `user-123:order-456` → routes to one shard
- **Multi-Key**: Multiple keys may route to different shards

### Cross-Shard Operations

**Limitations**:
- Cross-shard transactions not supported (no distributed transactions)
- Cross-shard queries require application-level coordination

**Patterns**:
1. **Saga Pattern**: For multi-shard transactions
2. **Eventual Consistency**: Accept eventual consistency for cross-shard updates
3. **Application-Level Coordination**: Use message queues or event sourcing

---

## Resharding Process

### Split Operation

**Use Case**: Shard grows too large, need to split into multiple shards

**Process**:
1. **Preparation**:
   - Create target shards (2+ new shards)
   - Allocate virtual nodes to target shards
   - Initialize target shards with schema

2. **Pre-copy**:
   - Scan source shard (table scan or index scan)
   - For each row:
     - Compute hash(shard_key)
     - Determine target shard
     - Insert into target shard
   - Track progress: `keys_migrated / total_keys`

3. **Delta Sync**:
   - Start CDC/WAL streaming from source
   - Apply changes to target shards
   - Continue until cutover

4. **Cutover**:
   - Update catalog: route new keys to target shards
   - Drain in-flight requests
   - Final delta sync
   - Mark source shard as read-only

5. **Validation**:
   - Compare row counts
   - Compare checksums (optional)
   - Verify no data loss

6. **Cleanup**:
   - Archive source shard (optional)
   - Remove source shard from catalog

**Downtime**: < 1 second (during cutover)

### Merge Operation

**Use Case**: Consolidate multiple small shards into one

**Process**:
1. **Preparation**:
   - Create target shard
   - Allocate virtual nodes from source shards to target

2. **Pre-copy**:
   - Scan all source shards in parallel
   - Copy all rows to target shard
   - Track progress per source shard

3. **Delta Sync**:
   - Stream changes from all source shards
   - Apply to target shard
   - Continue until cutover

4. **Cutover**:
   - Update catalog: route keys to target shard
   - Drain requests
   - Final delta sync

5. **Validation**:
   - Verify all data migrated
   - Check consistency

6. **Cleanup**:
   - Remove source shards from catalog

**Downtime**: < 1 second (during cutover)

### Resharding Job States

- **pending**: Job created, not started
- **precopy**: Bulk copying data
- **deltasync**: Syncing changes
- **cutover**: Switching routing
- **validation**: Verifying consistency
- **completed**: Successfully completed
- **failed**: Failed (can be retried)

---

## High Availability & Failover

### Replication Model

**Primary-Replica Architecture**:
- Each shard has 1 primary and N replicas (N >= 1)
- Writes go to primary, replicated to replicas
- Reads can go to replicas (eventual consistency) or primary (strong consistency)

**Replication Method**:
- PostgreSQL streaming replication (WAL-based)
- Asynchronous replication (default)
- Synchronous replication (optional, for stronger consistency)

### Failover Scenarios

#### 1. Primary Failure

**Detection**:
- Health controller detects primary down (3 consecutive failures)
- Timeout: 5 seconds per check

**Recovery**:
1. Select replica with lowest replication lag
2. Promote replica to primary
3. Update catalog
4. Routers refresh and route to new primary

**Downtime**: ~15-30 seconds (detection + promotion)

#### 2. Replica Failure

**Impact**: Minimal (reads may route to primary)

**Recovery**:
- Health controller marks replica as down
- Routers avoid routing to failed replica
- Replica can be replaced without downtime

#### 3. Router Failure

**Impact**: None (routers are stateless, load balancer routes to healthy routers)

**Recovery**:
- Load balancer detects failure
- Routes traffic to healthy routers
- Failed router can be replaced

#### 4. Manager Failure

**Impact**: No new operations (existing operations continue)

**Recovery**:
- Load balancer routes to healthy managers
- Failed manager can be replaced

#### 5. Catalog Failure

**Impact**: Routers use cached catalog (read-only mode)

**Recovery**:
- etcd cluster: Automatic failover to healthy nodes
- PostgreSQL: Failover to replica
- Routers refresh catalog when available

### Health Monitoring

**Metrics**:
- Primary/replica availability
- Replication lag (seconds)
- Connection pool utilization
- Query latency (p50, p95, p99)
- Error rates

**Alerts**:
- Primary down
- Replication lag > threshold
- High error rate
- High latency

---

## Security Architecture

### Authentication

**Method**: JWT-based authentication

**Flow**:
1. Client authenticates with credentials
2. Manager issues JWT token
3. Client includes token in requests
4. Router/Manager validates token

**Token Claims**:
- User ID
- Roles (admin, operator, viewer)
- Expiration time

### Authorization

**RBAC (Role-Based Access Control)**:

**Roles**:
- **admin**: Full access (create/delete shards, resharding)
- **operator**: Operational access (promote replicas, view status)
- **viewer**: Read-only access (view shards, metrics)

**Permissions**:
- Shard CRUD: admin only
- Resharding: admin only
- Replica promotion: operator+
- Query execution: all authenticated users
- Metrics: viewer+

### Audit Logging

**Logged Events**:
- Shard creation/deletion
- Resharding operations
- Replica promotions
- Authentication failures
- Authorization failures

**Log Format**:
```json
{
  "timestamp": "2024-01-01T12:00:00Z",
  "user": "admin@example.com",
  "action": "create_shard",
  "resource": "shard-01",
  "result": "success",
  "ip": "192.168.1.1"
}
```

### Network Security

**TLS/mTLS**:
- Configurable TLS for client connections
- Optional mTLS for internal communication
- Certificate management via config

**Network Isolation**:
- Routers and Managers in private network
- Only load balancer exposed publicly
- Database shards in private network

---

## Observability

### Metrics (Prometheus)

**Router Metrics**:
- `shard_router_requests_total`: Total requests
- `shard_router_request_duration_seconds`: Request latency
- `shard_router_errors_total`: Error count
- `shard_router_shard_requests_total`: Requests per shard
- `shard_router_connection_pool_size`: Connection pool size
- `shard_router_catalog_cache_hits`: Catalog cache hits

**Manager Metrics**:
- `shard_manager_operations_total`: Operations count
- `shard_manager_operation_duration_seconds`: Operation latency
- `shard_manager_reshard_jobs_active`: Active reshard jobs
- `shard_manager_reshard_progress`: Reshard progress

**Health Metrics**:
- `shard_health_status`: Shard health (1=healthy, 0=unhealthy)
- `shard_replication_lag_seconds`: Replication lag
- `shard_primary_up`: Primary availability (1=up, 0=down)
- `shard_replicas_up`: Number of replicas up

**Grafana Dashboards**:
- Request rate and latency
- Error rates
- Shard health and replication lag
- Resharding progress
- Connection pool utilization

### Logging

**Structured Logging** (zap):
- JSON format
- Log levels: DEBUG, INFO, WARN, ERROR
- Context: request ID, shard ID, user ID

**Log Categories**:
- Access logs (requests/responses)
- Error logs (failures, exceptions)
- Audit logs (security events)
- Operational logs (resharding, failover)

### Tracing

**OpenTelemetry Support** (optional):
- Distributed tracing across components
- Trace ID propagation
- Span annotations for operations

**Trace Points**:
- Router: Request handling, shard lookup, query execution
- Manager: Operations, resharding phases
- Re-sharder: Migration phases

---

## Scalability & Performance

### Horizontal Scaling

**Router Scaling**:
- Stateless design allows unlimited horizontal scaling
- Load balancer distributes requests
- Each router instance handles independent requests

**Manager Scaling**:
- Stateless design allows horizontal scaling
- Coordination via catalog (etcd/PostgreSQL)
- Leader election for resharding operations (optional)

**Shard Scaling**:
- Add shards to increase capacity
- Minimal data movement (~1/N)
- No downtime during shard addition

### Performance Characteristics

**Router Performance**:
- P50 latency: < 2ms (routing overhead)
- P95 latency: < 5ms
- P99 latency: < 10ms
- Throughput: 10,000+ QPS per instance

**Shard Performance**:
- Depends on underlying database
- PostgreSQL: 1,000-10,000 QPS per shard (depends on hardware)

**Catalog Performance**:
- etcd: 10,000+ reads/sec, 1,000+ writes/sec
- PostgreSQL: Depends on configuration

### Capacity Planning

**Shard Sizing**:
- Target: 100GB - 1TB per shard (PostgreSQL)
- Monitor: Disk usage, query latency, connection count

**Router Sizing**:
- CPU: 2-4 cores per 10,000 QPS
- Memory: 2-4 GB (catalog cache, connection pools)
- Network: 1 Gbps per 10,000 QPS

**Manager Sizing**:
- CPU: 1-2 cores (low traffic)
- Memory: 1-2 GB
- Network: Minimal

### Optimization Strategies

**Connection Pooling**:
- Per-shard connection pools
- Configurable pool size (default: 100)
- Connection reuse

**Catalog Caching**:
- In-memory cache with TTL
- Watch-based invalidation
- Reduces catalog lookups

**Query Optimization**:
- Prepared statements
- Connection reuse
- Batch operations (when possible)

---

## Deployment Architecture

### Container Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Load Balancer                         │
│              (nginx/HAProxy/Cloud LB)                     │
└───────────────┬───────────────────┬───────────────────────┘
                │                   │
        ┌───────▼───────┐   ┌───────▼───────┐
        │  Router Pods  │   │ Manager Pods  │
        │  (3+ replicas)│   │ (2+ replicas) │
        └───────┬───────┘   └───────┬───────┘
                │                   │
                └───────────┬───────┘
                            │
                    ┌───────▼───────┐
                    │  Catalog      │
                    │  (etcd cluster│
                    │   or PG HA)   │
                    └───────┬───────┘
                            │
                    ┌───────▼───────┐
                    │  Database     │
                    │  Shards       │
                    │  (Primary +   │
                    │   Replicas)   │
                    └───────────────┘
```

### Kubernetes Deployment

**Deployment Manifests**:
- Router: Deployment (3+ replicas), Service (ClusterIP), Ingress
- Manager: Deployment (2+ replicas), Service (ClusterIP)
- Catalog: StatefulSet (etcd) or Deployment (PostgreSQL)
- Shards: StatefulSet (PostgreSQL with persistent volumes)

**Resource Requirements**:
- Router: 2 CPU, 2GB RAM per pod
- Manager: 1 CPU, 1GB RAM per pod
- Catalog: Depends on etcd/PostgreSQL sizing

**High Availability**:
- Multiple replicas for stateless components
- Pod disruption budgets
- Health checks (liveness, readiness)

### Docker Compose (Development)

**Services**:
- etcd: Single node (development)
- router: Single instance
- manager: Single instance
- postgres-shard1: Example shard
- postgres-shard2: Example shard

**Networking**:
- Bridge network for service communication
- Port mapping for external access

---

## API Design

### Control Plane API (Manager)

**Base URL**: `http://manager:8081/api/v1`

#### Shard Management

**Create Shard**
```http
POST /shards
Content-Type: application/json
Authorization: Bearer <token>

{
  "name": "shard-01",
  "primary_endpoint": "postgres://user:pass@host:5432/db",
  "replicas": ["postgres://rep1:5432/db"],
  "vnode_count": 256
}
```

**List Shards**
```http
GET /shards
Authorization: Bearer <token>
```

**Get Shard**
```http
GET /shards/{id}
Authorization: Bearer <token>
```

**Delete Shard**
```http
DELETE /shards/{id}
Authorization: Bearer <token>
```

**Promote Replica**
```http
POST /shards/{id}/promote
Content-Type: application/json
Authorization: Bearer <token>

{
  "replica_endpoint": "postgres://rep1:5432/db"
}
```

#### Resharding

**Split Shard**
```http
POST /reshard/split
Content-Type: application/json
Authorization: Bearer <token>

{
  "source_shard_id": "shard-01",
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

**Merge Shards**
```http
POST /reshard/merge
Content-Type: application/json
Authorization: Bearer <token>

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

**Get Reshard Job Status**
```http
GET /reshard/jobs/{id}
Authorization: Bearer <token>
```

### Data Plane API (Router)

**Base URL**: `http://router:8080/v1`

#### Query Execution

**Execute Query**
```http
POST /execute
Content-Type: application/json
Authorization: Bearer <token>

{
  "shard_key": "user-123",
  "query": "SELECT * FROM users WHERE id = $1",
  "params": ["user-123"],
  "consistency": "strong"
}
```

**Response**:
```json
{
  "shard_id": "shard-01",
  "rows": [
    {"id": "user-123", "name": "John Doe"}
  ],
  "row_count": 1,
  "latency_ms": 5.2
}
```

**Get Shard for Key**
```http
GET /shard-for-key?key=user-123
Authorization: Bearer <token>
```

**Response**:
```json
{
  "shard_id": "shard-01",
  "shard_name": "shard-01",
  "hash_value": 1234567890
}
```

### Health Endpoints

**Router Health**
```http
GET /health
```

**Manager Health**
```http
GET /health
```

**Response**:
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime_seconds": 3600
}
```

### Metrics Endpoints

**Prometheus Metrics**
```http
GET /metrics
```

---

## Data Models

### Shard

```go
type Shard struct {
    ID              string    // Unique identifier
    Name            string    // Human-readable name
    HashRangeStart  uint64    // Start of hash range (for range-based)
    HashRangeEnd    uint64    // End of hash range (for range-based)
    PrimaryEndpoint string    // Primary database endpoint
    Replicas        []string  // Replica endpoints
    Status          string    // "active", "migrating", "readonly", "inactive"
    Version         int64     // Version for optimistic concurrency
    CreatedAt       time.Time
    UpdatedAt       time.Time
    VNodes          []VNode   // Virtual nodes assigned to this shard
}
```

### VNode (Virtual Node)

```go
type VNode struct {
    ID       uint64 // Virtual node ID
    ShardID  string // Shard owning this vnode
    Hash     uint64 // Hash value on the ring
}
```

### ShardCatalog

```go
type ShardCatalog struct {
    Version   int64   // Catalog version (incremented on updates)
    Shards    []Shard // All shards
    UpdatedAt time.Time
}
```

### ReshardJob

```go
type ReshardJob struct {
    ID            string    // Job ID
    Type          string    // "split" or "merge"
    SourceShards  []string  // Source shard IDs
    TargetShards  []string  // Target shard IDs
    Status        string    // "pending", "precopy", "deltasync", "cutover", "completed", "failed"
    Progress      float64   // 0.0 to 1.0
    StartedAt     time.Time
    CompletedAt   *time.Time
    ErrorMessage  string
    KeysMigrated  int64     // Number of keys migrated
    TotalKeys     int64     // Total keys to migrate
}
```

### ShardHealth

```go
type ShardHealth struct {
    ShardID        string        // Shard ID
    Status         string        // "healthy", "degraded", "unhealthy"
    ReplicationLag time.Duration // Replication lag
    LastCheck      time.Time     // Last health check time
    PrimaryUp      bool          // Primary availability
    ReplicasUp     []string      // Available replicas
    ReplicasDown   []string      // Unavailable replicas
}
```

### QueryRequest

```go
type QueryRequest struct {
    ShardKey    string        // Shard key for routing
    Query       string        // SQL query
    Params      []interface{} // Query parameters
    Consistency string        // "strong" or "eventual"
    Options     map[string]interface{} // Additional options
}
```

### QueryResponse

```go
type QueryResponse struct {
    ShardID   string        // Shard that executed the query
    Rows      []interface{} // Query results
    RowCount  int           // Number of rows
    LatencyMs float64       // Query latency in milliseconds
}
```

---

## Failure Scenarios & Recovery

### Scenario 1: Primary Database Failure

**Symptoms**:
- Health checks fail
- Queries to primary timeout
- Replication stops

**Recovery**:
1. Health controller detects failure (15-30 seconds)
2. Select healthy replica with lowest lag
3. Promote replica to primary
4. Update catalog
5. Routers refresh and route to new primary

**Downtime**: 15-30 seconds

**Prevention**:
- Multiple replicas per shard
- Regular health checks
- Monitoring and alerting

### Scenario 2: Router Failure

**Symptoms**:
- Load balancer detects router down
- Requests fail to that router

**Recovery**:
1. Load balancer routes to healthy routers
2. Failed router can be restarted/replaced
3. No data loss (stateless)

**Downtime**: None (if multiple routers)

**Prevention**:
- Multiple router instances
- Health checks
- Auto-scaling

### Scenario 3: Manager Failure

**Symptoms**:
- Control plane API unavailable
- New operations fail

**Recovery**:
1. Load balancer routes to healthy managers
2. Failed manager can be restarted/replaced
3. Existing operations continue

**Downtime**: None for data plane (control plane unavailable)

**Prevention**:
- Multiple manager instances
- Health checks

### Scenario 4: Catalog Failure

**Symptoms**:
- Routers cannot fetch catalog updates
- New shards not visible

**Recovery**:
1. Routers use cached catalog (read-only mode)
2. Catalog cluster fails over (etcd) or promotes replica (PostgreSQL)
3. Routers refresh catalog when available

**Downtime**: None (read-only mode)

**Prevention**:
- etcd cluster (3+ nodes) or PostgreSQL HA
- Regular backups

### Scenario 5: Resharding Failure

**Symptoms**:
- Reshard job status: "failed"
- Error message in job

**Recovery**:
1. Review error message
2. Fix underlying issue
3. Retry job (resumable from checkpoint)
4. Or rollback (if before cutover)

**Downtime**: None (if before cutover)

**Prevention**:
- Validate target shards before starting
- Monitor progress
- Test resharding in staging

### Scenario 6: Network Partition

**Symptoms**:
- Components cannot communicate
- Split-brain scenarios

**Recovery**:
1. etcd: Quorum-based decisions
2. Routers: Use cached catalog
3. Managers: Leader election
4. Resolve partition when network recovers

**Downtime**: Depends on partition duration

**Prevention**:
- Multiple network paths
- Quorum-based decisions
- Monitoring

---

## Operational Considerations

### Monitoring & Alerting

**Key Metrics to Monitor**:
- Request rate and latency (p50, p95, p99)
- Error rates (4xx, 5xx)
- Shard health status
- Replication lag
- Connection pool utilization
- Resharding progress
- Catalog version staleness

**Alerts**:
- Primary down
- Replication lag > 10 seconds
- Error rate > 1%
- P99 latency > 100ms
- Resharding job failed
- Catalog unavailable

### Backup & Recovery

**Backup Strategy**:
- Per-shard backups (PostgreSQL pg_dump or WAL archiving)
- Catalog backups (etcd snapshot or PostgreSQL dump)
- Regular backups (daily)
- Retention: 30 days

**Recovery Procedures**:
1. Restore shard from backup
2. Restore catalog if needed
3. Verify data consistency
4. Resume operations

### Capacity Planning

**Shard Capacity**:
- Monitor: Disk usage, query latency, connection count
- Threshold: 80% capacity → plan resharding
- Target: 100GB - 1TB per shard

**Router Capacity**:
- Monitor: CPU, memory, request rate
- Scale: Add routers when CPU > 70%
- Target: 10,000 QPS per router instance

**Catalog Capacity**:
- Monitor: Disk usage, request rate
- Scale: Add nodes (etcd) or upgrade (PostgreSQL)
- Target: < 10GB metadata

### Maintenance Windows

**Planned Maintenance**:
- Shard maintenance: Use replica promotion
- Router/Manager: Rolling updates (no downtime)
- Catalog: Rolling updates (etcd) or maintenance window (PostgreSQL)

**Resharding Windows**:
- Schedule during low-traffic periods
- Monitor during resharding
- Have rollback plan ready

### Disaster Recovery

**RTO (Recovery Time Objective)**: 1 hour
**RPO (Recovery Point Objective)**: 15 minutes

**DR Procedures**:
1. Restore catalog from backup
2. Restore shards from backups
3. Verify data consistency
4. Update endpoints if changed
5. Resume operations

**Cross-Region Replication**:
- Consider cross-region replication for critical shards
- Use PostgreSQL streaming replication or logical replication

---

## Conclusion

This system design provides a comprehensive, production-ready database sharding solution with:

- **Transparent Sharding**: Applications interact with a unified interface
- **Online Resharding**: Add/remove shards without downtime
- **High Availability**: Automatic failover and replica promotion
- **Security**: Authentication, authorization, and audit logging
- **Observability**: Comprehensive metrics, logging, and tracing
- **Scalability**: Horizontal scaling of all components

The architecture is designed for:
- **Reliability**: High availability and fault tolerance
- **Performance**: Low-latency routing and efficient data access
- **Operability**: Clear APIs, monitoring, and operational procedures
- **Security**: Multi-layered security with RBAC and audit logging

---

## References

- [Architecture Overview](../ARCHITECTURE.md)
- [API Documentation](./API.md)
- [Development Guide](./DEVELOPMENT.md)
- [Quick Start Guide](../QUICKSTART.md)

