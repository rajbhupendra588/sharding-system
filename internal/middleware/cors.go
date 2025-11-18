package middleware

import "net/http"

// CORS middleware to handle Cross-Origin Resource Sharing
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers before any other headers
		origin := r.Header.Get("Origin")
		
		// For development, allow all origins
		// In production, you should validate the origin against a whitelist
		if origin == "" {
			// If no origin header (same-origin request), allow all
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			// Echo back the origin for CORS requests
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight OPTIONS requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

