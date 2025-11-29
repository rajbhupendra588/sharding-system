# Kubernetes Deployment Requirements

When the Sharding System is installed in a Kubernetes cluster, it needs to perform several critical functions to operate effectively in a cloud-native environment.

## 🎯 Core Responsibilities

### 1. **Service Discovery** ✅ (Partially Implemented)

**What it needs to do:**
- Discover PostgreSQL instances running in the cluster
- Discover applications that use databases
- Track database connections and endpoints
- Update routing tables when services change

**Current Status:**
- ✅ `KubernetesDiscovery` service exists (`pkg/discovery/kubernetes.go`)
- ✅ Can discover Deployments and StatefulSets
- ✅ Extracts database connection info from environment variables
- ⚠️ **Needs**: Active integration in manager server startup

**Implementation Required:**
```go
// In manager server startup
discoveryService := discovery.NewKubernetesDiscovery(logger, registeredAppNames)
go discoveryService.StartDiscovery(ctx) // Continuous discovery loop
```

---

### 2. **Metrics Collection** ⚠️ (Needs Implementation)

**What it needs to do:**
- Collect metrics from PostgreSQL instances (query rate, CPU, memory, storage)
- Integrate with Prometheus for cluster-wide metrics
- Monitor pod resource usage (CPU, memory)
- Track database connection counts and latency

**Current Status:**
- ✅ `LoadMonitor` exists but returns zero metrics
- ✅ `MetricsCollector` interface defined
- ❌ **Missing**: Prometheus collector implementation
- ❌ **Missing**: PostgreSQL stats collector
- ❌ **Missing**: Kubernetes pod metrics collector

**Implementation Required:**
```go
// Prometheus Metrics Collector
type PrometheusCollector struct {
    promClient prometheus.Client
    logger     *zap.Logger
}

func (p *PrometheusCollector) CollectMetrics(ctx context.Context, shardID string) (*ShardMetrics, error) {
    // Query Prometheus for:
    // - postgres_stat_activity_count
    // - postgres_stat_database_blks_hit
    // - container_cpu_usage_seconds_total
    // - container_memory_usage_bytes
    // - postgres_stat_database_xact_commit
}

// PostgreSQL Stats Collector
type PostgreSQLStatsCollector struct {
    db *sql.DB
}

func (p *PostgreSQLStatsCollector) CollectMetrics(ctx context.Context, shardID string) (*ShardMetrics, error) {
    // Query pg_stat_activity, pg_stat_database, pg_stat_bgwriter
    // Calculate query rate, connection count, etc.
}
```

---

### 3. **Automatic Provisioning** ✅ (Implemented)

**What it needs to do:**
- Create PostgreSQL StatefulSets when databases are created
- Provision PersistentVolumeClaims for data storage
- Create Services for database access
- Manage database credentials via Secrets
- Wait for pods to be ready before marking shards as active

**Current Status:**
- ✅ `Operator` fully implemented (`pkg/operator/operator.go`)
- ✅ Creates StatefulSets, PVCs, Services, Secrets
- ✅ Waits for pod readiness
- ✅ Applies initial schema

**What it does:**
- Creates PostgreSQL StatefulSets with proper resource limits
- Provisions persistent storage
- Sets up networking (Services)
- Manages secrets for database credentials
- Monitors pod health

---

### 4. **Health Monitoring** ✅ (Implemented)

**What it needs to do:**
- Monitor PostgreSQL pod health
- Check database connectivity
- Track replication lag
- Detect failures and trigger failover

**Current Status:**
- ✅ `HealthController` exists (`pkg/health/controller.go`)
- ✅ Checks shard health periodically
- ✅ Detects replication lag
- ✅ Integrates with failover controller

**K8s Integration:**
- Uses Kubernetes pod status
- Checks liveness/readiness probes
- Monitors service endpoints

---

### 5. **Configuration Management** ✅ (Implemented)

**What it needs to do:**
- Read configuration from ConfigMaps
- Use Secrets for sensitive data (JWT, database passwords)
- Support environment variable overrides
- Handle configuration updates without restart

