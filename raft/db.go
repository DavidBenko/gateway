package raft

import (
	"github.com/AnyPresence/gateway/db"
	"github.com/AnyPresence/gateway/model"
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

// CreateProxyEndpoint asks the Raft server to create a proxy endpoint.
func (db *DB) CreateProxyEndpoint(endpoint model.ProxyEndpoint) error {
	if _, err := db.raft.Do(CreateProxyEndpointCommand(endpoint)); err != nil {
		return err
	}
	return nil
}

// ListProxyEndpoints returns all the model.ProxyEndpoint instances in
// the data store.
func (db *DB) ListProxyEndpoints() ([]model.ProxyEndpoint, error) {
	return db.backingDB.ListProxyEndpoints()
}

// GetProxyEndpointByName fetches a model.ProxyEndpoint based on its name.
func (db *DB) GetProxyEndpointByName(name string) (model.ProxyEndpoint, error) {
	return db.backingDB.GetProxyEndpointByName(name)
}

// GetProxyEndpointByPath fetches a model.ProxyEndpoint based on its path.
func (db *DB) GetProxyEndpointByPath(path string) (model.ProxyEndpoint, error) {
	return db.backingDB.GetProxyEndpointByPath(path)
}

// UpdateProxyEndpoint asks the Raft server to update a proxy endpoint
func (db *DB) UpdateProxyEndpoint(endpoint model.ProxyEndpoint) error {
	if _, err := db.raft.Do(UpdateProxyEndpointCommand(endpoint)); err != nil {
		return err
	}
	return nil
}

// DeleteProxyEndpointByName asks the Raft server to delete a proxy endpoint
func (db *DB) DeleteProxyEndpointByName(name string) error {
	if _, err := db.raft.Do(DeleteProxyEndpointByNameCommand(name)); err != nil {
		return err
	}
	return nil
}
