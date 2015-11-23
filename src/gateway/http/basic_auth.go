package http

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// NewHTTPBasicRouter wraps the router.
func NewHTTPBasicRouter(username, password, realm string, router Router) Router {
	if password == "" {
		return router
	}
	return &BasicAuthRouter{username, password, realm, router}
}

// BasicAuthRouter wraps all Handle calls in an HTTP Basic check.
type BasicAuthRouter struct {
	username string
	password string
	realm    string
	router   Router
}

// Handle wraps the handler in an AccessLoggingHandler for the router.
func (b *BasicAuthRouter) Handle(pattern string, handler http.Handler) *mux.Route {
	return b.router.Handle(pattern, b.Wrap(handler))
}

// Wrap provides the wrapped handling functionality.
func (b *BasicAuthRouter) Wrap(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if b.checkAuth(w, r) {
			handler.ServeHTTP(w, r)
			return
		}

		w.Header().Set("WWW-Authenticate",
			fmt.Sprintf(`Basic realm="%s"`, b.realm))
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 Unauthorized\n"))
	})
}

func (b *BasicAuthRouter) checkAuth(w http.ResponseWriter, r *http.Request) bool {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 {
		return false
	}

	bytes, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return false
	}

	pair := strings.SplitN(string(bytes), ":", 2)
	if len(pair) != 2 {
		return false
	}

	return (pair[0] == b.username) && (pair[1] == b.password)
}
