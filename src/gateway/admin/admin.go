package admin

import (
	"fmt"
	"log"
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
		RouteResource(&UsersController{BaseController{accountID: accountIDFromPath, userID: userIDDummy,
			auth: aphttp.AuthTypeSite}}, "/accounts/{accountID}/users", siteAdmin, db, conf)

		// sessions are unprotected to allow users to authenticate
		RouteSessions("/sessions", admin, db, conf)
	}

	// protected by requiring login (except dev mode)
	accountID := accountIDFromSession
	userID := userIDFromSession
	authAdmin := NewSessionAuthRouter(admin, []string{"OPTIONS"}, false)
	authAdminUser := NewSessionAuthRouter(admin, []string{"OPTIONS"}, true)
	if conf.DevMode {
		accountID = accountIDForDevMode(db)
		userID = userIDForDevMode(db)
		authAdmin = admin
		authAdminUser = admin
	}

	base := BaseController{conf: conf, accountID: accountID, userID: userID}

	RouteNotify(&NotifyController{BaseController: base}, "/notifications", authAdmin, db)

	if conf.EnableBroker {
		broker, err := newAggregator(conf)
		if err != nil {
			log.Fatal(err)
		}
		stream := &LogStreamController{base, broker}
		RouteLogStream(stream, "/logs/socket", authAdmin)
		RouteLogStream(stream, "/apis/{apiID}/logs/socket", authAdmin)
		RouteLogStream(stream, "/apis/{apiID}/proxy_endpoints/{endpointID}/logs/socket", authAdmin)
	}

	search := &LogSearchController{configuration.Elastic, base}
	RouteLogSearch(search, "/logs", authAdmin, db, conf)
	RouteLogSearch(search, "/apis/{apiID}/logs", authAdmin, db, conf)
	RouteLogSearch(search, "/apis/{apiID}/proxy_endpoints/{endpointID}/logs", authAdmin, db, conf)

	RouteResource(&UsersController{base}, "/users", authAdminUser, db, conf)
	RouteRegistration(&RegistrationController{base}, "/registrations", admin, db, conf)
	RoutePasswordReset(&PasswordResetController{configuration.SMTP, psconf, base}, "/password_reset", admin, db, conf)
	RoutePasswordResetConfirmation(&PasswordResetConfirmationController{base}, "/password_reset_confirmation", admin, db, conf)

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
	admin.Handle("/{path:.*}", http.HandlerFunc(adminStaticFileHandler(conf)))

	// also add a route to the base router so that if the user leaves off the trailing slash on the admin
	// path, the client is redirected to the path that includes the trailing slash. this allows the ember
	// front-end to play nicely with us.
	adminPath := strings.TrimRight(conf.PathPrefix, "/")
	router.HandleFunc(adminPath, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, fmt.Sprintf("%s/", adminPath), http.StatusMovedPermanently)
	})
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
