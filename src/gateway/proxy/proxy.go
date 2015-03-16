package proxy

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gateway/admin"
	"gateway/config"
	aphttp "gateway/http"
	"gateway/model"
	"gateway/proxy/vm"
	sql "gateway/sql"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx/types"
	"github.com/robertkrimen/otto"
)

// ProxyDataSource defines an interface to get data to serve the proxy
type ProxyDataSource interface {
	Endpoint(id int64) (*model.ProxyEndpoint, error)
	Libraries(apiID int64) ([]*model.Library, error)
}

// Server encapsulates the proxy server.
type Server struct {
	proxyConf   config.ProxyServer
	adminConf   config.ProxyAdmin
	router      *mux.Router
	proxyRouter *proxyRouter
	proxyData   ProxyDataSource
	db          *sql.DB
	httpClient  *http.Client
}

// NewServer builds a new proxy server.
func NewServer(proxyConfig config.ProxyServer, adminConfig config.ProxyAdmin, db *sql.DB) *Server {
	httpTimeout := time.Duration(proxyConfig.HTTPTimeout) * time.Second

	var source ProxyDataSource
	if proxyConfig.CacheAPIs {
		source = newCachingProxyDataSource(db)
	} else {
		source = newPassthroughProxyDataSource(db)
	}

	return &Server{
		proxyConf:  proxyConfig,
		adminConf:  adminConfig,
		router:     mux.NewRouter(),
		proxyData:  source,
		db:         db,
		httpClient: &http.Client{Timeout: httpTimeout},
	}
}

// Run runs the server.
func (s *Server) Run() {

	// Set up admin
	admin.Setup(s.router, s.db, s.adminConf)

	// Set up proxy
	s.proxyRouter = newProxyRouter(s.db)

	s.router.Handle("/{path:.*}",
		aphttp.AccessLoggingHandler(config.Proxy, s.proxyConf.RequestIDHeader,
			aphttp.ErrorCatchingHandler(s.proxyHandlerFunc))).
		MatcherFunc(s.isRoutedToEndpoint)

	s.router.NotFoundHandler = s.accessLoggingNotFoundHandler()

	// Run server
	listen := fmt.Sprintf("%s:%d", s.proxyConf.Host, s.proxyConf.Port)
	log.Printf("%s Server listening at %s", config.Proxy, listen)
	log.Fatalf("%s %v", config.System, http.ListenAndServe(listen, s.router))
}

func (s *Server) isRoutedToEndpoint(r *http.Request, rm *mux.RouteMatch) bool {
	var match mux.RouteMatch
	ok := s.proxyRouter.Match(r, &match)
	if ok {
		context.Set(r, aphttp.ContextMatchKey, &match)
	}
	return ok
}

