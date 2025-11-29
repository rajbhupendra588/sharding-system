# Fixes Applied for UI Issues

## Issues Fixed

### 1. Frontend API Configuration
- **Problem**: Frontend was trying to connect to `localhost:8081` and `localhost:8080` which don't work in Kubernetes
- **Fix**: Updated frontend to use relative URLs (`/api` and `/v1`) which nginx proxies to the backend services
- **Files Changed**:
  - `ui/src/core/config/constants.ts` - Changed default URLs to empty strings (relative)
  - `ui/src/core/config/app-config.ts` - Updated to handle relative URLs
  - Rebuilt frontend Docker image

### 2. etcd Service Discovery
- **Problem**: etcd pod wasn't becoming ready, so no endpoints were registered
- **Fix**: 
  - Changed readiness probe from HTTP to TCP socket check
  - Increased failure threshold to allow etcd more time to start
  - etcd is now ready and endpoints are registered

### 3. Manager/Router Connection to etcd
- **Problem**: Manager and Router couldn't resolve `etcd-0.etcd` DNS name
- **Fix**: Updated ConfigMap to use full service FQDN: `etcd.sharding-system.svc.cluster.local:2379`
- **Files Changed**:
  - `k8s/configmap.yaml` - Updated etcd endpoints

### 4. Nginx Proxy Configuration
- **Status**: Already configured correctly in `ui/nginx.conf`
- **Proxy Routes**:
  - `/api` → `http://manager:8081`
  - `/v1` → `http://router:8080`

## Current Status

✅ **Frontend**: 2/2 pods running
✅ **etcd**: 1/1 pod running and ready
✅ **Proxy**: 2/2 pods running
⏳ **Manager**: Restarting to pick up new etcd config
⏳ **Router**: Restarting to pick up new etcd config

## Access Points

- **Frontend UI**: http://localhost (port 80) or http://localhost:32030
- **Manager API**: http://localhost:8081
- **Router API**: http://localhost:8080

## Next Steps

1. Wait for Manager and Router pods to become ready (should happen automatically)
2. Once Manager is ready, the UI should be able to:
   - Show databases (from Manager API)
   - Show clusters (from Manager API)
   - Show client apps (from Manager API)
3. If issues persist, check:
   - `kubectl logs deployment/manager -n sharding-system`
   - `kubectl logs deployment/frontend -n sharding-system`
   - Browser console for API errors

## Verification Commands

```bash
# Check all pods
kubectl get pods -n sharding-system

# Check services
kubectl get svc -n sharding-system

# Check manager logs
kubectl logs -l app=manager -n sharding-system --tail=20

# Test frontend proxy
curl http://localhost/api/v1/health

# Check etcd endpoints
kubectl get endpoints -n sharding-system etcd
```