**Current Status:**
- ✅ ConfigMaps defined (`k8s/configmap.yaml`)
- ✅ Secrets defined (`k8s/secrets.yaml`)
- ✅ Environment variable support
- ⚠️ **Needs**: Hot-reload capability for config changes

**Configuration Sources:**
```yaml
# ConfigMap provides:
- etcd endpoints
- server settings
- sharding configuration
- observability settings

# Secrets provide:
- JWT_SECRET
- Database passwords
- TLS certificates
```

---

### 6. **RBAC & Security** ✅ (Implemented)

**What it needs to do:**
- Request appropriate Kubernetes permissions
- Use ServiceAccounts for authentication
- Limit access to required resources only
- Support multi-tenancy

**Current Status:**
- ✅ RBAC manifests (`k8s/rbac-discovery.yaml`)
- ✅ ServiceAccount defined
- ✅ Role and RoleBinding for discovery
- ✅ JWT-based API authentication

**Required Permissions:**
```yaml
# For discovery:
- get, list, watch: deployments, statefulsets, pods
- get, list: services, configmaps

# For operator:
- create, update, delete: statefulsets, services, pvcs, secrets
- get, list, watch: pods, statefulsets
```

---

### 7. **Networking** ✅ (Implemented)

**What it needs to do:**
- Expose services via Kubernetes Services
- Support Ingress for external access
- Handle internal service-to-service communication
- Support service mesh (Istio/Linkerd) if configured

**Current Status:**
- ✅ Services defined for Router and Manager
- ✅ ClusterIP for internal access
- ✅ NodePort/LoadBalancer support
- ⚠️ **Needs**: Ingress configuration examples

**Service Architecture:**
```
Router Service (ClusterIP) → Router Pods
Manager Service (ClusterIP) → Manager Pods
PostgreSQL Services (ClusterIP) → PostgreSQL Pods (per shard)
```

---

### 8. **Storage Management** ✅ (Implemented)

**What it needs to do:**
- Create PersistentVolumeClaims for database data
- Support different storage classes
- Handle storage expansion
- Backup to object storage (S3, GCS, etc.)

**Current Status:**
- ✅ PVC creation in operator
- ✅ Storage class support
- ✅ Backup service exists
- ⚠️ **Needs**: Object storage integration

---

### 9. **Auto-Scaling** ✅ (Implemented)

**What it needs to do:**
- Monitor shard load metrics
- Automatically split hot shards
- Scale PostgreSQL StatefulSets
- Rebalance data across shards

**Current Status:**
- ✅ Auto-split service implemented
- ✅ Hot shard detection
- ✅ Zero-downtime splitting
- ⚠️ **Needs**: Integration with K8s HPA (Horizontal Pod Autoscaler)

---

### 10. **Observability** ⚠️ (Needs Enhancement)

**What it needs to do:**
- Export Prometheus metrics
- Send logs to centralized logging (Loki, ELK)
- Integrate with distributed tracing (Jaeger, Zipkin)
- Provide Grafana dashboards

**Current Status:**
- ✅ Prometheus metrics endpoint (`/metrics`)
- ✅ Structured logging (zap)
- ❌ **Missing**: Log aggregation integration
- ❌ **Missing**: Tracing integration
- ✅ Grafana dashboard exists (`deploy/grafana/dashboard.json`)

---

## 📋 Implementation Checklist

### Immediate Requirements (Critical)

- [ ] **Integrate Kubernetes Discovery** - Enable automatic app discovery
- [ ] **Implement Prometheus Collector** - Real metrics collection
- [ ] **Add PostgreSQL Stats Collector** - Database-level metrics
- [ ] **Configure Service Monitoring** - Health checks and probes
- [ ] **Set up Log Aggregation** - Centralized logging

### High Priority (Important)

- [ ] **Ingress Configuration** - External access patterns
- [ ] **Config Hot-Reload** - Update config without restart
- [ ] **Object Storage Integration** - S3/GCS for backups
- [ ] **HPA Integration** - Kubernetes-native auto-scaling
- [ ] **Distributed Tracing** - Request tracing across services

