package admin

import (
	"github.com/AnyPresence/gateway/config"
	"github.com/AnyPresence/gateway/db"
	"github.com/AnyPresence/gateway/rest"
	"github.com/gorilla/mux"
)

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

// AddRoutes adds the admin routes to the specified router.
func AddRoutes(router *mux.Router, db db.DB, config config.ProxyAdmin) {
	admin := subrouter(router, config)

	(&rest.HTTPResource{Resource: &proxyEndpoint{db: db}}).Route(admin)

	admin.HandleFunc("/", adminHandler)
}
