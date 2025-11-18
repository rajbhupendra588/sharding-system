# Java/Quarkus Quick Start Guide

Quick guide to integrate the Sharding System into your Java + Quarkus microservice.

## Step 1: Build the Java Client

```bash
cd clients/java
mvn clean install
```

## Step 2: Add Dependency to Your Project

**Maven (`pom.xml`):**
```xml
<dependency>
    <groupId>com.sharding-system</groupId>
    <artifactId>sharding-client</artifactId>
    <version>1.0.0</version>
</dependency>
```

**Gradle (`build.gradle`):**
```gradle
implementation 'com.sharding-system:sharding-client:1.0.0'
```

## Step 3: Configure Application

Add to `src/main/resources/application.properties`:

```properties
sharding.router.url=http://localhost:8080
```

## Step 4: Create CDI Producer

```java
package com.example.service;

import com.sharding.system.client.ShardingClient;
import jakarta.enterprise.context.ApplicationScoped;
import jakarta.enterprise.inject.Produces;
import org.eclipse.microprofile.config.inject.ConfigProperty;

@ApplicationScoped
public class ShardingClientProducer {
    
    @ConfigProperty(name = "sharding.router.url", defaultValue = "http://localhost:8080")
    String routerUrl;
    
    @Produces
    @ApplicationScoped
    public ShardingClient shardingClient() {
        return new ShardingClient(routerUrl);
    }
}
```

## Step 5: Use in Your Service

```java
package com.example.service;

import com.sharding.system.client.ShardingClient;
import com.sharding.system.client.ShardingClientException;
import com.sharding.system.client.model.QueryResponse;
import jakarta.enterprise.context.ApplicationScoped;
import jakarta.inject.Inject;

@ApplicationScoped
public class UserService {
    
    @Inject
    ShardingClient shardingClient;
    
    public User getUserById(String userId) throws ShardingClientException {
        QueryResponse response = shardingClient.queryStrong(
            userId,  // shard key
            "SELECT id, name, email FROM users WHERE id = $1",
            userId   // query parameter
        );
        
        if (response.getRowCount() == 0) {
            return null;
        }
        
        // Map response to your domain object
        Map<String, Object> row = response.getRows().get(0);
        User user = new User();
        user.setId((String) row.get("id"));
        user.setName((String) row.get("name"));
        user.setEmail((String) row.get("email"));
        return user;
    }
    
    public void createUser(User user) throws ShardingClientException {
        shardingClient.queryStrong(
            user.getId(),
            "INSERT INTO users (id, name, email) VALUES ($1, $2, $3)",
            user.getId(),
            user.getName(),
            user.getEmail()
        );
    }
}
```

## Step 6: Use in REST Endpoint

```java
package com.example.resource;

import com.example.service.UserService;
import com.sharding.system.client.ShardingClientException;
import jakarta.inject.Inject;
import jakarta.ws.rs.*;
import jakarta.ws.rs.core.Response;

@Path("/users")
public class UserResource {
    
    @Inject
    UserService userService;
    
    @GET
    @Path("/{id}")
    public Response getUser(@PathParam("id") String id) {
        try {
            User user = userService.getUserById(id);
            return Response.ok(user).build();
        } catch (ShardingClientException e) {
            return Response.status(500)
                .entity("Error: " + e.getMessage())
                .build();
        }
    }
}
```

## Key Points

1. **Shard Key**: Always provide a shard key (e.g., user ID) for routing
2. **Query Parameters**: Use PostgreSQL-style `$1`, `$2`, etc.
3. **Consistency**: 
   - `queryStrong()` - Reads from primary (strong consistency)
   - `queryEventual()` - Can read from replica (eventual consistency)
4. **Error Handling**: Always catch `ShardingClientException`

## Complete Example

See `examples/quarkus-service/` for a complete working example.

## More Information

- [Full Integration Guide](./QUARKUS_INTEGRATION.md)
- [Java Client README](../clients/java/README.md)
- [API Documentation](./API.md)

