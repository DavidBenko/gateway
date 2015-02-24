package model

import "encoding/json"

// GetRoutes parses and returns the endpoint routes
func (e *ProxyEndpoint) GetRoutes() ([]*ProxyEndpointRoute, error) {
	var routes []*ProxyEndpointRoute
	err := json.Unmarshal(e.Routes, &routes)
	return routes, err
}

// ProxyEndpointRoute is a route on which the endpoint should be accessible.
type ProxyEndpointRoute struct {
	Path    string   `json:"path"`
	Methods []string `json:"methods"`
}

// Validate validates the model.
func (r *ProxyEndpointRoute) Validate() Errors {
	errors := make(Errors)
	if r.Path == "" {
		errors.add("path", "must not be blank")
	}
	if len(r.Methods) == 0 {
		errors.add("methods", "must not be empty")
	}
	return errors
}

// HandlesOptions returns whether this route handles the OPTIONS method explicitly.
func (r *ProxyEndpointRoute) HandlesOptions() bool {
	for _, method := range r.Methods {
		if method == "OPTIONS" {
			return true
		}
	}
	return false
}
