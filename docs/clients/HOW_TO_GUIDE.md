# How To: Java Low-Code Client - Complete User Guide

A comprehensive, step-by-step guide covering every use case for the Java Low-Code Client.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Basic CRUD Operations](#basic-crud-operations)
3. [Custom Queries](#custom-queries)
4. [Query Method Patterns](#query-method-patterns)
5. [Shard Key Handling](#shard-key-handling)
6. [Consistency Levels](#consistency-levels)
7. [Return Types](#return-types)
8. [Batch Operations](#batch-operations)
9. [Complex Queries](#complex-queries)
10. [Error Handling](#error-handling)
11. [Integration Patterns](#integration-patterns)
12. [Best Practices](#best-practices)
13. [Common Patterns](#common-patterns)
14. [Troubleshooting](#troubleshooting)

---

## Getting Started

### How To: Set Up Your First Entity

**Step 1:** Create your entity class with annotations:

```java
package com.example.model;

import com.sharding.system.client.annotation.*;

@Entity(table = "users")  // Table name in database
public class User {
    
    @ShardKey  // This field determines which shard the data goes to
    @Column(name = "id")
    private String id;
    
    @Column(name = "name")
    private String name;
    
    @Column(name = "email")
    private String email;
    
    @Column(name = "created_at")
    private java.time.LocalDateTime createdAt;
    
    // Constructors
    public User() {}
    
    public User(String id, String name, String email) {
        this.id = id;
        this.name = name;
        this.email = email;
        this.createdAt = java.time.LocalDateTime.now();
    }
    
    // Getters and Setters
    public String getId() { return id; }
    public void setId(String id) { this.id = id; }
    
    public String getName() { return name; }
    public void setName(String name) { this.name = name; }
    
    public String getEmail() { return email; }
    public void setEmail(String email) { this.email = email; }
    
    public java.time.LocalDateTime getCreatedAt() { return createdAt; }
    public void setCreatedAt(java.time.LocalDateTime createdAt) { this.createdAt = createdAt; }
}
```

**Key Points:**
- `@Entity(table = "users")` - Maps class to database table
- `@ShardKey` - Marks the field used for shard routing
- `@Column(name = "...")` - Optional, maps field to column (auto-converts camelCase if omitted)
- Field names can be anything - framework handles mapping

### How To: Create Your First Repository

**Step 1:** Create a repository interface:

```java
package com.example.repository;

import com.example.model.User;
import com.sharding.system.client.annotation.ShardingRepository;
import com.sharding.system.client.repository.CrudRepository;
import java.util.Optional;

@ShardingRepository(entity = User.class, table = "users")
public interface UserRepository extends CrudRepository<User, String> {
    // Extend CrudRepository to get: save, findById, deleteById, etc.
}
```

**Step 2:** Initialize and use:

```java
import com.sharding.system.client.config.ShardingClientAutoConfiguration;

public class UserService {
    private UserRepository userRepository;
    
    public void init() {
        ShardingClientAutoConfiguration config = new ShardingClientAutoConfiguration();
        config.setRouterUrl("http://localhost:8080");
        config.initialize();
        userRepository = config.getRepository(UserRepository.class);
    }
    
    public Optional<User> getUser(String id) {
        return userRepository.findById(id);  // That's it!
    }
}
```

---

## Basic CRUD Operations

### How To: Create (Insert) a Record

**Method 1: Using `save()`**

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    // save() is inherited from CrudRepository
}

// Usage:
User newUser = new User("user-123", "John Doe", "john@example.com");
User saved = userRepository.save(newUser);  // Automatically generates INSERT
```

**What happens:**
- Framework extracts shard key from `newUser.getId()`
- Generates: `INSERT INTO users (id, name, email, created_at) VALUES ($1, $2, $3, $4)`
- Executes query with strong consistency
- Returns the saved entity

**Method 2: Custom Insert Query**

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    @Query("INSERT INTO users (id, name, email) VALUES ($1, $2, $3) RETURNING *")
    User insertUser(String id, String name, String email);
}

// Usage:
User user = userRepository.insertUser("user-123", "John Doe", "john@example.com");
```

### How To: Read (Select) Records

**Method 1: Find by ID**

```java
// Repository (inherited from CrudRepository)
Optional<User> findById(@ShardKey String id);

// Usage:
Optional<User> user = userRepository.findById("user-123");
if (user.isPresent()) {
    System.out.println("Found: " + user.get().getName());
}
```

**Method 2: Find All in a Shard**

```java
// Repository (inherited from CrudRepository)
List<User> findAll(@ShardKey String shardKey);

// Usage:
List<User> allUsers = userRepository.findAll("user-123");  // Uses shard key for routing
```

**Method 3: Check Existence**

```java
// Repository (inherited from CrudRepository)
boolean existsById(@ShardKey String id);

// Usage:
if (userRepository.existsById("user-123")) {
    System.out.println("User exists");
}
```

**Method 4: Count Records**

```java
// Repository (inherited from CrudRepository)
long count(@ShardKey String shardKey);

// Usage:
long userCount = userRepository.count("user-123");
System.out.println("Total users: " + userCount);
```

### How To: Update a Record

**Method 1: Using `save()` (Upsert)**

```java
// Repository (inherited from CrudRepository)
User save(User entity);

// Usage:
User user = userRepository.findById("user-123").orElseThrow();
user.setEmail("newemail@example.com");
User updated = userRepository.save(user);  // Automatically generates UPDATE
```

**What happens:**
- Framework checks if entity exists (by ID)
- Generates: `UPDATE users SET name = $1, email = $2, created_at = $3 WHERE id = $4`
- Executes with strong consistency
- Returns updated entity

**Method 2: Custom Update Query**

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    @Query("UPDATE users SET email = $2 WHERE id = $1")
    void updateEmail(@ShardKey String id, String newEmail);
}

// Usage:
userRepository.updateEmail("user-123", "newemail@example.com");
```

**Method 3: Partial Update**

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    @Query("UPDATE users SET name = $2 WHERE id = $1 RETURNING *")
    User updateName(@ShardKey String id, String newName);
}

// Usage:
User updated = userRepository.updateName("user-123", "Jane Doe");
```

### How To: Delete Records

**Method 1: Delete by ID**

```java
// Repository (inherited from CrudRepository)
void deleteById(@ShardKey String id);

// Usage:
userRepository.deleteById("user-123");
```

**Method 2: Delete Entity**

```java
// Repository (inherited from CrudRepository)
void delete(User entity);

// Usage:
User user = userRepository.findById("user-123").orElseThrow();
userRepository.delete(user);
```

**Method 3: Delete All in Shard**

```java
// Repository (inherited from CrudRepository)
void deleteAll(@ShardKey String shardKey);

// Usage:
userRepository.deleteAll("user-123");  // Deletes all users in that shard
```

**Method 4: Custom Delete Query**

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    @Query("DELETE FROM users WHERE email = $1")
    void deleteByEmail(String email);
}

// Usage:
userRepository.deleteByEmail("john@example.com");
```

---

## Custom Queries

### How To: Write Custom SQL Queries

**Basic Custom Query:**

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    @Query("SELECT * FROM users WHERE email = $1")
    Optional<User> findByEmail(String email);
}
```

**Query with Multiple Parameters:**

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    @Query("SELECT * FROM users WHERE email = $1 AND status = $2")
    Optional<User> findByEmailAndStatus(String email, String status);
}
```

**Query with LIMIT:**

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    @Query("SELECT * FROM users WHERE status = $1 ORDER BY created_at DESC LIMIT $2")
    List<User> findRecentByStatus(String status, int limit);
}
```

**Query with LIKE:**

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    @Query("SELECT * FROM users WHERE name LIKE $1")
    List<User> findByNamePattern(String pattern);
}

// Usage:
List<User> users = userRepository.findByNamePattern("%John%");
```

**Query with JOIN:**

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    @Query("SELECT u.* FROM users u JOIN orders o ON u.id = o.user_id WHERE o.total > $1")
    List<User> findUsersWithLargeOrders(double minTotal);
}
```

**Query with Aggregation:**

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    @Query("SELECT COUNT(*) as count FROM users WHERE status = $1")
    long countByStatus(String status);
    
    @Query("SELECT AVG(age) as avg_age FROM users WHERE status = $1")
    double averageAgeByStatus(String status);
}
```

**Query Returning Single Value:**

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    @Query("SELECT COUNT(*) FROM users WHERE email = $1")
    int countByEmail(String email);
}
```

---

## Query Method Patterns

### How To: Use Method Name Patterns

The framework automatically generates SQL from method names:

**Pattern: `findBy{Field}`**

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    // Generates: SELECT * FROM users WHERE email = $1
    Optional<User> findByEmail(String email);
    
    // Generates: SELECT * FROM users WHERE name = $1
    Optional<User> findByName(String name);
    
    // Generates: SELECT * FROM users WHERE status = $1
    List<User> findByStatus(String status);
}
```

**Pattern: `findBy{Field}And{Field}`**

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    // Generates: SELECT * FROM users WHERE email = $1 AND status = $2
    Optional<User> findByEmailAndStatus(String email, String status);
    
    // Generates: SELECT * FROM users WHERE name = $1 AND email = $2
    List<User> findByNameAndEmail(String name, String email);
}
```

**Pattern: `findBy{Field}Or{Field}`**

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    // Note: For OR conditions, use @Query annotation
    @Query("SELECT * FROM users WHERE email = $1 OR name = $2")
    List<User> findByEmailOrName(String email, String name);
}
```

**Important:** Method name patterns work for simple `AND` conditions. For complex queries (OR, LIKE, JOIN, etc.), use `@Query` annotation.

---

## Shard Key Handling

### How To: Specify Shard Key in Methods

**Method 1: Using `@ShardKey` Annotation on Parameter**

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    // @ShardKey tells framework which parameter is the shard key
    Optional<User> findById(@ShardKey String id);
    
    @Query("SELECT * FROM users WHERE email = $1")
    Optional<User> findByEmail(@ShardKey String email);  // email is shard key
}
```

**Method 2: Shard Key from Entity**

```java
// When you pass an entity, framework extracts shard key from @ShardKey field
User save(User entity);  // Extracts from entity.getId()

// Usage:
User user = new User("user-123", "John", "john@example.com");
userRepository.save(user);  // Uses "user-123" as shard key
```

**Method 3: Shard Key for findAll/count/deleteAll**

```java
// These methods require a shard key parameter for routing
List<User> findAll(@ShardKey String shardKey);
long count(@ShardKey String shardKey);
void deleteAll(@ShardKey String shardKey);

// Usage:
List<User> users = userRepository.findAll("user-123");  // Any shard key from that shard
```

**Method 4: Extract Shard Key from Any Field**

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    // If email is also a shard key, you can use it
    @Query("SELECT * FROM users WHERE name = $1")
    Optional<User> findByName(@ShardKey String email, String name);
    // ↑ email parameter is used as shard key, name is used in WHERE clause
}
```

### How To: Handle Multiple Shard Keys

If your entity has multiple potential shard keys:

```java
@Entity(table = "orders")
public class Order {
    @ShardKey
    private String orderId;  // Primary shard key
    
    private String userId;   // Secondary shard key (for user-based queries)
    // ...
}

@ShardingRepository(entity = Order.class)
public interface OrderRepository extends CrudRepository<Order, String> {
    // Use orderId as shard key
    Optional<Order> findById(@ShardKey String orderId);
    
    // Use userId as shard key for user-based queries
    @Query("SELECT * FROM orders WHERE user_id = $1")
    List<Order> findByUserId(@ShardKey String userId);
}
```

---

## Consistency Levels

### How To: Use Strong Consistency

**Default Behavior:** Write operations and `findById` use strong consistency automatically.

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    // Strong consistency by default (reads from primary)
    Optional<User> findById(@ShardKey String id);
    
    // Explicit strong consistency
    @StrongConsistency
    @Query("SELECT * FROM users WHERE email = $1")
    Optional<User> findByEmail(String email);
}
```

**When to use:**
- Read-after-write scenarios
- Critical data that must be up-to-date
- Financial transactions
- User authentication

### How To: Use Eventual Consistency

**For Read-Heavy Operations:**

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    // Eventual consistency (can read from replica)
    @EventualConsistency
    @Query("SELECT * FROM users WHERE status = $1")
    List<User> findByStatus(String status);
    
    // Eventual consistency for list operations
    @EventualConsistency
    List<User> findAll(@ShardKey String shardKey);
}
```

**When to use:**
- Read-heavy operations
- Analytics queries
- Non-critical data
- Dashboard displays
- Reporting

**Performance Benefit:** Eventual consistency reads can be faster as they may read from replicas.

---

## Return Types

### How To: Return Single Entity

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    // Returns User or null
    User findById(@ShardKey String id);
    
    // Returns Optional<User>
    Optional<User> findByEmail(String email);
}
```

**Usage:**

```java
// Method 1: Direct return (can be null)
User user = userRepository.findById("user-123");
if (user != null) {
    // Handle user
}

// Method 2: Optional (recommended)
Optional<User> user = userRepository.findByEmail("john@example.com");
user.ifPresent(u -> {
    System.out.println("Found: " + u.getName());
});
```

### How To: Return List of Entities

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    List<User> findAll(@ShardKey String shardKey);
    
    @Query("SELECT * FROM users WHERE status = $1")
    List<User> findByStatus(String status);
}
```

**Usage:**

```java
List<User> users = userRepository.findByStatus("active");
for (User user : users) {
    System.out.println(user.getName());
}

// Or with streams
userRepository.findByStatus("active")
    .stream()
    .map(User::getName)
    .forEach(System.out::println);
```

### How To: Return Primitive Types

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    // Boolean
    boolean existsById(@ShardKey String id);
    
    // Long
    long count(@ShardKey String shardKey);
    
    // Custom count query
    @Query("SELECT COUNT(*) FROM users WHERE status = $1")
    long countByStatus(String status);
}
```

**Usage:**

```java
if (userRepository.existsById("user-123")) {
    System.out.println("User exists");
}

long total = userRepository.count("user-123");
System.out.println("Total users: " + total);
```

### How To: Return Void (No Return Value)

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    void deleteById(@ShardKey String id);
    
    @Query("UPDATE users SET status = $2 WHERE id = $1")
    void updateStatus(@ShardKey String id, String status);
}
```

**Usage:**

```java
userRepository.deleteById("user-123");
userRepository.updateStatus("user-123", "inactive");
```

---

## Batch Operations

### How To: Save Multiple Entities

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    // Inherited from CrudRepository
    List<User> saveAll(Iterable<User> entities);
}
```

**Usage:**

```java
List<User> users = Arrays.asList(
    new User("user-1", "John", "john@example.com"),
    new User("user-2", "Jane", "jane@example.com"),
    new User("user-3", "Bob", "bob@example.com")
);

List<User> saved = userRepository.saveAll(users);
```

**Note:** Each entity is saved individually. For true batch inserts, use a custom query:

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    @Query("INSERT INTO users (id, name, email) VALUES ($1, $2, $3), ($4, $5, $6), ($7, $8, $9)")
    void batchInsert(String id1, String name1, String email1,
                     String id2, String name2, String email2,
                     String id3, String name3, String email3);
}
```

### How To: Delete Multiple Entities

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    // Delete all in a shard
    void deleteAll(@ShardKey String shardKey);
    
    // Custom batch delete
    @Query("DELETE FROM users WHERE id IN ($1, $2, $3)")
    void deleteByIds(@ShardKey String shardKey, String id1, String id2, String id3);
}
```

