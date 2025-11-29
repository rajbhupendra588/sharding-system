# Quick Start Guide

## Prerequisites

- Kubernetes cluster (minikube, kind, or cloud)
- kubectl configured
- Docker
- Go 1.21+ (for Go apps)
- Maven 3.9+ (for Java apps)

## Quick Deploy

1. **Build all applications:**
   ```bash
   cd clients
   ./build-all.sh
   ```

2. **Deploy to Kubernetes:**
   ```bash
   ./deploy-all.sh
   ```

3. **Verify deployments:**
   ```bash
   kubectl get pods --all-namespaces | grep -E "(ecommerce|inventory|order|payment|user)"
   ```

4. **Test applications:**
   ```bash
   ./test-all.sh
   ```

## Manual Testing

### Ecommerce Service
```bash
kubectl port-forward -n ecommerce-cluster-1 svc/ecommerce-service 8080:8080
curl http://localhost:8080/health
curl http://localhost:8080/api/products
curl http://localhost:8080/api/products/stats
```

### Inventory Service
```bash
kubectl port-forward -n inventory-cluster-2 svc/inventory-service 8080:8080
curl http://localhost:8080/health
curl http://localhost:8080/api/inventory
curl http://localhost:8080/api/inventory/stats
```

### Order Service
```bash
kubectl port-forward -n orders-cluster-1 svc/order-service 8080:8080
curl http://localhost:8080/health
curl http://localhost:8080/api/orders
curl http://localhost:8080/api/orders/stats
```

### Payment Service
```bash
kubectl port-forward -n payments-cluster-2 svc/payment-service 8080:8080
curl http://localhost:8080/health
curl http://localhost:8080/api/payments
curl http://localhost:8080/api/payments/stats
```

### User Service
```bash
kubectl port-forward -n users-cluster-3 svc/user-service 8080:8080
curl http://localhost:8080/health
curl http://localhost:8080/api/users
curl http://localhost:8080/api/users/stats
```

## Scanning Databases

After deploying, register the cluster and scan databases:

1. **Register cluster** (if using local k8s):
   ```bash
   curl -X POST http://localhost:8081/api/v1/clusters \
     -H "Content-Type: application/json" \
     -d '{
       "name": "local-cluster",
       "type": "kubernetes",
       "provider": "local",
       "kubeconfig": "~/.kube/config",
       "context": "minikube"
     }'
   ```

2. **Scan all databases:**
   ```bash
   curl -X POST http://localhost:8081/api/v1/clusters/scan \
     -H "Content-Type: application/json" \
     -d '{
       "database_password": "postgres123"
     }'
   ```

3. **View metrics:**
   ```bash
   curl http://localhost:8081/metrics | grep postgres
   ```

## Cleanup

```bash
# Delete all namespaces
kubectl delete namespace ecommerce-cluster-1
kubectl delete namespace inventory-cluster-2
kubectl delete namespace orders-cluster-1
kubectl delete namespace payments-cluster-2
kubectl delete namespace users-cluster-3
```

