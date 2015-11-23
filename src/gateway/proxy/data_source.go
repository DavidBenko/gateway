package proxy

import (
	"gateway/config"
	"gateway/model"
	apsql "gateway/sql"

	"log"
	"sync"
)

type proxyDataSource interface {
	Endpoint(id int64) (*model.ProxyEndpoint, error)
	Libraries(apiID int64) ([]*model.Library, error)
}

type endpointPassthrough struct {
	db *apsql.DB
}

func newPassthroughProxyDataSource(db *apsql.DB) *endpointPassthrough {
	return &endpointPassthrough{db: db}
}

func (c *endpointPassthrough) Endpoint(id int64) (*model.ProxyEndpoint, error) {
	return model.FindProxyEndpointForProxy(c.db, id)
}

func (c *endpointPassthrough) Libraries(apiID int64) ([]*model.Library, error) {
	return model.AllLibrariesForProxy(c.db, apiID)
}

type endpointCache struct {
	db *apsql.DB

	mutex sync.RWMutex

	endpointIDs map[int64][]int64              //      apiID -> []endpointID
	endpoints   map[int64]*model.ProxyEndpoint // endpointID -> endpoint
	libraries   map[int64][]*model.Library     //      apiID -> []library
}

func newCachingProxyDataSource(db *apsql.DB) *endpointCache {
	cache := &endpointCache{db: db}
	cache.endpointIDs = make(map[int64][]int64)
	cache.endpoints = make(map[int64]*model.ProxyEndpoint)
	cache.libraries = make(map[int64][]*model.Library)
	db.RegisterListener(cache)
	return cache
}

func (c *endpointCache) Endpoint(id int64) (*model.ProxyEndpoint, error) {
	c.mutex.RLock()
	endpoint := c.endpoints[id]
	c.mutex.RUnlock()
	if endpoint != nil {
		return endpoint, nil
	}

	endpoint, err := model.FindProxyEndpointForProxy(c.db, id)
	if err != nil {
		return nil, err
	}

	c.mutex.Lock()
	c.endpoints[id] = endpoint
	c.endpointIDs[endpoint.APIID] = append(c.endpointIDs[endpoint.APIID], id)
	c.mutex.Unlock()

	return endpoint, nil
}

func (c *endpointCache) Libraries(apiID int64) ([]*model.Library, error) {
	c.mutex.RLock()
	libraries := c.libraries[apiID]
	c.mutex.RUnlock()
	if libraries != nil {
		return libraries, nil
	}

	libraries, err := model.AllLibrariesForProxy(c.db, apiID)
	if err != nil {
		return nil, err
	}

	c.mutex.Lock()
	c.libraries[apiID] = libraries
	c.mutex.Unlock()

	return libraries, nil
}

func (c *endpointCache) clearAPI(apiID int64) {
	log.Printf("%s Clearing API %d cache", config.System, apiID)

	del := false
	c.mutex.RLock()
	_, del = c.endpointIDs[apiID]
	if !del {
		_, del = c.libraries[apiID]
	}
	c.mutex.RUnlock()
	if !del {
		return
	}

	c.mutex.Lock()
	ids := c.endpointIDs[apiID]
	if ids != nil {
		for _, id := range c.endpointIDs[apiID] {
			delete(c.endpoints, id)
		}
	}
	delete(c.endpointIDs, apiID)
	delete(c.libraries, apiID)
	c.mutex.Unlock()
}

func (c *endpointCache) clearAll() {
	log.Printf("%s Clearing all API caches", config.System)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.endpointIDs = make(map[int64][]int64)
	c.endpoints = make(map[int64]*model.ProxyEndpoint)
	c.libraries = make(map[int64][]*model.Library)
}

func (c *endpointCache) Notify(n *apsql.Notification) {
	switch {
	case n.Table == "apis" && n.Event == apsql.Delete:
		fallthrough
	case n.Table == "environments" && (n.Event == apsql.Update || n.Event == apsql.Delete):
		fallthrough
	case n.Table == "libraries":
		fallthrough
	case n.Table == "proxy_endpoint_schemas":
		fallthrough
	case n.Table == "remote_endpoints" && (n.Event == apsql.Update || n.Event == apsql.Delete):
		fallthrough
	case n.Table == "proxy_endpoints":
		go c.clearAPI(n.APIID)
	}
}

func (c *endpointCache) Reconnect() {
	log.Printf("%s API cache notified of database reconnection", config.System)
	go c.clearAll()
}
