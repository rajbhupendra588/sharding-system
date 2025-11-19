package middleware

import (
	"net/http"
	"os"
	"strings"
	"sync"
)

var (
	// CORS configuration cache (MAANG standard: avoid repeated env reads)
	corsConfig     *corsConfigCache
	corsConfigOnce sync.Once
)

type corsConfigCache struct {
	allowedOrigins []string
	wildcard       bool
	mu             sync.RWMutex
}

// getCORSConfig returns cached CORS configuration
func getCORSConfig() *corsConfigCache {
	corsConfigOnce.Do(func() {
		allowedOriginsEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
		if allowedOriginsEnv == "" {
			// Default: allow all in development
			allowedOriginsEnv = "*"
		}

		config := &corsConfigCache{}
		if allowedOriginsEnv == "*" {
			config.wildcard = true
		} else {
			// Parse comma-separated origins
			origins := strings.Split(allowedOriginsEnv, ",")
			config.allowedOrigins = make([]string, 0, len(origins))
			for _, orig := range origins {
				trimmed := strings.TrimSpace(orig)
				if trimmed != "" {
					config.allowedOrigins = append(config.allowedOrigins, trimmed)
				}
			}
		}
		corsConfig = config
	})
	return corsConfig
}

// isOriginAllowed checks if an origin is allowed (MAANG standard: strict validation)
func (c *corsConfigCache) isOriginAllowed(origin string) (bool, string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if origin == "" {
		// Same-origin request - allow if wildcard or empty origin list means same-origin only
		if c.wildcard {
			return true, "*"
		}
		return false, ""
	}

	if c.wildcard {
		// Development mode - echo back origin
		return true, origin
	}

	// Production mode - strict whitelist matching
	for _, allowed := range c.allowedOrigins {
		if allowed == origin {
			return true, origin
		}
		// Support subdomain matching (e.g., *.example.com)
		if strings.HasPrefix(allowed, "*.") {
			domain := strings.TrimPrefix(allowed, "*.")
			if strings.HasSuffix(origin, "."+domain) || origin == domain {
				return true, origin
			}
		}
	}

	// Origin not in whitelist - reject
	return false, ""
}

// CORS middleware to handle Cross-Origin Resource Sharing (MAANG production standard)
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		config := getCORSConfig()

		// Check if origin is allowed
		allowed, allowedOrigin := config.isOriginAllowed(origin)
		
		if allowed && allowedOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			
			// Set credentials header only if not using wildcard (MAANG standard)
			if allowedOrigin != "*" {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		} else if !allowed && origin != "" {
			// Explicitly reject unauthorized origin (MAANG standard: fail secure)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, X-CSRF-Token")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type, X-Request-ID")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours (MAANG standard)

		// Handle preflight OPTIONS requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

