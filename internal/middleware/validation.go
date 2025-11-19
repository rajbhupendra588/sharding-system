package middleware

import (
	"net/http"
	"strings"
)

const (
	// DefaultMaxRequestSize is 10MB
	DefaultMaxRequestSize = 10 * 1024 * 1024
)

// RequestSizeLimit middleware limits the size of request bodies
func RequestSizeLimit(maxSize int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check Content-Length header
			if r.ContentLength > maxSize {
				http.Error(w, `{"error":{"code":"PAYLOAD_TOO_LARGE","message":"Request body too large"}}`, http.StatusRequestEntityTooLarge)
				return
			}
			
			// Limit the request body reader
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)
			
			next.ServeHTTP(w, r)
		})
	}
}

// ContentTypeValidation middleware validates Content-Type header
func ContentTypeValidation(allowedTypes []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip validation for GET, HEAD, OPTIONS, DELETE
			if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" || r.Method == "DELETE" {
				next.ServeHTTP(w, r)
				return
			}
			
			contentType := r.Header.Get("Content-Type")
			allowed := false
			for _, allowedType := range allowedTypes {
				if strings.Contains(contentType, allowedType) {
					allowed = true
					break
				}
			}
			
			if !allowed && len(allowedTypes) > 0 {
				http.Error(w, `{"error":{"code":"UNSUPPORTED_MEDIA_TYPE","message":"Content-Type not allowed"}}`, http.StatusUnsupportedMediaType)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

