package db

import "github.com/AnyPresence/gateway/model"

// DB defines the interface of a backing datastore.
type DB interface {
	// ProxyEndpoint
	CreateProxyEndpoint(endpoint model.ProxyEndpoint) error
	GetProxyEndpointByName(name string) (model.ProxyEndpoint, error)
	GetProxyEndpointByPath(path string) (model.ProxyEndpoint, error)
}