---

## Complex Queries

### How To: Use Subqueries

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    @Query("SELECT * FROM users WHERE id IN (SELECT user_id FROM orders WHERE total > $1)")
    List<User> findUsersWithOrdersAbove(double minTotal);
}
```

### How To: Use GROUP BY

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    @Query("SELECT status, COUNT(*) as count FROM users GROUP BY status")
    List<Map<String, Object>> countByStatus();
}
```

**Note:** For aggregation queries, you may need to return `Map<String, Object>` or create a DTO class.

### How To: Use CASE Statements

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    @Query("SELECT *, CASE WHEN age < 18 THEN 'minor' ELSE 'adult' END as category FROM users")
    List<User> findUsersWithCategory();
}
```

### How To: Use Window Functions

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    @Query("SELECT *, ROW_NUMBER() OVER (PARTITION BY status ORDER BY created_at) as row_num FROM users")
    List<User> findUsersWithRowNumbers();
}
```

---

## Error Handling

### How To: Handle Exceptions

**Basic Error Handling:**

```java
public class UserService {
    private UserRepository userRepository;
    
    public Optional<User> getUser(String id) {
        try {
            return userRepository.findById(id);
        } catch (RuntimeException e) {
            // Framework wraps ShardingClientException in RuntimeException
            System.err.println("Error fetching user: " + e.getMessage());
            e.printStackTrace();
            return Optional.empty();
        }
    }
}
```

