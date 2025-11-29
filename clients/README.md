# Client Applications

This directory contains 5 client applications (3 Go, 2 Java) with PostgreSQL databases for testing the multi-cluster database scanning and metrics collection features.

## Applications

1. **go-app-1** (E-commerce Service) - Namespace: `ecommerce-ns`
2. **go-app-2** (Users Service) - Namespace: `users-ns`
3. **go-app-3** (Orders Service) - Namespace: `orders-ns`
4. **java-app-1** (Products Service) - Namespace: `products-ns`
5. **java-app-2** (Inventory Service) - Namespace: `inventory-ns`

## Quick Start

### 1. Build All Applications

```bash
./build-and-test.sh
```

### 2. Build Docker Images and Deploy

```bash
./deploy-all.sh
```

### 3. Manual Deployment

Deploy individual applications:

```bash
# Deploy Go apps
kubectl apply -f go-app-1/k8s/deployment.yaml
kubectl apply -f go-app-2/k8s/deployment.yaml
kubectl apply -f go-app-3/k8s/deployment.yaml

# Deploy Java apps
kubectl apply -f java-app-1/k8s/deployment.yaml
kubectl apply -f java-app-2/k8s/deployment.yaml
```

### 4. Verify Deployments

```bash
# Check all namespaces
kubectl get pods -n ecommerce-ns
kubectl get pods -n users-ns
kubectl get pods -n orders-ns
kubectl get pods -n products-ns
kubectl get pods -n inventory-ns
```

### 5. Register Cluster and Scan

```bash
# Register the current Kubernetes cluster
curl -X POST http://localhost:8081/api/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "name": "local-cluster",
    "type": "on-premise",
    "provider": "local"
  }'

# Scan all clusters
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
```

## Database Connection

All applications use environment variables for database configuration:
- `DB_HOST` - Database host
- `DB_PORT` - Database port (default: 5432)
- `DB_USER` - Database user (default: postgres)
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name

The scanner discovers databases using:
- `DATABASE_URL` - Full database connection URL
- `DATABASE_HOST` - Database host
- `DATABASE_PORT` - Database port
- `DATABASE_NAME` - Database name
- `DATABASE_USER` - Database user

## Testing

Each application exposes a `/health` endpoint:

```bash
# Test ecommerce app
kubectl port-forward -n ecommerce-ns svc/ecommerce-app 8080:8080
curl http://localhost:8080/health

# Test users app
kubectl port-forward -n users-ns svc/users-app 8081:8080
curl http://localhost:8081/health
```

## Metrics Collection

After scanning, databases are automatically registered for metrics collection. Metrics are available at `/metrics` endpoint and include:

- Connection statistics
- Query performance metrics
- Replication statistics
- Table statistics
- Index statistics
- Lock statistics
- Background writer statistics

