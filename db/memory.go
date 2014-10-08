package db

import (
	"fmt"
	"sync"

	"github.com/AnyPresence/gateway/model"
)

// Memory implements an in-memory DB.
type Memory struct {
	mutex sync.RWMutex

	proxyEndpoints       map[string]model.ProxyEndpoint
	proxyEndpointsByPath map[string]model.ProxyEndpoint
}

// NewMemoryStore creates a new Memory data store.
func NewMemoryStore() *Memory {
	return &Memory{
		proxyEndpoints:       make(map[string]model.ProxyEndpoint),
		proxyEndpointsByPath: make(map[string]model.ProxyEndpoint),
	}
}

// CreateProxyEndpoint stores the model.ProxyEndpoint in the data store.
func (db *Memory) CreateProxyEndpoint(endpoint model.ProxyEndpoint) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	db.proxyEndpoints[endpoint.Name] = endpoint
	db.proxyEndpointsByPath[endpoint.Path] = endpoint
	return nil
}

// GetProxyEndpointByName fetches a model.ProxyEndpoint based on its name.
func (db *Memory) GetProxyEndpointByName(name string) (model.ProxyEndpoint, error) {
	return db.getProxyEndpoint(db.proxyEndpoints, name)
}

// GetProxyEndpointByPath fetches a model.ProxyEndpoint based on its path.
func (db *Memory) GetProxyEndpointByPath(path string) (model.ProxyEndpoint, error) {
	return db.getProxyEndpoint(db.proxyEndpointsByPath, path)
}

func (db *Memory) getProxyEndpoint(m map[string]model.ProxyEndpoint, key string) (model.ProxyEndpoint, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	endpoint, ok := m[key]
	if ok {
		return endpoint, nil
	}
	return model.ProxyEndpoint{},
		fmt.Errorf("No proxy endpoint exists for key '%s'", key)
}
