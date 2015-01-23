package http

import "net/http"

// AccessLoggingHandler logs general access notes about a request, plus
// sets up an ID in the context for other methods to use for logging.
func CORSAwareHandler(allow string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allow)
		handler.ServeHTTP(w, r)
	})
}

// CORSAwareRouter wraps all Handle calls in an CORSAwareHandler.
type CORSAwareRouter struct {
	allow  string
	router Router
}

// Handle wraps the handler in an CORSAwareHandler for the router.
func (r *CORSAwareRouter) Handle(pattern string, handler http.Handler) {
	r.router.Handle(pattern, CORSAwareHandler(r.allow, handler))
}

// NewCORSAwareRouter wraps the router.
func NewCORSAwareRouter(allow string, router Router) *CORSAwareRouter {
	return &CORSAwareRouter{allow: allow, router: router}
}
