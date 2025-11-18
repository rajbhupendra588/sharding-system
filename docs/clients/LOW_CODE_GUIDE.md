# Low-Code Java Client Guide

## Overview

The Sharding System Java Client now supports a **90-99% low-code approach**, dramatically reducing boilerplate code while maintaining full functionality. This guide shows you how to use the low-code features.

## Important: Naming Flexibility

**The framework does NOT enforce any naming conventions.** You can use **any names** you want for entities, repositories, services, and fields. The examples in this guide use common naming patterns (like `UserEntity`, `UserRepository`), but you can use whatever fits your project's conventions.

See [Naming Conventions Guide](./NAMING_CONVENTIONS.md) for examples with different naming styles.

## Before vs After Comparison

### BEFORE (High Code - ~100 lines)

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
        
        if (response.getRowCount() == 0) {
            return null;
        }
        
        return mapRowToUser(response.getRows().get(0));
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
    
    public void updateUser(User user) throws ShardingClientException {
        shardingClient.queryStrong(
            user.getId(),
            "UPDATE users SET name = $1, email = $2 WHERE id = $3",
            user.getName(),
            user.getEmail(),
            user.getId()
        );
    }
    
    public void deleteUser(String userId) throws ShardingClientException {
        shardingClient.queryStrong(
            userId,
            "DELETE FROM users WHERE id = $1",
            userId
        );
    }
    
    public User findByEmail(String email) throws ShardingClientException {
        QueryResponse response = shardingClient.queryStrong(
            email, // Using email as shard key
            "SELECT id, name, email FROM users WHERE email = $1",
            email
        );
        
        if (response.getRowCount() == 0) {
            return null;
        }
        
        return mapRowToUser(response.getRows().get(0));
    }
    
    private User mapRowToUser(Map<String, Object> row) {
        User user = new User();
        user.setId((String) row.get("id"));
        user.setName((String) row.get("name"));
        user.setEmail((String) row.get("email"));
        return user;
    }
}
```

### AFTER (Low Code - ~10 lines)

```java
// 1. Define Entity (with annotations)
@Entity(table = "users")
public class UserEntity {
    @ShardKey
    private String id;
    private String name;
    private String email;
    // getters/setters
}

// 2. Define Repository Interface (ZERO implementation!)
@ShardingRepository(entity = UserEntity.class)
public interface UserRepository extends CrudRepository<UserEntity, String> {
    Optional<UserEntity> findByEmail(String email);
}

// 3. Use in Service (ONE LINE per operation!)
@ApplicationScoped
public class UserService {
    @Inject
    UserRepository userRepository; // Auto-configured!
    
    public Optional<UserEntity> getUserById(String id) {
        return userRepository.findById(id); // ONE LINE!
    }
    
    public UserEntity createUser(UserEntity user) {
        return userRepository.save(user); // ONE LINE!
    }
    
    public void deleteUser(String id) {
        userRepository.deleteById(id); // ONE LINE!
    }
}
```

**Code Reduction: ~90%** 🎉

## Quick Start

### Step 1: Define Your Entity

**You can name your entity class anything!** Here's an example:

```java
package com.example.model;

import com.sharding.system.client.annotation.*;

@Entity(table = "users")
public class UserEntity {  // ← Can be named anything: User, Customer, Account, etc.
    @ShardKey
    @Column(name = "id")
    private String id;  // ← Field name can be anything: userId, accountId, etc.
    
    @Column(name = "name")
    private String name;
    
    @Column(name = "email")
    private String email;
    
    // Constructors, getters, setters...
}
```

**Annotations:**
- `@Entity(table = "users")` - Marks as entity, specifies table name (can be any table name)
- `@ShardKey` - Marks the shard key field (can be any field name)
- `@Column(name = "name")` - Maps field to column (optional, auto-converts camelCase)

**Note:** The class name `UserEntity` is just an example. You can use `User`, `Customer`, `Account`, `MyDomainModel`, or any name that fits your project's conventions. See [Naming Conventions](./NAMING_CONVENTIONS.md) for more examples.

### Step 2: Create Repository Interface

**You can name your repository interface anything!** Here's an example:

```java
package com.example.repository;

