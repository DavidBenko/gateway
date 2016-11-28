package proxy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"gateway/admin"
	"gateway/config"
	"gateway/core"
	aphttp "gateway/http"
	"gateway/logreport"
	"gateway/model"
	apvm "gateway/proxy/vm"
	sql "gateway/sql"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	stripe "github.com/stripe/stripe-go"
)

// Server encapsulates the proxy server.
type Server struct {
	*core.Core
	devMode     bool
	proxyConf   config.ProxyServer
	adminConf   config.ProxyAdmin
	conf        config.Configuration
	router      *mux.Router
	proxyRouter *proxyRouter
	proxyData   proxyDataSource
}

// NewServer builds a new proxy server.
func NewServer(conf config.Configuration, ownDb *sql.DB, warp *core.Core) *Server {
	var source proxyDataSource
	if conf.Proxy.CacheAPIs {
		source = newCachingProxyDataSource(ownDb)
	} else {
		source = newPassthroughProxyDataSource(ownDb)
	}

	return &Server{
		Core:      warp,
		devMode:   conf.DevMode(),
		proxyConf: conf.Proxy,
		adminConf: conf.Admin,
		conf:      conf,
		router:    mux.NewRouter(),
		proxyData: source,
	}
}

// Run runs the server.
func (s *Server) Run() {

	// system-wide health-check
	if s.proxyConf.HealthCheckPath != "" {
		s.router.Handle(s.proxyConf.HealthCheckPath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok\n"))
		}))
	}

	// Set up admin
	admin.Setup(s.router, s.OwnDb, s.Store, s.conf, s.Core)

	// Set up proxy
	s.proxyRouter = newProxyRouter(s.OwnDb)

	s.router.Handle("/{path:.*}",
		aphttp.AccessLoggingHandler(config.Proxy, s.proxyConf.RequestIDHeader,
			aphttp.ErrorCatchingHandler(s.proxyHandlerFunc))).
		MatcherFunc(s.isRoutedToEndpoint)

	s.router.NotFoundHandler = s.accessLoggingNotFoundHandler()

	// Run server
	listen := fmt.Sprintf("%s:%d", s.proxyConf.Host, s.proxyConf.Port)
	logreport.Printf("%s Server listening at %s", config.Proxy, listen)
	var adminHost string
	if len(strings.TrimSpace(s.adminConf.Host)) == 0 {
		adminHost = s.proxyConf.Host
	} else {
		adminHost = s.adminConf.Host
	}
	adminAvailable := fmt.Sprintf("%s:%d%s", adminHost, s.proxyConf.Port, s.adminConf.PathPrefix)
	logreport.Printf("%s Admin dashboard available at %s", config.Admin, adminAvailable)
	logreport.Fatalf("%s %v", config.System, http.ListenAndServe(listen, s.router))
}

func (s *Server) isRoutedToEndpoint(r *http.Request, rm *mux.RouteMatch) bool {
	var match mux.RouteMatch
	ok := s.proxyRouter.Match(r, &match)
	if ok {
		context.Set(r, aphttp.ContextMatchKey, &match)
	}
	return ok
}

