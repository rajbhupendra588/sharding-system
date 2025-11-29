# Complete Kubernetes Deployment Guide

This guide covers deploying the entire sharding-system and all client applications to Kubernetes.

## Overview

The deployment includes:
- **Sharding System Components**: etcd, Manager, Router, Proxy
- **Client Applications**: go-app-1, go-app-2, go-app-3, java-app-1, java-app-2, java-app-3

## Prerequisites

1. **Kubernetes Cluster** (version 1.21+)
   - Minikube, Kind, or cloud provider cluster (GKE, EKS, AKS)
   - At least 4 CPU cores and 8GB RAM available

2. **kubectl** configured and connected to your cluster
   ```bash
   kubectl cluster-info
   ```

3. **Docker images built** (or use a container registry)
   - Build images: `docker build -t <image-name>:latest .`
   - Or push to registry and update image names in YAML files

4. **Optional**: `openssl` for generating secure secrets

## Quick Start - Deploy Everything

### Automated Deployment (Recommended)

Deploy all components with a single command:

```bash
cd /path/to/sharding-system
./k8s/deploy-all.sh
```

This script will:
1. Deploy sharding-system components (etcd, manager, router, proxy)
2. Deploy all client applications
3. Wait for all deployments to be ready
4. Display service endpoints and status

### Step-by-Step Deployment

#### Step 1: Deploy Sharding System

```bash
cd k8s
./deploy.sh
```

Or manually:
```bash
# Create namespace
kubectl apply -f k8s/namespace.yaml

# Create secrets (update with your secure values first!)
kubectl apply -f k8s/secrets.yaml

# Create ConfigMap
kubectl apply -f k8s/configmap.yaml

# Set up RBAC
kubectl apply -f k8s/rbac-discovery.yaml

# Deploy etcd cluster
kubectl apply -f k8s/etcd-deployment.yaml

# Deploy Manager
kubectl apply -f k8s/manager-deployment.yaml

# Deploy Router
kubectl apply -f k8s/router-deployment.yaml

# Deploy Proxy
kubectl apply -f k8s/proxy-deployment.yaml
```

#### Step 2: Deploy Client Applications

```bash
cd clients

# Deploy Go applications
kubectl apply -f go-app-1/k8s/
kubectl apply -f go-app-2/k8s/
kubectl apply -f go-app-3/k8s/

# Deploy Java applications
kubectl apply -f java-app-1/k8s/
kubectl apply -f java-app-2/k8s/
kubectl apply -f java-app-3/k8s/
```

## Component Details

### Sharding System Components

#### etcd (Metadata Storage)
- **StatefulSet**: 3 replicas for high availability
- **Service**: Headless service for peer communication
- **Storage**: 1Gi per pod (configurable)
- **Namespace**: `sharding-system`

#### Manager
- **Deployment**: 2 replicas
- **Service**: LoadBalancer on port 8081
- **Metrics**: Port 9091
- **Namespace**: `sharding-system`
- **Endpoints**: 
  - API: `http://manager:8081` (cluster-internal)
  - External: Use LoadBalancer IP or port-forward

#### Router
- **Deployment**: 3 replicas
- **Service**: LoadBalancer on port 8080
- **Metrics**: Port 9090
- **Namespace**: `sharding-system`

#### Proxy
- **Deployment**: 2 replicas
- **Service**: LoadBalancer
- **Ports**: 
  - PostgreSQL: 5432
  - Admin API: 8082
- **Namespace**: `sharding-system`

### Client Applications

#### Go Applications
- **go-app-1** (E-Commerce): Namespace `ecommerce-ns`
- **go-app-2** (Users): Namespace `users-ns`
- **go-app-3** (Orders): Namespace `orders-ns`

Each includes:
- Application Deployment
- PostgreSQL StatefulSet
- Database Service
- Application Service

#### Java Applications
- **java-app-1** (Products): Namespace `products-ns`
- **java-app-2** (Inventory): Namespace `inventory-ns`
- **java-app-3** (User Service): Namespace `users-cluster-3`

Each includes:
- Application Deployment
- PostgreSQL StatefulSet (where applicable)
- Database Service
- Application Service

## Configuration

### Secrets

Before deploying, update `k8s/secrets.yaml`:

```bash
# Generate secure JWT secret
export JWT_SECRET=$(openssl rand -base64 32)

# Update secrets.yaml or pass as environment variable
export JWT_SECRET="your-secure-secret"
./k8s/deploy.sh
```

### ConfigMap

Edit `k8s/configmap.yaml` to customize:
- etcd endpoints (default: `etcd-0.etcd:2379`, etc.)
- Server timeouts
- Sharding strategy
- Log levels

### Image Configuration

