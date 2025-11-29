# Finding Your Applications and Databases

## Understanding De-registration

When you **de-register** a client application from the Sharding Manager:
- ✅ The application is **removed from the Sharding Manager's registry only**
- ✅ Your **actual application continues running** in Kubernetes
- ✅ Your **database and data remain intact**
- ✅ The application can still access its database directly

**De-registration does NOT delete anything** - it only removes the tracking/management relationship.

## How to Find Your Applications

### Option 1: Using kubectl (Kubernetes CLI)

#### List All Deployments
```bash
kubectl get deployments --all-namespaces
```

#### List All Pods
```bash
kubectl get pods --all-namespaces
```

#### List All Services
```bash
kubectl get services --all-namespaces
```

#### Find Applications with Sharding Labels
```bash
# Find deployments with sharding labels
kubectl get deployments --all-namespaces -l sharding.enabled=true

# Find pods with sharding labels
kubectl get pods --all-namespaces -l sharding.enabled=true
```

### Option 2: Using Kubernetes Dashboard

If you have Kubernetes Dashboard installed:

```bash
# Start the dashboard proxy
kubectl proxy

# Access at: http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/
```

Navigate to:
1. **Workloads** → **Deployments** to see all applications
2. **Workloads** → **Pods** to see running instances
3. **Config and Storage** → **Config Maps** to see configuration

### Option 3: Find Databases

#### PostgreSQL Databases in Kubernetes
```bash
# List PostgreSQL pods
kubectl get pods --all-namespaces -l app=postgresql

# List PostgreSQL services
kubectl get services --all-namespaces -l app=postgresql

# Get PostgreSQL connection details
kubectl get secret <postgres-secret-name> -n <namespace> -o yaml
```

#### External Databases
If your databases are external to Kubernetes:
- Check your cloud provider's database service (RDS, Cloud SQL, etc.)
- Check your infrastructure documentation
- Look at ConfigMaps or Secrets that contain database connection strings

## Re-registering an Application

### Method 1: Through the UI

1. Go to **Client Applications** page
2. Click **"Register New Application"**
3. Fill in the application details:
   - **Name**: Your application name
   - **Description**: What the application does
   - **Database Name**: The database it uses
   - **Database Connection Details**: Host, port, credentials
   - **Key Prefix**: (Optional) For multi-tenant setups
   - **Namespace**: (Optional) Kubernetes namespace

### Method 2: Using the API

```bash
curl -X POST http://localhost:8081/api/v1/client-apps \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-application",
    "description": "My Application Description",
    "database_name": "myapp_db",
    "database_host": "postgres.default.svc.cluster.local",
    "database_port": "5432",
    "database_user": "myapp_user",
    "database_password": "secret",
    "key_prefix": "tenant_",
    "namespace": "default"
  }'
```

### Method 3: Automatic Discovery (if enabled)

If Kubernetes discovery is configured:
1. Click **"Discover Apps"** button
2. Review discovered applications
3. Click **"Register"** on the applications you want to manage

## Finding Application Configuration

### Check ConfigMaps
```bash
# List all ConfigMaps
kubectl get configmaps --all-namespaces

# View a specific ConfigMap
kubectl describe configmap <configmap-name> -n <namespace>
```

### Check Secrets
```bash
# List all Secrets
kubectl get secrets --all-namespaces

# View a specific Secret (base64 encoded)
kubectl get secret <secret-name> -n <namespace> -o yaml
```

### Check Environment Variables
```bash
# Get environment variables from a pod
kubectl exec <pod-name> -n <namespace> -- env | grep -i db
```

## Common Database Connection Patterns

### Pattern 1: Database in Same Kubernetes Cluster
```
Host: <service-name>.<namespace>.svc.cluster.local
Port: 5432
Database: myapp_db
```

### Pattern 2: External Database (Cloud)
```
Host: mydb.abc123.us-east-1.rds.amazonaws.com
Port: 5432
Database: production_db
```

### Pattern 3: Database with Connection String
```
postgresql://username:password@host:port/database
```

## Troubleshooting

### "I can't find my application"

1. **Check if it's still running:**
   ```bash
   kubectl get pods --all-namespaces | grep <app-name>
   ```

2. **Check recent deployments:**
   ```bash
   kubectl get deployments --all-namespaces --sort-by=.metadata.creationTimestamp
   ```

3. **Check logs for clues:**
   ```bash
   kubectl logs <pod-name> -n <namespace>
   ```

### "I don't know my database credentials"

1. **Check Secrets:**
   ```bash
   kubectl get secrets -n <namespace>
   kubectl get secret <db-secret> -n <namespace> -o jsonpath='{.data}'
   ```

2. **Decode base64 values:**
   ```bash
   echo "<base64-value>" | base64 -d
   ```

3. **Check application configuration:**
   ```bash
   kubectl describe deployment <app-name> -n <namespace>
   ```

### "My application was working before de-registration"

Don't worry! De-registration only removes it from the Sharding Manager's tracking. Your application and database are still running. You just need to re-register it with the correct details.

## Best Practices

1. **Document your applications** - Keep a record of:
   - Application names and namespaces
   - Database connection details
   - Shard keys and strategies

2. **Use labels** - Tag your Kubernetes resources:
   ```yaml
   metadata:
     labels:
       app: my-application
       sharding.enabled: "true"
       sharding.database: "myapp_db"
   ```

3. **Export before de-registering:**
   ```bash
   # Get application details before de-registering
   curl http://localhost:8081/api/v1/client-apps/<app-id> > app-backup.json
   ```

4. **Use Infrastructure as Code** - Store your application and database configurations in Git using:
   - Kubernetes manifests
   - Helm charts
   - Terraform/Pulumi scripts

## Quick Reference Commands

```bash
# Find all applications in namespace
kubectl get all -n <namespace>

# Find PostgreSQL databases
kubectl get pods -l app=postgresql --all-namespaces

# Get database connection from environment
kubectl exec <pod-name> -- env | grep DATABASE

# List all resources with labels
kubectl get all --all-namespaces --show-labels

# Export application configuration
kubectl get deployment <app-name> -n <namespace> -o yaml > app-config.yaml
```

## Need Help?

If you're still having trouble finding your applications:

1. Check the Kubernetes cluster where they were deployed
2. Contact your DevOps/Platform team
3. Review your deployment scripts or CI/CD pipelines
4. Check your infrastructure documentation
5. Look at recent kubectl/helm commands in your shell history

Remember: **De-registration is reversible** - you can always re-register your application once you locate it!
