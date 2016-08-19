package proxy

import (
	"gateway/config"
	"gateway/logreport"
	"gateway/model"
	apsql "gateway/sql"

	"sync"
)

type proxyDataSource interface {
	Endpoint(id int64) (*model.ProxyEndpoint, error)
	Libraries(apiID int64) ([]*model.Library, error)
	Plan(accountID int64) (*model.Plan, error)
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

func (c *endpointPassthrough) Plan(accountID int64) (*model.Plan, error) {
	return model.FindPlanByAccountID(c.db, accountID)
}

type endpointCache struct {
	db *apsql.DB

	mutex sync.RWMutex

	endpointIDs map[int64][]int64              //      apiID -> []endpointID
	endpoints   map[int64]*model.ProxyEndpoint // endpointID -> endpoint
	libraries   map[int64][]*model.Library     //      apiID -> []library
	plans       map[int64]*model.Plan          // accountID -> plan
	accountIDs  map[int64][]int64              //      planID -> []accountID
}

func newCachingProxyDataSource(db *apsql.DB) *endpointCache {
	cache := &endpointCache{
		db:          db,
		endpointIDs: make(map[int64][]int64),
		endpoints:   make(map[int64]*model.ProxyEndpoint),
		libraries:   make(map[int64][]*model.Library),
		plans:       make(map[int64]*model.Plan),
		accountIDs:  make(map[int64][]int64),
	}

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

func (c *endpointCache) Plan(accountID int64) (*model.Plan, error) {
	c.mutex.RLock()
	plan := c.plans[accountID]
	c.mutex.RUnlock()
	if plan != nil {
		return plan, nil
	}

	plan, err := model.FindPlanByAccountID(c.db, accountID)
	if err != nil {
		return nil, err
	}

	c.mutex.Lock()
	c.plans[accountID] = plan
	c.accountIDs[plan.ID] = append(c.accountIDs[plan.ID], accountID)
	c.mutex.Unlock()

	return plan, nil
}

func (c *endpointCache) clearAPI(apiID int64) {
	logreport.Printf("%s Clearing API %d cache", config.System, apiID)

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

func (c *endpointCache) clearAccount(accountID int64) {
	logreport.Printf("%s Clearing Account %d cache", config.System, accountID)

	del := false
	c.mutex.RLock()
	plan, del := c.plans[accountID]
	c.mutex.RUnlock()
	if !del {
		return
	}

	c.mutex.Lock()

	for index, id := range c.accountIDs[plan.ID] {
		if id == accountID {
			c.accountIDs[plan.ID][index] = c.accountIDs[plan.ID][len(c.accountIDs[plan.ID])-1]
			c.accountIDs[plan.ID] = c.accountIDs[plan.ID][:len(c.accountIDs[plan.ID])-1]
		}
	}
	delete(c.plans, accountID)

	c.mutex.Unlock()
}

func (c *endpointCache) clearPlan(planID int64) {
	logreport.Printf("%s Clearing Plan %d cache", config.System, planID)

	del := false
	c.mutex.RLock()
	accountIDs, del := c.accountIDs[planID]
	c.mutex.RUnlock()
	if !del {
		return
	}

	c.mutex.Lock()
	if accountIDs != nil {
		for _, accountID := range accountIDs {
			delete(c.plans, accountID)
		}
	}
	delete(c.accountIDs, planID)

	c.mutex.Unlock()
}

func (c *endpointCache) clearAll() {
	logreport.Printf("%s Clearing all API caches", config.System)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.endpointIDs = make(map[int64][]int64)
	c.endpoints = make(map[int64]*model.ProxyEndpoint)
	c.libraries = make(map[int64][]*model.Library)
	c.plans = make(map[int64]*model.Plan)
	c.accountIDs = make(map[int64][]int64)
}

func (c *endpointCache) Notify(n *apsql.Notification) {
	switch {
	case n.Table == "accounts":
		go c.clearAccount(n.AccountID)
	case n.Table == "plans":
		go c.clearPlan(n.ID)
	case n.Table == "apis" && n.Event == apsql.Delete:
		fallthrough
	case n.Table == "environments" && (n.Event == apsql.Update || n.Event == apsql.Delete):
		fallthrough
	case n.Table == "libraries":
		fallthrough
	case n.Table == "proxy_endpoint_schemas":
		fallthrough
	case n.Table == "proxy_endpoint_components" && (n.Event == apsql.Update || n.Event == apsql.Delete):
		fallthrough
	case n.Table == "remote_endpoints" && (n.Event == apsql.Update || n.Event == apsql.Delete):
		fallthrough
	case n.Table == "proxy_endpoints":
		go c.clearAPI(n.APIID)
	}
}

func (c *endpointCache) Reconnect() {
	logreport.Printf("%s API cache notified of database reconnection", config.System)
	go c.clearAll()
}
