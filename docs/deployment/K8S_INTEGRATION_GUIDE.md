# Kubernetes Integration Guide

## Quick Start: What Happens When Deployed to K8s

When the Sharding System is deployed to a Kubernetes cluster, here's what it automatically does:

### 1. **Automatic Service Discovery** 🔍

The application discovers all PostgreSQL databases and applications in the cluster:

```go
// Automatically discovers:
- Deployments with database connections
- StatefulSets running PostgreSQL
- Services exposing databases
- ConfigMaps with database configs
```

**What you need to do:**
- Ensure RBAC permissions are granted (already in `k8s/rbac-discovery.yaml`)
- Applications should expose database info via environment variables or annotations

---

### 2. **Automatic Database Provisioning** 🗄️

When you create a database via API/UI, it automatically:

```yaml
Creates:
  - PostgreSQL StatefulSet (with replicas)
  - PersistentVolumeClaim (for data storage)
  - Service (for network access)
  - Secret (for database credentials)
  - ConfigMap (for configuration)
```

**What you need to do:**
- Ensure StorageClass exists for PVCs
- Set resource limits in database templates
- Configure backup storage (S3, GCS, etc.)

---

### 3. **Automatic Health Monitoring** ❤️

Continuously monitors:

```yaml
- Pod health (liveness/readiness)
- Database connectivity
- Replication lag
- Resource usage (CPU, memory, storage)
```

**What you need to do:**
- Configure Prometheus for metrics collection
- Set up alerting rules
- Configure log aggregation (Loki, ELK)

---

### 4. **Automatic Scaling** 📈

Automatically:

```yaml
- Detects hot shards (high load)
- Splits shards automatically
- Scales PostgreSQL pods
- Rebalances data
```

**What you need to do:**
- Configure metrics collection (Prometheus)
- Set scaling thresholds
- Monitor costs

---

## 🔧 Required Kubernetes Resources

### 1. **Namespace**

```bash
kubectl apply -f k8s/namespace.yaml
```

### 2. **Secrets**

```bash
# Generate JWT secret
export JWT_SECRET=$(openssl rand -base64 32)

# Update secrets.yaml and apply
kubectl apply -f k8s/secrets.yaml
```

### 3. **ConfigMap**

```bash
# Update configmap.yaml with your etcd endpoints
kubectl apply -f k8s/configmap.yaml
```

### 4. **RBAC**

```bash
# Grant permissions for discovery and operator
kubectl apply -f k8s/rbac-discovery.yaml
```

### 5. **Deployments**

```bash
kubectl apply -f k8s/manager-deployment.yaml
kubectl apply -f k8s/router-deployment.yaml
```

---

## 📊 What Gets Created Automatically

### For Each Database:

1. **StatefulSet** - PostgreSQL instance
   ```yaml
   apiVersion: apps/v1
   kind: StatefulSet
   metadata:
     name: {db-name}-shard-0
   spec:
     replicas: 1
     template:
       spec:
         containers:
         - name: postgres
           image: postgres:15
   ```

2. **PersistentVolumeClaim** - Data storage
   ```yaml
   apiVersion: v1
   kind: PersistentVolumeClaim
   metadata:
     name: {db-name}-shard-0-data
   spec:
     storageClassName: standard
     resources:
       requests:
         storage: 10Gi
   ```

3. **Service** - Network access
   ```yaml
   apiVersion: v1
   kind: Service
   metadata:
     name: {db-name}-shard-0
   spec:
     ports:
     - port: 5432
       targetPort: 5432
   ```

4. **Secret** - Database credentials
   ```yaml
   apiVersion: v1
   kind: Secret
   metadata:
     name: {db-name}-shard-0-credentials
   type: Opaque
   data:
     password: <base64-encoded>
   ```

---

## 🔍 Monitoring Integration

### Prometheus Scraping

The application exposes metrics at `/metrics`:

```yaml
# Prometheus ServiceMonitor
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: sharding-system
spec:
  selector:
    matchLabels:
      app: sharding-manager
  endpoints:
  - port: http
    path: /metrics
```

### Grafana Dashboard

Import the dashboard:
```bash
kubectl apply -f deploy/grafana/dashboard.json
```

---

## 🚨 Alerting Setup

### Example Alert Rules

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: sharding-alerts
spec:
  groups:
  - name: sharding
    rules:
    - alert: HighShardLoad
      expr: shard_query_rate > 1000
      for: 5m
      annotations:
        summary: "Shard {{ $labels.shard_id }} has high load"
    
    - alert: ShardDown
      expr: shard_health_status == 0
      for: 1m
      annotations:
        summary: "Shard {{ $labels.shard_id }} is down"
```

---

## 🔐 Security Best Practices

1. **Network Policies**
   ```yaml
   # Restrict pod-to-pod communication
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: sharding-network-policy
   spec:
     podSelector:
       matchLabels:
         app: sharding-manager
     policyTypes:
     - Ingress
     - Egress
   ```

2. **Pod Security Standards**
   ```yaml
   # Use restricted security context
   securityContext:
     runAsNonRoot: true
     runAsUser: 1000
     fsGroup: 1000
     seccompProfile:
       type: RuntimeDefault
   ```

3. **Secret Management**
   - Use external secret managers (Vault, Sealed Secrets)
   - Rotate secrets regularly
   - Use Kubernetes Secrets for temporary storage only

---

## 📈 Scaling Considerations

### Horizontal Scaling

The application supports:
- Multiple Manager replicas (with leader election)
- Multiple Router replicas (load balanced)
- Auto-scaling based on metrics

### Vertical Scaling

Adjust resource requests/limits:
```yaml
resources:
  requests:
    cpu: "500m"
    memory: "512Mi"
  limits:
    cpu: "2000m"
    memory: "2Gi"
```

---

## 🎯 Next Steps

1. **Deploy to K8s:**
   ```bash
   ./k8s/deploy.sh
   ```

2. **Configure Prometheus:**
   - Set up Prometheus operator
   - Configure ServiceMonitor
   - Import Grafana dashboard

3. **Set up Logging:**
   - Deploy Loki or ELK stack
   - Configure log aggregation
   - Set up log retention

4. **Configure Backups:**
   - Set up S3/GCS bucket
   - Configure backup schedules
   - Test restore procedures

5. **Monitor & Alert:**
   - Set up alerting rules
   - Configure notification channels
   - Test alerting flow

---

## ✅ Verification Checklist

After deployment, verify:

- [ ] Manager pod is running
- [ ] Router pod is running
- [ ] Services are accessible
- [ ] Discovery is finding applications
- [ ] Metrics are being collected
- [ ] Health checks are passing
- [ ] Logs are being aggregated
- [ ] Backups are working
- [ ] Auto-scaling is enabled
- [ ] Alerts are configured

---

**The application is production-ready for Kubernetes!** 🚀


