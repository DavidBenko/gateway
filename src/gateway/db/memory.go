package db

import (
	"fmt"
	"reflect"
	"sync"

	"gateway/model"
)

// Memory implements an in-memory DB.
type Memory struct {
	mutex sync.RWMutex

	storage map[string]map[interface{}]model.Model

	router model.Router
}

// NewMemoryStore creates a new Memory data store.
func NewMemoryStore() *Memory {
	return &Memory{
		storage: make(map[string]map[interface{}]model.Model),
	}
}

// Router returns this database's singleton router
func (db *Memory) Router() model.Router {
	return db.router
}

// UpdateRouter creates a new router from the passed script.
func (db *Memory) UpdateRouter(script string) (model.Router, error) {
	r := model.Router{Script: script}
	err := r.ParseRoutes()
	if err != nil {
		return r, err
	}
	db.router = r
	return r, nil
}

func (db *Memory) subMap(m model.Model) map[interface{}]model.Model {
	return db.subMapWithSuffix(m, "")
}

func (db *Memory) subMapForFieldName(m model.Model, fieldName string) map[interface{}]model.Model {
	return db.subMapWithSuffix(m, "_"+fieldName)
}

func (db *Memory) subMapWithSuffix(m model.Model, suffix string) map[interface{}]model.Model {
	key := m.CollectionName() + suffix
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
	db.mutex.Lock()
	defer db.mutex.Unlock()

	subMap := db.subMap(instance)
	if _, err := db.find(instance, subMap, "id", instance.ID()); err == nil {
		return fmt.Errorf("There is already a %s with ID() '%s'",
			instance.CollectionName(), instance.ID())
	}

	if ok, name, val := db.passesUniqueConstraints(instance); !ok {
		return fmt.Errorf("There is already a %s with %s '%s'",
			instance.CollectionName(), name, val)
	}

	subMap[instance.ID()] = instance
	db.createIndices(instance)

	return nil
}

// Get fetches an instance based on its ID().
func (db *Memory) Get(m model.Model, id interface{}) (model.Model, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	subMap := db.subMap(m)
	return db.find(m, subMap, "id", id)
}

// Find finds an instance based on an indexed field.
func (db *Memory) Find(m model.Model, findByFieldName string, id interface{}) (model.Model, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	var subMap map[interface{}]model.Model
	db.doOnIndexedFields(m, func(indexedFieldName string, fieldValue string) {
		if indexedFieldName == findByFieldName {
			subMap = db.subMapForFieldName(m, indexedFieldName)
		}
	})
	return db.find(m, subMap, findByFieldName, id)
}

// Update updates the instance in the data store.
// Returns an error if the instance is not already in the store; Insert instead.
func (db *Memory) Update(instance model.Model) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	subMap := db.subMap(instance)

	oldInstance, err := db.find(instance, subMap, "id", instance.ID())
	if err != nil {
		return err
	}

	if ok, name, val := db.passesUniqueConstraints(instance); !ok {
		return fmt.Errorf("There is already a %s with %s '%s'",
			instance.CollectionName(), name, val)
	}

	subMap[instance.ID()] = instance
	db.updateIndices(oldInstance, instance)

	return nil
}

// Delete deletes the instance from the data store.
// Returns an error if the instance is not already in the store.
func (db *Memory) Delete(m model.Model, id interface{}) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	subMap := db.subMap(m)

	oldInstance, err := db.find(m, subMap, "id", id)
	if err != nil {
		return err
	}

	delete(subMap, id)
	db.deleteIndices(oldInstance)

	return nil
}

// find does a lookup on a specific submap, but does not lock.
// Calling functions must lock for safety.
func (db *Memory) find(
	m model.Model,
	subMap map[interface{}]model.Model,
	keyName string,
	key interface{},
) (model.Model, error) {
	instance, ok := subMap[key]
	if ok {
		return instance, nil
	}
	return m, fmt.Errorf("There are no %s with %s '%s'", m.CollectionName(), keyName, key)
}

func (db *Memory) createIndices(instance model.Model) {
	db.doOnIndexedFields(instance, func(fieldName string, fieldValue string) {
		indexedSubMap := db.subMapForFieldName(instance, fieldName)
		indexedSubMap[fieldValue] = instance
	})
}

func (db *Memory) deleteIndices(instance model.Model) {
	db.doOnIndexedFields(instance, func(fieldName string, fieldValue string) {
		indexedSubMap := db.subMapForFieldName(instance, fieldName)
		delete(indexedSubMap, fieldValue)
	})
}

func (db *Memory) updateIndices(old model.Model, new model.Model) {
	db.deleteIndices(old)
	db.createIndices(new)
}

func (db *Memory) doOnIndexedFields(
	instance model.Model,
	command func(fieldName string, fieldValue string),
) {
	db.doOnTaggedFields(instance, model.FieldTagIndexed, command)
}

func (db *Memory) passesUniqueConstraints(instance model.Model) (ok bool, name string, val interface{}) {
	ok = true
	db.doOnTaggedFields(instance, model.FieldTagUnique, func(fieldName string, fieldValue string) {
		indexedSubMap := db.subMapForFieldName(instance, fieldName)
		storedInstance, exists := indexedSubMap[fieldValue]
		if exists && (storedInstance.ID() != instance.ID()) {
			ok = false
			name = fieldName
			val = fieldValue
		}
	})
	return
}

func (db *Memory) doOnTaggedFields(
	instance model.Model,
	tagName string,
	command func(fieldName string, fieldValue string),
) {
	t := reflect.TypeOf(instance)
	n := t.NumField()
	for i := 0; i < n; i++ {
		field := t.Field(i)
		tagged := field.Tag.Get(tagName) != ""
		if tagged {
			value := reflect.ValueOf(instance).FieldByName(field.Name)

			// For non-string fields, the value may not be what you expect
			command(field.Name, value.String())
		}
	}
}
