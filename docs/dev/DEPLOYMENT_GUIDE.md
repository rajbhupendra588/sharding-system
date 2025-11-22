# Deployment Guide

This guide covers deploying the Sharding System in various environments, from local development to production Kubernetes clusters.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Local Deployment](#local-deployment)
3. [Docker Deployment](#docker-deployment)
4. [Kubernetes Deployment](#kubernetes-deployment)
5. [Production Considerations](#production-considerations)
6. [Monitoring and Observability](#monitoring-and-observability)
7. [Troubleshooting](#troubleshooting)

## Prerequisites

### For Local Deployment
- Docker and Docker Compose
- 4GB+ RAM available
- Ports available: 8080, 8081, 2389, 3000

### For Production Deployment
- Kubernetes cluster (1.21+)
- kubectl configured
- Helm 3.x installed
- etcd cluster or PostgreSQL for metadata
- Database instances for shards

## Local Deployment

### Quick Start with Docker Compose

The easiest way to get started is using Docker Compose:

```bash
# Clone the repository
git clone https://github.com/your-org/sharding-system.git
cd sharding-system

# Start all services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f
```

This starts:
- **etcd**: Metadata store (port 2389)
- **Router**: Query router (port 8080)
- **Manager**: Shard manager (port 8081)
- **PostgreSQL shards**: Example shard databases (ports 5432, 5433)

### Access Services

- **Router API**: `http://localhost:8080`
- **Manager API**: `http://localhost:8081`
- **Web UI**: `http://localhost:3000` (if UI is included)
- **Metrics**: 
  - Router: `http://localhost:9090/metrics`
  - Manager: `http://localhost:9091/metrics`

### Stopping Services

```bash
# Stop all services
docker-compose down

# Stop and remove volumes (clears data)
docker-compose down -v
```

### Manual Local Deployment

If you prefer to run services manually:

1. **Start etcd**
   ```bash
   docker run -d \
     --name etcd \
     -p 2389:2379 \
     quay.io/coreos/etcd:v3.5.10
   ```

2. **Build Binaries**
   ```bash
   make build-backend
   ```

3. **Start Router**
   ```bash
   ./bin/router
   ```

4. **Start Manager**
   ```bash
   export JWT_SECRET="your-secret-key-min-32-chars"
   ./bin/manager
   ```

## Docker Deployment

### Building Docker Images

```bash
# Build Router image
docker build -f Dockerfile.router -t sharding-router:latest .

# Build Manager image
docker build -f Dockerfile.manager -t sharding-manager:latest .
```

### Running Individual Containers

**Router:**
```bash
docker run -d \
  --name sharding-router \
  -p 8080:8080 \
  -p 9090:9090 \
  -e CONFIG_PATH=/root/configs/router.json \
  -e METADATA_ENDPOINTS=etcd:2379 \
  --network sharding-network \
  sharding-router:latest
```

**Manager:**
```bash
docker run -d \
  --name sharding-manager \
  -p 8081:8081 \
  -p 9091:9091 \
  -e CONFIG_PATH=/root/configs/manager.json \
  -e METADATA_ENDPOINTS=etcd:2379 \
  -e JWT_SECRET="your-secret-key-min-32-chars" \
  --network sharding-network \
  sharding-manager:latest
```

### Using Docker Compose

See the `docker-compose.yml` file for a complete setup. Customize as needed:

```yaml
services:
  router:
    build:
      context: .
      dockerfile: Dockerfile.router
    environment:
      - CONFIG_PATH=/root/configs/router.json
      - METADATA_ENDPOINTS=etcd:2379
    ports:
      - "8080:8080"
      - "9090:9090"
    depends_on:
      - etcd
```

## Kubernetes Deployment

### Prerequisites

- Kubernetes cluster (1.21+)
- kubectl configured
- Helm 3.x installed
- etcd cluster or PostgreSQL available

### Using Helm Charts

The project includes Helm charts for easy Kubernetes deployment:

1. **Add Helm Repository** (if applicable)
   ```bash
   helm repo add sharding-system https://charts.example.com/sharding-system
   helm repo update
   ```

2. **Deploy with Default Values**
   ```bash
   helm install sharding-system ./deploy/helm/sharding-system
   ```

3. **Deploy with Custom Values**
   ```bash
   helm install sharding-system ./deploy/helm/sharding-system \
     -f custom-values.yaml
   ```

### Custom Values Example

Create `custom-values.yaml`:

```yaml
manager:
  replicaCount: 3
  resources:
    limits:
      cpu: 1000m
      memory: 1Gi
    requests:
      cpu: 500m
      memory: 512Mi

router:
  replicaCount: 5
  resources:
    limits:
      cpu: 1000m
      memory: 1Gi

etcd:
  endpoints: ["etcd-0.etcd:2379", "etcd-1.etcd:2379", "etcd-2.etcd:2379"]

config:
  security:
    enableRBAC: true
  sharding:
    vnodeCount: 512
```

### Manual Kubernetes Deployment

If not using Helm:

1. **Create Namespace**
   ```bash
   kubectl create namespace sharding-system
   ```

2. **Deploy etcd** (if not using external etcd)
   ```bash
   kubectl apply -f k8s/namespace.yaml
   # Deploy etcd operator or statefulset
   ```

3. **Create ConfigMap**
   ```bash
   kubectl apply -f k8s/configmap.yaml
   ```

4. **Deploy Manager**
   ```bash
   kubectl apply -f k8s/manager-deployment.yaml
   ```

5. **Deploy Router**
   ```bash
   kubectl apply -f k8s/router-deployment.yaml
   ```

### Service Exposure

**For Development:**
```bash
kubectl port-forward svc/sharding-manager 8081:8081
kubectl port-forward svc/sharding-router 8080:8080
```

**For Production:**
- Use LoadBalancer or Ingress
- Configure TLS/SSL
- Set up proper DNS

### Scaling

**Scale Router:**
```bash
kubectl scale deployment sharding-router --replicas=5 -n sharding-system
```

**Scale Manager:**
```bash
kubectl scale deployment sharding-manager --replicas=3 -n sharding-system
```

**Note:** Manager scaling requires leader election (not yet implemented).

## Production Considerations

### High Availability

1. **etcd Cluster**
   - Deploy 3+ etcd nodes
   - Use etcd operator for management
   - Regular backups

2. **Router Replicas**
   - Deploy multiple router replicas
   - Use load balancer for distribution
   - Stateless design enables easy scaling

3. **Manager Replicas**
   - Currently single instance recommended
   - Future: Leader election for HA

4. **Database Shards**
   - Configure replicas for each shard
   - Set up automatic failover
   - Regular backups

### Security

1. **Enable RBAC**
   ```json
   {
     "security": {
       "enable_rbac": true
     }
   }
   ```

2. **Set JWT Secret**
   ```bash
   export JWT_SECRET="strong-secret-min-32-chars"
   ```

3. **Enable TLS**
   ```json
   {
     "security": {
       "enable_tls": true
     }
   }
   ```

4. **Network Policies**
   - Restrict network access
   - Use Kubernetes NetworkPolicies
   - Limit ingress/egress

5. **Secrets Management**
   - Use Kubernetes Secrets
   - Or external secret manager (Vault, AWS Secrets Manager)
   - Never commit secrets

### Performance Tuning

1. **Connection Pooling**
   ```json
   {
     "sharding": {
       "max_connections": 200,
       "connection_ttl": "10m"
     }
   }
   ```

2. **Virtual Nodes**
   ```json
   {
     "sharding": {
       "vnode_count": 512
     }
   }
   ```

3. **Resource Limits**
   ```yaml
   resources:
     limits:
       cpu: 2000m
       memory: 2Gi
     requests:
       cpu: 1000m
       memory: 1Gi
   ```

### Monitoring

1. **Prometheus**
   - Scrape metrics from `/metrics` endpoints
   - Configure alerts for:
     - High latency (p95 > 100ms)
     - Error rate > 1%
     - Shard health issues

2. **Grafana Dashboards**
   - Use provided dashboard JSON
   - Monitor key metrics:
     - Query latency
     - Throughput
     - Shard health
     - Resharding progress

3. **Logging**
   - Centralized logging (ELK, Loki)
   - Structured JSON logs
   - Log aggregation and analysis

### Backup and Recovery

1. **Metadata Backup**
   - Regular etcd backups
   - Export shard catalog
   - Store backups securely

2. **Shard Backup**
   - Database-level backups
   - Point-in-time recovery
   - Test restore procedures

3. **Disaster Recovery**
   - Document recovery procedures
   - Regular DR drills
   - Multi-region deployment (future)

## Monitoring and Observability

### Health Checks

**Manager:**
```bash
curl http://localhost:8081/api/v1/health
```

**Router:**
```bash
curl http://localhost:8080/v1/health
```

### Metrics Endpoints

**Manager Metrics:**
```bash
curl http://localhost:9091/metrics
```

**Router Metrics:**
```bash
curl http://localhost:9090/metrics
```

### Key Metrics

- `sharding_queries_total`: Total queries processed
- `sharding_query_duration_seconds`: Query latency
- `sharding_shard_health`: Shard health status
- `sharding_reshard_jobs`: Resharding job status

### Logging

**View Logs:**
```bash
# Docker Compose
docker-compose logs -f router
docker-compose logs -f manager

# Kubernetes
kubectl logs -f deployment/sharding-router -n sharding-system
kubectl logs -f deployment/sharding-manager -n sharding-system
```

**Log Levels:**
- `debug`: Verbose logging
- `info`: Standard logging (production)
- `warn`: Warnings only
- `error`: Errors only

## Troubleshooting

### Common Issues

#### Services Won't Start

**Problem:** Containers exit immediately

**Solutions:**
1. Check logs: `docker-compose logs`
2. Verify etcd is running: `docker-compose ps etcd`
3. Check configuration files
4. Verify ports are not in use

#### Cannot Connect to etcd

**Problem:** "connection refused" errors

**Solutions:**
1. Verify etcd is running: `docker-compose ps etcd`
2. Check endpoint in config: `localhost:2389`
3. Verify network connectivity
4. Check firewall rules

#### Authentication Errors

**Problem:** "Authentication required" errors

**Solutions:**
1. Verify RBAC is enabled
2. Check JWT_SECRET is set
3. Obtain valid token via login endpoint
4. Include token in Authorization header

#### High Latency

**Problem:** Queries are slow

**Solutions:**
1. Check database performance
2. Verify shard health
3. Review connection pool settings
4. Check network latency
5. Monitor metrics for bottlenecks

### Debugging

**Enable Debug Logging:**
```json
{
  "observability": {
    "log_level": "debug"
  }
}
```

**Check Service Status:**
```bash
# Docker
docker-compose ps

# Kubernetes
kubectl get pods -n sharding-system
kubectl describe pod <pod-name> -n sharding-system
```

**View Resource Usage:**
```bash
# Docker
docker stats

# Kubernetes
kubectl top pods -n sharding-system
```

## Additional Resources

- [Configuration Guide](../user/CONFIGURATION_GUIDE.md)
- [Architecture Documentation](../architecture/ARCHITECTURE.md)
- [API Reference](../api/API_REFERENCE.md)
- [User Guide](../user/USER_GUIDE.md)

