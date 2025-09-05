package middleware

import (
    "net"
    "net/http"
    "sync"
    "time"

    "golang.org/x/time/rate"
)

// Basic per-IP rate limiter.
var visitors = make(map[string]*rate.Limiter)
var mu sync.Mutex

func getLimiter(ip string) *rate.Limiter {
    mu.Lock()
    defer mu.Unlock()
    limiter, exists := visitors[ip]
    if !exists {
        limiter = rate.NewLimiter(5, 10) // 5 req/sec, burst 10
        visitors[ip] = limiter
        // optional: cleanup goroutine
        go func() {
            time.Sleep(10 * time.Minute)
            mu.Lock()
            delete(visitors, ip)
            mu.Unlock()
        }()
    }
    return limiter
}

func RateLimit(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip, _, _ := net.SplitHostPort(r.RemoteAddr)
        limiter := getLimiter(ip)
        if !limiter.Allow() {
            http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}