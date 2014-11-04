package raft

import (
	"log"
	"reflect"

	"gateway/db"
	"gateway/model"

	"github.com/goraft/raft"
)

// DB is a Raft-aware backing datastore
type DB struct {
	backingDB db.DB
	raft      raft.Server
}

// NewRaftDB creates a new Raft-aware datastore based on another data store.
func NewRaftDB(backingDB db.DB, server raft.Server) *DB {
	return &DB{
		backingDB: backingDB,
		raft:      server,
	}
}

// Router returns this database's singleton router.
func (db *DB) Router() model.Router {
	return db.backingDB.Router()
}

// UpdateRouter creates a new router from the passed script.
func (db *DB) UpdateRouter(script string) (model.Router, error) {
	r, err := db.raft.Do(NewUpdateRouterCommand(script))
	return r.(model.Router), err
}

// NextID returns the next id to use when inserting a model.
func (db *DB) NextID(m model.Model) interface{} {
	return db.backingDB.NextID(m)
}

// List returns all instances in the data store.
func (db *DB) List(instance model.Model) ([]model.Model, error) {
	return db.backingDB.List(instance)
}

// Insert asks the Raft server to insert a persisted instance.
func (db *DB) Insert(instance model.Model) error {
	if _, err := db.raft.Do(newCommand(Insert, instance)); err != nil {
		return err
	}
	return nil
}

// Get fetches a model instance based on its name.
func (db *DB) Get(m model.Model, id interface{}) (model.Model, error) {
	return db.backingDB.Get(m, id)
}

// Find finds a model instance based on an indexed field.
func (db *DB) Find(m model.Model, findByFieldName string, id interface{}) (model.Model, error) {
	return db.backingDB.Find(m, findByFieldName, id)
}

// Update asks the Raft server to update a persisted instance.
func (db *DB) Update(instance model.Model) error {
	if _, err := db.raft.Do(newCommand(Update, instance)); err != nil {
		return err
	}
	return nil
}

// Delete asks the Raft server to delete a persisted instance.
func (db *DB) Delete(m model.Model, id interface{}) error {
	instance, err := db.backingDB.Get(m, id)
	if err != nil {
		return err
	}
	if _, err := db.raft.Do(newCommand(Delete, instance)); err != nil {
		return err
	}
	return nil
}

func newCommand(action DBWriteAction, instance model.Model) raft.Command {
	switch instance := instance.(type) {
	case *model.ProxyEndpoint:
		return NewProxyEndpointDBCommand(action, instance)
	case *model.Library:
		return NewLibraryDBCommand(action, instance)
	}
	log.Fatalf("Could not create DB write command for instance of type %s",
		reflect.TypeOf(instance))
	return raft.NOPCommand{}
}
