# Quarkus Sharding Example Service

This is a complete example of a Quarkus microservice using the Sharding System.

## Prerequisites

1. Java 11+
2. Maven 3.8+
3. Sharding System running (see main [QUICKSTART.md](../../QUICKSTART.md))

## Setup

### 1. Build the Sharding Client Library

First, build and install the Java client library:

```bash
cd ../../clients/java
mvn clean install
```

### 2. Start the Sharding System

Start the sharding router and manager:

```bash
cd ../../..
docker-compose up -d
```

### 3. Create a Shard

Create a shard using the Manager API:

```bash
curl -X POST http://localhost:8081/api/v1/shards \
  -H "Content-Type: application/json" \
  -d '{
    "name": "shard-01",
    "primary_endpoint": "postgres://postgres:postgres@postgres-shard1:5432/shard1",
    "replicas": [],
    "vnode_count": 256
  }'
```

### 4. Create the Users Table

Connect to the PostgreSQL shard and create the users table:

```bash
docker exec -it postgres-shard1 psql -U postgres -d shard1
```

```sql
CREATE TABLE users (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL
);
```

## Running the Service

### Development Mode

```bash
./mvnw quarkus:dev
```

The service will be available at `http://localhost:8080`

### Production Mode

```bash
./mvnw clean package
java -jar target/quarkus-app/quarkus-run.jar
```

## API Endpoints

### Create User

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "id": "user-123",
    "name": "John Doe",
    "email": "john@example.com"
  }'
```

### Get User

```bash
curl http://localhost:8080/users/user-123
```

### Update User

```bash
curl -X PUT http://localhost:8080/users/user-123 \
  -H "Content-Type: application/json" \
  -d '{
    "id": "user-123",
    "name": "Jane Doe",
    "email": "jane@example.com"
  }'
```

### Delete User

```bash
curl -X DELETE http://localhost:8080/users/user-123
```

### List Users (for a specific shard key)

```bash
curl "http://localhost:8080/users?shardKey=user-123"
```

### Get Shard for Key

```bash
curl http://localhost:8080/users/shard/user-123
```

## Project Structure

```
quarkus-service/
├── src/
│   ├── main/
│   │   ├── java/com/example/
│   │   │   ├── model/
│   │   │   │   └── User.java          # Domain model
│   │   │   ├── resource/
│   │   │   │   └── UserResource.java  # REST endpoints
│   │   │   └── service/
│   │   │       ├── ShardingClientProducer.java  # CDI producer
│   │   │       └── UserService.java   # Business logic
│   │   └── resources/
│   │       └── application.properties # Configuration
│   └── test/
└── pom.xml
```

## Key Features Demonstrated

1. **CDI Integration**: Using `@Produces` to create `ShardingClient` bean
2. **Dependency Injection**: Injecting `ShardingClient` into services
3. **REST API**: JAX-RS endpoints using the sharding client
4. **Error Handling**: Proper exception handling and error responses
5. **Configuration**: Using MicroProfile Config for router URL

## Next Steps

- Add validation using Bean Validation
- Add metrics using Micrometer
- Add retry logic using Resilience4j
- Add caching for frequently accessed data
- Add OpenAPI documentation

