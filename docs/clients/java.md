# Sharding System Java Client

Java client library for interacting with the Sharding System Router API.

## Installation

### Maven

Add the dependency to your `pom.xml`:

```xml
<dependency>
    <groupId>com.sharding-system</groupId>
    <artifactId>sharding-client</artifactId>
    <version>1.0.0</version>
</dependency>
```

### Gradle

Add the dependency to your `build.gradle`:

```gradle
implementation 'com.sharding-system:sharding-client:1.0.0'
```

## Usage

### Basic Usage

```java
import com.sharding.system.client.ShardingClient;
import com.sharding.system.client.ShardingClientException;
import com.sharding.system.client.model.QueryResponse;

// Create a client
ShardingClient client = new ShardingClient("http://localhost:8080");

try {
    // Get shard for a key
    String shardId = client.getShardForKey("user-123");
    System.out.println("Shard ID: " + shardId);
    
    // Execute a query with strong consistency
    QueryResponse response = client.queryStrong(
        "user-123",
        "SELECT * FROM users WHERE id = $1",
        "user-123"
    );
    
    System.out.println("Rows: " + response.getRowCount());
    System.out.println("Latency: " + response.getLatencyMs() + " ms");
    
    // Access results
    for (Map<String, Object> row : response.getRows()) {
        System.out.println("User: " + row.get("name"));
    }
    
} catch (ShardingClientException e) {
    e.printStackTrace();
} finally {
    client.close();
}
```

### Consistency Levels

**Strong Consistency** (reads from primary):
```java
QueryResponse response = client.queryStrong(
    "user-123",
    "SELECT * FROM users WHERE id = $1",
    "user-123"
);
```

**Eventual Consistency** (can read from replica):
```java
QueryResponse response = client.queryEventual(
    "user-123",
    "SELECT COUNT(*) FROM users"
);
```

### Query Parameters

The client supports positional parameters (PostgreSQL-style `$1`, `$2`, etc.):

```java
QueryResponse response = client.queryStrong(
    "user-123",
    "SELECT * FROM users WHERE id = $1 AND status = $2",
    "user-123",
    "active"
);
```

## API Reference

### ShardingClient

#### Constructor

```java
ShardingClient(String routerUrl)
```

Creates a new client instance pointing to the router URL.

#### Methods

**getShardForKey(String key)**
- Returns the shard ID for a given key
- Throws `ShardingClientException` on failure

**query(String shardKey, String query, List<Object> params, String consistency)**
- Executes a query with specified consistency level
- Returns `QueryResponse`
- Throws `ShardingClientException` on failure

**queryStrong(String shardKey, String query, Object... params)**
- Executes a query with strong consistency
- Convenience method with varargs

**queryEventual(String shardKey, String query, Object... params)**
- Executes a query with eventual consistency
- Convenience method with varargs

**close()**
- Closes the HTTP client and releases resources

## Error Handling

All methods throw `ShardingClientException` on failure. Always wrap calls in try-catch:

```java
try {
    QueryResponse response = client.queryStrong("user-123", "SELECT * FROM users WHERE id = $1", "user-123");
    // Process response
} catch (ShardingClientException e) {
    // Handle error
    logger.error("Query failed", e);
}
```

## Building from Source

```bash
cd clients/java
mvn clean install
```

This will build the JAR and install it to your local Maven repository.

