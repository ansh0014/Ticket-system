package middleware

import (
    "context"
    "net/http"
    "os"
    "strings"

    "github.com/ansh0014/api/internal"
)	

type ctxKey string

const userKey ctxKey = "user"

// JWTExtract extracts sub from JWT and injects X-User-ID header for upstreams.
// It does NOT block requests — upstream services decide on auth enforcement.
func JWTExtract(next http.Handler) http.Handler {
    secret := os.Getenv("JWT_SECRET")
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        auth := r.Header.Get("Authorization")
        if auth != "" && strings.HasPrefix(auth, "Bearer ") && secret != "" {
            tokenString := strings.TrimPrefix(auth, "Bearer ")
            claims, err := internal.ParseToken(tokenString, secret)
            if err == nil {
                if sub, ok := claims["sub"].(string); ok && sub != "" {
                    r = r.WithContext(context.WithValue(r.Context(), userKey, sub))
                    r.Header.Set("X-User-ID", sub)
                }
            }
            _ = err
        }
        next.ServeHTTP(w, r)
    })
}

// RequireAuth enforces presence of X-User-ID (use for internal routes if needed)
func RequireAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Header.Get("X-User-ID") == "" {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}