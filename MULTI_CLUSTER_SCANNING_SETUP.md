# Multi-Cluster Database Scanning Setup

This document describes the setup for multi-cluster database scanning with PostgreSQL metrics collection.

## Overview

The system now supports:
1. **Multi-cluster scanning** - Discover and scan databases across multiple Kubernetes clusters
2. **5 client applications** - 3 Go apps and 2 Java apps, each with PostgreSQL databases
3. **Real-time metrics** - PostgreSQL statistics exposed at `/metrics` endpoint
4. **Automatic registration** - Discovered databases are automatically registered for metrics collection

## Client Applications

### Go Applications
1. **go-app-1** (E-commerce Service)
   - Namespace: `ecommerce-ns`
   - Database: `ecommerce_db`
   - Port: 8080

2. **go-app-2** (Users Service)
   - Namespace: `users-ns`
   - Database: `users_db`
   - Port: 8080

3. **go-app-3** (Orders Service)
   - Namespace: `orders-ns`
   - Database: `orders_db`
   - Port: 8080

### Java Applications
4. **java-app-1** (Products Service)
   - Namespace: `products-ns`
   - Database: `products_db`
   - Port: 8080

5. **java-app-2** (Inventory Service)
   - Namespace: `inventory-ns`
   - Database: `inventory_db`
   - Port: 8080

## Quick Start

### 1. Build All Applications

```bash
# Build client applications
cd clients
./build-and-test.sh

# Build main manager service
cd ..
make build-backend
```

### 2. Deploy Client Applications

```bash
# Deploy all applications
cd clients
./deploy-all.sh

# Or deploy individually
kubectl apply -f go-app-1/k8s/deployment.yaml
kubectl apply -f go-app-2/k8s/deployment.yaml
kubectl apply -f go-app-3/k8s/deployment.yaml
kubectl apply -f java-app-1/k8s/deployment.yaml
kubectl apply -f java-app-2/k8s/deployment.yaml
```

### 3. Start Manager Service

```bash
# Start manager (make sure etcd is running)
make start-backend

# Or run directly
./bin/manager
```

### 4. Register Kubernetes Cluster

```bash
curl -X POST http://localhost:8081/api/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "name": "local-cluster",
    "type": "on-premise",
    "provider": "local"
  }'
```

### 5. Scan All Clusters

```bash
curl -X POST http://localhost:8081/api/v1/clusters/scan \
  -H "Content-Type: application/json" \
  -d '{
    "deep_scan": true
  }'
```

### 6. View Metrics

```bash
# View Prometheus metrics
curl http://localhost:8081/metrics

# Metrics include:
# - Connection statistics
# - Query performance metrics
# - Replication statistics
# - Table statistics
# - Index statistics
# - Lock statistics
# - Background writer statistics
```

## Architecture

### Discovery Flow

1. **Cluster Registration**: Register Kubernetes clusters via API
2. **Application Discovery**: System discovers applications in all namespaces
3. **Database Detection**: Extracts database connection info from environment variables
4. **Database Scanning**: Performs deep scan of discovered databases
5. **Metrics Registration**: Automatically registers databases for metrics collection

### Metrics Collection

- **PrometheusCollector**: Collects standard Prometheus metrics
- **PostgresStatsCollector**: Collects detailed PostgreSQL statistics
- **Automatic Registration**: Databases are registered when discovered during scan
- **Real-time Updates**: Metrics are collected every 30 seconds

### Environment Variables for Discovery

Applications should expose these environment variables for discovery:

- `DATABASE_URL` - Full database connection URL
- `DATABASE_HOST` - Database host
- `DATABASE_PORT` - Database port
- `DATABASE_NAME` - Database name
- `DATABASE_USER` - Database user

Or alternatively:

- `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`

## API Endpoints

### Cluster Management

- `POST /api/v1/clusters` - Register a cluster
- `GET /api/v1/clusters` - List all clusters
- `GET /api/v1/clusters/{id}` - Get cluster details
- `DELETE /api/v1/clusters/{id}` - Delete a cluster

### Scanning

- `POST /api/v1/clusters/scan` - Scan all clusters (or specific clusters)
- `GET /api/v1/clusters/scan/results` - Get scan results

### Metrics

- `GET /metrics` - Prometheus metrics endpoint

## Database Statistics Collected

### Connection Statistics
- Active connections
- Idle connections
- Waiting connections
- Max connections
- Connection usage percentage

### Query Statistics
- Total queries
- Queries per second
- Average query time
- Cache hit ratio

### Replication Statistics
- Replication lag
- Replica count
- WAL position
- Replica status

### Table Statistics
- Total tables
- Total rows
- Live/dead tuples
- Sequential vs index scans
- Largest tables

### Index Statistics
- Total indexes
- Index size
- Index hit ratio
- Unused indexes

### Lock Statistics
- Total locks
- Granted/waiting locks
- Deadlocks
- Locks by type/mode

### Background Writer Statistics
- Checkpoints
- Buffer statistics
- Write statistics

## Troubleshooting

### Applications Not Discovered

1. Check that applications have database environment variables set
2. Verify namespace is not in system namespaces list
3. Check application labels and annotations

### Metrics Not Showing

1. Verify database scan completed successfully
2. Check database connection credentials
3. Ensure PostgreSQL is accessible from manager service
4. Check logs for registration errors

### Build Issues

```bash
# Clean and rebuild
make clean
make build-backend

# For client apps
cd clients
./build-and-test.sh
```

## Next Steps

1. Deploy applications to different Kubernetes clusters
2. Register multiple clusters
3. Scan all clusters
4. View metrics in Prometheus/Grafana
5. Set up alerts based on metrics

