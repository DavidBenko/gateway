package admin

import (
	"net/http"

	"gateway/config"
	aphttp "gateway/http"
	sql "gateway/sql"

	"github.com/gorilla/mux"
)

// Setup sets up the session and adds admin routes.
func Setup(router *mux.Router, db *sql.DB, conf config.ProxyAdmin, psconf config.ProxyServer) {
	var admin aphttp.Router
	admin = aphttp.NewAccessLoggingRouter(config.Admin, conf.RequestIDHeader,
		subrouter(router, conf))

	if conf.CORSEnabled {
		admin = aphttp.NewCORSAwareRouter(conf.CORSOrigin, admin)
	}

	if !conf.DevMode {
		setupSessions(conf)

		// siteAdmin is additionally protected for the site owner
		siteAdmin := aphttp.NewHTTPBasicRouter(conf.Username, conf.Password, conf.Realm, admin)
		RouteResource(&AccountsController{}, "/accounts", siteAdmin, db, conf)
		RouteResource(&UsersController{BaseController{accountID: accountIDFromPath}}, "/accounts/{accountID}/users", siteAdmin, db, conf)

		// sessions are unprotected to allow users to authenticate
		RouteSessions("/sessions", admin, db, conf)
	}

	// protected by requiring login (except dev mode)
	accountID := accountIDFromSession
	authAdmin := NewSessionAuthRouter(admin, []string{"OPTIONS"})
	if conf.DevMode {
		accountID = accountIDForDevMode(db)
		authAdmin = admin
	}

	base := BaseController{conf: conf, accountID: accountID}

	RouteNotify("/notify", authAdmin, db)

	RouteResource(&UsersController{base}, "/users", authAdmin, db, conf)

	apisController := &APIsController{base}
	RouteAPIExport(apisController, "/apis/{id}/export", authAdmin, db, conf)
	RouteAPIImport(apisController, "/apis/import", authAdmin, db, conf)
	RouteResource(apisController, "/apis", authAdmin, db, conf)

	testController := &TestController{base, psconf}
	RouteTest(testController, "/apis/{apiID}/proxy_endpoints/{endpointID}/tests/{testID}/test", authAdmin, db, conf)

	RouteResource(&HostsController{base}, "/apis/{apiID}/hosts", authAdmin, db, conf)
	RouteResource(&EnvironmentsController{base}, "/apis/{apiID}/environments", authAdmin, db, conf)
	RouteResource(&LibrariesController{base}, "/apis/{apiID}/libraries", authAdmin, db, conf)
	RouteResource(&EndpointGroupsController{base}, "/apis/{apiID}/endpoint_groups", authAdmin, db, conf)
	RouteResource(&RemoteEndpointsController{base}, "/apis/{apiID}/remote_endpoints", authAdmin, db, conf)
	RouteResource(&ProxyEndpointsController{base}, "/apis/{apiID}/proxy_endpoints", authAdmin, db, conf)

	// static assets for self-hosted systems
	admin.Handle("/{path:.*}", http.HandlerFunc(adminStaticFileHandler(conf)))
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
