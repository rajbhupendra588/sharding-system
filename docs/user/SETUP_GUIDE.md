# Initial Setup Guide

## Admin Credentials Setup

When you first start the Sharding System, you need to set up your admin credentials. The system does **not** create default users - you must configure your own admin account.

## Setup Process

### Step 1: Start the System

Start the Manager service (which handles authentication):

```bash
# Using Docker Compose
docker-compose up -d manager

# Or using Kubernetes
kubectl apply -f k8s/manager-deployment.yaml
```

### Step 2: Check Setup Status

Verify that the system requires setup (no users exist):

```bash
curl http://localhost:8081/api/v1/auth/setup
```

If setup is required, you'll get a response indicating the system needs initialization.

### Step 3: Create Admin Account

Create your first admin account using the setup endpoint:

```bash
curl -X POST http://localhost:8081/api/v1/auth/setup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "myadmin",
    "password": "SecurePassword123!"
  }'
```

**Response:**
```json
{
  "message": "System setup completed successfully",
  "username": "myadmin",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Important Notes:**
- Username must not be empty
- Password must be at least 8 characters
- This endpoint only works when no users exist
- The first user created will have admin role
- Maximum of 2 admin users allowed in the system

### Step 4: Login

Use the token from the setup response, or login normally:

```bash
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "myadmin",
    "password": "SecurePassword123!"
  }'
```

## Admin User Limits

### Maximum 2 Admins

The system enforces a **maximum of 2 admin users** at any time. This ensures:

- Security: Limits administrative access points
- Accountability: Clear responsibility boundaries
- Compliance: Meets security best practices

### Creating Additional Admins

To create a second admin user, you must:

1. Be logged in as an existing admin
2. Use the user management API (if implemented) or directly insert into the database
3. The system will reject attempts to create a third admin

**Example Error:**
```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "maximum of 2 admin users allowed (current: 2)"
  }
}
```

## Security Best Practices

1. **Strong Passwords**: Use passwords with:
   - Minimum 8 characters
   - Mix of uppercase, lowercase, numbers, and symbols
   - Not based on dictionary words

2. **Secure Storage**: Store admin credentials securely:
   - Use environment variables or secrets management
   - Never commit credentials to version control
   - Rotate passwords regularly

3. **Access Control**: 
   - Limit admin access to necessary personnel only
   - Use the 2-admin limit to maintain accountability
   - Monitor admin actions through audit logs

4. **Token Management**:
   - Tokens expire after 24 hours
   - Store tokens securely
   - Rotate tokens regularly

## Troubleshooting

### Setup Endpoint Returns "Already Initialized"

If you see this error, it means users already exist. You can:

1. Check existing users (requires admin access)
2. Reset the database (development only):
   ```sql
   DELETE FROM users;
   ```

### Cannot Create Second Admin

If you're trying to create a second admin and getting an error:

1. Check current admin count:
   ```bash
   # Requires admin authentication
   curl -H "Authorization: Bearer YOUR_TOKEN" \
     http://localhost:8081/api/v1/users/admins/count
   ```

2. If you have 2 admins, you must:
   - Remove an existing admin first, OR
   - Update an existing admin's role

### Database Connection Issues

If setup fails due to database connection:

1. Verify `USER_DATABASE_DSN` environment variable is set
2. Check database is accessible
3. Verify database schema is initialized

## Environment Variables

Configure the user database connection:

```bash
export USER_DATABASE_DSN="postgres://user:password@localhost:5432/sharding_users?sslmode=disable"
```

Or set in `configs/manager.json`:

```json
{
  "security": {
    "user_database_dsn": "postgres://user:password@localhost:5432/sharding_users?sslmode=disable"
  }
}
```

## Next Steps

After setting up admin credentials:

1. **Configure Shards**: Create your first shard
2. **Start Router**: Start the router service
3. **Test Connection**: Verify client applications can connect
4. **Monitor**: Set up monitoring and alerting

See the [User Guide](USER_GUIDE.md) for more information.