### Nice to Have (Enhancements)

- [ ] **Service Mesh Integration** - Istio/Linkerd support
- [ ] **Multi-Region Support** - Cross-cluster sharding
- [ ] **Disaster Recovery** - Automated failover across regions
- [ ] **Cost Optimization** - Spot instance support
- [ ] **GitOps Integration** - ArgoCD/Flux support

---

## 🔧 Configuration Examples

### Environment Variables for K8s

```yaml
env:
  # Kubernetes
  - name: KUBERNETES_NAMESPACE
    valueFrom:
      fieldRef:
        fieldPath: metadata.namespace
  
  # Discovery
  - name: DISCOVERY_ENABLED
    value: "true"
  - name: DISCOVERY_INTERVAL
    value: "30s"
  
  # Metrics
  - name: PROMETHEUS_ENABLED
    value: "true"
  - name: PROMETHEUS_URL
    value: "http://prometheus:9090"
  
  # Storage
  - name: BACKUP_STORAGE_TYPE
    value: "s3"
  - name: S3_BUCKET
    value: "sharding-backups"
```

### ServiceAccount & RBAC

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sharding-manager
  namespace: sharding-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: sharding-operator
rules:
  - apiGroups: ["apps"]
    resources: ["statefulsets", "deployments"]
    verbs: ["get", "list", "watch", "create", "update", "delete"]
  - apiGroups: [""]
    resources: ["pods", "services", "pvc", "secrets"]
    verbs: ["get", "list", "watch", "create", "update", "delete"]
```

---

## 🚀 Deployment Flow

When deployed to Kubernetes, the application should:

1. **Startup Sequence:**
   ```
   Manager Pod Starts
   → Load ConfigMap/Secrets
   → Connect to etcd
   → Initialize Kubernetes Client
   → Start Discovery Service
   → Start Metrics Collectors
   → Start Health Monitoring
   → Start Auto-Scale Service
   → Ready for requests
   ```

2. **Runtime Operations:**
   - Continuously discover new applications
   - Collect metrics from all shards
   - Monitor health and trigger failover
   - Auto-scale based on load
   - Handle pod restarts gracefully

3. **Shutdown Sequence:**
   - Stop accepting new requests
   - Complete in-flight operations
   - Save state to etcd
   - Gracefully terminate

---

## 📊 Monitoring & Alerting

### Required Metrics

- **Application Metrics:**
  - Request rate per shard
  - Error rate
  - Latency (p50, p95, p99)
  - Active connections

- **Infrastructure Metrics:**
  - Pod CPU/Memory usage
  - Storage usage
  - Network I/O
  - Pod restart count

- **Business Metrics:**
  - Databases created/deleted
  - Shards split/merged
  - Failover events
  - Backup success rate

### Alerting Rules

```yaml
# Example Prometheus alerts
- alert: HighShardLoad
  expr: shard_query_rate > 1000
  for: 5m
  
- alert: ShardDown
  expr: shard_health_status == 0
  for: 1m
  
- alert: HighReplicationLag
  expr: replication_lag_seconds > 10
  for: 2m
```

---

## 🔐 Security Considerations

1. **Network Policies:** Restrict pod-to-pod communication
2. **Pod Security Standards:** Use restricted security context
3. **Secret Management:** Use external secret managers (Vault, Sealed Secrets)
4. **TLS:** Enable mTLS for service-to-service communication
5. **Image Security:** Scan container images for vulnerabilities
6. **RBAC:** Follow principle of least privilege

---

## 📝 Summary

The application is **well-prepared** for Kubernetes deployment with:
- ✅ Operator for automatic provisioning
- ✅ Discovery service for finding apps
- ✅ Health monitoring and failover
- ✅ Auto-scaling capabilities
- ✅ RBAC and security

**Key gaps to fill:**
- ⚠️ Real metrics collection (Prometheus/PostgreSQL)
- ⚠️ Active discovery integration
- ⚠️ Log aggregation
- ⚠️ Distributed tracing

The foundation is solid - mainly needs integration and configuration for production K8s environments.


