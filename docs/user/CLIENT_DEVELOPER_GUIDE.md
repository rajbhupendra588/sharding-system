# Client Application Developer Guide

## Overview

This guide explains what you, as a client application developer, need to do to integrate your application with the Sharding System.

## Important: Shards Are NOT Auto-Created

**Key Point**: When you use the client library (Java or Go), shards are **NOT automatically created**. The library only:
- Fetches existing shard configuration from the Manager
- Routes your queries to existing shards
- Does NOT create or provision shards

**You must create shards separately** before your application can use them.

## Step-by-Step Integration Process

### Step 1: Ensure Sharding System is Running

Before integrating your application, make sure the Sharding System is running:

```bash
# Check if Manager is running
curl http://localhost:8081/health

# Check if Router is running
curl http://localhost:8080/health
```

**Manager** (Port 8081): Manages shards, creates shards, handles configuration
**Router** (Port 8080): Routes your queries to the correct shard

### Step 2: Register Your Client Application

Register your application with the Sharding System so it can track which shards you're using.

#### Option A: Via Web UI (Recommended for First Time)

1. Open the Web UI: `http://localhost:3000`
2. Navigate to **Client Apps** in the sidebar
3. Click **Register Client App**
4. Fill in:
   - **Name**: Your application name (e.g., "E-commerce Service")
   - **Description**: Brief description
   - **Key Prefix**: Optional prefix for your shard keys (e.g., "ecommerce:")
5. Click **Register**
6. **Save the Client App ID** - you'll need it later

#### Option B: Via API

```bash
curl -X POST http://localhost:8081/api/v1/client-apps \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Application",
    "description": "My application description",
    "key_prefix": "myapp:"
  }'
```

**Response:**
```json
{
  "id": "client-app-uuid-here",
  "name": "My Application",
  "status": "active",
  ...
}
```

**Save the `id` field** - this is your Client App ID.

### Step 3: Create Shards for Your Application

Shards must be created **before** your application can use them. You have three options:

#### Option A: Create Shards via Web UI

1. Go to **Shards** page in the Web UI
2. Click **Create Shard**
3. Fill in:
   - **Name**: Shard name (e.g., "myapp-shard-1")
   - **Client App ID**: The ID from Step 2
   - **Primary Endpoint**: Database connection string (e.g., `postgresql://localhost:5432/mydb`)
   - **Replicas**: Optional replica endpoints
   - **Database Details**: Host, port, database name, username, password
4. Click **Create**
5. Repeat for additional shards

#### Option B: Create Shards via API

```bash
curl -X POST http://localhost:8081/api/v1/shards \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "myapp-shard-1",
    "client_app_id": "YOUR_CLIENT_APP_ID",
    "primary_endpoint": "postgresql://localhost:5432/mydb",
    "host": "localhost",
    "port": 5432,
    "database": "mydb",
    "username": "postgres",
    "password": "password",
    "vnode_count": 256
  }'
```

#### Option C: Create Database with Auto-Sharding

The Database Service API can create a database and automatically create shards:

```bash
curl -X POST http://localhost:8081/api/v1/databases \
  -H "Content-Type: application/json" \
  -d '{
    "name": "myapp-db",
    "display_name": "My Application Database",
    "template": "starter",
    "shard_key": "id"
  }'
```

This will:
1. Create a client app automatically
2. Create shards asynchronously based on the template
3. Return a connection string you can use

### Step 4: Initialize Database Schemas on Shards

**Important**: You must create your database tables/schemas on **each shard** separately.

```bash
# For each shard database
psql -h localhost -U postgres -d shard1 < your-schema.sql
psql -h localhost -U postgres -d shard2 < your-schema.sql
# ... repeat for all shards
```

Or use your migration tool (Flyway, Liquibase, etc.) to run migrations on each shard.

### Step 5: Add Client Library to Your Application

#### For Java Applications

**Add dependency** (Maven):
```xml
<dependency>
    <groupId>com.sharding-system</groupId>
    <artifactId>sharding-client</artifactId>
    <version>1.0.0</version>
</dependency>
```

**Initialize the client:**
```java
import com.sharding.system.client.ShardingClient;

// Initialize client
ShardingClient client = new ShardingClient.Builder()
    .managerUrl("http://localhost:8081")
    .clientAppId("YOUR_CLIENT_APP_ID")  // From Step 2
    .build();

// Use the client
List<Map<String, Object>> results = client.query(
    "user-123",  // shard key
    "SELECT * FROM users WHERE id = ?",
    "user-123"
);
```

#### For Go Applications

**Add dependency:**
```bash
go get github.com/sharding-system/pkg/client
```

**Use the client:**
```go
import "github.com/sharding-system/pkg/client"

// Create client
client := client.NewClient("http://localhost:8080")

// Execute query
result, err := client.QueryStrong(
    "user-123",  // shard key
    "SELECT * FROM users WHERE id = $1",
    "user-123",
)
```

