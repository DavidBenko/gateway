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

type KeyStore struct {
	cache     cache.Cacher
	nameIDMap map[string]int64
	db        *apsql.DB
	sync.RWMutex
}

type KeyDataSourceCriteria struct {
	AccountID int64
	Name      string
}

type storedKey struct {
	Parsed interface{}
	Key    *model.Key
}

func NewKeyStore(db *apsql.DB, cacheSize int) *KeyStore {
	nim := make(map[string]int64)
	evictFn := func(key, value interface{}) {
		stored := value.(*storedKey)
		delete(nim, cacheName(stored.Key.AccountID, stored.Key.Name))
	}

	return &KeyStore{
		db:        db,
		cache:     cache.NewLRUCache(cacheSize, evictFn),
		nameIDMap: nim,
	}
}

func (k *KeyStore) Get(criteria interface{}) (interface{}, bool) {
	c := criteria.(*KeyDataSourceCriteria)
	k.RLock()
	// Check if we can map the accountID and name to a key ID
	if id, ok := k.nameIDMap[cacheName(c.AccountID, c.Name)]; ok {
		if value, ok := k.cache.Get(id); ok {
			k.RUnlock()
			return value.(*storedKey).Parsed, ok
		}
	}
	k.RUnlock()

	key, e := model.FindKeyByAccountIdAndName(c.AccountID, c.Name, k.db)
	if e != nil {
		logreport.Printf("%s Error getting key %s: %s\n", config.Admin, c.Name, e)
		return nil, false
	}

	if key != nil {
		validKey, err := key.GetParsedKey()
		if err != nil {
			logreport.Printf("%s Error parsing key %s: %s\n", config.Admin, c.Name, err.Error())
			return nil, false
		}

		k.put(&storedKey{Parsed: validKey, Key: key})

		return validKey, true
	}
	return nil, false
}

func (k *KeyStore) put(storedKey *storedKey) {
	k.Lock()
	defer k.Unlock()
	k.cache.Add(storedKey.Key.ID, storedKey)
	k.nameIDMap[cacheName(storedKey.Key.AccountID, storedKey.Key.Name)] = storedKey.Key.ID
}

//Listener interface

// Notifiy satisfies the Listener interface.
func (k *KeyStore) Notify(n *apsql.Notification) {
	if n.Table != "keys" {
		return
	}

	switch n.Event {
	case apsql.Delete:
		k.Lock()
		defer k.Unlock()
		id := n.Messages[0]
		k.cache.Remove(id)
	}
}

// Reconnect satisfies the Listener interface.
func (k *KeyStore) Reconnect() {
	k.Lock()
	defer k.Unlock()
	k.cache.Purge()
}

func cacheName(accountID int64, name string) string {
	return fmt.Sprintf("%d:%s", accountID, name)
}
