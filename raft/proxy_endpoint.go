package raft

import (
	"github.com/AnyPresence/gateway/db"
	"github.com/AnyPresence/gateway/model"
	goraft "github.com/goraft/raft"
)

// CreateProxyEndpointCommand creates a ProxyEndpoint.
type createProxyEndpointCommand struct {
	Endpoint model.ProxyEndpoint `json:"endpoint"`
}

// CreateProxyEndpointCommand returns a new CreateProxyEndpointCommand.
func CreateProxyEndpointCommand(endpoint model.ProxyEndpoint) *createProxyEndpointCommand {
	return &createProxyEndpointCommand{Endpoint: endpoint}
}

// CommandName is the name of the command in the Raft log.
func (c *createProxyEndpointCommand) CommandName() string {
	return "create_proxy_endpoint"
}

// Apply creates the ProxyEndpoint in the data store.
func (c *createProxyEndpointCommand) Apply(server goraft.Server) (interface{}, error) {
	db := server.Context().(db.DB)
	return nil, db.CreateProxyEndpoint(c.Endpoint)
}

// UpdateProxyEndpointCommand updates a ProxyEndpoint.
type updateProxyEndpointCommand struct {
	Endpoint model.ProxyEndpoint `json:"endpoint"`
}

// UpdateProxyEndpointCommand returns a new UpdateProxyEndpointCommand.
func UpdateProxyEndpointCommand(endpoint model.ProxyEndpoint) *updateProxyEndpointCommand {
	return &updateProxyEndpointCommand{Endpoint: endpoint}
}

// CommandName is the name of the command in the Raft log.
func (c *updateProxyEndpointCommand) CommandName() string {
	return "update_proxy_endpoint"
}

// Apply updates the ProxyEndpoint in the data store.
func (c *updateProxyEndpointCommand) Apply(server goraft.Server) (interface{}, error) {
	db := server.Context().(db.DB)
	return nil, db.UpdateProxyEndpoint(c.Endpoint)
}

// DeleteProxyEndpointCommand deletes a ProxyEndpoint.
type deleteProxyEndpointByNameCommand struct {
	Name string `json:"name"`
}

// DeleteProxyEndpointByNameCommand returns a new DeleteProxyEndpointCommand.
func DeleteProxyEndpointByNameCommand(name string) *deleteProxyEndpointByNameCommand {
	return &deleteProxyEndpointByNameCommand{Name: name}
}

// CommandName is the name of the command in the Raft log.
func (c *deleteProxyEndpointByNameCommand) CommandName() string {
	return "delete_proxy_endpoint"
}

// Apply deletes the ProxyEndpoint in the data store.
func (c *deleteProxyEndpointByNameCommand) Apply(server goraft.Server) (interface{}, error) {
	db := server.Context().(db.DB)
	return nil, db.DeleteProxyEndpointByName(c.Name)
}