import com.example.model.UserEntity;  // ← Your entity class (any name)
import com.sharding.system.client.annotation.*;
import com.sharding.system.client.repository.CrudRepository;
import java.util.Optional;
import java.util.List;

@ShardingRepository(entity = UserEntity.class, table = "users")
public interface UserRepository extends CrudRepository<UserEntity, String> {
    // ↑ Can be named anything: UserRepo, UserDAO, IUserService, etc.
    
    // Auto-generated query: SELECT * FROM users WHERE email = $1
    Optional<UserEntity> findByEmail(String email);
    // ↑ Method name can be anything: getByEmail, lookupByEmail, etc.
    
    // Custom query with automatic mapping
    @Query("SELECT * FROM users WHERE name LIKE $1 ORDER BY name LIMIT $2")
    List<UserEntity> findByNameLike(String pattern, int limit);
    
    // Eventual consistency for read-heavy operations
    @EventualConsistency
    @Query("SELECT * FROM users WHERE status = $1")
    List<UserEntity> findByStatus(String status);
}
```

**Note:** Repository interface names can be anything (`UserRepository`, `UserRepo`, `UserDAO`, etc.). Method names can also be anything (except CRUD methods from `CrudRepository`). See [Naming Conventions](./NAMING_CONVENTIONS.md) for more examples.

**Available CRUD Methods (from CrudRepository):**
- `save(T entity)` - Insert or update
- `findById(@ShardKey ID id)` - Find by ID
- `existsById(@ShardKey ID id)` - Check existence
- `findAll(@ShardKey String shardKey)` - Find all in shard
- `count(@ShardKey String shardKey)` - Count entities
- `deleteById(@ShardKey ID id)` - Delete by ID
- `delete(T entity)` - Delete entity
- `deleteAll(@ShardKey String shardKey)` - Delete all in shard

### Step 3: Configure and Use

**For Quarkus:**

```java
@ApplicationScoped
public class UserService {
    private ShardingClientAutoConfiguration config;
    private UserRepository userRepository;
    
    @PostConstruct
    public void init() {
        config = new ShardingClientAutoConfiguration();
        config.setRouterUrl("http://localhost:8080");
        config.initialize();
        userRepository = config.getRepository(UserRepository.class);
    }
    
    public Optional<UserEntity> getUser(String id) {
        return userRepository.findById(id); // That's it!
    }
}
```

**For Standard Java:**

```java
public class UserService {
    private UserRepository userRepository;
    
    public UserService() {
        ShardingClient client = new ShardingClient("http://localhost:8080");
        ShardingClientAutoConfiguration config = new ShardingClientAutoConfiguration();
        config.setRouterUrl("http://localhost:8080");
        config.initialize();
        userRepository = config.getRepository(UserRepository.class);
    }
    
    public Optional<UserEntity> getUser(String id) {
        return userRepository.findById(id);
    }
}
```

## Query Method Name Patterns

The framework automatically generates SQL from method names:

| Method Name Pattern | Generated SQL |
|---------------------|---------------|
| `findById(String id)` | `SELECT * FROM table WHERE id = $1` |
| `findByEmail(String email)` | `SELECT * FROM table WHERE email = $1` |
| `findByEmailAndStatus(String email, String status)` | `SELECT * FROM table WHERE email = $1 AND status = $2` |
| `save(UserEntity user)` | `INSERT INTO table (...) VALUES (...)` |
| `deleteById(String id)` | `DELETE FROM table WHERE id = $1` |
| `existsById(String id)` | `SELECT COUNT(*) > 0 FROM table WHERE id = $1` |
| `count(String shardKey)` | `SELECT COUNT(*) FROM table` |

## Custom Queries

Use `@Query` annotation for custom SQL:

```java
@Query("SELECT * FROM users WHERE age > $1 AND status = $2 ORDER BY created_at DESC LIMIT $3")
List<UserEntity> findActiveUsersOlderThan(int minAge, String status, int limit);
```

## Consistency Levels

Control read consistency:

```java
// Strong consistency (default for writes, reads from primary)
@StrongConsistency
Optional<UserEntity> findById(String id);

