package proxy

import (
	"gateway/config"
	aphttp "gateway/http"
	"gateway/model"
	apsql "gateway/sql"

	"fmt"
	logger "log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

type proxyRouter struct {
	db               *apsql.DB
	hostsRouter      *mux.Router
	accountIDs       map[int64]int64
	hostsRouterMutex sync.RWMutex

	apiRouters      map[int64]*mux.Router
	apiRoutersMutex sync.RWMutex
	methods         map[int64]map[string]map[string]bool
	merged          map[string]map[string]bool
}

func newProxyRouter(db *apsql.DB) *proxyRouter {
	router := &proxyRouter{db: db}
	router.rebuildAll()
	db.RegisterListener(router)
	return router
}

func (r *proxyRouter) Match(request *http.Request, match *mux.RouteMatch) bool {
	defer r.hostsRouterMutex.RUnlock()
	r.hostsRouterMutex.RLock()

	var hostMatch mux.RouteMatch
	if ok := r.hostsRouter.Match(request, &hostMatch); !ok {
		return false
	}

	defer r.apiRoutersMutex.RUnlock()
	r.apiRoutersMutex.RLock()

	apiIDString := hostMatch.Route.GetName()
	apiID, err := strconv.ParseInt(apiIDString, 10, 64)
	if err != nil {
		logger.Fatalf("%s Error converting APIID to int64: %v", config.System, err)
	}
	router, ok := r.apiRouters[apiID]
	if !ok {
		return false
	}

	context.Set(request, aphttp.ContextAccountIDKey, r.accountIDs[apiID])
	context.Set(request, aphttp.ContextAPIIDKey, apiID)
	matched := router.Match(request, match)
	if match.Route != nil {
		endpointID, err := strconv.ParseInt(match.Route.GetName(), 10, 64)
		if err != nil {
			logger.Fatalf("%s Error converting EndpointID to int64: %v", config.System, err)
		}
		context.Set(request, aphttp.ContextEndpointIDKey, endpointID)
	}
	return matched
}

func (r *proxyRouter) rebuildAll() error {
	err := r.rebuildHosts()
	if err != nil {
		return err
	}
	return r.rebuildAPIRouters()
}

func (r *proxyRouter) rebuildHosts() error {
	logger.Printf("%s Rebuilding hosts router", config.System)

	hosts, err := model.AllHosts(r.db)
	if err != nil {
		logger.Printf("%s Error fetching hosts to route: %v",
			config.System, err)
		return err
	}

	router, accountIDs := mux.NewRouter(), make(map[int64]int64)
	for _, host := range hosts {
		route := router.NewRoute()
		route.Name(strconv.FormatInt(host.APIID, 10))
		route.Host(host.Hostname)
		accountIDs[host.APIID] = host.AccountID
	}

	apis, err := model.AllAPIs(r.db)
	if err != nil {
		logger.Printf("%s Error fetching apis: %v",
			config.System, err)
		return err
	}
	for _, api := range apis {
		route := router.NewRoute()
		route.Name(strconv.FormatInt(api.ID, 10))
		route.Host(fmt.Sprintf("%v.example.com", api.ID))
	}

	defer r.hostsRouterMutex.Unlock()
	r.hostsRouterMutex.Lock()
	r.hostsRouter = router
	r.accountIDs = accountIDs

	return nil
}

func merge(methods map[int64]map[string]map[string]bool) map[string]map[string]bool {
	merged := make(map[string]map[string]bool)
	for _, api := range methods {
		for path, route := range api {
			if _, ok := merged[path]; !ok {
				merged[path] = make(map[string]bool)
			}
			for method := range route {
				merged[path][method] = true
			}
		}
	}
	return merged
}

func (r *proxyRouter) rebuildAPIRouters() error {
	logger.Printf("%s Rebuilding all API routers", config.System)

	proxyEndpoints, err := model.AllProxyEndpointsForRouting(r.db)
	if err != nil {
		logger.Printf("%s Error fetching proxy endpoints for all APIs to route: %v",
			config.System, err)
		return err
	}

	routers := make(map[int64]*mux.Router)
	methods := make(map[int64]map[string]map[string]bool)
	for _, endpoint := range proxyEndpoints {
		router, ok := routers[endpoint.APIID]
		if !ok {
			router = mux.NewRouter()
			routers[endpoint.APIID] = router
		}
		if _, ok := methods[endpoint.APIID]; !ok {
			methods[endpoint.APIID] = make(map[string]map[string]bool)
		}
		addProxyEndpointRoutes(endpoint, router, methods[endpoint.APIID])
	}

	defer r.apiRoutersMutex.Unlock()
	r.apiRoutersMutex.Lock()
	r.apiRouters = routers
	r.methods = methods
	r.merged = merge(methods)

	return nil
}

func (r *proxyRouter) rebuildAPIRouterForAPIID(apiID int64) error {
	logger.Printf("%s Rebuilding API router for API %d", config.System, apiID)

	proxyEndpoints, err := model.AllProxyEndpointsForRoutingForAPIID(r.db, apiID)
	if err != nil {
		logger.Printf("%s Error fetching proxy endpoints for API %d to route: %v",
			config.System, apiID, err)
		return err
	}

	router := mux.NewRouter()
	methods := make(map[string]map[string]bool)
	for _, endpoint := range proxyEndpoints {
		addProxyEndpointRoutes(endpoint, router, methods)
	}

	defer r.apiRoutersMutex.Unlock()
	r.apiRoutersMutex.Lock()
	r.apiRouters[apiID] = router
	r.methods[apiID] = methods
	r.merged = merge(r.methods)

	return nil
}

func (r *proxyRouter) deleteAPIRouterForAPIID(apiID int64) error {
	logger.Printf("%s Deleting API router for API %d", config.System, apiID)

	defer r.apiRoutersMutex.Unlock()
	r.apiRoutersMutex.Lock()
	delete(r.apiRouters, apiID)
	delete(r.methods, apiID)
	r.merged = merge(r.methods)

	return nil
}

func isTest(r *http.Request, rm *mux.RouteMatch) bool {
	context.Set(r, aphttp.ContextTest, true)
	return true
}

func addProxyEndpointRoutes(endpoint *model.ProxyEndpoint, router *mux.Router,
	apiMethods map[string]map[string]bool) error {
	routes, err := endpoint.GetRoutes()
	if err != nil {
		logger.Printf("%s Error getting proxy endpoint %d routes: %v",
			config.System, endpoint.ID, err)
		return err
	}

	addRoute := func(proxyRoute *model.ProxyEndpointRoute, prefix string) *mux.Route {
		route := router.NewRoute()
		route.Name(strconv.FormatInt(endpoint.ID, 10))
		route.Path(prefix + proxyRoute.Path)

		methods := proxyRoute.Methods
		if apiMethods[proxyRoute.Path] == nil {
			apiMethods[proxyRoute.Path] = make(map[string]bool)
		}
		for _, method := range proxyRoute.Methods {
			apiMethods[proxyRoute.Path][method] = true
		}
		if endpoint.CORSEnabled && !proxyRoute.HandlesOptions() {
			methods = append(methods, "OPTIONS")
			apiMethods[proxyRoute.Path]["OPTIONS"] = true
		}
		route.Methods(methods...)

		return route
	}

	for _, proxyRoute := range routes {
		if endpoint.Active {
			addRoute(proxyRoute, "")
		}
		/*the testing interface*/
		addRoute(proxyRoute, "/justapis/test").MatcherFunc(isTest)
	}

	return nil
}

func (r *proxyRouter) Notify(n *apsql.Notification) {
	switch {
	case n.Table == "accounts" && n.Event == apsql.Delete:
		go r.rebuildHosts()
		go r.rebuildAPIRouters()
	case n.Table == "apis" && n.Event == apsql.Delete:
		go r.rebuildHosts()
		go r.deleteAPIRouterForAPIID(n.APIID)
	case n.Table == "apis" && n.Event == apsql.Insert:
		go r.rebuildHosts()
	case n.Table == "hosts":
		go r.rebuildHosts()
	case n.Table == "proxy_endpoints":
		go r.rebuildAPIRouterForAPIID(n.APIID)
	}
}

func (r *proxyRouter) Reconnect() {
	logger.Printf("%s Proxy notified of database reconnection", config.System)
	go r.rebuildAll()
}
