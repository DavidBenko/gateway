package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Router defines the minimal interface of *mux.Router necessary for routing.
type Router interface {
	Handle(pattern string, handler http.Handler) *mux.Route
}
