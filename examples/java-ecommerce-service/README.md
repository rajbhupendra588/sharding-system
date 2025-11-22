# Java E-Commerce Service with Database Sharding

A production-ready Spring Boot application demonstrating how to use the Sharding System to build scalable, high-performance microservices.

## Overview

This application showcases an e-commerce service that leverages database sharding to:
- **Scale horizontally** as user base grows
- **Improve query performance** by distributing data across multiple shards
- **Co-locate related data** (users and their orders) for efficient queries
- **Handle high-volume transactions** without database bottlenecks

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│              Java E-Commerce Service (Spring Boot)          │
│                      Port: 8082                             │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        │ HTTP/REST
                        ▼
┌─────────────────────────────────────────────────────────────┐
│              Sharding Router (Data Plane)                    │
│                      Port: 8080                             │
└───────────────────────┬─────────────────────────────────────┘
                        │
        ┌───────────────┼───────────────┐
        ▼               ▼               ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│   Shard 1    │ │   Shard 2    │ │   Shard N   │
│  (PostgreSQL)│ │  (PostgreSQL)│ │ (PostgreSQL)│
└─────────────┘ └─────────────┘ └─────────────┘
```

## Key Features

### 1. **Sharding by User ID**
- Users are sharded by `user_id`
- Orders are co-located with users (sharded by `user_id`)
- Enables efficient queries like "get all orders for user X"

### 2. **Production-Ready Components**
- ✅ Spring Boot 3.2 with Java 17
- ✅ RESTful API with OpenAPI/Swagger documentation
- ✅ Comprehensive error handling
- ✅ Health checks and metrics (Prometheus)
- ✅ Structured logging
- ✅ Input validation
- ✅ Database migration scripts

### 3. **Sharding Benefits Demonstrated**

#### Horizontal Scaling
- Add more shards as user base grows
- No single database bottleneck
- Linear scalability

#### Performance
- Queries hit only one shard (not entire database)
- Fast user order history (co-located data)
- Lower latency for user-specific operations

#### Co-location
- User and their orders on same shard
- Efficient joins within a shard
- Fast order history queries

#### Fault Isolation
- Shard failures don't affect entire system
- Better availability and resilience

## Prerequisites

- Java 17 or higher
- Maven 3.6+
- Sharding System running (Router on port 8080, Manager on port 8081)
- PostgreSQL databases configured as shards

## Quick Start

### 1. Build the Application

```bash
cd examples/java-ecommerce-service
mvn clean install
```

### 2. Configure Sharding Router URL

Set the environment variable or update `application.yml`:

```bash
export SHARDING_ROUTER_URL=http://localhost:8080
```

Or edit `src/main/resources/application.yml`:
```yaml
sharding:
  router:
    url: http://localhost:8080
```

### 3. Initialize Database Schemas

Run the SQL migration scripts on each shard database:

```bash
# For each shard database
psql -h localhost -U postgres -d shard1 < src/main/resources/db/migration/001_create_users_table.sql
psql -h localhost -U postgres -d shard1 < src/main/resources/db/migration/002_create_orders_table.sql
psql -h localhost -U postgres -d shard1 < src/main/resources/db/migration/003_create_order_items_table.sql
psql -h localhost -U postgres -d shard1 < src/main/resources/db/migration/004_create_products_table.sql

# Repeat for shard2, shard3, etc.
```

### 4. Run the Application

```bash
mvn spring-boot:run
```

The application will start on `http://localhost:8082`

## API Documentation

Once the application is running, access:
- **Swagger UI**: http://localhost:8082/swagger-ui.html
- **OpenAPI JSON**: http://localhost:8082/api-docs

## API Endpoints

### Users
- `POST /api/v1/users` - Create a new user
- `GET /api/v1/users/{id}` - Get user by ID
- `PUT /api/v1/users/{id}` - Update user
- `DELETE /api/v1/users/{id}` - Delete user (soft delete)
- `GET /api/v1/users/{id}/shard` - Get shard information for user

### Orders
- `POST /api/v1/orders` - Create a new order
- `GET /api/v1/orders/{orderId}?userId={userId}` - Get order by ID
- `GET /api/v1/orders/user/{userId}` - Get all orders for a user
- `PUT /api/v1/orders/{orderId}/status?userId={userId}` - Update order status
- `GET /api/v1/orders/user/{userId}/statistics` - Get order statistics

