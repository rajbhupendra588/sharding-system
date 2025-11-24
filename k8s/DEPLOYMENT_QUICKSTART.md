# Quick Deployment Guide

## One-Command Deployment

```bash
# Make sure you're in the project root
cd /path/to/sharding-system

# Run the deployment script
./k8s/deploy.sh
```

That's it! The script will:
1. ✅ Create namespace
2. ✅ Set up secrets (auto-generates JWT if not provided)
3. ✅ Create ConfigMap
4. ✅ Configure RBAC for Kubernetes discovery
5. ✅ Deploy Manager and Router
6. ✅ Wait for pods to be ready
7. ✅ Show status and access information

## Prerequisites Check

Before deploying, ensure:

```bash
# 1. kubectl is installed and configured
kubectl version --client

# 2. You can connect to your cluster
kubectl cluster-info

# 3. You have cluster admin permissions (for RBAC)
kubectl auth can-i create clusterrolebindings
```

## Custom Configuration

### Set Custom JWT Secret

```bash
export JWT_SECRET="your-secure-random-string-min-32-chars"
./k8s/deploy.sh
```

### Update Image Names

Edit `k8s/manager-deployment.yaml` and `k8s/router-deployment.yaml`:

```yaml
image: your-registry/sharding-manager:v1.0.0
image: your-registry/sharding-router:v1.0.0
```

### Update etcd Endpoints

Edit `k8s/configmap.yaml`:

```yaml
"endpoints": ["your-etcd-0:2379", "your-etcd-1:2379", "your-etcd-2:2379"]
```

## Access Services

### Option 1: LoadBalancer (if supported)

```bash
# Get external IPs
kubectl get svc -n sharding-system

# Access APIs
curl http://<MANAGER_IP>:8081/api/v1/health
curl http://<ROUTER_IP>:8080/v1/health
```

### Option 2: Port Forwarding

```bash
# Terminal 1 - Manager
kubectl port-forward svc/manager 8081:8081 -n sharding-system

# Terminal 2 - Router  
kubectl port-forward svc/router 8080:8080 -n sharding-system

# Test
curl http://localhost:8081/api/v1/health
curl http://localhost:8080/v1/health
```

## Verify Deployment

```bash
# Check all pods are running
kubectl get pods -n sharding-system

# Check services
kubectl get svc -n sharding-system

# View logs
kubectl logs -f deployment/manager -n sharding-system
kubectl logs -f deployment/router -n sharding-system
```

## Kubernetes Discovery

The system automatically discovers applications in your cluster! Access the UI and click "Discover Apps" to see all applications and their databases.

## Troubleshooting

### Pods not starting?

```bash
# Check pod status
kubectl describe pod <pod-name> -n sharding-system

# Check logs
kubectl logs <pod-name> -n sharding-system
```

### Can't access services?

```bash
# Check service endpoints
kubectl get endpoints -n sharding-system

# Verify port forwarding
kubectl port-forward svc/manager 8081:8081 -n sharding-system
```

### Discovery not working?

```bash
# Check RBAC
kubectl get clusterrolebinding sharding-discovery-binding

# Test permissions
kubectl auth can-i list deployments \
  --as=system:serviceaccount:sharding-system:sharding-manager \
  --all-namespaces
```

## Cleanup

```bash
# Remove everything
kubectl delete -f k8s/ -n sharding-system

# Remove RBAC
kubectl delete clusterrole sharding-discovery
kubectl delete clusterrolebinding sharding-discovery-binding
```

## Next Steps

1. **Access the UI**: Set up port forwarding and access the web interface
2. **Discover Apps**: Use the "Discover Apps" feature to auto-register applications
3. **Create Shards**: Register client applications and create shards for them
4. **Monitor**: Set up Prometheus/Grafana for metrics

For detailed information, see `k8s/README.md`.

