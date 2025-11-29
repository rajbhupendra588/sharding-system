# Cluster Registration and Scanning Guide

## Problem
The scanning feature in the UI is disabled because you need to:
1. **First register a cluster**
2. **Then select it** (check the checkbox)
3. **Then you can scan**

## Solution: Manual Cluster Registration via API

### Option 1: Register a Local/Minikube Cluster (Simplest)

If you're running a local Kubernetes cluster (like minikube, kind, or Docker Desktop K8s), you can register it with minimal configuration:

```bash
curl -X POST http://localhost:8081/api/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "name": "local-cluster",
    "type": "onprem",
    "provider": "local",
    "kubeconfig": "",
    "context": "",
    "endpoint": ""
  }'
```

**Response:**
```json
{
  "id": "wXEzVZkw1sqcIKh63ZvcxQ==",
  "name": "local-cluster",
  "type": "onprem",
  "status": "active",
  "created_at": "2025-11-27T...",
  "updated_at": "2025-11-27T..."
}
```

### Option 2: Register with Kubeconfig Path

If you have a kubeconfig file:

```bash
curl -X POST http://localhost:8081/api/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-cluster",
    "type": "onprem",
    "kubeconfig": "/path/to/your/kubeconfig",
    "context": "your-context-name"
  }'
```

### Option 3: Register with Kubeconfig Content (Base64)

```bash
# First, encode your kubeconfig
KUBECONFIG_B64=$(cat ~/.kube/config | base64)

curl -X POST http://localhost:8081/api/v1/clusters \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"my-cluster\",
    \"type\": \"onprem\",
    \"kubeconfig\": \"$KUBECONFIG_B64\",
    \"context\": \"your-context-name\"
  }"
```

### Option 4: Register Cloud Cluster (AWS EKS, GCP GKE, Azure AKS)

```bash
curl -X POST http://localhost:8081/api/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-eks-cluster",
    "type": "cloud",
    "provider": "aws",
    "endpoint": "https://your-cluster-endpoint.eks.amazonaws.com",
    "credentials": {
      "access_key_id": "your-access-key",
      "secret_access_key": "your-secret-key",
      "region": "us-east-1"
    }
  }'
```

## After Registration

### 1. Verify Cluster is Registered

```bash
curl http://localhost:8081/api/v1/clusters | python3 -m json.tool
```

### 2. Use the UI to Scan

1. Go to the **Multi-Cluster Database Scanner** page in the UI
2. You should now see your registered cluster in the table
3. **Check the checkbox** next to the cluster(s) you want to scan
4. Click **"Start Scan"** button (it will now be enabled)

### 3. Or Trigger Scan via API

```bash
# Scan all clusters
curl -X POST http://localhost:8081/api/v1/clusters/scan \
  -H "Content-Type: application/json" \
  -d '{
    "deep_scan": false
  }'

# Scan specific cluster(s)
curl -X POST http://localhost:8081/api/v1/clusters/scan \
  -H "Content-Type: application/json" \
  -d '{
    "cluster_ids": ["wXEzVZkw1sqcIKh63ZvcxQ=="],
    "deep_scan": false
  }'
```

### 4. Check Scan Results

```bash
# Get all scan results
curl http://localhost:8081/api/v1/clusters/scan/results | python3 -m json.tool

# Get results for specific cluster
curl "http://localhost:8081/api/v1/clusters/scan/results?cluster_id=wXEzVZkw1sqcIKh63ZvcxQ==" | python3 -m json.tool
```

## Quick Start Example

Here's a complete example for a local cluster:

```bash
# 1. Register the cluster
CLUSTER_RESPONSE=$(curl -s -X POST http://localhost:8081/api/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "name": "local-cluster",
    "type": "onprem"
  }')

echo "Cluster registered: $CLUSTER_RESPONSE"

# 2. Extract cluster ID (if needed)
CLUSTER_ID=$(echo $CLUSTER_RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin)['id'])")

# 3. Trigger scan
curl -X POST http://localhost:8081/api/v1/clusters/scan \
  -H "Content-Type: application/json" \
  -d "{
    \"cluster_ids\": [\"$CLUSTER_ID\"],
    \"deep_scan\": false
  }"

# 4. Wait a few seconds, then check results
sleep 5
curl http://localhost:8081/api/v1/clusters/scan/results | python3 -m json.tool
```

## Troubleshooting

### If scanning still doesn't work in UI:

1. **Refresh the page** - The cluster list might need to refresh
2. **Check browser console** - Look for any JavaScript errors
3. **Verify cluster is listed**: `curl http://localhost:8081/api/v1/clusters`
4. **Check manager logs**: `tail -f logs/manager.log`

### If you get "cluster not found" errors:

- Make sure the cluster ID is correct
- Verify the cluster is registered: `curl http://localhost:8081/api/v1/clusters`

### If scan returns no databases:

- Make sure your Kubernetes cluster has databases running
- Check that the cluster connection is working
- Try a "deep scan" by setting `"deep_scan": true` in the scan request

## Notes

- **Scan results are stored in memory** - They will be lost when the manager restarts
- **To persist databases**, you need to create them manually after scanning
- The discovered databases will appear in the **Databases** page after scanning



