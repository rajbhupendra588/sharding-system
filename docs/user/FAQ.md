# Frequently Asked Questions (FAQ)

## General

### What is the Sharding System?

The Sharding System is a production-ready, self-contained database sharding service that provides transparent routing, online resharding, replication management, health monitoring, and comprehensive observability. It separates the data plane (request routing) from the control plane (configuration and management).

### What problem does it solve?

It solves the problem of scaling databases horizontally by:
- Distributing data across multiple database instances (shards)
- Providing transparent routing so applications don't need to know which shard contains their data
- Enabling online resharding without downtime
- Managing replication and failover automatically

### How does it differ from database replication?

- **Replication**: Copies the same data to multiple servers (high availability, read scaling)
- **Sharding**: Distributes different data across multiple servers (write scaling, storage scaling)

The Sharding System can use replication within each shard for high availability, while sharding provides horizontal scaling.

### What databases are supported?

Currently, the system works with any SQL database that supports standard SQL queries (PostgreSQL, MySQL, etc.). The shard databases themselves are independent - you can use different database types for different shards if needed.

### Is it production-ready?

Yes, the system is designed for production use with features like:
- High availability and failover
- Health monitoring
- Comprehensive metrics
- Security (RBAC, audit logging)
- Online resharding

However, always test thoroughly in your environment before production deployment.

## Getting Started

### How do I access the Web UI?

Navigate to `http://localhost:3000` in your browser after starting the services. If running via Docker Compose, ensure the UI service is included.

### What are the default ports?

- **Router**: 8080 (API), 9090 (metrics)
- **Manager**: 8081 (API), 9091 (metrics)
- **etcd**: 2389 (client), 2390 (peer)
- **Web UI**: 3000

### How do I create my first shard?

1. **Via Web UI:**
   - Navigate to "Shards" section
   - Click "Create Shard"
   - Enter shard name and database endpoints
   - Submit

2. **Via API:**
   ```bash
   curl -X POST http://localhost:8081/api/v1/shards \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer <token>" \
     -d '{
       "name": "shard-1",
       "primary_endpoint": "postgresql://user:pass@localhost:5432/shard1",
       "replicas": [],
       "vnode_count": 256
     }'
   ```

### What is a shard key?

A shard key is a value used to determine which shard stores your data. The system computes a hash of the shard key to route queries to the correct shard. Examples:
- User ID: `"user-123"`
- Order ID: `"order-456"`
- Tenant ID: `"tenant-789"`

**Important:** Choose shard keys with high cardinality and even distribution.

## Configuration

### How do I configure the system?

Configuration is done via JSON files:
- `configs/manager.json`: Manager service configuration
- `configs/router.json`: Router service configuration

See the [Configuration Guide](./CONFIGURATION_GUIDE.md) for details.

### Can I use environment variables?

Yes, some settings can be overridden:
- `JWT_SECRET`: Required when RBAC is enabled
- `USER_DATABASE_DSN`: User database connection string
- `CONFIG_PATH`: Path to configuration file

### How do I enable authentication?

Set `enable_rbac: true` in the security configuration and provide a `JWT_SECRET` environment variable (minimum 32 characters).

## Sharding

### How does sharding work?

1. Client provides a shard key with the query
2. Router computes hash of the shard key (Murmur3)
3. Hash is mapped to a virtual node (vnode)
4. Vnode belongs to a physical shard
5. Query is routed to that shard

### What is consistent hashing?

Consistent hashing is a technique that minimizes data movement when adding or removing shards. Instead of `hash % n` (which requires moving most data when `n` changes), consistent hashing only requires moving ~1/n of the data.

### What are virtual nodes (vnodes)?

Virtual nodes allow each physical shard to be represented by multiple points on the hash ring. This provides:
- Better load balancing
- Smoother resharding operations
- Reduced hotspots

Default is 256 vnodes per shard (configurable).

### How do I choose a good shard key?

Choose shard keys with:
- **High cardinality**: Many unique values
- **Even distribution**: Avoid keys that create hotspots
- **Query relevance**: Keys you frequently query by

**Good examples:** User ID, Order ID, UUID
**Bad examples:** Country (if 90% users are from one country), Status (limited values)

### Can I change the shard key after data is loaded?

No, changing the shard key would require moving all data. Choose your shard key carefully before loading data.

## Resharding

### What is resharding?

Resharding is the process of changing the number of shards or redistributing data. The system supports:
- **Split**: Dividing one shard into multiple shards
- **Merge**: Combining multiple shards into one

### How does online resharding work?

The resharding process:
1. **Precopy**: Copy historical data to new shards
2. **Deltasync**: Synchronize incremental changes
3. **Cutover**: Switch traffic to new shards
4. **Complete**: Old shards can be removed

All steps happen without downtime.

### How long does resharding take?

Depends on:
- Amount of data to migrate
- Network bandwidth
- Database performance
- Number of keys

Monitor the resharding job status via the API to track progress.

### Can I cancel a resharding job?

