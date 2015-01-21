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
