package admin

import (
	"net/http"

	"gateway/config"
	aphttp "gateway/http"
	sql "gateway/sql"

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

// Setup sets up the session and adds admin routes.
func Setup(router *mux.Router, db *sql.DB, conf config.ProxyAdmin) {
	setupSessions(conf)

	var admin aphttp.Router
	admin = aphttp.NewAccessLoggingRouter(config.Admin, subrouter(router, conf))

	if conf.CORSEnabled {
		admin = aphttp.NewCORSAwareRouter(conf.CORSOrigin, admin)
	}

	// siteAdmin is additionally protected for the site owner
	siteAdmin := aphttp.NewHTTPBasicRouter(conf.Username, conf.Password, conf.Realm, admin)
	RouteResource(&AccountsController{}, "/accounts", siteAdmin, db, conf)
	RouteResource(&UsersController{BaseController{}, accountIDFromPath}, "/accounts/{accountID}/users", siteAdmin, db, conf)

	// sessions are unprotected to allow users to authenticate
	RouteSessions("/sessions", admin, db, conf)

	// protected by requiring login
	authAdmin := NewSessionAuthRouter(admin)
	RouteResource(&UsersController{BaseController{}, accountIDFromSession}, "/users", authAdmin, db, conf)
	RouteResource(&APIsController{}, "/apis", authAdmin, db, conf)
	RouteResource(&HostsController{}, "/apis/{apiID}/hosts", authAdmin, db, conf)
	RouteResource(&EnvironmentsController{}, "/apis/{apiID}/environments", authAdmin, db, conf)
	RouteResource(&EndpointGroupsController{}, "/apis/{apiID}/endpoint_groups", authAdmin, db, conf)
	RouteResource(&RemoteEndpointsController{}, "/apis/{apiID}/remote_endpoints", authAdmin, db, conf)
	RouteResource(&ProxyEndpointsController{}, "/apis/{apiID}/proxy_endpoints", authAdmin, db, conf)

	// static assets for self-hosted systems
	admin.Handle("/{path:.*}", http.HandlerFunc(adminStaticFileHandler))
}