### Step 6: Use Shard Keys in Your Queries

When making queries, always provide a **shard key** that determines which shard to use:

```java
// Good: Using shard key
client.query("user-123", "SELECT * FROM users WHERE id = ?", "user-123");

// Good: Using consistent prefix
client.query("ecommerce:order-456", "SELECT * FROM orders WHERE id = ?", "order-456");
```

**Best Practices:**
- Use consistent shard keys (e.g., user ID, order ID)
- If you registered with a key prefix, use it consistently
- Choose shard keys with high cardinality (many unique values)
- Co-locate related data (e.g., user and their orders on same shard)

### Step 7: Verify Everything Works

1. **Check shards are created:**
   ```bash
   curl http://localhost:8081/api/v1/shards?client_app_id=YOUR_CLIENT_APP_ID
   ```

2. **Test a query:**
   ```bash
   curl -X POST http://localhost:8080/v1/execute \
     -H "Content-Type: application/json" \
     -d '{
       "shard_key": "user-123",
       "query": "SELECT * FROM users WHERE id = $1",
       "params": ["user-123"],
       "consistency": "strong"
     }'
   ```

3. **Check your client app in the UI:**
   - Go to **Client Apps** page
   - Find your application
   - Verify it shows the correct number of shards
   - Check that request count is increasing

## Complete Example Workflow

Here's a complete example for a new application:

```bash
# 1. Register client app
CLIENT_APP_ID=$(curl -X POST http://localhost:8081/api/v1/client-apps \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My E-commerce App",
    "key_prefix": "ecommerce:"
  }' | jq -r '.id')

echo "Client App ID: $CLIENT_APP_ID"

# 2. Create first shard
curl -X POST http://localhost:8081/api/v1/shards \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"ecommerce-shard-1\",
    \"client_app_id\": \"$CLIENT_APP_ID\",
    \"host\": \"localhost\",
    \"port\": 5432,
    \"database\": \"ecommerce_db\",
    \"username\": \"postgres\",
    \"password\": \"password\"
  }"

# 3. Create database schema on the shard
psql -h localhost -U postgres -d ecommerce_db < schema.sql

# 4. Now your application can use the sharding client
# The client will automatically fetch shard configuration
```

## Common Questions

### Q: Do I need to create shards manually every time?

**A**: Yes, shards must be created before use. However, you can:
- Use the Database Service API to auto-create multiple shards
- Use Kubernetes Operator to auto-provision shards
- Write scripts to automate shard creation

### Q: What if I don't register my client app?

**A**: The system can auto-discover your app after 10+ requests, but it's better to register explicitly for:
- Better tracking and monitoring
- Proper shard association
- Clear identification in the UI

### Q: Can I use the same shard for multiple applications?

**A**: Technically yes, but **not recommended**. Each client app should have its own shards for:
- Better isolation
- Independent scaling
- Clear ownership and tracking

### Q: How do I add more shards later?

**A**: Simply create new shards via API or UI and associate them with your client app. The client library will automatically pick them up on the next refresh.

### Q: What happens if a shard fails?

**A**: The Manager monitors shard health. If a primary fails:
- Replicas can be promoted to primary
- The router automatically uses the new primary
- Your application continues working (with possible brief interruption)

## Troubleshooting

### Problem: "No shards found for client app"

**Solution:**
1. Verify shards are created: `GET /api/v1/shards?client_app_id=YOUR_ID`
2. Check shards are associated with your client app
3. Wait a few seconds for the client library to refresh

### Problem: "Client application not found"

**Solution:**
1. Verify you registered the client app: `GET /api/v1/client-apps`
2. Check you're using the correct Client App ID
3. Ensure the client app status is "active"

### Problem: "Query fails with connection error"

**Solution:**
1. Verify database is running and accessible
2. Check database credentials are correct
3. Ensure schema is created on the shard
4. Test database connection directly: `psql -h HOST -U USER -d DATABASE`

### Problem: "Shard map refresh fails"

**Solution:**
1. Verify Manager is running: `curl http://localhost:8081/health`
2. Check Manager URL is correct in client configuration
3. Verify network connectivity between client and Manager

## Next Steps

- [API Reference](../api/API_REFERENCE.md) - Full API documentation
- [Architecture Guide](../architecture/ARCHITECTURE.md) - Understand how it works
- [Java Client Examples](../../examples/java-ecommerce-service/) - Working examples
- [Go Client Examples](../../examples/client_example.go) - Go examples

## Summary Checklist

- [ ] Sharding System is running (Manager + Router)
- [ ] Client application is registered
- [ ] Shards are created and associated with your client app
- [ ] Database schemas are created on all shards
- [ ] Client library is added to your application
- [ ] Client is initialized with correct Manager URL and Client App ID
- [ ] Queries are using shard keys correctly
- [ ] Application is tested and working

