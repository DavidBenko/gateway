package model

// Library holds reusable scripts for the proxy.
type Library struct {
	Name   string
	Script string
}

// ProxyEndpoint represents an endpoint that the Gateway should handle.
type ProxyEndpoint struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Method string `json:"method"`
	Script string `json:"script"`
}

// Service represents a remote service the Gateway has access to.
type Service struct {
	Name string
}

// ServiceResource represents a resource on a remote Service.
type ServiceResource struct {
	Name string
}
