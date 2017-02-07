package vm

import (
	"fmt"
	"gateway/core/cache"
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
	APIID     int64
	Codename  string
}

func NewRemoteEndpointStore(db *apsql.DB, cacheSize int) *RemoteEndpointStore {
	cim := make(map[string]int64)
	evictFn := func(key, value interface{}) {
		stored := value.(*model.RemoteEndpoint)
		delete(cim, cacheKey(stored.AccountID, stored.APIID, stored.Codename))
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

	if id, hasKey := r.codenameIDMap[cacheKey(c.AccountID, c.APIID, c.Codename)]; hasKey {
		if value, ok := r.cache.Get(id); ok {
			r.RUnlock()
			return value, true
		}
	}
	r.RUnlock()

	endpoint, err := model.FindRemoteEndpointForAccountIDApiIDAndCodename(r.db, c.AccountID, c.APIID, c.Codename)
	if err != nil {
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
	r.codenameIDMap[cacheKey(endpoint.AccountID, endpoint.APIID, endpoint.Codename)] = endpoint.ID
}

func cacheKey(accountID, APIID int64, codename string) string {
	return fmt.Sprintf("%d:%d:%s", accountID, APIID, codename)
}

func (r *RemoteEndpointStore) Notify(n *apsql.Notification) {
	if n.Table != "remote_endpoints" {
		return
	}

	switch n.Event {
	case apsql.Update:
		fallthrough
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