// Eventual consistency (reads from replica, faster)
@EventualConsistency
List<UserEntity> findAll(String shardKey);
```

## Shard Key Extraction

The framework automatically extracts shard keys:

1. **From @ShardKey annotated parameter:**
   ```java
   Optional<User> findById(@ShardKey String id);
   ```

2. **From entity field:**
   ```java
   void save(UserEntity user); // Extracts from user.getId()
   ```

3. **From @ShardKey annotated field:**
   ```java
   @Entity
   public class UserEntity {
       @ShardKey
       private String id; // Automatically used as shard key
   }
   ```

## Return Types

Supported return types with automatic mapping:

- `T` - Single entity
- `Optional<T>` - Optional entity
- `List<T>` - List of entities
- `boolean` - For `existsById()`
- `long` - For `count()`
- `void` - For delete operations

## Complete Example

See `examples/quarkus-service/` for a complete working example:

- `UserEntity.java` - Entity definition (example name - you can use any name)
- `UserRepository.java` - Repository interface (example name - you can use any name)
- `LowCodeUserService.java` - Service using repository (example name - you can use any name)
- `LowCodeUserResource.java` - REST endpoint (example name - you can use any name)

**Remember:** All names in the examples are just examples. Use whatever naming conventions fit your project!

## Benefits

✅ **90-99% Less Code** - Write interfaces, not implementations  
✅ **Zero Boilerplate** - No manual SQL, no manual mapping  
✅ **Type Safety** - Compile-time checking  
✅ **Auto-Generated Queries** - From method names  
✅ **Automatic Mapping** - QueryResponse → Entity  
✅ **Shard Key Detection** - Automatic extraction  
✅ **Consistency Control** - Annotations for strong/eventual  
✅ **Custom Queries** - When you need them  

## Migration Guide

To migrate from high-code to low-code:

1. **Add annotations to your entity:**
   ```java
   @Entity(table = "users")
   public class User {
       @ShardKey
       private String id;
       // ...
   }
   ```

2. **Create repository interface:**
   ```java
   @ShardingRepository(entity = User.class)
   public interface UserRepository extends CrudRepository<User, String> {
       // Add your custom methods
   }
   ```

3. **Replace service code:**
   ```java
   // OLD
   QueryResponse response = shardingClient.queryStrong(...);
   User user = mapRowToUser(response.getRows().get(0));
   
   // NEW
   Optional<User> user = userRepository.findById(id);
   ```

4. **Remove manual mapping code** - No longer needed!

## Advanced Features

### Batch Operations

```java
List<UserEntity> users = Arrays.asList(user1, user2, user3);
userRepository.saveAll(users); // Batch insert
```

### Complex Queries

```java
@Query("SELECT u.* FROM users u JOIN orders o ON u.id = o.user_id WHERE o.total > $1")
List<UserEntity> findUsersWithLargeOrders(double minTotal);
```

### Transaction Support

For transactions, use the underlying `ShardingClient`:

```java
// Future: Transaction support will be added
```

## Troubleshooting

**Issue:** "No shard key field found"  
**Solution:** Add `@ShardKey` annotation to your entity's ID field

**Issue:** "Cannot determine shard key for method"  
**Solution:** Add `@ShardKey` annotation to method parameter

**Issue:** Mapping fails  
**Solution:** Ensure column names match (use `@Column` annotation if needed)

## Next Steps

- See [Naming Conventions](./NAMING_CONVENTIONS.md) - Learn about naming flexibility
- See [Java Client Reference](./java.md) for detailed API documentation
- Check [Quarkus Integration](./QUARKUS_INTEGRATION.md) for framework-specific setup
- Explore [Examples](../examples/) for more use cases

