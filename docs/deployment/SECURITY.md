# Security Guide

This document outlines the security features and best practices for the Sharding System.

## Authentication

### Password Security

- **Hashing**: Passwords are hashed using bcrypt with cost factor 10
- **Storage**: Password hashes are stored securely (never plain text)
- **Verification**: Secure comparison using bcrypt

### Account Protection

- **Lockout**: Accounts are locked after 5 failed login attempts
- **Lockout Duration**: 15 minutes
- **Automatic Unlock**: Account unlocks after lockout period expires

### JWT Tokens

- **Algorithm**: HS256 (HMAC-SHA256)
- **Expiration**: 24 hours
- **Secret**: Must be 32+ characters (set via `JWT_SECRET` env var)
- **Validation**: Tokens are validated on every request

## Authorization

### Role-Based Access Control (RBAC)

**Roles**:
- **admin**: Full access to all operations
- **operator**: Read, create, update operations
- **viewer**: Read-only access

**Permissions**:
- Shard CRUD: admin only
- Resharding: admin only
- Replica promotion: operator+
- Query execution: all authenticated users
- Metrics: viewer+

## Input Validation

### Request Size Limits

- **Maximum Size**: 10MB per request
- **Enforcement**: Automatic rejection of oversized requests
- **Error**: Returns 413 Payload Too Large

### Content-Type Validation

- **Required**: `application/json` for POST/PUT/PATCH requests
- **Enforcement**: Automatic validation on all mutation endpoints
- **Error**: Returns 415 Unsupported Media Type

## CORS Configuration

### Development Mode

```bash
export CORS_ALLOWED_ORIGINS="*"
```

Allows all origins (for development only).

### Production Mode

```bash
export CORS_ALLOWED_ORIGINS="https://app.example.com,https://admin.example.com"
```

Comma-separated list of allowed origins.

## Environment Variables

### Required for Production

```bash
# JWT Secret (32+ characters)
JWT_SECRET="your-super-secret-jwt-key-minimum-32-characters"

# CORS Allowed Origins
CORS_ALLOWED_ORIGINS="https://yourdomain.com"
```

### Optional

```bash
# Configuration path
CONFIG_PATH="configs/manager.json"

# Log level
LOG_LEVEL="info"
```

## Security Best Practices

### 1. Secrets Management

- ✅ Never commit secrets to version control
- ✅ Use environment variables or secrets management (Vault, AWS Secrets Manager)
- ✅ Rotate secrets regularly
- ✅ Use different secrets for each environment

### 2. Network Security

- ✅ Enable TLS in production (`enable_tls: true`)
- ✅ Use private networks for internal communication
- ✅ Implement firewall rules
- ✅ Use VPN for administrative access

### 3. Monitoring

- ✅ Monitor failed login attempts
- ✅ Alert on suspicious activity
- ✅ Log all authentication events
- ✅ Track account lockouts

### 4. User Management

- ✅ Change default passwords immediately
- ✅ Implement password complexity requirements
- ✅ Enable password expiration policies
- ✅ Implement user management API (future enhancement)

## Security Checklist

### Before Production Deployment

- [ ] Set strong `JWT_SECRET` (32+ characters)
- [ ] Restrict `CORS_ALLOWED_ORIGINS` to production domains
- [ ] Enable TLS (`enable_tls: true`)
- [ ] Enable RBAC (`enable_rbac: true`)
- [ ] Change default user passwords
- [ ] Set up monitoring and alerting
- [ ] Configure firewall rules
- [ ] Review audit logs regularly
- [ ] Set up backup and recovery
- [ ] Test authentication flow end-to-end

## Incident Response

### If JWT Secret is Compromised

1. Immediately rotate `JWT_SECRET`
2. Invalidate all existing tokens
3. Force all users to re-authenticate
4. Review audit logs for suspicious activity

### If Account is Locked

1. Wait for 15-minute lockout period
2. Or reset account lockout counter (requires admin access)
3. Review failed login attempts in logs

## Compliance

### Security Standards

- ✅ Password hashing: OWASP compliant (bcrypt)
- ✅ JWT tokens: RFC 7519 compliant
- ✅ CORS: W3C CORS specification compliant
- ✅ Input validation: OWASP Top 10 compliant

## Additional Resources

- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [JWT Best Practices](https://datatracker.ietf.org/doc/html/rfc8725)
- [CORS Security](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)

