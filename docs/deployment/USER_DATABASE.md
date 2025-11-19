# User Database Setup (MAANG Production Standard)

## Overview

The sharding system now supports PostgreSQL-backed user storage, following MAANG (Meta, Amazon, Apple, Netflix, Google) production standards.

## Features

- ✅ **Database-backed user storage** - Persistent user data
- ✅ **Connection pooling** - Optimized for high concurrency
- ✅ **In-memory caching** - Fast user lookups
- ✅ **Account lockout** - Automatic after 5 failed attempts
- ✅ **Audit logging** - Track login attempts and last login
- ✅ **Automatic schema creation** - Self-initializing
- ✅ **Graceful fallback** - Falls back to in-memory if DB unavailable

## Database Schema

```sql
CREATE TABLE users (
    username VARCHAR(255) PRIMARY KEY,
    password_hash VARCHAR(255) NOT NULL,
    roles JSONB NOT NULL DEFAULT '[]'::jsonb,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP,
    failed_login_attempts INTEGER NOT NULL DEFAULT 0,
    locked_until TIMESTAMP
);

CREATE INDEX idx_users_active ON users(active) WHERE active = true;
CREATE INDEX idx_users_locked ON users(locked_until) WHERE locked_until IS NOT NULL;
```

## Configuration

### Option 1: Environment Variable

```bash
export USER_DATABASE_DSN="postgres://user:password@localhost:5432/sharding_users?sslmode=disable"
```

### Option 2: Config File

Update `configs/manager.json`:

```json
{
  "security": {
    "user_database_dsn": "postgres://user:password@localhost:5432/sharding_users?sslmode=disable"
  }
}
```

## Setup Instructions

### 1. Create PostgreSQL Database

```bash
# Connect to PostgreSQL
psql -U postgres

# Create database
CREATE DATABASE sharding_users;

# Create user (optional)
CREATE USER sharding_user WITH PASSWORD 'secure_password';
GRANT ALL PRIVILEGES ON DATABASE sharding_users TO sharding_user;
```

### 2. Configure DSN

```bash
# Production example
export USER_DATABASE_DSN="postgres://sharding_user:secure_password@db.example.com:5432/sharding_users?sslmode=require"

# Development example
export USER_DATABASE_DSN="postgres://postgres:postgres@localhost:5432/sharding_users?sslmode=disable"
```

### 3. Start Application

The schema will be automatically created on first run. Default users will be inserted if the table is empty.

## Default Users

The system automatically creates these users on first run:

- **admin** / admin123 - Full admin access
- **operator** / operator123 - Operator role
- **viewer** / viewer123 - Viewer role

**⚠️ IMPORTANT**: Change these passwords immediately in production!

## Connection Pool Settings

Following MAANG standards:

- **Max Open Connections**: 25
- **Max Idle Connections**: 5
- **Connection Max Lifetime**: 5 minutes
- **Connection Max Idle Time**: 1 minute

## Account Lockout

- **Failed Attempts**: 5 attempts
- **Lockout Duration**: 15 minutes
- **Automatic Unlock**: After lockout period expires

## Monitoring

### Check User Status

```sql
SELECT username, active, failed_login_attempts, locked_until, last_login_at
FROM users
WHERE username = 'admin';
```

### View Recent Logins

```sql
SELECT username, last_login_at, failed_login_attempts
FROM users
WHERE last_login_at IS NOT NULL
ORDER BY last_login_at DESC
LIMIT 10;
```

### Check Locked Accounts

```sql
SELECT username, locked_until, failed_login_attempts
FROM users
WHERE locked_until IS NOT NULL AND locked_until > NOW();
```

## Production Best Practices

1. **Use SSL/TLS**: Always use `sslmode=require` in production
2. **Separate Database**: Use a dedicated database for user storage
3. **Backup Regularly**: Include user database in backup strategy
4. **Monitor Connections**: Track connection pool usage
5. **Rotate Passwords**: Implement password rotation policy
6. **Audit Logs**: Review failed login attempts regularly

## Troubleshooting

### Database Connection Failed

If the database is unavailable, the system will:
1. Log a warning
2. Fall back to in-memory user store
3. Continue operating (users will be lost on restart)

### Schema Creation Failed

Ensure the database user has CREATE TABLE permissions:

```sql
GRANT CREATE ON DATABASE sharding_users TO sharding_user;
```

### Performance Issues

If experiencing slow queries:

1. Check indexes are created:
   ```sql
   \d users
   ```

2. Analyze query performance:
   ```sql
   EXPLAIN ANALYZE SELECT * FROM users WHERE username = 'admin';
   ```

3. Monitor connection pool:
   ```sql
   SELECT count(*) FROM pg_stat_activity WHERE datname = 'sharding_users';
   ```

## Migration from In-Memory Store

To migrate existing users:

1. Export users from in-memory store (if possible)
2. Insert into database:
   ```sql
   INSERT INTO users (username, password_hash, roles, active)
   VALUES ('admin', '$2a$10$...', '["admin"]'::jsonb, true);
   ```

3. Restart application with `USER_DATABASE_DSN` set

## Security Considerations

- ✅ Passwords are hashed with bcrypt (cost 10)
- ✅ Account lockout prevents brute force attacks
- ✅ Failed attempts are tracked and logged
- ✅ Connection strings are masked in logs
- ✅ Database credentials should be in secrets management

## Example Docker Compose

```yaml
services:
  postgres-users:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: sharding_users
      POSTGRES_USER: sharding_user
      POSTGRES_PASSWORD: ${USER_DB_PASSWORD}
    volumes:
      - user-db-data:/var/lib/postgresql/data
    networks:
      - sharding-network

  manager:
    environment:
      USER_DATABASE_DSN: "postgres://sharding_user:${USER_DB_PASSWORD}@postgres-users:5432/sharding_users?sslmode=disable"
```

