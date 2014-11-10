package http

import "net/http"

// Router defines the minimal interface of *mux.Router necessary for routing.
type Router interface {
	Handle(pattern string, handler http.Handler)
}
