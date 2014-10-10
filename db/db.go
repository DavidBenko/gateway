package db

import "github.com/AnyPresence/gateway/model"

// DB defines the interface of a backing datastore.
type DB interface {
	// ProxyEndpoint
	ListProxyEndpoints() ([]model.ProxyEndpoint, error)
	CreateProxyEndpoint(endpoint model.ProxyEndpoint) error
	GetProxyEndpointByName(name string) (model.ProxyEndpoint, error)
	GetProxyEndpointByPath(path string) (model.ProxyEndpoint, error)
	UpdateProxyEndpoint(endpoint model.ProxyEndpoint) error
	DeleteProxyEndpointByName(name string) error
}
