package http

import (
	"net/http"
	"strings"
)

// CORSHeaderHandler sets the core CORS headers before delegating to the handler.
func CORSHeaderHandler(allow string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allow)
		w.Header().Set("Access-Control-Request-Headers", "*")
		w.Header().Set("Access-Control-Max-Age", "600")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		handler.ServeHTTP(w, r)
	})
}

// CORSOptionsHandler sets only the CORS allow methods header.
func CORSOptionsHandler(methods []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ", "))
		w.WriteHeader(http.StatusOK)
	})
}

// CORSAwareRouter wraps all Handle calls in an CORSAwareHandler.
type CORSAwareRouter struct {
	allow  string
	router Router
}

// Handle wraps the handler in an CORSHeaderHandler for the router.
func (r *CORSAwareRouter) Handle(pattern string, handler http.Handler) {
	r.router.Handle(pattern, CORSHeaderHandler(r.allow, handler))
}

// NewCORSAwareRouter wraps the router.
func NewCORSAwareRouter(allow string, router Router) *CORSAwareRouter {
	return &CORSAwareRouter{allow: allow, router: router}
}
