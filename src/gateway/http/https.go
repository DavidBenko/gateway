package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

type forceSSLRouter struct {
	router          Router
	protoHeaderName string
}

// NewForceSSLRouter creates a new Router that wraps an existing Router and
// issues a redirect to the equivalent HTTPS URL, if any requests have been
// received over plain HTTP
func NewForceSSLRouter(router Router, protoHeaderName string) Router {
	return &forceSSLRouter{router: router, protoHeaderName: protoHeaderName}
}

// Handle wraps the handler in a forceSSLHandler
func (f *forceSSLRouter) Handle(pattern string, handler http.Handler) *mux.Route {
	return f.router.Handle(pattern, forceSSLHandler(handler, f.protoHeaderName))
}

func forceSSLHandler(handler http.Handler, protoHeaderName string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if protoHeaderName == "" && r.TLS == nil {
			http.Redirect(w, r, httpsURL(r), http.StatusMovedPermanently)
			return
		}

		if len(r.Header[protoHeaderName]) == 0 || r.Header[protoHeaderName][0] != "https" {
			http.Redirect(w, r, httpsURL(r), http.StatusMovedPermanently)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

func httpsURL(req *http.Request) string {
	u := req.URL
	u.Scheme = "https"
	if req.Host != "" {
		u.Host = req.Host
	}
	return u.String()
}