**Detailed Error Handling:**

```java
import com.sharding.system.client.ShardingClientException;

public class UserService {
    private UserRepository userRepository;
    
    public Optional<User> getUser(String id) {
        try {
            return userRepository.findById(id);
        } catch (RuntimeException e) {
            if (e.getCause() instanceof ShardingClientException) {
                ShardingClientException sce = (ShardingClientException) e.getCause();
                // Handle specific sharding errors
                handleShardingError(sce);
            }
            throw e;
        }
    }
    
    private void handleShardingError(ShardingClientException e) {
        // Log error, send alert, etc.
        logger.error("Sharding error: " + e.getMessage(), e);
    }
}
```

**Error Handling Best Practices:**

```java
public class UserService {
    private UserRepository userRepository;
    
    public User createUser(User user) {
        try {
            // Validate input
            validateUser(user);
            
            // Check if exists
            if (userRepository.existsById(user.getId())) {
                throw new IllegalArgumentException("User already exists");
            }
            
            // Save
            return userRepository.save(user);
            
        } catch (IllegalArgumentException e) {
            // Business logic errors
            throw e;
        } catch (RuntimeException e) {
            // Infrastructure errors
            logger.error("Failed to create user", e);
            throw new ServiceException("Failed to create user", e);
        }
    }
}
```

