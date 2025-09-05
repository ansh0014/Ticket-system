package pkg

import (
    "net/http"
    "net/http/httputil"
    "net/url"
    "sort"
    "strings"
)

// ProxyMap routes prefixes to reverse proxies
type ProxyMap struct {
    prefixes []string
    proxies  map[string]*httputil.ReverseProxy
}

// NewProxyMap creates reverse proxies for each prefix -> target
func NewProxyMap(m map[string]string) *ProxyMap {
    pm := &ProxyMap{
        proxies: map[string]*httputil.ReverseProxy{},
    }
    for prefix, target := range m {
        u, err := url.Parse(target)
        if err != nil {
            continue
        }
        rp := httputil.NewSingleHostReverseProxy(u)
        orig := rp.Director
        p := prefix
        baseHost := u.Host
        rp.Director = func(req *http.Request) {
            orig(req)
            // strip prefix so upstream receives path without "/booking" etc.
            req.URL.Path = strings.TrimPrefix(req.URL.Path, p)
            if req.URL.Path == "" {
                req.URL.Path = "/"
            }
            // set Host to upstream host
            req.Host = baseHost
        }
        pm.proxies[prefix] = rp
        pm.prefixes = append(pm.prefixes, prefix)
    }
    // sort prefixes by length desc so longest match wins
    sort.Slice(pm.prefixes, func(i, j int) bool {
        return len(pm.prefixes[i]) > len(pm.prefixes[j])
    })
    return pm
}

// Route finds best matching proxy for request path
func (pm *ProxyMap) Route(r *http.Request) *httputil.ReverseProxy {
    path := r.URL.Path
    for _, prefix := range pm.prefixes {
        if strings.HasPrefix(path, prefix) {
            return pm.proxies[prefix]
        }
    }
    return nil
}