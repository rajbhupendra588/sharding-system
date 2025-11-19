# Production Deployment Guide

This guide provides comprehensive instructions for deploying the Sharding System to production environments.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Architecture Overview](#architecture-overview)
3. [Environment Configuration](#environment-configuration)
4. [Docker Deployment](#docker-deployment)
5. [Kubernetes Deployment](#kubernetes-deployment)
6. [Security Hardening](#security-hardening)
7. [Monitoring & Observability](#monitoring--observability)
8. [Backup & Recovery](#backup--recovery)
9. [Scaling Guidelines](#scaling-guidelines)
10. [Troubleshooting](#troubleshooting)

## Prerequisites

### Required Components

- **etcd**: Version 3.5+ for metadata storage
- **PostgreSQL**: Version 12+ for shard databases
- **Load Balancer**: For high availability
- **Monitoring**: Prometheus + Grafana
- **Logging**: Centralized logging solution (ELK, Loki, etc.)

### Infrastructure Requirements

- **CPU**: Minimum 2 cores per service instance
- **Memory**: Minimum 4GB RAM per service instance
- **Storage**: SSD recommended for etcd and databases
- **Network**: Low latency between services (<10ms)

## Architecture Overview

```
┌─────────────────┐
│   Load Balancer │
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
┌───▼───┐ ┌──▼───┐
│Router │ │Router│ (Multiple instances)
└───┬───┘ └──┬───┘
    │        │
    └────┬───┘
         │
    ┌────▼────┐
    │  etcd   │ (Cluster)
    └────┬────┘
         │
┌────────▼────────┐
│   Manager       │ (Multiple instances)
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
┌───▼───┐ ┌──▼───┐
│Shard1 │ │Shard2│ (PostgreSQL clusters)
└───────┘ └──────┘
```

## Environment Configuration

### Backend Configuration

Create production configuration files:

**configs/router.prod.json**:
```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8080,
    "read_timeout": "30s",
    "write_timeout": "30s",
    "idle_timeout": "120s"
  },
  "metadata": {
    "type": "etcd",
    "endpoints": ["etcd-1:2379", "etcd-2:2379", "etcd-3:2379"],
    "timeout": "5s"
  },
  "sharding": {
    "strategy": "hash",
    "hash_function": "murmur3",
    "vnode_count": 256,
    "replica_policy": "replica_ok",
    "max_connections": 100,
    "connection_ttl": "5m"
  },
  "security": {
    "enable_tls": true,
    "enable_rbac": true,
    "jwt_secret": "${JWT_SECRET}"
  },
  "observability": {
    "metrics_port": 9090,
    "enable_tracing": true,
    "log_level": "info"
  }
}
```

**configs/manager.prod.json**:
```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8081,
    "read_timeout": "30s",
    "write_timeout": "30s",
    "idle_timeout": "120s"
  },
  "metadata": {
    "type": "etcd",
    "endpoints": ["etcd-1:2379", "etcd-2:2379", "etcd-3:2379"],
    "timeout": "5s"
  },
  "security": {
    "enable_tls": true,
    "enable_rbac": true,
    "jwt_secret": "${JWT_SECRET}",
    "audit_log_path": "/var/log/sharding/audit.log"
  },
  "observability": {
    "metrics_port": 9091,
    "enable_tracing": true,
    "log_level": "info"
  }
}
```

### Environment Variables

Set these environment variables:

```bash
# JWT Secret (generate a strong secret)
export JWT_SECRET="your-super-secret-jwt-key-min-32-chars"

# Database credentials (use secrets management)
export DB_USERNAME="sharding_user"
export DB_PASSWORD="secure-password"

# etcd endpoints
export ETCD_ENDPOINTS="etcd-1:2379,etcd-2:2379,etcd-3:2379"

# Log level
export LOG_LEVEL="info"
```

## Docker Deployment

### Build Images

```bash
# Build router image
docker build -f Dockerfile.router -t sharding-router:latest .

# Build manager image
docker build -f Dockerfile.manager -t sharding-manager:latest .
```

### Docker Compose for Production

Create `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  etcd:
    image: quay.io/coreos/etcd:v3.5.9
    environment:
      - ETCD_NAME=etcd1
      - ETCD_DATA_DIR=/etcd-data
      - ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:2379
      - ETCD_ADVERTISE_CLIENT_URLS=http://etcd:2379
    volumes:
      - etcd-data:/etcd-data
    networks:
      - sharding-network

  router:
    image: sharding-router:latest
    ports:
      - "8080:8080"
      - "9090:9090"
    environment:
      - CONFIG_PATH=/app/configs/router.prod.json
      - JWT_SECRET=${JWT_SECRET}
    volumes:
      - ./configs:/app/configs:ro
    depends_on:
      - etcd
    networks:
      - sharding-network
    restart: unless-stopped
    deploy:
      replicas: 3

  manager:
    image: sharding-manager:latest
    ports:
      - "8081:8081"
      - "9091:9091"
    environment:
      - CONFIG_PATH=/app/configs/manager.prod.json
      - JWT_SECRET=${JWT_SECRET}
    volumes:
      - ./configs:/app/configs:ro
      - audit-logs:/var/log/sharding
    depends_on:
      - etcd
    networks:
      - sharding-network
    restart: unless-stopped
    deploy:
      replicas: 2

volumes:
  etcd-data:
  audit-logs:

networks:
  sharding-network:
    driver: bridge
```

### Deploy

```bash
docker-compose -f docker-compose.prod.yml up -d
```

## Kubernetes Deployment

### Namespace

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: sharding-system
```

### ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: sharding-config
  namespace: sharding-system
data:
  router.json: |
    {
      "server": {
        "host": "0.0.0.0",
        "port": 8080
      },
      "metadata": {
        "endpoints": ["etcd-0.etcd:2379", "etcd-1.etcd:2379", "etcd-2.etcd:2379"]
      }
    }
  manager.json: |
    {
      "server": {
        "host": "0.0.0.0",
        "port": 8081
      },
      "metadata": {
        "endpoints": ["etcd-0.etcd:2379", "etcd-1.etcd:2379", "etcd-2.etcd:2379"]
      }
    }
```

### Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: sharding-secrets
  namespace: sharding-system
type: Opaque
stringData:
  jwt-secret: "your-jwt-secret-here"
  db-password: "your-db-password"
```

### Router Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: router
  namespace: sharding-system
spec:
  replicas: 3
  selector:
    matchLabels:
      app: router
  template:
    metadata:
      labels:
        app: router
    spec:
      containers:
      - name: router
        image: sharding-router:latest
        ports:
        - containerPort: 8080
        - containerPort: 9090
        env:
        - name: CONFIG_PATH
          value: "/app/configs/router.json"
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: sharding-secrets
              key: jwt-secret
        volumeMounts:
        - name: config
          mountPath: /app/configs
          readOnly: true
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /v1/health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /v1/health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: sharding-config
```

### Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: router
  namespace: sharding-system
spec:
  selector:
    app: router
  ports:
  - name: http
    port: 8080
    targetPort: 8080
  - name: metrics
    port: 9090
    targetPort: 9090
  type: LoadBalancer
```

## Security Hardening

### TLS Configuration

1. Generate certificates:
```bash
# Generate CA
openssl genrsa -out ca.key 4096
openssl req -new -x509 -days 365 -key ca.key -out ca.crt

# Generate server certificate
openssl genrsa -out server.key 4096
openssl req -new -key server.key -out server.csr
openssl x509 -req -days 365 -in server.csr -CA ca.crt -CAkey ca.key -out server.crt
```

2. Update configuration to use TLS:
```json
{
  "security": {
    "enable_tls": true,
    "tls_cert_path": "/etc/ssl/certs/server.crt",
    "tls_key_path": "/etc/ssl/private/server.key"
  }
}
```

### Network Security

- Use firewall rules to restrict access
- Implement network policies in Kubernetes
- Use VPN or private networks for internal communication
- Enable rate limiting on API endpoints

### Authentication

- Use strong JWT secrets (minimum 32 characters)
- Implement token rotation
- Set appropriate token expiration times
- Use HTTPS for all communications

## Monitoring & Observability

### Prometheus Configuration

```yaml
scrape_configs:
  - job_name: 'router'
    static_configs:
      - targets: ['router:9090']
  - job_name: 'manager'
    static_configs:
      - targets: ['manager:9091']
```

### Key Metrics to Monitor

- Request rate (QPS)
- Latency (p50, p95, p99)
- Error rate
- Connection pool usage
- Shard health status
- Resharding job progress

### Alerting Rules

```yaml
groups:
  - name: sharding_alerts
    rules:
      - alert: HighErrorRate
        expr: rate(shard_queries_total{status="error"}[5m]) > 0.1
        for: 5m
        annotations:
          summary: "High error rate detected"
      
      - alert: ShardDown
        expr: shard_health_status == 0
        for: 1m
        annotations:
          summary: "Shard is down"
```

## Backup & Recovery

### etcd Backup

```bash
# Backup etcd
etcdctl snapshot save /backup/etcd-snapshot.db \
  --endpoints=etcd-1:2379,etcd-2:2379,etcd-3:2379

# Restore from backup
etcdctl snapshot restore /backup/etcd-snapshot.db \
  --data-dir=/etcd-data-restored
```

### Database Backup

Use PostgreSQL native backup tools:
```bash
# Backup shard database
pg_dump -h shard-host -U user -d shard_db > shard_backup.sql

# Restore
psql -h shard-host -U user -d shard_db < shard_backup.sql
```

## Scaling Guidelines

### Horizontal Scaling

- **Router**: Scale based on QPS (target: <1000 QPS per instance)
- **Manager**: Scale based on resharding operations (usually 2-3 instances)
- **etcd**: Use 3 or 5 nodes for HA

### Vertical Scaling

- Monitor CPU and memory usage
- Increase resources if consistently >80% utilization
- Use connection pooling to optimize database connections

## Troubleshooting

### Common Issues

1. **Connection refused to etcd**
   - Check etcd cluster health
   - Verify network connectivity
   - Check firewall rules

2. **High latency**
   - Check database connection pool
   - Monitor network latency
   - Review query performance

3. **Resharding stuck**
   - Check resharding job logs
   - Verify database connectivity
   - Check available disk space

### Log Locations

- Router logs: `/var/log/sharding/router.log`
- Manager logs: `/var/log/sharding/manager.log`
- Audit logs: `/var/log/sharding/audit.log`

### Debug Mode

Enable debug logging:
```json
{
  "observability": {
    "log_level": "debug"
  }
}
```

## Next Steps

1. Set up monitoring dashboards
2. Configure alerting
3. Implement backup automation
4. Set up CI/CD pipeline
5. Perform load testing
6. Document runbooks for common operations