---

## Integration Patterns

### How To: Integrate with Quarkus

**Step 1: Create CDI Producer**

```java
package com.example.config;

import com.sharding.system.client.ShardingClient;
import com.sharding.system.client.config.ShardingClientAutoConfiguration;
import com.sharding.system.client.repository.ShardingRepositoryFactory;
import com.example.repository.UserRepository;
import jakarta.enterprise.context.ApplicationScoped;
import jakarta.enterprise.inject.Produces;
import org.eclipse.microprofile.config.inject.ConfigProperty;

@ApplicationScoped
public class ShardingConfiguration {
    
    @ConfigProperty(name = "sharding.router.url", defaultValue = "http://localhost:8080")
    String routerUrl;
    
    @Produces
    @ApplicationScoped
    public ShardingClient shardingClient() {
        return new ShardingClient(routerUrl);
    }
    
    @Produces
    @ApplicationScoped
    public UserRepository userRepository(ShardingClient shardingClient) {
        return ShardingRepositoryFactory.createRepository(shardingClient, UserRepository.class);
    }
}
```

**Step 2: Use in Service**

```java
package com.example.service;

import com.example.repository.UserRepository;
import jakarta.enterprise.context.ApplicationScoped;
import jakarta.inject.Inject;

@ApplicationScoped
public class UserService {
    
    @Inject
    UserRepository userRepository;
    
    public Optional<User> getUser(String id) {
        return userRepository.findById(id);
    }
}
```

