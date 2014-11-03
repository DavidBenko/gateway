package proxy

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"gateway/config"
	"gateway/db"
	aphttp "gateway/http"
	"gateway/model"
	"gateway/proxy/admin"
	"gateway/proxy/vm"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/robertkrimen/otto"
)

// Server encapsulates the proxy server.
type Server struct {
	proxyConf config.ProxyServer
	adminConf config.ProxyAdmin
	db        db.DB
	router    *mux.Router
}

// NewServer builds a new proxy server.
func NewServer(proxyConfig config.ProxyServer, adminConfig config.ProxyAdmin, db db.DB) *Server {
	return &Server{
		proxyConf: proxyConfig,
		adminConf: adminConfig,
		db:        db,
		router:    mux.NewRouter(),
	}
}

// Run runs the server.
func (s *Server) Run() {
	// Set up admin
	admin.AddRoutes(s.router, s.db, s.adminConf)

	// Set up proxy
	s.router.Handle("/{path:.*}",
		aphttp.AccessLoggingHandler(config.Proxy,
			aphttp.ErrorCatchingHandler(s.proxyHandlerFunc))).
		MatcherFunc(s.isRoutedToProxyEndpoint)

	s.router.NotFoundHandler = accessLoggingNotFoundHandler()

	// Run server
	listen := fmt.Sprintf("%s:%d", s.proxyConf.Host, s.proxyConf.Port)
	log.Printf("%s Server listening at %s", config.Proxy, listen)
	log.Fatalf("%s %v", config.System, http.ListenAndServe(listen, s.router))
}

func (s *Server) isRoutedToProxyEndpoint(r *http.Request, rm *mux.RouteMatch) bool {
	router := s.db.Router().MUXRouter
	if router == nil {
		return false
	}

	var match mux.RouteMatch
	ok := router.Match(r, &match)
	if ok {
		context.Set(r, aphttp.ContextMatchKey, &match)
	}
	return ok
}

func (s *Server) proxyHandlerFunc(w http.ResponseWriter, r *http.Request) aphttp.Error {
	start := time.Now()

	match := context.Get(r, aphttp.ContextMatchKey).(*mux.RouteMatch)
	requestID := context.Get(r, aphttp.ContextRequestIDKey).(string)

	modelEndpoint, err := s.db.Find(&model.ProxyEndpoint{}, "Name",
		match.Route.GetName())
	if err != nil {
		return aphttp.NewServerError(err)
	}

	endpoint := modelEndpoint.(*model.ProxyEndpoint)

	incomingJSON, err := proxyRequestJSON(r, match.Vars)
	if err != nil {
		return aphttp.NewServerError(err)
	}

	var scripts = []interface{}{
		endpoint.Script,
		"main(JSON.parse(__ap_proxyRequestJSON));",
	}

	vm, err := vm.NewVM(requestID)
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

	aphttp.AddHeaders(w.Header(), response.Headers)
	w.WriteHeader(response.StatusCode)
	w.Write([]byte(response.Body))

	total := time.Since(start)
	proxied := vm.ProxiedRequestsDuration
	processing := total - proxied
	log.Printf("%s [req %s] [time] %v (processing %v, requests %v)",
		config.Proxy, requestID, total, processing, proxied)

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
