# Kubernetes Deployment Guide

This directory contains all Kubernetes manifests for deploying the Sharding System to a Kubernetes cluster.

## Prerequisites

- Kubernetes cluster (1.21+)
- `kubectl` configured and connected to your cluster
- `openssl` or `base64` for generating secrets (optional)

## Quick Start

### Option 1: Automated Deployment (Recommended)

```bash
# Set JWT secret (optional - will be auto-generated if not set)
export JWT_SECRET="your-secure-random-string-at-least-32-chars"

# Run deployment script
./k8s/deploy.sh
```

### Option 2: Manual Deployment

Deploy resources in order:

```bash
# 1. Create namespace
kubectl apply -f k8s/namespace.yaml

# 2. Create secrets (update secrets.yaml first!)
kubectl apply -f k8s/secrets.yaml

# 3. Create ConfigMap
kubectl apply -f k8s/configmap.yaml

# 4. Set up RBAC for Kubernetes discovery
kubectl apply -f k8s/rbac-discovery.yaml

# 5. Deploy Manager
kubectl apply -f k8s/manager-deployment.yaml

# 6. Deploy Router
kubectl apply -f k8s/router-deployment.yaml
```

## Files Overview

### Core Deployment Files

- **`namespace.yaml`** - Creates the `sharding-system` namespace
- **`secrets.yaml`** - Contains sensitive data (JWT secrets, etc.)
- **`configmap.yaml`** - Contains application configuration
- **`manager-deployment.yaml`** - Manager service deployment and service
- **`router-deployment.yaml`** - Router service deployment and service
- **`rbac-discovery.yaml`** - RBAC for Kubernetes application discovery

### Deployment Script

- **`deploy.sh`** - Automated deployment script that handles all steps

## Configuration

### Secrets

Before deploying, update `secrets.yaml` with your secure values:

```bash
# Generate a secure JWT secret
openssl rand -base64 32

# Edit secrets.yaml and replace the jwt-secret value
vim k8s/secrets.yaml
```

Or set it as an environment variable:

```bash
export JWT_SECRET="your-secure-random-string"
./k8s/deploy.sh
```

### ConfigMap

Edit `configmap.yaml` to customize:
- etcd endpoints
- Database connection settings
- Logging configuration
- Other application settings

### Image Configuration

Update the image names in deployment files if using a custom registry:

```yaml
# In manager-deployment.yaml and router-deployment.yaml
image: your-registry/sharding-manager:latest
image: your-registry/sharding-router:latest
```

## Kubernetes Discovery

The system includes automatic discovery of applications running in Kubernetes namespaces. This requires:

1. **RBAC Permissions** - Already configured in `rbac-discovery.yaml`
   - ServiceAccount: `sharding-manager`
   - ClusterRole: `sharding-discovery`
   - ClusterRoleBinding: `sharding-discovery-binding`

2. **Permissions Include**:
   - List/get namespaces
   - List/get deployments and statefulsets
   - List/get pods
   - Read configmaps and secrets (for database discovery)

## Accessing Services

### LoadBalancer (Production)

If your cluster supports LoadBalancer services, the services will get external IPs:

```bash
# Get service IPs
kubectl get services -n sharding-system

# Access Manager API
curl http://<MANAGER_IP>:8081/api/v1/health

# Access Router API
curl http://<ROUTER_IP>:8080/v1/health
```

### Port Forwarding (Development/Testing)

```bash
# Manager API
kubectl port-forward svc/manager 8081:8081 -n sharding-system

# Router API
kubectl port-forward svc/router 8080:8080 -n sharding-system
```

### Ingress (Alternative)

You can create an Ingress resource to expose services via a domain name:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: sharding-ingress
  namespace: sharding-system
spec:
  rules:
  - host: sharding.example.com
    http:
      paths:
      - path: /manager
        pathType: Prefix
        backend:
          service:
            name: manager
            port:
              number: 8081
      - path: /router
        pathType: Prefix
        backend:
          service:
            name: router
            port:
              number: 8080
```

## Monitoring

### View Logs

```bash
# Manager logs
kubectl logs -f deployment/manager -n sharding-system

# Router logs
kubectl logs -f deployment/router -n sharding-system

# All pods
kubectl logs -f -l app=manager -n sharding-system
```

### Check Pod Status

```bash
kubectl get pods -n sharding-system
kubectl describe pod <pod-name> -n sharding-system
```

### Metrics

Metrics are exposed on ports 9091 (Manager) and 9090 (Router):

```bash
# Port forward metrics endpoint
kubectl port-forward svc/manager 9091:9091 -n sharding-system
curl http://localhost:9091/metrics
```

## Scaling

### Scale Manager

```bash
kubectl scale deployment manager --replicas=3 -n sharding-system
```

### Scale Router

```bash
kubectl scale deployment router --replicas=5 -n sharding-system
```

**Note**: Manager scaling requires leader election (not yet implemented). For now, keep manager replicas at 1-2.

## Troubleshooting

### Pods Not Starting

```bash
# Check pod events
kubectl describe pod <pod-name> -n sharding-system

# Check logs
kubectl logs <pod-name> -n sharding-system

# Common issues:
# - Image pull errors: Check image name and registry access
# - Config errors: Check ConfigMap and Secrets
# - RBAC errors: Check ServiceAccount and permissions
```

### Kubernetes Discovery Not Working

```bash
# Check ServiceAccount
kubectl get serviceaccount sharding-manager -n sharding-system

# Check ClusterRoleBinding
kubectl get clusterrolebinding sharding-discovery-binding

# Test permissions
kubectl auth can-i list deployments --as=system:serviceaccount:sharding-system:sharding-manager --all-namespaces
```

### Service Not Accessible

```bash
# Check service endpoints
kubectl get endpoints -n sharding-system

# Check service selector matches pod labels
kubectl get pods -l app=manager -n sharding-system
kubectl get service manager -n sharding-system -o yaml
```

## Cleanup

To remove all resources:

```bash
# Delete all resources
kubectl delete -f k8s/ -n sharding-system

# Or delete namespace (removes everything)
kubectl delete namespace sharding-system

# Remove RBAC (ClusterRole and ClusterRoleBinding)
kubectl delete clusterrole sharding-discovery
kubectl delete clusterrolebinding sharding-discovery-binding
```

## Production Considerations

1. **Secrets Management**: Use a secrets management system (e.g., Sealed Secrets, External Secrets Operator)
2. **Resource Limits**: Adjust CPU/memory limits based on your workload
3. **High Availability**: 
   - Deploy multiple router replicas
   - Use etcd cluster (not single instance)
   - Consider using StatefulSet for manager if persistence is needed
4. **Monitoring**: Set up Prometheus and Grafana for metrics
5. **Backup**: Regular backups of etcd data
6. **Security**: 
   - Enable RBAC in application config
   - Use TLS for all communications
   - Restrict network policies
7. **Image Registry**: Use a private registry for production images

## Support

For issues or questions:
- Check logs: `kubectl logs -f deployment/manager -n sharding-system`
- Review configuration: `kubectl get configmap sharding-config -n sharding-system -o yaml`
- Check RBAC: `kubectl describe clusterrolebinding sharding-discovery-binding`