func (s *Server) proxyHandler(w http.ResponseWriter, r *http.Request) (
	response *core.ProxyResponse, logs *bytes.Buffer, httpErr aphttp.Error) {
	start := time.Now()

	var vm *apvm.ProxyVM

	match := context.Get(r, aphttp.ContextMatchKey).(*mux.RouteMatch)
	requestID := context.Get(r, aphttp.ContextRequestIDKey).(string)
	logPrefix := context.Get(r, aphttp.ContextLogPrefixKey).(string)
	accountID := context.Get(r, aphttp.ContextAccountIDKey).(int64)

	logs = &bytes.Buffer{}
	logPrint := logreport.PrintfCopier(logs)

	defer func() {
		if httpErr != nil {
			s.logError(logPrint, logPrefix, httpErr, r)
		}
		s.logDuration(vm, logPrint, logPrefix, start)
	}()

	proxyEndpointID, err := strconv.ParseInt(match.Route.GetName(), 10, 64)
	if err != nil {
		httpErr = s.httpError(err)
		return
	}

	proxyEndpoint, err := s.proxyData.Endpoint(proxyEndpointID)
	if err != nil {
		httpErr = s.httpError(err)
		return
	}

	proxyEndpoint.AccountID = accountID

	libraries, err := s.proxyData.Libraries(proxyEndpoint.APIID)
	if err != nil {
		httpErr = s.httpError(err)
		return
	}

	codeTimeout := s.proxyConf.CodeTimeout
	if stripe.Key != "" {
		plan, err := s.proxyData.Plan(accountID)
		if err != nil {
			httpErr = s.httpError(err)
			return
		}
		if plan.JavascriptTimeout < codeTimeout {
			codeTimeout = plan.JavascriptTimeout
		}
	}

	logPrint("%s [route] %s", logPrefix, proxyEndpoint.Name)

	if r.Method == "OPTIONS" {
		route, err := s.matchingRouteForOptions(proxyEndpoint, r)
		if err != nil {
			httpErr = s.httpError(err)
			return
		}
		if !route.HandlesOptions() {
			err := s.corsOptionsHandlerFunc(w, r, proxyEndpoint, route, requestID)
			if err != nil {
				httpErr = s.httpError(err)
			}
			return
		}
	}

	request, err := proxyRequestJSON(r, requestID, match.Vars)
	if err != nil {
		httpErr = s.httpError(err)
		return
	}

	if schema := proxyEndpoint.Schema; schema != nil && schema.RequestSchema != "" {
		err := s.ProcessSchema(proxyEndpoint.Schema.RequestSchema, request.Body)
		if err != nil {
			if err.Error() == "EOF" {
				httpErr = aphttp.NewError(errors.New("a json document is required in the request"), 422)
				return
			}
			httpErr = aphttp.NewError(err, 400)
			return
		}
	}

	vm, err = apvm.NewVM(logPrint, logPrefix, w, r, s.proxyConf, s.OwnDb, proxyEndpoint, libraries, codeTimeout, s.Core.KeyStore)
	if err != nil {
		httpErr = s.httpError(err)
		return
	}

	incomingJSON, err := request.Marshal()
	if err != nil {
		httpErr = s.httpError(err)
		return
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
		httpErr = s.httpError(err)
		return
	}

	if err = s.RunComponents(&vm.CoreVM, proxyEndpoint.Components); err != nil {
		if err.Error() == "JavaScript took too long to execute" {
			logPrint("%s [timeout] JavaScript execution exceeded %ds timeout threshold", logPrefix, codeTimeout)
		}
		httpErr = s.httpJavascriptError(err, proxyEndpoint.Environment)
		return
	}

	responseObject, err := vm.Run("response;")
	if err != nil {
		httpErr = s.httpError(err)
		return
	}
	responseJSON, err := s.ObjectJSON(&vm.CoreVM, responseObject)
	if err != nil {
		httpErr = s.httpError(err)
		return
	}
	response, err = core.ProxyResponseFromJSON(responseJSON)
	if err != nil {
		httpErr = s.httpError(err)
		return
	}

	if schema := proxyEndpoint.Schema; schema != nil &&
		(schema.ResponseSchema != "" ||
			(schema.ResponseSameAsRequest && schema.RequestSchema != "")) {
		responseSchema := schema.ResponseSchema
		if schema.ResponseSameAsRequest {
			responseSchema = schema.RequestSchema
		}
		err := s.ProcessSchema(responseSchema, response.Body)
		if err != nil {
			if err.Error() == "EOF" {
				httpErr = aphttp.NewError(errors.New("a json document is required in the response"), 500)
				return
			}
			httpErr = aphttp.NewError(err, 500)
			return
		}
	}

	if proxyEndpoint.CORSEnabled {
		s.addCORSCommonHeaders(w, proxyEndpoint)
	}
	response.Headers["Content-Length"] = len(response.Body)
	aphttp.AddHeaders(w.Header(), response.Headers)

	return
}

func (s *Server) proxyHandlerFunc(w http.ResponseWriter, r *http.Request) aphttp.Error {
	logPrefix := context.Get(r, aphttp.ContextLogPrefixKey).(string)
	test, _ := context.Get(r, aphttp.ContextTest).(bool)

	response, logs, httpErr := s.proxyHandler(w, r)

	if test {
		responseBody, status := "", ""
		if httpErr != nil {
			responseBody = fmt.Sprintf("%s\n", httpErr.String())
			status = fmt.Sprintf("%v", httpErr.Code())
		} else if response != nil {
			responseBody = response.Body
			status = fmt.Sprintf("%v", response.StatusCode)
		}
		response := aphttp.TestResponse{
			Body:   responseBody,
			Log:    logs.String(),
			Status: status,
		}

		body, err := json.Marshal(&response)
		if err != nil {
			logreport.Printf("%s [error] %s", logPrefix, err)
			return s.httpError(err)
		}

		w.Write(body)
	} else if httpErr != nil {
		return httpErr
	} else if response != nil {
		w.WriteHeader(response.StatusCode)
		w.Write([]byte(response.Body))
	}
	return nil
}

func (s *Server) httpError(err error) aphttp.Error {
	if !s.devMode {
		return aphttp.DefaultServerError()
	}

	return aphttp.NewServerError(err)
}

func (s *Server) httpJavascriptError(err error, env *model.Environment) aphttp.Error {
	if env == nil {
		return s.httpError(err)
	}

	if env.ShowJavascriptErrors {
		return aphttp.NewServerError(err)
	}

	return aphttp.DefaultServerError()
}

func (s *Server) accessLoggingNotFoundHandler() http.Handler {
	return aphttp.AccessLoggingHandler(config.Proxy, s.proxyConf.RequestIDHeader,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		}))
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
	requestID string) error {

	s.addCORSCommonHeaders(w, endpoint)
	methods := []string{}
	for method, _ := range s.proxyRouter.merged[route.Path] {
		methods = append(methods, method)
	}
	sort.Strings(methods)
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

func (s *Server) logError(logPrint logreport.Logf, logPrefix string, err aphttp.Error, r *http.Request) {
	errString := "Unknown Error"
	lines := strings.Split(err.String(), "\n")
	if len(lines) > 0 {
		errString = lines[0]
	}
	logPrint("%s [error] %s\n%v", logPrefix, errString, r)
}

func (s *Server) logDuration(vm *apvm.ProxyVM, logPrint logreport.Logf, logPrefix string, start time.Time) {
	var proxiedRequestsDuration time.Duration
	if vm != nil {
		proxiedRequestsDuration = vm.ProxiedRequestsDuration
	}

	total := time.Since(start)
	processing := total - proxiedRequestsDuration
	logPrint("%s [time] %v (processing %v, requests %v)",
		logPrefix, total, processing, proxiedRequestsDuration)
}
