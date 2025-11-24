# Multi-Tenancy and Client Application Management

## Overview

The Sharding System is designed to handle multiple client applications in a Kubernetes cluster efficiently. This document explains how the system manages 100+ client applications and the architectural patterns used.

## Architecture for Multiple Client Applications

### Current Architecture (Shared Sharding System)

The system follows a **shared infrastructure** model where multiple client applications connect to the same Shard Router and Manager:

```
┌─────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                        │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Client     │  │   Client     │  │   Client     │      │
│  │      App 1   │  │      App 2   │  │      App 100 │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
│         │                 │                 │              │
│         └─────────────────┼─────────────────┘              │
│                           │                                │
│                  ┌────────▼─────────┐                      │
│                  │  Shard Router    │                      │
│                  │  (Data Plane)    │                      │
│                  └────────┬─────────┘                      │
│                           │                                │
│                  ┌────────▼─────────┐                      │
│                  │ Shard Manager   │                      │
│                  │ (Control Plane) │                      │
│                  └────────┬─────────┘                      │
│                           │                                │
│         ┌──────────────────┼──────────────────┐            │
│         │                  │                  │            │
│    ┌────▼────┐       ┌────▼────┐       ┌────▼────┐      │
│    │ Shard 1 │       │ Shard 2 │       │ Shard N  │      │
│    └─────────┘       └─────────┘       └─────────┘      │
└─────────────────────────────────────────────────────────────┘
```

### How It Works

1. **Stateless Router**: The Shard Router is stateless and horizontally scalable. It can handle requests from any number of client applications.

2. **Shard Key-Based Routing**: Each client application uses **shard keys** in their requests. The router computes a hash of the shard key to determine which shard handles the request.

3. **No Client Isolation Required**: Since routing is based on shard keys (not client identity), multiple client applications can share the same shard infrastructure. Each application's data is naturally isolated by their shard key namespace.

4. **Horizontal Scaling**: 
   - Router instances can be scaled independently
   - Shards can be added/removed dynamically
   - No single point of failure

## Handling 100 Client Applications

### Capacity Planning

The system can handle 100+ client applications because:

1. **Stateless Design**: The router doesn't maintain per-client state
2. **Efficient Routing**: Hash-based routing is O(1) complexity
3. **Connection Pooling**: Shared connection pools reduce resource usage
4. **Horizontal Scaling**: Add more router instances as needed

### Performance Characteristics

- **Throughput**: Each router instance can handle thousands of requests per second
- **Latency**: Sub-millisecond routing overhead
- **Scalability**: Linear scaling with router instances

### Resource Requirements (Example)

For 100 client applications:

```
Router Instances: 3-5 (for high availability)
Manager Instances: 2-3 (for redundancy)
Shards: Variable (based on data volume, not client count)
```

## Multi-Tenancy Models

### Model 1: Shared Infrastructure (Current)

**Pros:**
- Efficient resource utilization
- Simple deployment
- Easy to manage
- Cost-effective

**Cons:**
- All clients share the same shard catalog
- No per-client resource limits (can be added)
- All clients see the same shard topology

**Use Case:** Most common scenario where clients have similar requirements and trust levels.

### Model 2: Tenant Isolation (Future Enhancement)

For scenarios requiring strict isolation, you can implement tenant-specific configurations:

```go
type Tenant struct {
    ID          string
    Name        string
    ShardIDs    []string  // Tenant-specific shard assignments
    Config      map[string]interface{}
}
```

**Implementation:**
- Each tenant gets assigned specific shards
- Router routes based on tenant ID + shard key
- Manager enforces tenant boundaries

**Use Case:** Multi-tenant SaaS where different customers need isolation.

## Best Practices for Multiple Client Applications

### 1. Shard Key Design

Each client application should use **unique shard key namespaces**:

```go
// Client App 1
shardKey := "app1:user:123"

// Client App 2  
shardKey := "app2:order:456"

// Client App 100
shardKey := "app100:product:789"
```

### 2. Connection Management

- Use connection pooling in client libraries
- Implement retry logic with exponential backoff
- Set appropriate timeouts

### 3. Monitoring

- Monitor router throughput and latency
- Track shard health across all clients
- Set up alerts for capacity thresholds

### 4. Resource Limits

Consider implementing:
- Per-client rate limiting
- Per-client connection limits
- Per-client shard quotas

## Scaling Strategy

### Vertical Scaling
- Increase router instance resources (CPU/memory)
- Increase shard database capacity

### Horizontal Scaling
- Add more router instances (stateless, easy)
- Add more shards (requires resharding)
- Add more manager instances (for redundancy)

### Auto-Scaling

In Kubernetes, use Horizontal Pod Autoscaler (HPA):

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: router-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: router
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

## Example: 100 Client Applications

### Deployment Architecture

```yaml
# Router Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: router
spec:
  replicas: 5  # Scale based on load
  template:
    spec:
      containers:
      - name: router
        resources:
          requests:
            cpu: 500m
            memory: 512Mi
          limits:
            cpu: 2000m
            memory: 2Gi

# Manager Deployment  
apiVersion: apps/v1
kind: Deployment
metadata:
  name: manager
spec:
  replicas: 2  # For redundancy
```

### Expected Performance

With 5 router instances:
- **Capacity**: ~50,000 requests/second (10k per instance)
- **Latency**: <5ms p99
- **Client Applications**: 100+ (no hard limit)

### Monitoring Metrics

Key metrics to monitor:
- `router_requests_total` - Total requests across all clients
- `router_request_duration_seconds` - Request latency
- `router_active_connections` - Active connections
- `shard_health_status` - Health of each shard

## FAQ

### Q: Can client applications interfere with each other?

**A:** No. Each client uses shard keys to route to shards. Data isolation is maintained by shard key namespacing. However, they share the same infrastructure, so high load from one client can affect others (mitigated by rate limiting).

### Q: How do I limit resources per client application?

**A:** Currently, the system doesn't enforce per-client limits. You can:
1. Implement rate limiting at the router level
2. Use Kubernetes resource quotas per namespace
3. Implement tenant isolation (future enhancement)

### Q: What if one client application needs more shards?

**A:** The system supports dynamic shard creation. Create additional shards and the router will automatically distribute load. Use resharding to migrate data if needed.

### Q: How do I ensure high availability for 100 clients?

**A:** 
- Deploy multiple router instances (recommended: 3-5)
- Deploy multiple manager instances (recommended: 2-3)
- Use shard replication for data redundancy
- Implement health checks and automatic failover

## Future Enhancements

1. **Tenant Management API**: Explicit tenant registration and management
2. **Per-Tenant Shard Assignment**: Assign specific shards to tenants
3. **Resource Quotas**: Per-tenant limits on requests, connections, shards
4. **Tenant-Level Metrics**: Separate metrics per tenant
5. **Tenant Isolation**: Network-level isolation between tenants

