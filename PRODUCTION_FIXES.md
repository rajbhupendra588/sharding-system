# Production Security Fixes

This document provides the fixes needed to make the system production-ready.

## Fix 1: Password Hashing

### Add bcrypt dependency
```bash
go get golang.org/x/crypto/bcrypt
```

### Update auth_handler.go
```go
import "golang.org/x/crypto/bcrypt"

// User represents a user in the system
type User struct {
    Username     string
    PasswordHash string
    Roles        []string
}

// In production, load from database
var users = map[string]*User{
    "admin": {
        Username:     "admin",
        PasswordHash: "$2a$10$...", // bcrypt hash of password
        Roles:        []string{"admin"},
    },
    // ... other users
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    // ... existing code ...
    
    user, exists := users[req.Username]
    if !exists {
        http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"Invalid credentials"}}`, http.StatusUnauthorized)
        return
    }
    
    // Verify password
    err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
    if err != nil {
        http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"Invalid credentials"}}`, http.StatusUnauthorized)
        return
    }
    
    // Generate token with user roles
    token, err := h.authManager.GenerateToken(user.Username, user.Roles)
    // ... rest of code ...
}
```

## Fix 2: Require JWT_SECRET

### Update manager.go
```go
jwtSecret := os.Getenv("JWT_SECRET")
if jwtSecret == "" {
    logger.Fatal("JWT_SECRET environment variable is required")
}
if len(jwtSecret) < 32 {
    logger.Fatal("JWT_SECRET must be at least 32 characters")
}
```

## Fix 3: Enable Auth Middleware

### Update manager.go
```go
// Enable auth if RBAC is enabled
if cfg.Security.EnableRBAC {
    muxRouter.Use(middleware.AuthMiddleware(authManager))
}
```

## Fix 4: Rate Limiting

### Create rate_limit.go
```go
package middleware

import (
    "net/http"
    "sync"
    "time"
)

type rateLimiter struct {
    requests map[string][]time.Time
    mu       sync.Mutex
    limit    int
    window   time.Duration
}

func RateLimitMiddleware(limit int, window time.Duration) func(http.Handler) http.Handler {
    rl := &rateLimiter{
        requests: make(map[string][]time.Time),
        limit:    limit,
        window:   window,
    }
    
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ip := r.RemoteAddr
            rl.mu.Lock()
            defer rl.mu.Unlock()
            
            now := time.Now()
            // Clean old requests
            validRequests := []time.Time{}
            for _, t := range rl.requests[ip] {
                if now.Sub(t) < rl.window {
                    validRequests = append(validRequests, t)
                }
            }
            
            if len(validRequests) >= rl.limit {
                http.Error(w, `{"error":{"code":"RATE_LIMIT","message":"Too many requests"}}`, http.StatusTooManyRequests)
                return
            }
            
            validRequests = append(validRequests, now)
            rl.requests[ip] = validRequests
            
            next.ServeHTTP(w, r)
        })
    }
}
```

## Fix 5: Restrict CORS

### Update cors.go
```go
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")
            
            // Check if origin is allowed
            allowed := false
            for _, allowedOrigin := range allowedOrigins {
                if origin == allowedOrigin {
                    allowed = true
                    break
                }
            }
            
            if allowed {
                w.Header().Set("Access-Control-Allow-Origin", origin)
            } else if len(allowedOrigins) == 0 {
                // Development mode - allow all
                w.Header().Set("Access-Control-Allow-Origin", "*")
            }
            
            // ... rest of CORS headers ...
        })
    }
}
```

## Summary

**Critical fixes needed**:
1. ✅ Password hashing (bcrypt)
2. ✅ Require JWT_SECRET
3. ✅ Enable auth middleware
4. ✅ Rate limiting
5. ✅ CORS restrictions

**Estimated implementation time**: 2-3 hours

