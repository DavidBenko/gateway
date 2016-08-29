package core

import (
	"gateway/logreport"
	"gateway/model"
	apsql "gateway/sql"
	"sync"
)

type KeyStore struct {
	Keys map[int64]map[string]*StoredKey
	db   *apsql.DB
	sync.RWMutex
}

type StoredKey struct {
	Key interface{}
	ID  int64
}

func NewKeyStore(db *apsql.DB) *KeyStore {
	keys := make(map[int64]map[string]*StoredKey)
	return &KeyStore{Keys: keys, db: db}
}

func (k *KeyStore) GetKey(accountID int64, name string) (interface{}, bool) {
	k.RLock()

	if accountKeys, ok := k.Keys[accountID]; ok {
		if key, exists := accountKeys[name]; exists {
			k.RUnlock()
			return key.Key, true
		}
	}
	k.RUnlock()

	key, e := model.FindKeyByAccountIdAndName(accountID, name, k.db)
	if e != nil {
		logreport.Println(e)
		return nil, false
	}

	if key != nil {
		validKey, err := key.GetParsedKey()
		if err != nil {
			logreport.Printf("\n[keys] error parsing key %s: %s\n", name, err.Error())
			return nil, false
		}

		k.PutKey(accountID, name, &StoredKey{Key: validKey, ID: key.ID})

		return validKey, true
	}
	return nil, false
}

// PutKey stores a new key in the KeyStore. Replaces an existing key if present.
// Method is threadsafe.
func (k *KeyStore) PutKey(accountID int64, name string, key *StoredKey) {
	k.Lock()
	defer k.Unlock()
	if _, ok := k.Keys[accountID]; !ok {
		k.Keys[accountID] = make(map[string]*StoredKey)
	}

	k.Keys[accountID][name] = key
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

		keyID := n.Messages[0]

		if accountKeys, ok := k.Keys[n.AccountID]; ok {
			for k, key := range accountKeys {
				if key.ID == keyID {
					delete(accountKeys, k)
					break
				}
			}
		}
	}
}

// Reconnect satisfies the Listener interface.
func (k *KeyStore) Reconnect() {
	k.Lock()
	defer k.Unlock()

	// Just recreate the Keys map and let it rebuild organically.
	new := make(map[int64]map[string]*StoredKey)
	k.Keys = new
}
