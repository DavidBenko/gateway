package proxy

import (
	"fmt"
	"log"
	"net/http"

	"gateway/config"
	"gateway/db"
	aphttp "gateway/http"
	"gateway/model"
	"gateway/proxy/admin"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/robertkrimen/otto"
)

type contextKey int

const (
	contextMatchKey contextKey = iota
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
		aphttp.ErrorCatchingHandler(s.proxyHandlerFunc)).
		MatcherFunc(s.isRoutedToProxyEndpoint)

	// Run server
	listen := fmt.Sprintf("%s:%d", s.proxyConf.Host, s.proxyConf.Port)
	log.Println("Proxy server listening at:", listen)
	log.Fatal(http.ListenAndServe(listen, s.router))
}

func (s *Server) isRoutedToProxyEndpoint(r *http.Request, rm *mux.RouteMatch) bool {
	router := s.db.Router().MUXRouter
	if router == nil {
		return false
	}

	var match mux.RouteMatch
	ok := router.Match(r, &match)
	if ok {
		context.Set(r, contextMatchKey, &match)
	}
	return ok
}

func (s *Server) proxyHandlerFunc(w http.ResponseWriter, r *http.Request) aphttp.Error {
	match := context.Get(r, contextMatchKey).(*mux.RouteMatch)

	modelEndpoint, err := s.db.Find(model.ProxyEndpoint{}, "Name",
		match.Route.GetName())
	if err != nil {
		return aphttp.NewServerError(err)
	}

	endpoint := modelEndpoint.(model.ProxyEndpoint)

	incomingJSON, err := proxyRequestJSON(r, match.Vars)
	if err != nil {
		return aphttp.NewServerError(err)
	}

	var scripts = []interface{}{
		endpoint.Script,
		"main(JSON.parse(__ap_proxyRequestJSON));",
	}

	vm, err := s.newVM()
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
	return nil
}