If using a container registry, update image names:

```yaml
# In deployment files
image: your-registry/sharding-manager:v1.0.0
imagePullPolicy: Always  # Change from IfNotPresent
```

For local development with Minikube/Kind:
```bash
# Load images into cluster
minikube image load sharding-manager:latest
minikube image load sharding-router:latest
minikube image load sharding-proxy:latest
```

## Accessing Services

### Port Forwarding (Development)

```bash
# Manager API
kubectl port-forward svc/manager 8081:8081 -n sharding-system

# Router API
kubectl port-forward svc/router 8080:8080 -n sharding-system

# Proxy (PostgreSQL)
kubectl port-forward svc/proxy 5432:5432 -n sharding-system

# Proxy Admin
kubectl port-forward svc/proxy 8082:8082 -n sharding-system

# Client Applications
kubectl port-forward svc/ecommerce-app 8080:8080 -n ecommerce-ns
kubectl port-forward svc/users-app 8080:8080 -n users-ns
# ... etc
```

### LoadBalancer (Production)

If your cluster supports LoadBalancer services:
```bash
# Get external IPs
kubectl get services -n sharding-system

# Access via external IPs
curl http://<manager-ip>:8081/api/v1/health
curl http://<router-ip>:8080/v1/health
```

### Ingress (Production)

For production, set up Ingress:
```bash
kubectl apply -f k8s/ingress.yaml
```

## Monitoring and Debugging

### Check Deployment Status

```bash
# All sharding-system components
kubectl get all -n sharding-system

# All client applications
kubectl get deployments,services --all-namespaces | grep -E "ecommerce|users|orders|products|inventory"

# Pod status
kubectl get pods --all-namespaces
```

### View Logs

```bash
# Manager logs
kubectl logs -f deployment/manager -n sharding-system

# Router logs
kubectl logs -f deployment/router -n sharding-system

# Proxy logs
kubectl logs -f deployment/proxy -n sharding-system

# etcd logs
kubectl logs -f etcd-0 -n sharding-system

# Client application logs
kubectl logs -f deployment/ecommerce-app -n ecommerce-ns
```

### Describe Resources

```bash
# Get detailed info about a deployment
kubectl describe deployment manager -n sharding-system

# Check events
kubectl get events -n sharding-system --sort-by='.lastTimestamp'
```

### Troubleshooting

**Pods not starting:**
```bash
# Check pod status
kubectl get pods -n sharding-system
kubectl describe pod <pod-name> -n sharding-system

# Check logs
kubectl logs <pod-name> -n sharding-system
```

**Services not accessible:**
```bash
# Verify service endpoints
kubectl get endpoints -n sharding-system

# Test connectivity from within cluster
kubectl run -it --rm debug --image=busybox --restart=Never -n sharding-system -- wget -O- http://manager:8081/api/v1/health
```

**etcd cluster issues:**
```bash
# Check etcd cluster status
kubectl exec -it etcd-0 -n sharding-system -- etcdctl endpoint health

# Check etcd member list
kubectl exec -it etcd-0 -n sharding-system -- etcdctl member list
```

## Scaling

### Horizontal Pod Autoscaling

HPA is configured for Manager and Router. Check status:
```bash
kubectl get hpa -n sharding-system
```

### Manual Scaling

```bash
# Scale Manager
kubectl scale deployment manager --replicas=3 -n sharding-system

# Scale Router
kubectl scale deployment router --replicas=5 -n sharding-system

# Scale Proxy
kubectl scale deployment proxy --replicas=3 -n sharding-system
```

## Cleanup

### Remove All Components

```bash
# Remove sharding-system
kubectl delete -f k8s/ -n sharding-system
kubectl delete namespace sharding-system

# Remove client applications
kubectl delete namespace ecommerce-ns users-ns orders-ns products-ns inventory-ns users-cluster-3
```

### Remove Specific Component

```bash
# Remove only proxy
kubectl delete -f k8s/proxy-deployment.yaml

# Remove specific client app
kubectl delete -f clients/go-app-1/k8s/
```

## Production Considerations

1. **Resource Limits**: Adjust CPU/memory requests and limits based on workload
2. **Storage**: Use persistent volumes for etcd and databases
3. **Security**: 
   - Enable TLS for etcd
   - Use secrets management (Vault, Sealed Secrets)
   - Enable network policies
4. **Monitoring**: Set up Prometheus and Grafana
5. **Backup**: Regular backups of etcd and database data
6. **High Availability**: Ensure multiple replicas and proper pod distribution

## Next Steps

- Configure Ingress for external access
- Set up monitoring and alerting
- Configure backup strategies
- Review and adjust resource limits
- Set up CI/CD pipelines for automated deployments

