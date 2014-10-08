package command

import (
	"github.com/AnyPresence/gateway/db"
	"github.com/AnyPresence/gateway/model"
	"github.com/goraft/raft"
)

// CreateProxyEndpointCommand creates a ProxyEndpoint.
type CreateProxyEndpointCommand struct {
	Endpoint model.ProxyEndpoint `json:"endpoint"`
}

// CreateProxyEndpoint returns a new CreateProxyEndpointCommand.
func CreateProxyEndpoint(endpoint model.ProxyEndpoint) *CreateProxyEndpointCommand {
	return &CreateProxyEndpointCommand{Endpoint: endpoint}
}

// CommandName is the name of the command in the Raft log.
func (c *CreateProxyEndpointCommand) CommandName() string {
	return "create_proxy_endpoint"
}

// Apply creates the ProxyEndpoint in the data store.
func (c *CreateProxyEndpointCommand) Apply(server raft.Server) (interface{}, error) {
	db := server.Context().(db.DB)
	return nil, db.CreateProxyEndpoint(c.Endpoint)
}
