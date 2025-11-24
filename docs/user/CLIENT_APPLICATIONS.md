# Client Applications Management

## Overview

The Sharding System now includes a **Client Applications** feature that allows you to track and manage which client applications are using the sharding system. This addresses the question: "For which client application has sharding been done?"

## Features

### 1. Client Application Registration

You can register client applications that use the sharding system. Each client application can be configured with:

- **Name**: Human-readable name (e.g., "E-commerce Service", "User Management API")
- **Description**: Optional description of the client application
- **Key Prefix**: Optional prefix pattern for shard keys (e.g., "app1:", "ecommerce:")

### 2. Usage Tracking

The system automatically tracks:

- **Shards Used**: Which shards each client application is using
- **Request Count**: Total number of requests made by the client
- **Last Seen**: When the client last made a request
- **Status**: Active or inactive

### 3. UI Dashboard

The UI now includes:

- **Client Applications Page**: View all registered client applications
- **Dashboard Widget**: See total client applications count
- **Client Details**: View shards used, request counts, and activity

## How It Works

### Shard Key Pattern Matching

When a client application makes requests with shard keys, the system identifies the client by matching the shard key prefix:

```
Client App 1: Uses shard keys like "app1:user:123", "app1:order:456"
Client App 2: Uses shard keys like "app2:product:789", "app2:cart:101"
```

If a client is registered with key prefix `"app1:"`, all shard keys starting with `"app1:"` will be associated with that client.

### Manual Registration

1. Navigate to **Client Apps** in the sidebar
2. Click **Register Client App**
3. Fill in:
   - Name: `"E-commerce Service"`
   - Description: `"Main e-commerce application"`
   - Key Prefix: `"ecommerce:"` (optional)
4. Click **Register**

### Automatic Discovery

The system can also auto-discover client applications by analyzing shard key patterns. Clients with significant usage (10+ requests) will be automatically registered.

## UI Usage

### Viewing Client Applications

1. Go to **Client Apps** from the sidebar
2. See all registered client applications in a card-based layout
3. Each card shows:
   - Client name and description
   - Status (active/inactive)
   - Number of shards used
   - Total request count
   - Last seen timestamp
   - List of shards (clickable links)

### Dashboard Integration

The Dashboard now shows:
- **Client Applications** stat card with total count
- Quick link to Client Apps page

### Filtering and Search

- Use the search box to filter client applications by name or description
- Real-time filtering as you type

## API Endpoints

### List Client Applications

```bash
GET /api/v1/client-apps
```

Response:
```json
[
  {
    "id": "uuid",
    "name": "E-commerce Service",
    "description": "Main e-commerce application",
    "status": "active",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "last_seen": "2024-01-01T12:00:00Z",
    "shard_ids": ["shard-1", "shard-2"],
    "request_count": 1234,
    "key_prefix": "ecommerce:"
  }
]
```

### Get Client Application

```bash
GET /api/v1/client-apps/{id}
```

### Register Client Application

```bash
POST /api/v1/client-apps
Content-Type: application/json

{
  "name": "E-commerce Service",
  "description": "Main e-commerce application",
  "key_prefix": "ecommerce:"
}
```

## Best Practices

### 1. Use Consistent Key Prefixes

Each client application should use a unique, consistent prefix for shard keys:

```go
// Good: Consistent prefix
shardKey := "ecommerce:user:123"
shardKey := "ecommerce:order:456"

// Bad: Inconsistent patterns
shardKey := "user:123"  // No prefix
shardKey := "order:456" // No prefix
```

### 2. Register Clients Early

Register client applications when they start using the sharding system to enable proper tracking.

### 3. Use Descriptive Names

Use clear, descriptive names for client applications:
- ✅ "E-commerce Service"
- ✅ "User Management API"
- ❌ "App1"
- ❌ "Service"

### 4. Monitor Usage

Regularly check the Client Apps page to:
- See which clients are actively using the system
- Identify clients with high request counts
- Monitor shard distribution across clients

## Example: Multiple Client Applications

### Scenario

You have 3 client applications in your Kubernetes cluster:

1. **E-commerce Service** - Uses keys like `"ecommerce:user:123"`
2. **Analytics Service** - Uses keys like `"analytics:event:456"`
3. **Notification Service** - Uses keys like `"notif:message:789"`

### Setup

1. Register each client application with its key prefix:
   ```bash
   # Register E-commerce Service
   POST /api/v1/client-apps
   {
     "name": "E-commerce Service",
     "key_prefix": "ecommerce:"
   }

   # Register Analytics Service
   POST /api/v1/client-apps
   {
     "name": "Analytics Service",
     "key_prefix": "analytics:"
   }

   # Register Notification Service
   POST /api/v1/client-apps
   {
     "name": "Notification Service",
     "key_prefix": "notif:"
   }
   ```

2. View in UI:
   - Navigate to **Client Apps**
   - See all 3 applications listed
   - Each shows which shards they're using
   - Track request counts and activity

### Result

The UI will show:
- **3 Client Applications** total
- Each client's shard usage
- Request patterns
- Activity status

## Troubleshooting

### Client Not Showing Up

**Problem**: Client application is not appearing in the list.

**Solutions**:
1. Check if the client is registered: `GET /api/v1/client-apps`
2. Verify key prefix matches shard key patterns
3. Ensure client has made requests (for auto-discovery)

### Shards Not Associated

**Problem**: Client shows 0 shards even though it's making requests.

**Solutions**:
1. Verify key prefix is correctly configured
2. Check that shard keys match the prefix pattern
3. Wait a few moments for tracking to update

### Request Count Not Updating

**Problem**: Request count remains at 0.

**Solutions**:
1. Ensure client is making requests through the router
2. Verify key prefix matches request patterns
3. Check router logs for tracking errors

## Future Enhancements

Planned features:
- Per-client metrics and dashboards
- Client-specific rate limiting
- Client resource quotas
- Automatic client discovery from router logs
- Client health monitoring

