package vm

import (
	"fmt"
	"gateway/config"
	"gateway/core/cache"
	"gateway/logreport"
	"gateway/model"
	apsql "gateway/sql"
	"sync"
)

type RemoteEndpointStore struct {
	cache         cache.Cacher
	codenameIDMap map[string]int64
	db            *apsql.DB
	sync.RWMutex
}

type RemoteEndpointStoreCriteria struct {
	AccountID int64
	Codename  string
}

func NewRemoteEndpointStore(db *apsql.DB, cacheSize int) *RemoteEndpointStore {
	cim := make(map[string]int64)
	evictFn := func(key, value interface{}) {
		stored := value.(*model.RemoteEndpoint)
		delete(cim, cacheKey(stored.AccountID, stored.Codename))
	}

	return &RemoteEndpointStore{
		db:            db,
		cache:         cache.NewLRUCache(cacheSize, evictFn),
		codenameIDMap: cim,
	}
}

func (r *RemoteEndpointStore) Get(criteria interface{}) (interface{}, bool) {
	c := criteria.(*RemoteEndpointStoreCriteria)
	r.RLock()

	if id, ok := r.codenameIDMap[cacheKey(c.AccountID, c.Codename)]; ok {
		if value, ok := r.cache.Get(id); ok {
			r.RUnlock()
			return value, ok
		}
	}
	r.RUnlock()

	endpoint, err := model.FindRemoteEndpointForAccountIDAndCodename(r.db, c.Codename, c.AccountID)
	if err != nil {
		logreport.Printf("%s Error getting remote endpoint %s: %s\n", config.Admin, c.Codename, err)
		return nil, false
	}

	if endpoint != nil {
		r.put(endpoint)

		return endpoint, true
	}
	return nil, false
}

func (r *RemoteEndpointStore) put(endpoint *model.RemoteEndpoint) {
	r.Lock()
	defer r.Unlock()
	r.cache.Add(endpoint.ID, endpoint)
	r.codenameIDMap[cacheKey(endpoint.AccountID, endpoint.Codename)] = endpoint.ID
}

func cacheKey(accountID int64, codename string) string {
	return fmt.Sprintf("%d:%s", accountID, codename)
}

func (r *RemoteEndpointStore) Notify(n *apsql.Notification) {
	if n.Table != "remote_endpoints" {
		return
	}

	switch n.Event {
	case apsql.Delete:
		r.Lock()
		defer r.Unlock()
		id := n.Messages[0]
		r.cache.Remove(id)
	}
}

func (r *RemoteEndpointStore) Reconnect() {
	r.Lock()
	defer r.Unlock()
	r.cache.Purge()
}
