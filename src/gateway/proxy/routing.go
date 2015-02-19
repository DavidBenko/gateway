package proxy

import (
	"gateway/config"
	aphttp "gateway/http"
	"gateway/model"
	apsql "gateway/sql"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

type proxyRouter struct {
	db               *apsql.DB
	hostsRouter      *mux.Router
	hostsRouterMutex sync.RWMutex

	apiRouters      map[int64]*mux.Router
	apiRoutersMutex sync.RWMutex
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
		log.Fatalf("%s Error converting APIID to int64: %v", config.System, err)
	}
	router, ok := r.apiRouters[apiID]
	if !ok {
		return false
	}
	context.Set(request, aphttp.ContextAPIIDKey, apiID)

	return router.Match(request, match)
}

func (r *proxyRouter) rebuildAll() error {
	err := r.rebuildHosts()
	if err != nil {
		return err
	}
	return r.rebuildAPIRouters()
}

func (r *proxyRouter) rebuildHosts() error {
	log.Printf("%s Rebuilding hosts router", config.System)

	hosts, err := model.AllHosts(r.db)
	if err != nil {
		log.Printf("%s Error fetching hosts to route: %v",
			config.System, err)
		return err
	}

	router := mux.NewRouter()
	for _, host := range hosts {
		route := router.NewRoute()
		route.Name(strconv.FormatInt(host.APIID, 10))
		route.Host(host.Name)
	}

	defer r.hostsRouterMutex.Unlock()
	r.hostsRouterMutex.Lock()
	r.hostsRouter = router

	return nil
}

func (r *proxyRouter) rebuildAPIRouters() error {
	log.Printf("%s Rebuilding all API routers", config.System)

	proxyEndpoints, err := model.AllActiveProxyEndpointsForRouting(r.db)
	if err != nil {
		log.Printf("%s Error fetching proxy endpoints for all APIs to route: %v",
			config.System, err)
		return err
	}

	routers := make(map[int64]*mux.Router)
	for _, endpoint := range proxyEndpoints {
		router, ok := routers[endpoint.APIID]
		if !ok {
			router = mux.NewRouter()
			routers[endpoint.APIID] = router
		}

		addProxyEndpointRoutes(endpoint, router)
	}

	defer r.apiRoutersMutex.Unlock()
	r.apiRoutersMutex.Lock()
	r.apiRouters = routers

	return nil
}

func (r *proxyRouter) rebuildAPIRouterForAPIID(apiID int64) error {
	log.Printf("%s Rebuilding API router for API %d", config.System, apiID)

	proxyEndpoints, err := model.AllActiveProxyEndpointsForRoutingForAPIID(r.db, apiID)
	if err != nil {
		log.Printf("%s Error fetching proxy endpoints for API %d to route: %v",
			config.System, apiID, err)
		return err
	}

	router := mux.NewRouter()
	for _, endpoint := range proxyEndpoints {
		addProxyEndpointRoutes(endpoint, router)
	}

	defer r.apiRoutersMutex.Unlock()
	r.apiRoutersMutex.Lock()
	r.apiRouters[apiID] = router

	return nil
}

func (r *proxyRouter) deleteAPIRouterForAPIID(apiID int64) error {
	log.Printf("%s Deleting API router for API %d", config.System, apiID)

	defer r.apiRoutersMutex.Unlock()
	r.apiRoutersMutex.Lock()
	delete(r.apiRouters, apiID)

	return nil
}

func addProxyEndpointRoutes(endpoint *model.ProxyEndpoint, router *mux.Router) error {
	routes, err := endpoint.GetRoutes()
	if err != nil {
		log.Printf("%s Error getting proxy endpoint %d routes: %v",
			config.System, endpoint.ID, err)
		return err
	}

	for _, proxyRoute := range routes {
		route := router.NewRoute()
		route.Name(strconv.FormatInt(endpoint.ID, 10))
		route.Path(proxyRoute.Path)
		route.Methods(proxyRoute.Methods...)
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
	case n.Table == "hosts":
		go r.rebuildHosts()
	case n.Table == "proxy_endpoints":
		go r.rebuildAPIRouterForAPIID(n.APIID)
	}
}

func (r *proxyRouter) Reconnect() {
	log.Printf("%s Proxy notified of database reconnection", config.System)
	go r.rebuildAll()
}
