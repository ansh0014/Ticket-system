package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ansh0014/booking/utils"
)

// AuthMiddleware extracts the user ID from the Authorization header
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// Skip auth for public endpoints
			if isPublicEndpoint(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			utils.UnauthorizedResponse(w, "Authorization header is required")
			return
		}

		// Check if the header has the Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.UnauthorizedResponse(w, "Invalid authorization format")
			return
		}

		token := parts[1]
		// In a real system, you would validate this token
		// For now, we'll assume the token is the user ID
		userID := token

		// Add user ID to context
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ServiceInjector injects services into the request context
func ServiceInjector(services map[string]interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add services to context
			ctx := r.Context()
			for name, service := range services {
				ctx = context.WithValue(ctx, name+"Service", service)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RecoverMiddleware recovers from panics
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				utils.ServerErrorResponse(w, "Internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware logs requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log request
		// In a real system, you would use a proper logger
		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware handles CORS
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isPublicEndpoint checks if the endpoint is public
func isPublicEndpoint(path string) bool {
	publicEndpoints := []string{
		"/health",
		"/api/platforms/flight/search",
		"/api/platforms/railway/search",
		"/api/platforms/event/search",
		"/api/platforms/movie/search",
		"/api/platforms/movie",
	}

	for _, endpoint := range publicEndpoints {
		if strings.HasPrefix(path, endpoint) {
			return true
		}
	}

	return false
}