func (s *Server) proxyHandlerFunc(w http.ResponseWriter, r *http.Request) (httpErr aphttp.Error) {
	start := time.Now()

	match := context.Get(r, aphttp.ContextMatchKey).(*mux.RouteMatch)
	requestID := context.Get(r, aphttp.ContextRequestIDKey).(string)
	logPrefix := context.Get(r, aphttp.ContextLogPrefixKey).(string)

	var proxiedRequestsDuration time.Duration
	defer func() {
		if httpErr != nil {
			log.Printf("%s [error] %s", logPrefix, httpErr.String())
		}
		total := time.Since(start)
		processing := total - proxiedRequestsDuration
		log.Printf("%s [time] %v (processing %v, requests %v)",
			logPrefix, total, processing, proxiedRequestsDuration)
	}()

	proxyEndpointID, err := strconv.ParseInt(match.Route.GetName(), 10, 64)
	if err != nil {
		return aphttp.NewServerError(err)
	}

	proxyEndpoint, err := s.proxyData.Endpoint(proxyEndpointID)
	if err != nil {
		return aphttp.NewServerError(err)
	}

	libraries, err := s.proxyData.Libraries(proxyEndpoint.APIID)
	if err != nil {
		return aphttp.NewServerError(err)
	}

	log.Printf("%s [route] %s", logPrefix, proxyEndpoint.Name)

	if r.Method == "OPTIONS" {
		route, err := s.matchingRouteForOptions(proxyEndpoint, r)
		if err != nil {
			return aphttp.NewServerError(err)
		}
		if !route.HandlesOptions() {
			return s.corsOptionsHandlerFunc(w, r, proxyEndpoint, route, requestID)
		}
	}

	vm, err := vm.NewVM(logPrefix, w, r, s.proxyConf, s.db, proxyEndpoint, libraries)
	if err != nil {
		return aphttp.NewServerError(err)
	}

	incomingJSON, err := proxyRequestJSON(r, requestID, match.Vars)
	if err != nil {
		return aphttp.NewServerError(err)
	}
	vm.Set("__ap_proxyRequestJSON", incomingJSON)
	scripts := []interface{}{
		"var request = JSON.parse(__ap_proxyRequestJSON);",
		"var response = new AP.HTTP.Response();",
	}
	scripts = append(scripts,
		fmt.Sprintf("var session = new AP.Session(%s);",
			strconv.Quote(proxyEndpoint.Environment.SessionName)))

	if _, err := vm.RunAll(scripts); err != nil {
		return aphttp.NewServerError(err)
	}

	if err = s.runComponents(vm, proxyEndpoint.Components); err != nil {
		return aphttp.NewServerError(err)
	}

	responseObject, err := vm.Run("response;")
	if err != nil {
		return aphttp.NewServerError(err)
	}
	responseJSON, err := s.objectJSON(vm, responseObject)
	if err != nil {
		return aphttp.NewServerError(err)
	}
	response, err := proxyResponseFromJSON(responseJSON)
	if err != nil {
		return aphttp.NewServerError(err)
	}
	proxiedRequestsDuration = vm.ProxiedRequestsDuration

	if proxyEndpoint.CORSEnabled {
		s.addCORSCommonHeaders(w, proxyEndpoint)
	}

	response.Headers["Content-Length"] = len(response.Body)
	delete(response.Headers, "Content-Encoding")
	aphttp.AddHeaders(w.Header(), response.Headers)
	w.WriteHeader(response.StatusCode)
	w.Write([]byte(response.Body))
	return nil
}

func (s *Server) objectJSON(vm *vm.ProxyVM, object otto.Value) (string, error) {
	jsJSON, err := vm.Object("JSON")
	if err != nil {
		return "", err
	}
	result, err := jsJSON.Call("stringify", object)
	if err != nil {
		return "", err
	}
	return result.String(), nil
}

func (s *Server) accessLoggingNotFoundHandler() http.Handler {
	return aphttp.AccessLoggingHandler(config.Proxy, s.proxyConf.RequestIDHeader,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		}))
}

func (s *Server) runStoredJSONScript(vm *vm.ProxyVM, jsonScript types.JsonText) error {
	script, err := strconv.Unquote(string(jsonScript))
	if err != nil || script == "" {
		return err
	}
	_, err = vm.Run(script)
	return err
}

func (s *Server) matchingRouteForOptions(endpoint *model.ProxyEndpoint,
	r *http.Request) (*model.ProxyEndpointRoute, error) {
	routes, err := endpoint.GetRoutes()
	if err != nil {
		return nil, err
	}
	for _, proxyRoute := range routes {
		route := &mux.Route{}
		route.Path(proxyRoute.Path)
		methods := proxyRoute.Methods
		if !proxyRoute.HandlesOptions() {
			methods = append(methods, "OPTIONS")
		}
		route.Methods(methods...)
		var match mux.RouteMatch
		if route.Match(r, &match) {
			return proxyRoute, nil
		}
	}
	return nil, errors.New("No route matched")
}

func (s *Server) corsOptionsHandlerFunc(w http.ResponseWriter, r *http.Request,
	endpoint *model.ProxyEndpoint, route *model.ProxyEndpointRoute,
	requestID string) aphttp.Error {

	s.addCORSCommonHeaders(w, endpoint)
	methods := route.Methods
	methods = append(methods, "OPTIONS")
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ", "))
	return nil
}

func (s *Server) addCORSCommonHeaders(w http.ResponseWriter,
	endpoint *model.ProxyEndpoint) {

	api := endpoint.API

	w.Header().Set("Access-Control-Allow-Origin", api.CORSAllowOrigin)
	w.Header().Set("Access-Control-Request-Headers", api.CORSRequestHeaders)
	w.Header().Set("Access-Control-Allow-Headers", api.CORSAllowHeaders)
	w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", api.CORSMaxAge))

	if api.CORSAllowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
}
