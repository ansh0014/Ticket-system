package routes

import (
    "net/http"

    "github.com/ansh0014/api/pkg"

    "github.com/gorilla/mux"
)

// NewRouter returns a router that forwards matching paths to the proxy map.
func NewRouter(pm *pkg.ProxyMap) http.Handler {
    r := mux.NewRouter()

    // health
    r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("ok"))
    }).Methods("GET")

    // catch-all: forward to upstream based on prefix
    r.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        p := pm.Route(r)
        if p == nil {
            http.NotFound(w, r)
            return
        }
        p.ServeHTTP(w, r)
    }))

    return r
}