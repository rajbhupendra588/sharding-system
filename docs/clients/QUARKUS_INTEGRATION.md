# Quarkus Integration Guide

This guide shows you how to integrate the Sharding System into your existing Java + Quarkus microservice application.

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Installation](#installation)
4. [Configuration](#configuration)
5. [Basic Integration](#basic-integration)
6. [Advanced Usage](#advanced-usage)
7. [Best Practices](#best-practices)
8. [Example Service](#example-service)

---

## Overview

The Sharding System provides transparent database sharding for your Quarkus application. Instead of connecting directly to a single database, your application connects to the Sharding Router, which automatically routes queries to the appropriate shard based on a shard key.

### Architecture

```
┌─────────────────────────────────┐
│   Your Quarkus Microservice      │
│                                  │
│  ┌──────────────────────────┐   │
│  │  ShardingClient          │   │
│  │  (Java Client Library)    │   │
│  └──────────┬───────────────┘   │
└─────────────┼───────────────────┘
              │ HTTP
              ▼
┌─────────────────────────────────┐
│   Sharding Router               │
│   (Data Plane)                  │
└──────────┬──────────────────────┘
           │
    ┌──────┴──────┬──────────┐
    ▼             ▼          ▼
┌─────────┐  ┌─────────┐  ┌─────────┐
│ Shard 1 │  │ Shard 2 │  │ Shard N │
│ (DB)    │  │ (DB)    │  │ (DB)    │
└─────────┘  └─────────┘  └─────────┘
```

---

## Prerequisites

- Java 11 or higher
- Quarkus 2.0 or higher
- Maven or Gradle
- Sharding System Router running (see [QUICKSTART.md](../QUICKSTART.md))

---

## Installation

### Step 1: Build the Java Client Library

First, build and install the Java client library to your local Maven repository:

```bash
cd clients/java
mvn clean install
```

### Step 2: Add Dependency to Your Quarkus Project

Add the dependency to your `pom.xml`:

```xml
<dependency>
    <groupId>com.sharding-system</groupId>
    <artifactId>sharding-client</artifactId>
    <version>1.0.0</version>
</dependency>
```

Or if using Gradle, add to `build.gradle`:

```gradle
implementation 'com.sharding-system:sharding-client:1.0.0'
```

---

## Configuration

### Application Properties

Add configuration to `src/main/resources/application.properties`:

```properties
# Sharding Router URL
sharding.router.url=http://localhost:8080

# Optional: Connection timeout (milliseconds)
sharding.router.timeout=30000

# Optional: Enable/disable sharding (useful for local development)
sharding.enabled=true
```

### Environment Variables

You can also configure via environment variables:

```bash
export SHARDING_ROUTER_URL=http://router:8080
export SHARDING_ENABLED=true
```

---

## Basic Integration

### Step 1: Create a Quarkus CDI Bean

Create a CDI bean that provides the `ShardingClient`:

```java
package com.example.service;

import com.sharding.system.client.ShardingClient;
import io.quarkus.arc.DefaultBean;
import jakarta.enterprise.context.ApplicationScoped;
import jakarta.enterprise.inject.Produces;
import org.eclipse.microprofile.config.inject.ConfigProperty;

@ApplicationScoped
public class ShardingClientProducer {
    
    @ConfigProperty(name = "sharding.router.url", defaultValue = "http://localhost:8080")
    String routerUrl;
    
    @Produces
    @ApplicationScoped
    @DefaultBean
    public ShardingClient shardingClient() {
        return new ShardingClient(routerUrl);
    }
}
```

### Step 2: Inject and Use in Your Service

```java
package com.example.service;

import com.sharding.system.client.ShardingClient;
import com.sharding.system.client.ShardingClientException;
import com.sharding.system.client.model.QueryResponse;
import jakarta.enterprise.context.ApplicationScoped;
import jakarta.inject.Inject;
import java.util.List;
import java.util.Map;

@ApplicationScoped
public class UserService {
    
    @Inject
    ShardingClient shardingClient;
    
    public User getUserById(String userId) throws ShardingClientException {
        // Execute query - router automatically routes to correct shard
        QueryResponse response = shardingClient.queryStrong(
            userId,  // shard key
            "SELECT id, name, email FROM users WHERE id = $1",
            userId   // query parameter
        );
        
        if (response.getRowCount() == 0) {
            return null;
        }
        
        // Convert result to domain object
        Map<String, Object> row = response.getRows().get(0);
        User user = new User();
        user.setId((String) row.get("id"));
        user.setName((String) row.get("name"));
        user.setEmail((String) row.get("email"));
        
        return user;
    }
    
    public void createUser(User user) throws ShardingClientException {
        shardingClient.queryStrong(
            user.getId(),  // shard key
            "INSERT INTO users (id, name, email) VALUES ($1, $2, $3)",
            user.getId(),
            user.getName(),
            user.getEmail()
        );
    }
    
    public List<User> getUsersByStatus(String status) throws ShardingClientException {
        // Note: This query will only search the shard for the first user
        // For cross-shard queries, you need to query each shard separately
        // or use a different approach (see Advanced Usage)
        
        QueryResponse response = shardingClient.queryEventual(
            "default",  // Use a default key if you need to query a specific shard
            "SELECT id, name, email FROM users WHERE status = $1",
            status
        );
        
        return response.getRows().stream()
            .map(row -> {
                User user = new User();
                user.setId((String) row.get("id"));
                user.setName((String) row.get("name"));
                user.setEmail((String) row.get("email"));
                return user;
            })
            .toList();
    }
}
```

### Step 3: Use in REST Endpoints

```java
package com.example.resource;

import com.example.service.UserService;
import com.sharding.system.client.ShardingClientException;
import jakarta.inject.Inject;
import jakarta.ws.rs.*;
import jakarta.ws.rs.core.MediaType;
import jakarta.ws.rs.core.Response;

@Path("/users")
@Produces(MediaType.APPLICATION_JSON)
@Consumes(MediaType.APPLICATION_JSON)
public class UserResource {
    
    @Inject
    UserService userService;
    
    @GET
    @Path("/{id}")
    public Response getUser(@PathParam("id") String id) {
        try {
            User user = userService.getUserById(id);
            if (user == null) {
                return Response.status(Response.Status.NOT_FOUND).build();
            }
            return Response.ok(user).build();
        } catch (ShardingClientException e) {
            return Response.status(Response.Status.INTERNAL_SERVER_ERROR)
                .entity(new ErrorResponse(e.getMessage()))
                .build();
        }
    }
    
    @POST
    public Response createUser(User user) {
        try {
            userService.createUser(user);
            return Response.status(Response.Status.CREATED).build();
        } catch (ShardingClientException e) {
            return Response.status(Response.Status.INTERNAL_SERVER_ERROR)
                .entity(new ErrorResponse(e.getMessage()))
                .build();
        }
    }
}
```

---

## Advanced Usage

### 1. Connection Pooling and Resource Management

For better resource management, use a `@PreDestroy` method to close the client:

```java
@ApplicationScoped
public class ShardingClientProducer {
    
    @Inject
    @ConfigProperty(name = "sharding.router.url")
    String routerUrl;
    
    private ShardingClient client;
    
    @Produces
    @ApplicationScoped
    public ShardingClient shardingClient() {
        if (client == null) {
            client = new ShardingClient(routerUrl);
        }
        return client;
    }
    
    @PreDestroy
    void cleanup() {
        if (client != null) {
            try {
                client.close();
            } catch (IOException e) {
                // Log error
            }
        }
    }
}
```

### 2. Retry Logic with Resilience4j

Add resilience4j for retry logic:

```xml
<dependency>
    <groupId>io.github.resilience4j</groupId>
    <artifactId>resilience4j-quarkus</artifactId>
</dependency>
```

```java
@ApplicationScoped
public class UserService {
    
    @Inject
    ShardingClient shardingClient;
    
    @Retry(name = "sharding-retry")
    @CircuitBreaker(name = "sharding-circuit-breaker")
    public User getUserById(String userId) throws ShardingClientException {
        return shardingClient.queryStrong(
            userId,
            "SELECT id, name, email FROM users WHERE id = $1",
            userId
        );
    }
}
```

Configure in `application.properties`:

```properties
resilience4j.retry.instances.sharding-retry.maxAttempts=3
resilience4j.retry.instances.sharding-retry.waitDuration=500ms

resilience4j.circuitbreaker.instances.sharding-circuit-breaker.failureRateThreshold=50
resilience4j.circuitbreaker.instances.sharding-circuit-breaker.waitDurationInOpenState=10s
```

### 3. Cross-Shard Queries

For queries that need to search across all shards, you can:

**Option A: Query Each Shard**

```java
public List<User> getAllUsers() throws ShardingClientException {
    // First, get list of all shards (requires Manager API access)
    // Then query each shard
    List<User> allUsers = new ArrayList<>();
    
    for (String shardId : getAllShardIds()) {
        QueryResponse response = shardingClient.queryEventual(
            shardId,
            "SELECT id, name, email FROM users"
        );
        
        // Convert and add to list
        response.getRows().forEach(row -> {
            User user = mapRowToUser(row);
            allUsers.add(user);
        });
    }
    
    return allUsers;
}
```

**Option B: Use Event Sourcing or CQRS**

For complex cross-shard queries, consider:
- Event sourcing to maintain a read model
- CQRS with a materialized view
- Search index (Elasticsearch, etc.)

### 4. Transaction Management

**Important**: The Sharding System does not support distributed transactions. For multi-shard operations:

**Option A: Saga Pattern**

```java
@ApplicationScoped
public class OrderService {
    
    @Inject
    ShardingClient shardingClient;
    
    public void createOrder(Order order) throws ShardingClientException {
        // Step 1: Create order in order shard
        shardingClient.queryStrong(
            order.getId(),
            "INSERT INTO orders (id, user_id, total) VALUES ($1, $2, $3)",
            order.getId(), order.getUserId(), order.getTotal()
        );
        
        // Step 2: Update user balance in user shard
        try {
            shardingClient.queryStrong(
                order.getUserId(),
                "UPDATE users SET balance = balance - $1 WHERE id = $2",
                order.getTotal(), order.getUserId()
            );
        } catch (ShardingClientException e) {
            // Compensate: Delete order
            shardingClient.queryStrong(
                order.getId(),
                "DELETE FROM orders WHERE id = $1",
                order.getId()
            );
            throw e;
        }
    }
}
```

**Option B: Event-Driven Architecture**

Use events for eventual consistency:

```java
public void createOrder(Order order) throws ShardingClientException {
    // Create order
    shardingClient.queryStrong(
        order.getId(),
        "INSERT INTO orders (id, user_id, total) VALUES ($1, $2, $3)",
        order.getId(), order.getUserId(), order.getTotal()
    );
    
    // Publish event
    eventPublisher.publish(new OrderCreatedEvent(order));
}

@ConsumeEvent("order-created")
public void handleOrderCreated(OrderCreatedEvent event) {
    // Update user balance asynchronously
    // If this fails, use compensating action
}
```

### 5. Monitoring and Observability

Add Micrometer metrics:

```java
@ApplicationScoped
public class UserService {
    
    @Inject
    ShardingClient shardingClient;
    
    @Inject
    MeterRegistry meterRegistry;
    
    public User getUserById(String userId) throws ShardingClientException {
        Timer.Sample sample = Timer.start(meterRegistry);
        
        try {
            QueryResponse response = shardingClient.queryStrong(
                userId,
                "SELECT id, name, email FROM users WHERE id = $1",
                userId
            );
            
            sample.stop(Timer.builder("sharding.query.duration")
                .tag("operation", "getUserById")
                .register(meterRegistry));
            
            meterRegistry.counter("sharding.query.count",
                "operation", "getUserById",
                "status", "success").increment();
            
            return mapResponseToUser(response);
        } catch (ShardingClientException e) {
            sample.stop(Timer.builder("sharding.query.duration")
                .tag("operation", "getUserById")
                .register(meterRegistry));
            
            meterRegistry.counter("sharding.query.count",
                "operation", "getUserById",
                "status", "error").increment();
            
            throw e;
        }
    }
}
```

---

## Best Practices

### 1. Shard Key Selection

- **Use high-cardinality keys**: User IDs, order IDs, etc.
- **Avoid hot keys**: Keys with disproportionate traffic
- **Co-locate related data**: Store user and their orders in the same shard (use user_id as shard key)

### 2. Query Patterns

- **Single-key queries**: Best performance (routed to one shard)
- **Multi-key queries**: Route to same shard if possible (use same shard key)
- **Cross-shard queries**: Avoid when possible, use alternatives (events, search index)

### 3. Consistency Levels

- **Strong consistency**: Use for writes and critical reads
- **Eventual consistency**: Use for non-critical reads (counts, aggregations)

### 4. Error Handling

Always handle `ShardingClientException`:

```java
try {
    QueryResponse response = shardingClient.queryStrong(...);
    // Process response
} catch (ShardingClientException e) {
    logger.error("Sharding query failed", e);
    // Retry, fallback, or return error
    throw new ServiceException("Failed to execute query", e);
}
```

### 5. Performance

- **Connection pooling**: The client uses HTTP connection pooling automatically
- **Batch operations**: When possible, batch multiple operations
- **Caching**: Cache frequently accessed data (use Quarkus Cache)

---

## Example Service

See the complete example in `examples/quarkus-service/` directory.

### Running the Example

1. Start the Sharding System:
```bash
docker-compose up -d
```

2. Create a shard (see [QUICKSTART.md](../QUICKSTART.md))

3. Run the Quarkus service:
```bash
cd examples/quarkus-service
./mvnw quarkus:dev
```

4. Test the API:
```bash
# Create a user
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"id": "user-123", "name": "John Doe", "email": "john@example.com"}'

# Get user
curl http://localhost:8080/users/user-123
```

---

## Migration from Direct Database Access

### Before (Direct Database)

```java
@ApplicationScoped
public class UserService {
    
    @Inject
    AgroalDataSource dataSource;
    
    public User getUserById(String userId) throws SQLException {
        try (Connection conn = dataSource.getConnection();
             PreparedStatement stmt = conn.prepareStatement(
                 "SELECT id, name, email FROM users WHERE id = ?")) {
            stmt.setString(1, userId);
            ResultSet rs = stmt.executeQuery();
            // ... map to User
        }
    }
}
```

### After (Sharding System)

```java
@ApplicationScoped
public class UserService {
    
    @Inject
    ShardingClient shardingClient;
    
    public User getUserById(String userId) throws ShardingClientException {
        QueryResponse response = shardingClient.queryStrong(
            userId,
            "SELECT id, name, email FROM users WHERE id = $1",
            userId
        );
        // ... map to User
    }
}
```

**Key Changes**:
1. Replace `AgroalDataSource` with `ShardingClient`
2. Change parameter placeholders from `?` to `$1`, `$2`, etc. (PostgreSQL style)
3. Handle `ShardingClientException` instead of `SQLException`
4. Use `QueryResponse` instead of `ResultSet`

---

## Troubleshooting

### Common Issues

**1. Connection Refused**
- Check that the Sharding Router is running
- Verify `sharding.router.url` configuration

**2. Query Fails**
- Check that shards are created and active
- Verify shard key is correct
- Check router logs

**3. Performance Issues**
- Monitor router metrics
- Check shard health
- Consider connection pooling configuration

---

## Additional Resources

- [API Documentation](./API.md)
- [System Design](./SYSTEM_DESIGN.md)
- [Quick Start Guide](../QUICKSTART.md)
- [Java Client README](../clients/java/README.md)

