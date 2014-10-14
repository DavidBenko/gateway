package db

import (
	"fmt"
	"reflect"
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
	if _, err := db.Get(instance, instance.ID()); err == nil {
		return fmt.Errorf("There is already a %s with ID() '%s'",
			instance.CollectionName(), instance.ID())
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	subMap := db.subMap(instance)
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
	var oldInstance model.Model
	oldInstance, err := db.Get(instance, instance.ID())
	if err != nil {
		return err
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	subMap := db.subMap(instance)
	subMap[instance.ID()] = instance
	db.updateIndices(oldInstance, instance)

	return nil
}

// Delete deletes the instance from the data store.
// Returns an error if the instance is not already in the store.
func (db *Memory) Delete(m model.Model, id interface{}) error {
	var oldInstance model.Model
	oldInstance, err := db.Get(m, id)
	if err != nil {
		return err
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	subMap := db.subMap(m)
	delete(subMap, id)
	db.deleteIndices(oldInstance)

	return nil
}

func (db *Memory) find(
	m model.Model,
	subMap map[interface{}]model.Model,
	keyName string,
	key interface{},
) (model.Model, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	instance, ok := subMap[key]
	if ok {
		return instance, nil
	}
	return m, fmt.Errorf("There are no %s with %s '%v'", m.CollectionName(), keyName, key)
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
	t := reflect.TypeOf(instance)
	n := t.NumField()
	for i := 0; i < n; i++ {
		field := t.Field(i)
		indexed := field.Tag.Get(fieldTagIndexed) == fieldTagIndexedTrue
		if indexed {
			value := reflect.ValueOf(instance).FieldByName(field.Name)

			// TODO: For non-string indexed fields, this will fail
			command(field.Name, value.String())
		}
	}
}
