package model

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/mux"
	"github.com/robertkrimen/otto"
)

// Router holds the script that generates the proxy routes,
// as well as the parsed out mux.Router.
type Router struct {
	Script    string      `json:"script"`
	MUXRouter *mux.Router `json:"-"`
}

type jsRoute struct {
	Name    string   `json:"_name"`
	Path    string   `json:"_path"`
	Methods []string `json:"_methods"`
}

// ParseRoutes parses the script and stores the resulting the mux.Router.
func (r *Router) ParseRoutes() error {
	vm := otto.New()

	var files = []string{"route.js", "router.js"}
	var scripts = make([]interface{}, 0)
	for _, filename := range files {
		fileJS, err := Asset(filename)
		if err != nil {
			return err
		}
		scripts = append(scripts, fileJS)
	}
	scripts = append(scripts, r.Script)
	scripts = append(scripts, "router.routeData();")

	var result otto.Value
	for _, script := range scripts {
		var err error
		result, err = vm.Run(script)
		if err != nil {
			return err
		}
	}

	var jsRoutes []jsRoute
	if err := json.Unmarshal([]byte(result.String()), &jsRoutes); err != nil {
		return err
	}

	router := mux.NewRouter()
	for _, route := range jsRoutes {
		addRoute(route, router)
	}
	r.MUXRouter = router

	return nil
}

func addRoute(js jsRoute, router *mux.Router) error {
	route := router.NewRoute()

	if js.Name == "" {
		return fmt.Errorf("All routes must route to a proxy endpoint by name.")
	}
	route.Name(js.Name)

	if js.Path != "" {
		route.Path(js.Path)
	}

	if len(js.Methods) > 0 {
		route.Methods(js.Methods...)
	}

	return nil
}