### How To: Integrate with Spring Boot

**Step 1: Create Configuration Class**

```java
package com.example.config;

import com.sharding.system.client.ShardingClient;
import com.sharding.system.client.config.ShardingClientAutoConfiguration;
import com.example.repository.UserRepository;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class ShardingConfig {
    
    @Value("${sharding.router.url:http://localhost:8080}")
    private String routerUrl;
    
    @Bean
    public ShardingClient shardingClient() {
        return new ShardingClient(routerUrl);
    }
    
    @Bean
    public UserRepository userRepository(ShardingClient shardingClient) {
        ShardingClientAutoConfiguration config = new ShardingClientAutoConfiguration();
        config.setRouterUrl(routerUrl);
        config.initialize();
        return config.getRepository(UserRepository.class);
    }
}
```

**Step 2: Use in Service**

```java
package com.example.service;

import com.example.repository.UserRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

@Service
public class UserService {
    
    @Autowired
    private UserRepository userRepository;
    
    public Optional<User> getUser(String id) {
        return userRepository.findById(id);
    }
}
```

### How To: Use in Standard Java Application

```java
public class Application {
    private static UserRepository userRepository;
    
    public static void main(String[] args) {
        // Initialize
        ShardingClientAutoConfiguration config = new ShardingClientAutoConfiguration();
        config.setRouterUrl("http://localhost:8080");
        config.initialize();
        userRepository = config.getRepository(UserRepository.class);
        
        // Use
        Optional<User> user = userRepository.findById("user-123");
        user.ifPresent(u -> System.out.println("Found: " + u.getName()));
        
        // Cleanup
        Runtime.getRuntime().addShutdownHook(new Thread(() -> {
            config.close();
        }));
    }
}
```

