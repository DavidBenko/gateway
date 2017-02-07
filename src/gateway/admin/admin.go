package admin

import (
	"fmt"
	"net/http"
	"strings"

	"gateway/config"
	"gateway/core"
	aphttp "gateway/http"
	"gateway/logreport"
	"gateway/model"
	sql "gateway/sql"
	"gateway/store"

	"github.com/gorilla/mux"
	stripe "github.com/stripe/stripe-go"
)

var (
	defaultDomain string
)

// Setup sets up the session and adds admin routes.
func Setup(router *mux.Router, db *sql.DB, s store.Store, configuration config.Configuration, c *core.Core) {
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
		RouteResource(&UsersController{BaseController: BaseController{accountID: accountIDFromPath, userID: userIDDummy,
			auth: aphttp.AuthTypeSite}}, "/accounts/{accountID}/users", siteAdmin, db, conf)
		// plans allow users to see plans when registering
		RouteResource(&PlansController{}, "/plans", admin, db, conf)
		// sessions are unprotected to allow users to authenticate
		RouteSessions("/sessions", admin, db, conf)
	}

	defaultDomain = psconf.Domain

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

	base := BaseController{conf: conf, accountID: accountID, userID: userID,
		SMTP: configuration.SMTP, ProxyServer: psconf}

	RouteNotify(&NotifyController{BaseController: base}, "/notifications", authAdmin, db, s)

	// Expose binary info to the front end.
	RouteInfo(&InfoController{}, "/info", admin, db, conf)
	if conf.EnableBroker {
		err := newAggregator(conf)
		if err != nil {
			logreport.Fatal(err)
		}
	}
	stream := &LogStreamController{base}
	RouteLogStream(stream, "/logs/socket", authAdmin)
	RouteLogStream(stream, "/apis/{apiID}/logs/socket", authAdmin)
	RouteLogStream(stream, "/apis/{apiID}/proxy_endpoints/{endpointID}/logs/socket", authAdmin)
	RouteLogStream(stream, "/apis/{apiID}/jobs/{endpointID}/logs/socket", authAdmin)
	RouteLogStream(stream, "/timers/{timerID}/logs/socket", authAdmin)

	search := &LogSearchController{configuration.Elastic, base}
	RouteLogSearch(search, "/logs", authAdmin, db, conf)
	RouteLogSearch(search, "/apis/{apiID}/logs", authAdmin, db, conf)
	RouteLogSearch(search, "/apis/{apiID}/proxy_endpoints/{endpointID}/logs", authAdmin, db, conf)
	RouteLogSearch(search, "/apis/{apiID}/jobs/{endpointID}/logs", authAdmin, db, conf)
	RouteLogSearch(search, "/timers/{timerID}/logs", authAdmin, db, conf)

	RouteSingularResource(&AccountController{BaseController: base}, "/account", authAdminUser, db, conf)
	RouteResource(&UsersController{BaseController: base}, "/users", authAdminUser, db, conf)
	if conf.EnableRegistration {
		RouteRegistration(&RegistrationController{base}, "/registrations", admin, db, conf)
		RouteConfirmation(&ConfirmationController{base}, "/registration_confirmation", admin, db, conf)
		// This is here to handle older confirmation emails.
		RouteConfirmation(&ConfirmationController{base}, "/confirmation", admin, db, conf)
	}
	RoutePasswordReset(&PasswordResetController{base}, "/password_reset", admin, db, conf)
	RoutePasswordResetCheck(&PasswordResetCheckController{base}, "/password_reset_check", admin, db, conf)
	RoutePasswordResetConfirmation(&PasswordResetConfirmationController{base}, "/password_reset_confirmation", admin, db, conf)
	if stripe.Key != "" {
		RouteSubscriptions(&SubscriptionsController{base}, "/subscriptions", admin, db, conf)
	}

	apisController := &APIsController{BaseController: base}
	RouteAPIExport(apisController, "/apis/{id}/export", authAdmin, db, conf)
	RouteResource(apisController, "/apis", authAdmin, db, conf)

	testController := &TestController{base, psconf, c}
	RouteTest(testController, "/apis/{apiID}/proxy_endpoints/{endpointID}/tests/{testID}/test", authAdmin, db, conf)

	jobTestController := &JobTestController{base, c}
	RouteJobTest(jobTestController, "/apis/{apiID}/jobs/{endpointID}/tests/{testID}/test", authAdmin, db, conf)

	RouteKeys(&KeysController{base}, "/keys", authAdmin, db, conf)
	RouteSketch(&SketchController{base}, "/sketch", authAdmin, db, conf)
	RouteResource(&HostsController{BaseController: base}, "/apis/{apiID}/hosts", authAdmin, db, conf)
	RouteResource(&EnvironmentsController{BaseController: base}, "/apis/{apiID}/environments", authAdmin, db, conf)
	RouteRootEnvironments(&RootEnvironmentsController{BaseController: base}, "/environments", authAdmin, db, conf)
	RouteResource(&LibrariesController{BaseController: base}, "/apis/{apiID}/libraries", authAdmin, db, conf)
	RouteResource(&EndpointGroupsController{BaseController: base}, "/apis/{apiID}/endpoint_groups", authAdmin, db, conf)
	RouteRootEndpointGroups(&RootEndpointGroupsController{BaseController: base}, "/endpoint_groups", authAdmin, db, conf)
	RouteResource(&RemoteEndpointsController{BaseController: base}, "/apis/{apiID}/remote_endpoints", authAdmin, db, conf)
	RouteRootRemoteEndpoints(&RootRemoteEndpointsController{BaseController: base}, "/remote_endpoints", authAdmin, db, conf)
	RouteResource(&ProxyEndpointsController{BaseController: base, Type: model.ProxyEndpointTypeHTTP}, "/apis/{apiID}/proxy_endpoints", authAdmin, db, conf)
	RouteRootJobs(&RootJobsController{BaseController: base}, "/jobs", authAdmin, db, conf)
	RouteResource(&JobsController{BaseController: base, Type: model.ProxyEndpointTypeJob}, "/apis/{apiID}/jobs", authAdmin, db, conf)
	RouteResource(&JobTestsController{BaseController: base}, "/apis/{apiID}/jobs/{jobID}/tests", authAdmin, db, conf)
	RouteResource(&ProxyEndpointSchemasController{BaseController: base}, "/apis/{apiID}/proxy_endpoints/{endpointID}/schemas", authAdmin, db, conf)
	RouteResource(&ProxyEndpointChannelsController{BaseController: base}, "/apis/{apiID}/proxy_endpoints/{endpointID}/channels", authAdmin, db, conf)
	scratchPadController := &MetaScratchPadsController{ScratchPadsController{BaseController: base}, c}
	RouteScratchPads(scratchPadController, "/apis/{apiID}/remote_endpoints/{endpointID}/environment_data/{environmentDataID}/scratch_pads", authAdmin, db, conf)
	pushChannelsController := &MetaPushChannelsController{PushChannelsController{BaseController: base}, c}
	RoutePushChannels(pushChannelsController, "/push_channels", authAdmin, db, conf)
	RouteResource(&PushChannelMessagesController{BaseController: base}, "/push_channels/{pushChannelID}/push_channel_messages", authAdmin, db, conf)
	RouteResource(&PushDevicesController{BaseController: base}, "/push_channels/{pushChannelID}/push_devices", authAdmin, db, conf)
	RouteResource(&PushMessagesController{BaseController: base}, "/push_channels/{pushChannelID}/push_devices/{pushDeviceID}/push_messages", authAdmin, db, conf)
	RouteResource(&PushDevicesController{BaseController: base}, "/push_devices", authAdmin, db, conf)
	RouteResource(&PushChannelMessagesController{BaseController: base}, "/push_channel_messages", authAdmin, db, conf)
	RouteResource(&SharedComponentsController{BaseController: base}, "/apis/{apiID}/shared_components", authAdmin, db, conf)
	RouteResource(&TimersController{BaseController: base}, "/timers", authAdmin, db, conf)

	if configuration.RemoteEndpoint.CustomFunctionEnabled {
		customFunctionTestController := &CustomFunctionTestController{base}
		RouteCustomFunctionTest(customFunctionTestController, "/apis/{apiID}/custom_functions/{customFunctionID}/tests/{testID}/test", authAdmin, db, conf)

		RouteResource(&CustomFunctionsController{BaseController: base}, "/apis/{apiID}/custom_functions", authAdmin, db, conf)
		RouteResource(&CustomFunctionFilesController{BaseController: base}, "/apis/{apiID}/custom_functions/{customFunctionID}/files", authAdmin, db, conf)
		RouteResource(&CustomFunctionTestsController{BaseController: base}, "/apis/{apiID}/custom_functions/{customFunctionID}/tests", authAdmin, db, conf)
		RouteCustomFunctionBuild(&CustomFunctionBuildController{BaseController: base}, "/apis/{apiID}/custom_functions/{customFunctionID}/build", authAdmin, db, conf)
	}

	RouteStoreResource(&StoreCollectionsController{base, s}, "/store_collections", authAdmin, conf)
	RouteStoreResource(&StoreObjectsController{base, s}, "/store_collections/{collectionID}/store_objects", authAdmin, conf)

	RouteResource(&RemoteEndpointTypesController{BaseController: base}, "/remote_endpoint_types", authAdmin, db, conf)

	// static assets for self-hosted systems
	admin.Handle("/{path:.*}", http.HandlerFunc(adminStaticFileHandler(conf)))

	// also add a route to the base router so that if the user leaves off the trailing slash on the admin
	// path, the client is redirected to the path that includes the trailing slash. this allows the ember
	// front-end to play nicely with us.
	adminPath := strings.TrimRight(conf.PathPrefix, "/")
	router.HandleFunc(adminPath, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, fmt.Sprintf("%s/", adminPath), http.StatusMovedPermanently)
	})

	var public aphttp.Router
	public = aphttp.NewAccessLoggingRouter(config.Proxy, conf.RequestIDHeader,
		router)
	if conf.CORSEnabled {
		public = aphttp.NewCORSAwareRouter(conf.CORSOrigin, public)
	}
	matcher := newHostMatcher(db)
	RouteSwagger(&SwaggerController{matcher}, "/swagger.json", public, db, conf)
	RoutePush(&PushController{matcher, c}, "/push", public, db, conf)
	RouteMQTTProxy(&MQTTProxyController{base, configuration.Push}, "/mqtt", public, conf)
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
