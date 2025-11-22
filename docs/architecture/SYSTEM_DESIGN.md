# System Design

## Design Principles

### 1. Scalability
- **Horizontal Scaling**: Add more shards to increase capacity linearly
- **Stateless Components**: Router is stateless, enabling easy horizontal scaling
- **Efficient Routing**: Hash-based routing with O(1) lookup complexity
- **Connection Pooling**: Reuse database connections to minimize overhead

### 2. Availability
- **Replication**: Each shard can have multiple read replicas
- **Automatic Failover**: Replica promotion when primary fails
- **Health Monitoring**: Continuous health checks for all shards
- **Graceful Degradation**: System continues operating with partial failures
- **Zero-Downtime Operations**: Online resharding without service interruption

### 3. Consistency
- **Metadata Consistency**: Strong consistency for shard catalog (via etcd/PostgreSQL)
- **Data Consistency**: Configurable per-query (strong vs eventual)
- **Eventual Consistency**: Default for read replicas to improve performance
- **Strong Consistency**: Available for critical reads requiring latest data

### 4. Performance
- **Low Latency**: Efficient hash computation and routing
- **High Throughput**: Parallel query execution across shards
- **Connection Management**: Pooled connections with TTL
- **Caching**: Local caching of shard catalog in Router

### 5. Observability
- **Comprehensive Metrics**: Prometheus-compatible metrics
- **Structured Logging**: JSON logs with configurable levels
- **Health Endpoints**: Multiple health check endpoints
- **Audit Trail**: Complete audit logging for security events

### 6. Security
- **Authentication**: JWT-based authentication
- **Authorization**: Role-Based Access Control (RBAC)
- **Encryption**: TLS/SSL support for data in transit
- **Input Validation**: Comprehensive request validation
- **Audit Logging**: All operations logged for compliance

## Architecture Patterns

### Data Plane / Control Plane Separation

The system separates **data plane** (Router) from **control plane** (Manager):

- **Data Plane (Router)**: Handles high-frequency query routing
  - Stateless and horizontally scalable
  - Optimized for low latency
  - Caches shard catalog locally
  
- **Control Plane (Manager)**: Manages configuration and lifecycle
  - Stateful operations (shard creation, resharding)
  - Lower frequency, higher latency acceptable
  - Maintains authoritative shard catalog

**Benefits:**
- Independent scaling of components
- Clear separation of concerns
- Better fault isolation

### Consistent Hashing

Uses consistent hashing with virtual nodes to:
- Distribute data evenly across shards
- Minimize data movement during resharding
- Avoid hotspots and load imbalance

**Implementation:**
- Murmur3 hash function
- Virtual nodes (vnodes) for better distribution
- Hash ring for shard assignment

### Smart Client Pattern

Client libraries can cache shard catalog locally:
- Reduces latency (no lookup on every request)
- Reduces load on Manager
- Automatic refresh on catalog updates

### Event-Driven Updates

Shard catalog updates propagate asynchronously:
- Manager updates metadata store
- Router polls or receives notifications
- Eventual consistency acceptable for metadata

## Technologies

### Core Stack
- **Language**: Go 1.21+
  - High performance
  - Excellent concurrency support
  - Strong standard library
  
- **Metadata Storage**: 
  - **etcd** (recommended): Distributed key-value store
  - **PostgreSQL** (alternative): Relational database
  
- **Communication**: 
  - **REST API**: HTTP/JSON for all operations
  - **gRPC**: Not currently used (future consideration)

### Supporting Technologies
- **Authentication**: JWT (JSON Web Tokens)
- **Metrics**: Prometheus format
- **Logging**: Structured JSON (zap logger)
- **UI**: React + TypeScript
- **Containerization**: Docker
- **Orchestration**: Kubernetes (via Helm)

## Data Models

### Shard Model
```go
type Shard struct {
    ID              string    // Unique shard identifier
    Name            string    // Human-readable name
    HashRangeStart  uint64    // Start of hash range
    HashRangeEnd    uint64    // End of hash range
    PrimaryEndpoint string    // Primary database connection string
    Replicas        []string  // Read replica endpoints
    Status          string    // active, migrating, readonly, inactive
    Version         int64     // Version for optimistic concurrency
    VNodes          []VNode   // Virtual nodes for consistent hashing
}
```

### Query Request Model
```go
type QueryRequest struct {
    ShardKey    string        // Key to determine shard
    Query       string        // SQL query
    Params      []interface{} // Query parameters
    Consistency string        // "strong" or "eventual"
    Options     map[string]interface{}
}
```