### Products
- `POST /api/v1/products` - Create a new product
- `GET /api/v1/products/{id}` - Get product by ID
- `PUT /api/v1/products/{id}` - Update product
- `PUT /api/v1/products/{id}/stock` - Update product stock
- `GET /api/v1/products/category/{category}` - Get products by category

### Sharding Demo
- `GET /api/v1/demo/shard-distribution?sampleSize=10` - Show shard distribution
- `GET /api/v1/demo/benefits` - Explain sharding benefits
- `GET /api/v1/demo/shard-for-key/{key}` - Get shard for a key

## Example Usage

### Create a User

```bash
curl -X POST http://localhost:8082/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "id": "user-123",
    "username": "johndoe",
    "email": "john@example.com",
    "fullName": "John Doe",
    "phoneNumber": "+1234567890",
    "address": "123 Main St, City, State"
  }'
```

### Create an Order

```bash
curl -X POST http://localhost:8082/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-123",
    "totalAmount": 99.99,
    "status": "PENDING",
    "shippingAddress": "123 Main St, City, State",
    "paymentMethod": "CREDIT_CARD",
    "items": [
      {
        "productId": "prod-1",
        "productName": "Widget",
        "quantity": 2,
        "unitPrice": 49.99,
        "totalPrice": 99.98
      }
    ]
  }'
```

### Get User Orders (Efficient - Single Shard Query)

```bash
curl http://localhost:8082/api/v1/orders/user/user-123
```

### Check Shard Distribution

```bash
curl http://localhost:8082/api/v1/demo/shard-distribution?sampleSize=20
```

## How Sharding Helps

### 1. **Scalability**
As your user base grows from thousands to millions, you can add more shards without rewriting your application code. The sharding router automatically distributes new users across available shards.

### 2. **Performance**
- **Before Sharding**: Query for user orders scans entire database (millions of rows)
- **After Sharding**: Query hits only one shard (thousands of rows per shard)
- **Result**: 10-100x faster queries

### 3. **Co-location Benefits**
By sharding orders by `user_id`, all orders for a user are stored on the same shard as the user. This enables:
- Fast order history queries (single shard)
- Efficient joins (user + orders)
- Potential for shard-local transactions

### 4. **Fault Isolation**
If one shard fails, only users on that shard are affected. Other shards continue operating normally.

### 5. **Cost Efficiency**
- Scale individual shards based on load
- Use smaller, cheaper database instances
- Better resource utilization

## Monitoring

### Health Checks
- **Application Health**: http://localhost:8082/actuator/health
- **Sharding System Health**: Included in health endpoint

### Metrics
- **Prometheus Metrics**: http://localhost:8082/actuator/prometheus
- **Application Metrics**: http://localhost:8082/actuator/metrics

### Logs
Application logs are written to `logs/ecommerce-service.log` with rotation.

## Production Considerations

### 1. **Connection Pooling**
The application uses the sharding client's built-in connection pooling. For production:
- Configure pool size based on load
- Monitor connection pool metrics
- Set appropriate timeouts

### 2. **Error Handling**
- Implement retry logic for transient failures
- Use circuit breakers for resilience
- Monitor error rates and alert on thresholds

### 3. **Caching**
Consider adding caching layers:
- User data cache (Redis)
- Product catalog cache
- Shard routing cache (already handled by router)

### 4. **Monitoring**
- Set up Prometheus + Grafana dashboards
- Monitor query latency per shard
- Track shard distribution and load
- Alert on shard failures

### 5. **Security**
- Use HTTPS in production
- Implement authentication/authorization
- Validate and sanitize all inputs
- Use parameterized queries (already implemented)

### 6. **Database Migrations**
- Use a proper migration tool (Flyway/Liquibase) in production
- Test migrations on staging first
- Have rollback plans ready

## Troubleshooting

### Application won't start
- Check if sharding router is running on port 8080
- Verify `SHARDING_ROUTER_URL` environment variable
- Check application logs: `logs/ecommerce-service.log`

### Queries failing
- Verify database schemas are created on all shards
- Check sharding router logs
- Verify shard connectivity from router

### Performance issues
- Check shard distribution (use `/api/v1/demo/shard-distribution`)
- Monitor query latency metrics
- Consider adding read replicas for high-read workloads

## Learn More

- [Sharding System Architecture](../../docs/architecture/ARCHITECTURE.md)
- [API Documentation](../../docs/api/API_ENDPOINTS.md)
- [Java Client Guide](../../docs/clients/JAVA_QUICKSTART.md)
- [Production Deployment Guide](../../docs/deployment/PRODUCTION.md)

## License

This is a sample application demonstrating sharding capabilities.

