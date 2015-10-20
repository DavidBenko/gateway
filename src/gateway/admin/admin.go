package admin

import (
	"net/http"
	"strings"

	"gateway/config"
	aphttp "gateway/http"
	sql "gateway/sql"

	"github.com/gorilla/mux"
)

// Setup sets up the session and adds admin routes.
func Setup(router *mux.Router, db *sql.DB, configuration config.Configuration) {
	conf, psconf := configuration.Admin, configuration.Proxy
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
		RouteResource(&UsersController{BaseController{accountID: accountIDFromPath, userID: userIDDummy}},
			"/accounts/{accountID}/users", siteAdmin, db, conf)

		// sessions are unprotected to allow users to authenticate
		RouteSessions("/sessions", admin, db, conf)
	}

	// protected by requiring login (except dev mode)
	accountID := accountIDFromSession
	userID := userIDFromSession
	authAdmin := NewSessionAuthRouter(admin, []string{"OPTIONS"})
	if conf.DevMode {
		accountID = accountIDForDevMode(db)
		userID = userIDForDevMode(db)
		authAdmin = admin
	}

	base := BaseController{conf: conf, accountID: accountID, userID: userID}

	RouteNotify(&NotifyController{BaseController: base}, "/notifications", authAdmin, db)

	aggregator = newAggregator(conf)
	RouteLogging("/logs/socket", authAdmin)
	RouteLogging("/apis/{apiID}/logs/socket", authAdmin)
	RouteLogging("/apis/{apiID}/proxy_endpoints/{endpointID}/logs/socket", authAdmin)

	search := &LogSearchController{configuration.Elastic, base}
	RouteLogSearch(search, "/logs", authAdmin, db, conf)
	RouteLogSearch(search, "/apis/{apiID}/logs", authAdmin, db, conf)
	RouteLogSearch(search, "/apis/{apiID}/proxy_endpoints/{endpointID}/logs", authAdmin, db, conf)

	RouteResource(&UsersController{base}, "/users", authAdmin, db, conf)

	apisController := &APIsController{base}
	RouteAPIExport(apisController, "/apis/{id}/export", authAdmin, db, conf)
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
	adminStaticFileHandler := http.HandlerFunc(adminStaticFileHandler(conf))
	admin.Handle("/{path:.*}", adminStaticFileHandler)

	// also add a route to the base router so that if the user leaves off the trailing slash on the admin
	// path, the adminStaticFileHandler still serves the request.
	adminPath := strings.TrimRight(conf.PathPrefix, "/")
	router.Handle(adminPath, adminStaticFileHandler)
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
