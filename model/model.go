package model

// Library holds reusable scripts for the proxy.
type Library struct {
	Name   string
	Script string
}

// ProxyEndpoint represents an endpoint that the Gateway should handle.
type ProxyEndpoint struct {
	IDField int64  `json:"id"`
	Name    string `json:"name"`
	Path    string `json:"path" index:"true"`
	Method  string `json:"method"`
	Script  string `json:"script"`
}

// CollectionName returns a system-unique name that can be used to reference
// collections of this model, e.g. for URLs or database table names.
func (p ProxyEndpoint) CollectionName() string {
	return "proxy_endpoints"
}

// ID returns the id of the endpoint.
func (p ProxyEndpoint) ID() interface{} {
	return p.IDField
}

// Service represents a remote service the Gateway has access to.
type Service struct {
	Name string
}

// ServiceResource represents a resource on a remote Service.
type ServiceResource struct {
	Name string
}
