package db

import (
	"fmt"
	"sync"

	"github.com/AnyPresence/gateway/model"
)

// Memory implements an in-memory DB.
type Memory struct {
	mutex sync.RWMutex

	storage map[string]map[interface{}]model.Model
}

// NewMemoryStore creates a new Memory data store.
func NewMemoryStore() *Memory {
	return &Memory{
		storage: make(map[string]map[interface{}]model.Model),
	}
}

func (db *Memory) subMap(m model.Model) map[interface{}]model.Model {
	key := m.CollectionName()
	subMap, ok := db.storage[key]
	if !ok {
		subMap = make(map[interface{}]model.Model)
		db.storage[key] = subMap
	}
	return subMap
}

// List returns all instances that share the type with the model.
func (db *Memory) List(instance model.Model) ([]interface{}, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	subMap := db.subMap(instance)
	list := make([]interface{}, 0, len(subMap))
	for _, value := range subMap {
		list = append(list, value)
	}
	return list, nil
}

// Insert stores the instance in the data store.
// Returns an error if the instance is already in the store; Update instead.
func (db *Memory) Insert(instance model.Model) error {
	if _, err := db.Get(instance, instance.ID()); err == nil {
		return fmt.Errorf("There is already a %s with ID() '%s'",
			instance.CollectionName(), instance.ID())
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	subMap := db.subMap(instance)
	subMap[instance.ID()] = instance
	return nil
}

// Get fetches an instance based on its ID().
func (db *Memory) Get(m model.Model, id interface{}) (model.Model, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	subMap := db.subMap(m)
	instance, ok := subMap[id]
	if ok {
		return instance, nil
	}
	return m, fmt.Errorf("There are no %s with id '%v'", m.CollectionName(), id)
}

// Update updates the instance in the data store.
// Returns an error if the instance is not already in the store; Insert instead.
func (db *Memory) Update(instance model.Model) error {
	if _, err := db.Get(instance, instance.ID()); err != nil {
		return err
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	subMap := db.subMap(instance)
	subMap[instance.ID()] = instance
	return nil
}

// Delete deletes the instance from the data store.
// Returns an error if the instance is not already in the store.
func (db *Memory) Delete(m model.Model, id interface{}) error {
	if _, err := db.Get(m, id); err != nil {
		return err
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	subMap := db.subMap(m)
	delete(subMap, id)
	return nil
}