### Resharding Job Model
```go
type ReshardJob struct {
    ID            string    // Unique job identifier
    Type          string    // "split" or "merge"
    SourceShards  []string  // Source shard IDs
    TargetShards  []string  // Target shard IDs
    Status        string    // pending, precopy, deltasync, cutover, completed, failed
    Progress      float64   // 0.0 to 1.0
    KeysMigrated  int64     // Number of keys migrated
    TotalKeys     int64     // Total keys to migrate
}
```

## Design Decisions

### Why Hash-Based Sharding?
- **Even Distribution**: Ensures balanced load across shards
- **Predictable Routing**: O(1) lookup time
- **Scalability**: Easy to add/remove shards with consistent hashing

### Why Consistent Hashing?
- **Minimal Data Movement**: Only ~1/n keys move when adding a shard
- **Virtual Nodes**: Better load balancing than simple consistent hashing
- **Smooth Resharding**: Gradual migration without downtime

### Why Separate Router and Manager?
- **Performance**: Router optimized for high-frequency operations
- **Scalability**: Can scale Router independently based on query load
- **Reliability**: Control plane failures don't affect data plane

### Why etcd for Metadata?
- **Consistency**: Strong consistency guarantees
- **Performance**: Low latency for reads
- **Reliability**: Distributed, fault-tolerant
- **Simplicity**: Simple key-value API

### Why REST over gRPC?
- **Simplicity**: Easier to debug and test
- **Compatibility**: Works with any HTTP client
- **Observability**: Standard HTTP metrics and logs
- **Future**: Can add gRPC later if needed

## Scalability Limits

### Current Limits
- **Shards**: No hard limit (tested up to 1000+ shards)
- **Connections**: Configurable per shard (default: 100)
- **Query Size**: 10MB request limit (configurable)
- **Concurrent Queries**: Limited by connection pool size

### Scaling Strategies
1. **Add More Shards**: Increase capacity linearly
2. **Add Router Replicas**: Scale data plane horizontally
3. **Add Manager Replicas**: Scale control plane (with leader election)
4. **Optimize Hash Function**: Use faster hash algorithms if needed
5. **Connection Pooling**: Tune pool sizes based on load

## Failure Modes and Handling

### Router Failure
- **Impact**: Query routing unavailable
- **Mitigation**: Multiple router instances behind load balancer
- **Recovery**: Automatic failover via load balancer

### Manager Failure
- **Impact**: Cannot create/update shards, but existing queries continue
- **Mitigation**: Manager replicas with leader election
- **Recovery**: New leader takes over, catalog remains in metadata store

### Metadata Store Failure
- **Impact**: Cannot update shard catalog, Router uses cached catalog
- **Mitigation**: etcd cluster with quorum, PostgreSQL with replication
- **Recovery**: Restore from backup or rebuild from shard state

### Shard Failure
- **Impact**: Queries to that shard fail
- **Mitigation**: Replica promotion, health monitoring
- **Recovery**: Promote replica or restore from backup

### Network Partition
- **Impact**: Components cannot communicate
- **Mitigation**: Router uses cached catalog, continues serving queries
- **Recovery**: Automatic reconnection when partition heals

## Performance Characteristics

### Latency
- **Router Lookup**: < 1ms (cached catalog)
- **Hash Computation**: < 0.1ms (Murmur3)
- **Query Execution**: Depends on database (typically 1-100ms)
- **Total Overhead**: < 2ms (excluding database query time)

### Throughput
- **Router**: 10,000+ queries/second (single instance)
- **Manager**: 100+ operations/second (shard management)
- **Scales Linearly**: Add more routers for higher throughput

### Resource Usage
- **Router Memory**: ~50MB base + ~1MB per shard
- **Manager Memory**: ~100MB base + ~2MB per shard
- **CPU**: Low (< 5% under normal load)
- **Network**: Minimal (only metadata updates)

## Future Enhancements

### Planned Features
1. **Cross-Shard Queries**: Scatter-gather queries across multiple shards
2. **Global Secondary Indexes**: Index non-shard-key fields
3. **Automatic Rebalancing**: Auto-split hot shards
4. **Query Caching**: Cache frequent queries
5. **Read-Through Cache**: Redis integration for hot data

### Under Consideration
1. **gRPC API**: For better performance in high-throughput scenarios
2. **GraphQL API**: For flexible querying
3. **Multi-Region Support**: Geo-distributed sharding
4. **Streaming Replication**: Real-time data replication
5. **Backup/Restore**: Automated backup and restore operations

