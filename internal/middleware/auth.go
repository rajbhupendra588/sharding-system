package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/sharding-system/pkg/security"
)

// AuthMiddleware creates authentication middleware
func AuthMiddleware(authManager *security.AuthManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for public endpoints
			publicPaths := []string{
				"/health",
				"/v1/health",
				"/api/v1/health",
				"/metrics",
				"/api/v1/auth/login",
				"/swagger/",
			}

			path := r.URL.Path
			for _, publicPath := range publicPaths {
				// Handle paths that already end with /
				if strings.HasSuffix(publicPath, "/") {
					if strings.HasPrefix(path, publicPath) {
						next.ServeHTTP(w, r)
						return
					}
				} else if path == publicPath || strings.HasPrefix(path, publicPath+"/") {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"Missing authorization header"}}`, http.StatusUnauthorized)
				return
			}

			// Check Bearer token format
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"Invalid authorization header format"}}`, http.StatusUnauthorized)
				return
			}

			token := parts[1]

			// Validate token
			claims, err := authManager.ValidateToken(token)
			if err != nil {
				http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"Invalid or expired token"}}`, http.StatusUnauthorized)
				return
			}

			// Add claims to request context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "username", claims.Username)
			ctx = context.WithValue(ctx, "roles", claims.Roles)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
