package proxy

import (
	"encoding/json"
	"fmt"

	"gateway/proxy/routerjs"

	"github.com/gorilla/mux"
	"github.com/robertkrimen/otto"
)

type jsRoute struct {
	Methods    []string          `json:"_methods"`
	Schemes    []string          `json:"_schemes"`
	Host       string            `json:"_host"`
	Path       string            `json:"_path"`
	PathPrefix string            `json:"_pathPrefix"`
	Headers    map[string]string `json:"_headers"`
	Queries    map[string]string `json:"_queries"`
	Name       string            `json:"_name"`
}

// ParseRoutes parses the script and stores the resulting the mux.Router.
func ParseRoutes(script *otto.Script) (*mux.Router, error) {
	vm := otto.New()

	var files = []string{"route.js", "router.js"}
	var scripts = make([]interface{}, 0)
	for _, filename := range files {
		fileJS, err := routerjs.Asset(filename)
		if err != nil {
			return nil, err
		}
		scripts = append(scripts, fileJS)
	}
	scripts = append(scripts, script)
	scripts = append(scripts, "router.routeData();")

	var result otto.Value
	for _, script := range scripts {
		var err error
		result, err = vm.Run(script)
		if err != nil {
			return nil, err
		}
	}

	var jsRoutes []jsRoute
	if err := json.Unmarshal([]byte(result.String()), &jsRoutes); err != nil {
		return nil, err
	}

	router := mux.NewRouter()
	for _, route := range jsRoutes {
		addRoute(route, router)
	}

	return router, nil
}

func addRoute(js jsRoute, router *mux.Router) error {
	route := router.NewRoute()

	if js.Name == "" {
		return fmt.Errorf("All routes must route to a proxy endpoint by name.")
	}
	route.Name(js.Name)

	if len(js.Methods) > 0 {
		route.Methods(js.Methods...)
	}

	if len(js.Schemes) > 0 {
		route.Schemes(js.Schemes...)
	}

	if js.Host != "" {
		route.Host(js.Host)
	}

	if js.Path != "" {
		route.Path(js.Path)
	}

	if js.PathPrefix != "" {
		route.PathPrefix(js.PathPrefix)
	}

	if len(js.Headers) > 0 {
		var pairs []string
		for k, v := range js.Headers {
			pairs = append(pairs, k, v)
		}
		route.Headers(pairs...)
	}

	if len(js.Queries) > 0 {
		var pairs []string
		for k, v := range js.Queries {
			pairs = append(pairs, k, v)
		}
		route.Queries(pairs...)
	}

	return nil
}
