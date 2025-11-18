# Architecture Overview

## System Components

### 1. Shard Router (Data Plane)
- **Purpose**: Routes queries to appropriate shards based on shard key
- **Port**: 8080 (HTTP API), 9090 (Metrics)
- **Responsibilities**:
  - Key-to-shard mapping using consistent hashing
  - Connection pooling to shard databases
  - Read/write routing (primary vs replica)
  - Query execution and result aggregation

### 2. Shard Manager (Control Plane)
- **Purpose**: Manages shard lifecycle and resharding operations
- **Port**: 8081 (HTTP API), 9091 (Metrics)
- **Responsibilities**:
  - Shard creation/deletion
  - Resharding orchestration (split/merge)
  - Replica promotion
  - Catalog management

### 3. Metadata Store (Catalog)
- **Purpose**: Source of truth for shard mappings
- **Implementation**: etcd (can be PostgreSQL)
- **Stores**:
  - Shard metadata (endpoints, ranges, status)
  - Virtual node mappings
  - Catalog version for consistency

### 4. Re-sharder
- **Purpose**: Handles data migration during resharding
- **Process**:
  1. Pre-copy: Bulk copy existing data
  2. Delta sync: Capture changes during copy
  3. Cutover: Switch routing to new shards
  4. Validation: Verify data consistency

### 5. Health Controller
- **Purpose**: Monitors shard health and handles failover
- **Checks**:
  - Primary/replica availability
  - Replication lag
  - Automatic failover on primary failure

### 6. Client Library
- **Purpose**: Lightweight library for microservices
- **Features**:
  - Shard key resolution
  - Query execution helpers
  - Consistency level selection

## Data Flow

### Query Flow
```
Client → Router → Catalog (shard lookup) → Shard DB → Response
```

### Resharding Flow
```
Manager → Create Target Shards → Re-sharder:
  1. Pre-copy data
  2. Delta sync (CDC/WAL)
  3. Cutover (update catalog)
  4. Validate
```

### Failover Flow
```
Health Controller → Detect Primary Down → 
  Manager → Promote Replica → Update Catalog → 
  Router (auto-refresh) → Route to New Primary
```

## Sharding Strategy

### Consistent Hashing
- Uses virtual nodes (default: 256 per shard)
- Hash function: Murmur3 (default) or xxHash
- Minimal data movement when adding/removing shards

### Key Routing
- Single-key operations: Route to owning shard
- Multi-key operations: Route to same shard if possible
- Cross-shard operations: Application-level coordination

## Consistency Model

- **Per-shard**: Strong ACID consistency (native DB)
- **Cross-shard**: Eventual consistency or application-level transactions (Saga pattern)
- **Read consistency**: Configurable (strong vs eventual)

## Security

- **Authentication**: JWT-based
- **Authorization**: Role-based access control (RBAC)
- **Audit**: All control plane operations logged
- **TLS**: Configurable mTLS for internal communication

## Observability

- **Metrics**: Prometheus-compatible (QPS, latency, errors)
- **Tracing**: OpenTelemetry support (optional)
- **Logging**: Structured logging with zap

## Scalability

- **Horizontal**: Add shards to scale
- **Router**: Stateless, can be scaled horizontally
- **Manager**: Can be scaled (coordination via etcd)
- **Connection pooling**: Per-shard connection limits

## Failure Handling

- **Shard failure**: Automatic failover to replica
- **Router failure**: Stateless, can restart/replace
- **Catalog failure**: Read-only mode with cached mappings
- **Migration failure**: Rollback support, resumable jobs

