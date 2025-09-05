package handler

import (
    "net"
    "net/http"
    "strings"
    "time"

    "github.com/ansh0014/api/pkg"
)

// Handler holds the proxy map and implements HTTP handling for gateway forwarding.
type Handler struct {
    pm *pkg.ProxyMap
}

// New creates a new proxy handler.
func New(pm *pkg.ProxyMap) http.Handler {
    return &Handler{pm: pm}
}

// ServeHTTP routes requests to the appropriate upstream reverse proxy.
// It also sets common proxy headers (X-Forwarded-For, X-Real-IP, X-Request-ID) if missing.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    p := h.pm.Route(r)
    if p == nil {
        http.NotFound(w, r)
        return
    }

    // X-Real-IP / X-Forwarded-For
    if ip := realIP(r); ip != "" {
        existing := r.Header.Get("X-Forwarded-For")
        if existing == "" {
            r.Header.Set("X-Forwarded-For", ip)
        } else if !strings.Contains(existing, ip) {
            r.Header.Set("X-Forwarded-For", existing+", "+ip)
        }
        if r.Header.Get("X-Real-IP") == "" {
            r.Header.Set("X-Real-IP", ip)
        }
    }

    // X-Request-ID
    if r.Header.Get("X-Request-ID") == "" {
        r.Header.Set("X-Request-ID", generateReqID())
    }

    // Forward to matched reverse proxy
    p.ServeHTTP(w, r)
}

func realIP(r *http.Request) string {
    // try common headers
    if h := r.Header.Get("X-Real-IP"); h != "" {
        return h
    }
    if h := r.Header.Get("X-Forwarded-For"); h != "" {
        // may be comma separated list, take first
        parts := strings.Split(h, ",")
        return strings.TrimSpace(parts[0])
    }
    // fallback to remote address
    if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
        return ip
    }
    return ""
}

func generateReqID() string {
    return time.Now().UTC().Format("20060102T150405.000000000")
}