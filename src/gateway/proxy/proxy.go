package proxy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"gateway/config"
	"gateway/db"
	"gateway/model"
	"gateway/proxy/admin"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/robertkrimen/otto"
)

type contextKey int

const (
	contextEndpointKey contextKey = iota
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
	s.router.HandleFunc("/{path:.*}", proxyHandlerFunc).
		MatcherFunc(s.hasRegisteredProxyEndpoint)

	// Run server
	listen := fmt.Sprintf("%s:%d", s.proxyConf.Host, s.proxyConf.Port)
	log.Println("Proxy server listening at:", listen)
	log.Fatal(http.ListenAndServe(listen, s.router))
}

func (s *Server) hasRegisteredProxyEndpoint(r *http.Request, rm *mux.RouteMatch) bool {
	path := r.URL.Path[1:]
	endpoint, err := s.db.Find(model.ProxyEndpoint{}, "Path", path)
	if err != nil {
		return false
	}
	context.Set(r, contextEndpointKey, endpoint)
	return true
}

func proxyHandlerFunc(w http.ResponseWriter, r *http.Request) {
	endpoint := context.Get(r, contextEndpointKey).(model.ProxyEndpoint)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{}

	req, err := http.NewRequest(r.Method, "http://localhost:4567", nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Body = ioutil.NopCloser(massageBody(bytes.NewBuffer(body), endpoint.Script))

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	newBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	newRespBody := massageBody(bytes.NewBuffer(newBody), endpoint.Script)

	fmt.Fprint(w, newRespBody.String())
}

func massageBody(body *bytes.Buffer, src interface{}) *bytes.Buffer {
	vm := otto.New()

	vm.Set("body", body.String())

	_, err := vm.Run(src)
	if err != nil {
		log.Fatal(err)
	}

	newBodyRaw, err := vm.Get("body")
	if err != nil {
		log.Fatal(err)
	}

	newBody, err := newBodyRaw.ToString()
	if err != nil {
		log.Fatal(err)
	}

	return bytes.NewBufferString(newBody)
}
