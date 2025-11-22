# Quick Start Guide

Get the Java E-Commerce Service running in 5 minutes!

## Prerequisites

- Java 17+
- Maven 3.6+
- Sharding System running (see main README)

## Step 1: Build the Application

```bash
cd examples/java-ecommerce-service
mvn clean install
```

## Step 2: Set Up Database Schemas

Run these SQL scripts on each shard database:

```bash
# Example for shard1
psql -h localhost -U postgres -d shard1 -f src/main/resources/db/migration/001_create_users_table.sql
psql -h localhost -U postgres -d shard1 -f src/main/resources/db/migration/002_create_orders_table.sql
psql -h localhost -U postgres -d shard1 -f src/main/resources/db/migration/003_create_order_items_table.sql
psql -h localhost -U postgres -d shard1 -f src/main/resources/db/migration/004_create_products_table.sql
```

## Step 3: Configure and Run

```bash
export SHARDING_ROUTER_URL=http://localhost:8080
mvn spring-boot:run
```

## Step 4: Test the API

### Create a User
```bash
curl -X POST http://localhost:8082/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "id": "user-001",
    "username": "alice",
    "email": "alice@example.com",
    "fullName": "Alice Smith"
  }'
```

### Create an Order
```bash
curl -X POST http://localhost:8082/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-001",
    "totalAmount": 49.99,
    "status": "PENDING",
    "shippingAddress": "123 Main St",
    "items": [{
      "productId": "prod-1",
      "productName": "Widget",
      "quantity": 1,
      "unitPrice": 49.99,
      "totalPrice": 49.99
    }]
  }'
```

### Get User Orders
```bash
curl http://localhost:8082/api/v1/orders/user/user-001
```

### See Sharding in Action
```bash
curl http://localhost:8082/api/v1/demo/shard-distribution?sampleSize=10
```

## Access Documentation

- Swagger UI: http://localhost:8082/swagger-ui.html
- Health Check: http://localhost:8082/actuator/health

That's it! You're ready to explore the sharding system.

