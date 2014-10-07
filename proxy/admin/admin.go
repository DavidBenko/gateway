package admin

import (
	"fmt"
	"net/http"

	"github.com/AnyPresence/gateway/config"
	"github.com/gorilla/mux"
)

// AddRoutes adds the admin routes to the specified router.
func AddRoutes(router *mux.Router, config config.ProxyAdmin) {
	admin := subrouter(router, config)
	admin.HandleFunc("/", adminHandler)
}

func subrouter(router *mux.Router, config config.ProxyAdmin) *mux.Router {
	adminRoute := router.NewRoute()
	if config.Host != "" {
		adminRoute = adminRoute.Host(config.Host)
	}
	if config.PathPrefix != "" {
		adminRoute = adminRoute.PathPrefix(config.PathPrefix)
	}
	return adminRoute.Subrouter()
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "This is an admin page.\n")
}