Currently, resharding jobs cannot be cancelled once started. Ensure you have sufficient resources before starting a resharding operation.

## Query Execution

### How do I execute a query?

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

### What is the difference between strong and eventual consistency?

- **Strong Consistency**: Reads from primary database (latest data, higher latency)
- **Eventual Consistency**: Reads from replica (may be slightly stale, lower latency)

Use strong consistency for critical reads, eventual for read-heavy workloads.

### Can I query across multiple shards?

Currently, queries must target a single shard. Cross-shard queries are planned for future releases. Design your schema to keep related data on the same shard.

### What SQL features are supported?

The system supports standard SQL queries. However:
- Queries must target a single shard
- Joins across shards are not supported
- Transactions across shards are not supported

## Troubleshooting

### Why is my shard not showing up?

1. Check shard creation logs
2. Verify shard status: `GET /api/v1/shards/{id}`
3. Ensure shard status is "active"
4. Check Router has refreshed catalog

### Why are queries failing?

1. **Check shard key**: Ensure shard key is provided
2. **Verify shard exists**: Check shard is active
3. **Database connectivity**: Verify database is accessible
4. **Query syntax**: Ensure SQL is valid
5. **Check logs**: Review Router and Manager logs

### How do I reset my password?

If using RBAC with a user database:
1. Connect to user database directly
2. Update password hash
3. Or use password reset functionality (if implemented)

For development, you can disable RBAC temporarily.

### Why is authentication failing?

1. Verify RBAC is enabled: Check config
2. Check JWT_SECRET is set: `echo $JWT_SECRET`
3. Verify token is valid: Check expiration
4. Include token in header: `Authorization: Bearer <token>`

### Services won't start

1. **Check logs**: `docker-compose logs` or `kubectl logs`
2. **Verify etcd**: Ensure etcd is running
3. **Check ports**: Ensure ports are not in use
4. **Verify config**: Check configuration files are valid
5. **Check resources**: Ensure sufficient memory/CPU

### High query latency

1. **Database performance**: Check database metrics
2. **Network latency**: Verify network connectivity
3. **Connection pool**: Check connection pool settings
4. **Query optimization**: Review query performance
5. **Shard health**: Verify shards are healthy

### Resharding job stuck

1. **Check job status**: `GET /api/v1/reshard/jobs/{id}`
2. **Review error message**: Check for specific errors
3. **Database connectivity**: Verify databases are accessible
4. **Disk space**: Ensure sufficient disk space
5. **Review logs**: Check Manager logs for details

## Performance

### What is the performance overhead?

The routing overhead is minimal:
- Hash computation: < 0.1ms
- Catalog lookup: < 1ms (cached)
- Total overhead: < 2ms (excluding database query time)

### How many queries per second can it handle?

Depends on:
- Hardware resources
- Database performance
- Query complexity
- Number of router instances

A single router instance can handle 10,000+ queries/second. Scale horizontally by adding more router instances.

### How do I scale the system?

1. **Add more shards**: Increase storage and write capacity
2. **Add router replicas**: Scale query throughput
3. **Add database replicas**: Scale read capacity
4. **Optimize queries**: Add indexes, optimize SQL

### What are the resource requirements?

**Minimum (Development):**
- Router: 128MB RAM, 0.1 CPU
- Manager: 256MB RAM, 0.1 CPU

**Recommended (Production):**
- Router: 512MB-1GB RAM, 0.5-1 CPU per instance
- Manager: 512MB-1GB RAM, 0.5-1 CPU

## Security

### Is the system secure?

The system includes:
- JWT-based authentication
- Role-Based Access Control (RBAC)
- TLS/SSL support
- Input validation
- Audit logging

However, security is a shared responsibility. Follow security best practices:
- Enable RBAC in production
- Use strong JWT secrets
- Enable TLS for production
- Regularly update dependencies
- Monitor audit logs

### How do I enable TLS?

Set `enable_tls: true` in the security configuration and provide TLS certificates.

### What is audit logging?

Audit logging records all operations for compliance and security:
- Shard creation/deletion
- Resharding operations
- Authentication events
- Configuration changes

Logs are written to the path specified in `audit_log_path`.

## Support

### Where can I get help?

1. **Documentation**: Check the [Documentation Index](../README.md)
2. **FAQ**: This document
3. **Issues**: Open an issue on GitHub
4. **Discussions**: Use GitHub Discussions

### How do I report a bug?

Open an issue on GitHub with:
- Clear description
- Steps to reproduce
- Expected vs actual behavior
- Environment details
- Relevant logs

### How do I request a feature?

Open an issue on GitHub with:
- Feature description
- Use case
- Proposed solution (if any)
- Benefits

## Additional Resources

- [User Guide](./USER_GUIDE.md)
- [Configuration Guide](./CONFIGURATION_GUIDE.md)
- [API Reference](../api/API_REFERENCE.md)
- [Architecture Documentation](../architecture/ARCHITECTURE.md)
- [Developer Guide](../dev/DEVELOPER_GUIDE.md)