---

## Best Practices

### How To: Design Your Entities

**✅ DO:**

```java
@Entity(table = "users")
public class User {
    @ShardKey
    private String id;  // Use meaningful shard key
    
    private String name;
    private String email;
    
    // Include timestamps
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;
    
    // Use appropriate types
    private BigDecimal balance;  // Not double for money
    private LocalDate birthDate;  // Not String for dates
}
```

**❌ DON'T:**

```java
@Entity
public class User {
    // Missing @ShardKey annotation
    private String id;
    
    // Using wrong types
    private String balance;  // Should be BigDecimal
    private String birthDate;  // Should be LocalDate
}
```

### How To: Design Your Repositories

**✅ DO:**

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    // Use Optional for single results
    Optional<User> findByEmail(String email);
    
    // Use List for multiple results
    List<User> findByStatus(String status);
    
    // Use descriptive method names
    List<User> findActiveUsersCreatedAfter(LocalDateTime date);
    
    // Use @Query for complex queries
    @Query("SELECT * FROM users WHERE ...")
    List<User> findComplexQuery(...);
}
```

**❌ DON'T:**

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    // Don't return null directly
    User findByEmail(String email);  // Use Optional instead
    
    // Don't use vague names
    List<User> find(String email);  // Use findByEmail
    
    // Don't forget @ShardKey for routing
    Optional<User> findById(String id);  // Should be @ShardKey String id
}
```

### How To: Handle Transactions

Currently, the framework doesn't support transactions. For transactional operations:

```java
// Use the underlying ShardingClient for transaction support
public class TransactionalUserService {
    private ShardingClient shardingClient;
    
    public void transferMoney(String fromUserId, String toUserId, BigDecimal amount) {
        // Manual transaction handling
        try {
            // Debit
            shardingClient.queryStrong(fromUserId, 
                "UPDATE accounts SET balance = balance - $1 WHERE user_id = $2",
                amount, fromUserId);
            
            // Credit
            shardingClient.queryStrong(toUserId,
                "UPDATE accounts SET balance = balance + $1 WHERE user_id = $2",
                amount, toUserId);
        } catch (Exception e) {
            // Rollback logic
            throw new TransactionException("Transfer failed", e);
        }
    }
}
```

---

## Common Patterns

### How To: Implement Pagination

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    @Query("SELECT * FROM users WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3")
    List<User> findByStatusPaginated(String status, int limit, int offset);
}

// Usage:
int page = 0;
int pageSize = 10;
List<User> users = userRepository.findByStatusPaginated("active", pageSize, page * pageSize);
```

### How To: Implement Soft Delete

```java
@Entity(table = "users")
public class User {
    @ShardKey
    private String id;
    
    private boolean deleted;  // Soft delete flag
    private LocalDateTime deletedAt;
    // ...
}

