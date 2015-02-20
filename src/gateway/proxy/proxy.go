package proxy

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
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

// Server encapsulates the proxy server.
type Server struct {
	proxyConf   config.ProxyServer
	adminConf   config.ProxyAdmin
	router      *mux.Router
	proxyRouter *proxyRouter
	db          *sql.DB
}

// NewServer builds a new proxy server.
func NewServer(proxyConfig config.ProxyServer, adminConfig config.ProxyAdmin, db *sql.DB) *Server {
	return &Server{
		proxyConf: proxyConfig,
		adminConf: adminConfig,
		router:    mux.NewRouter(),
		db:        db,
	}
}

// Run runs the server.
func (s *Server) Run() {

	// Set up admin
	admin.Setup(s.router, s.db, s.adminConf)

	// Set up proxy
	s.proxyRouter = newProxyRouter(s.db)

	s.router.Handle("/{path:.*}",
		aphttp.AccessLoggingHandler(config.Proxy,
			aphttp.ErrorCatchingHandler(s.proxyHandlerFunc))).
		MatcherFunc(s.isRoutedToEndpoint)

	s.router.NotFoundHandler = accessLoggingNotFoundHandler()

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

func (s *Server) proxyHandlerFunc(w http.ResponseWriter, r *http.Request) aphttp.Error {
	start := time.Now()

	match := context.Get(r, aphttp.ContextMatchKey).(*mux.RouteMatch)
	requestID := context.Get(r, aphttp.ContextRequestIDKey).(string)

	var proxiedRequestsDuration time.Duration
	defer func() {
		total := time.Since(start)
		processing := total - proxiedRequestsDuration
		log.Printf("%s [req %s] [time] %v (processing %v, requests %v)",
			config.Proxy, requestID, total, processing, proxiedRequestsDuration)
	}()

	proxyEndpointID, err := strconv.ParseInt(match.Route.GetName(), 10, 64)
	if err != nil {
		return aphttp.NewServerError(err)
	}

	proxyEndpoint, err := model.FindProxyEndpointForProxy(s.db, proxyEndpointID)
	if err != nil {
		return aphttp.NewServerError(err)
	}

	log.Printf("%s [req %s] [route] %s", config.Proxy, requestID, proxyEndpoint.Name)

	vm, err := vm.NewVM(requestID, w, r, s.proxyConf, s.db, proxyEndpoint)
	if err != nil {
		return aphttp.NewServerError(err)
	}

	/* TODO: Let's see if we can bind the request in directly? */
	incomingJSON, err := proxyRequestJSON(r, match.Vars)
	if err != nil {
		return aphttp.NewServerError(err)
	}
	vm.Set("__ap_proxyRequestJSON", incomingJSON)
	scripts := []interface{}{
		"var request = JSON.parse(__ap_proxyRequestJSON);",
		"var response = new AP.HTTP.Response();",
	}
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

	if s.proxyConf.RequestIDHeader != "" {
		response.Headers[s.proxyConf.RequestIDHeader] = requestID
	}

	response.Headers["Content-Length"] = len(response.Body)
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

func accessLoggingNotFoundHandler() http.Handler {
	return aphttp.AccessLoggingHandler(config.Proxy,
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
