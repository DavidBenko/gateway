package proxy

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"gateway/config"
	aphttp "gateway/http"
	"gateway/proxy/admin"
	"gateway/proxy/keys"
	"gateway/proxy/vm"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/robertkrimen/otto"
)

// Server encapsulates the proxy server.
type Server struct {
	proxyConf   config.ProxyServer
	adminConf   config.ProxyAdmin
	router      *mux.Router
	proxyRouter *mux.Router
	scripts     map[string]*otto.Script
}

// NewServer builds a new proxy server.
func NewServer(proxyConfig config.ProxyServer, adminConfig config.ProxyAdmin) *Server {
	scripts, err := scriptsFromFilesystem(proxyConfig)
	if err != nil {
		log.Fatalf("%s Could not setup proxy code: %v", config.Proxy, err)
	}

	routes, ok := scripts["routes"]
	if !ok {
		log.Fatalf("%s Top level code path must contain routes.js with routing code", config.Proxy)
	}

	proxyRouter, err := ParseRoutes(routes)
	if err != nil {
		log.Fatalf("%s Could not parse routes.js: %v", config.Proxy, err)
	}

	if proxyConfig.WatchRestart {
		go watchForRestarts(proxyConfig.CodePath)
	}

	return &Server{
		proxyConf:   proxyConfig,
		adminConf:   adminConfig,
		router:      mux.NewRouter(),
		proxyRouter: proxyRouter,
		scripts:     scripts,
	}
}

// Run runs the server.
func (s *Server) Run() {

	// Set up admin
	admin.AddRoutes(s.router, s.adminConf)

	// Set up proxy
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

	controller := match.Route.GetName()

	log.Printf("%s [req %s] [route] %s", config.Proxy, requestID, controller)

	endpointKey, err := keys.ScriptKeyForCodeString(controller)
	if err != nil {
		return aphttp.NewServerError(err)
	}

	endpoint, ok := s.scripts[endpointKey]
	// FIXME: I think we can guarantee this particular bit won't happen at startup
	if !ok {
		err = fmt.Errorf("No code found for '%s'", controller)
		return aphttp.NewServerError(err)
	}

	incomingJSON, err := proxyRequestJSON(r, match.Vars)
	if err != nil {
		return aphttp.NewServerError(err)
	}

	handleScript := fmt.Sprintf("App.handle(JSON.parse(__ap_proxyRequestJSON), %s);", controller)

	var scripts = []interface{}{
		s.scripts["app"], // FIXME Ensure existance & check validity in setup
		endpoint,
		handleScript,
	}

	vm, err := vm.NewVM(requestID, w, r, s.proxyConf, s.scripts)
	if err != nil {
		return aphttp.NewServerError(err)
	}

	vm.Set("__ap_proxyRequestJSON", incomingJSON)

	var result otto.Value
	for _, script := range scripts {
		var err error
		result, err = vm.Run(script)
		if err != nil {
			return aphttp.NewServerError(err)
		}
	}

	responseJSON, err := s.objectJSON(vm, result)
	if err != nil {
		return aphttp.NewServerError(err)
	}
	response, err := proxyResponseFromJSON(responseJSON)
	if err != nil {
		return aphttp.NewServerError(err)
	}
	proxiedRequestsDuration = vm.ProxiedRequestsDuration

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