@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    @Query("UPDATE users SET deleted = true, deleted_at = NOW() WHERE id = $1")
    void softDelete(@ShardKey String id);
    
    @Query("SELECT * FROM users WHERE deleted = false AND status = $1")
    List<User> findActiveUsers(String status);
}
```

### How To: Implement Audit Logging

```java
@Entity(table = "users")
public class User {
    @ShardKey
    private String id;
    
    private String name;
    private LocalDateTime createdAt;
    private String createdBy;
    private LocalDateTime updatedAt;
    private String updatedBy;
    // ...
}

@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    @Query("UPDATE users SET name = $2, updated_at = NOW(), updated_by = $3 WHERE id = $1")
    void updateWithAudit(@ShardKey String id, String name, String updatedBy);
}
```

### How To: Implement Search

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    @Query("SELECT * FROM users WHERE name ILIKE $1 OR email ILIKE $1 LIMIT $2")
    List<User> search(String query, int limit);
    
    @Query("SELECT * FROM users WHERE name ILIKE $1 OR email ILIKE $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3")
    List<User> searchPaginated(String query, int limit, int offset);
}
```

---

## Troubleshooting

### Problem: "No shard key field found"

**Solution:** Add `@ShardKey` annotation to your entity's ID field:

```java
@Entity(table = "users")
public class User {
    @ShardKey  // ← Add this
    private String id;
    // ...
}
```

### Problem: "Cannot determine shard key for method"

**Solution:** Add `@ShardKey` annotation to method parameter:

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    Optional<User> findById(@ShardKey String id);  // ← Add @ShardKey
}
```

### Problem: Mapping fails - column not found

**Solution:** Use `@Column` annotation to map field to column:

```java
@Entity(table = "users")
public class User {
    @Column(name = "user_id")  // ← Explicit column mapping
    private String id;
    
    @Column(name = "full_name")  // ← Explicit column mapping
    private String name;
}
```

### Problem: Query returns empty results

**Check:**
1. Shard key is correct
2. Table name matches
3. Column names match
4. Data exists in the shard

**Debug:**

```java
// Check if entity exists
boolean exists = userRepository.existsById("user-123");
System.out.println("Exists: " + exists);

// Check shard key
String shardId = shardingClient.getShardForKey("user-123");
System.out.println("Shard ID: " + shardId);
```

### Problem: Performance issues

**Solutions:**

1. **Use eventual consistency for reads:**
```java
@EventualConsistency
List<User> findAll(@ShardKey String shardKey);
```

2. **Add LIMIT to queries:**
```java
@Query("SELECT * FROM users WHERE status = $1 LIMIT 100")
List<User> findByStatus(String status);
```

3. **Use indexes** (database level)

4. **Batch operations** instead of individual saves

---

## Quick Reference

### Common Repository Methods

```java
// Create
T save(T entity);
List<T> saveAll(Iterable<T> entities);

// Read
Optional<T> findById(@ShardKey ID id);
boolean existsById(@ShardKey ID id);
List<T> findAll(@ShardKey String shardKey);
long count(@ShardKey String shardKey);

// Update
T save(T entity);  // Upsert

// Delete
void deleteById(@ShardKey ID id);
void delete(T entity);
void deleteAll(@ShardKey String shardKey);
```

### Common Annotations

```java
@Entity(table = "table_name")           // Entity class
@ShardKey                               // Shard key field/parameter
@Column(name = "column_name")           // Column mapping
@ShardingRepository(entity = X.class)   // Repository interface
@Query("SQL query")                     // Custom query
@StrongConsistency                      // Strong consistency
@EventualConsistency                    // Eventual consistency
```

### Common Return Types

```java
T                    // Single entity (can be null)
Optional<T>          // Single entity (null-safe)
List<T>              // Multiple entities
boolean              // Existence check
long                 // Count
void                 // No return value
```

---

## Conclusion

This guide covers all major use cases for the Java Low-Code Client. For more information:

- [Low-Code Guide](./LOW_CODE_GUIDE.md) - Overview and concepts
- [Naming Conventions](./NAMING_CONVENTIONS.md) - Naming flexibility
- [Java Client Reference](./java.md) - API documentation
- [Quarkus Integration](./QUARKUS_INTEGRATION.md) - Framework integration

Happy coding! 🚀

