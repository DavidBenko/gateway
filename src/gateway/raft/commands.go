package raft

import (
	"fmt"
	"log"

	"gateway/db"
	"gateway/model"

	goraft "github.com/goraft/raft"
)

func init() {
	log.Print("Registering Raft commands")

	goraft.RegisterCommand(&UpdateRouterCommand{})

	goraft.RegisterCommand(&ProxyEndpointDBCommand{})
	goraft.RegisterCommand(&LibraryDBCommand{})
}

// UpdateRouterCommand is a DBCommand to modify ProxyEndpoint instances.
type UpdateRouterCommand struct {
	Script string `json:"script"`
}

// NewUpdateRouterCommand returns a command to update the router.
func NewUpdateRouterCommand(script string) *UpdateRouterCommand {
	return &UpdateRouterCommand{Script: script}
}

// CommandName is the name of the command in the Raft log.
func (c *UpdateRouterCommand) CommandName() string {
	return "UpdateRouter"
}

// Apply runs the update router action against the data store.
func (c *UpdateRouterCommand) Apply(server goraft.Server) (interface{}, error) {
	db := server.Context().(db.DB)
	return db.UpdateRouter(c.Script)
}

// DBWriteAction is a write action a data store can do.
type DBWriteAction string

const (
	// Insert actions insert items into the data store.
	Insert = "insert"
	// Update actions update items in the data store.
	Update = "update"
	// Delete actions delete items from the data store.
	Delete = "delete"
)

// DBCommand is a helper struct for Raft actions that modify the data store.
type DBCommand struct {
	Action DBWriteAction `json:"action"`
}

// Apply insterts the instance in the data store.
func (c *DBCommand) Apply(server goraft.Server, instance model.Model) (interface{}, error) {
	db := server.Context().(db.DB)
	switch c.Action {
	case Insert:
		return nil, db.Insert(instance)
	case Update:
		return nil, db.Update(instance)
	case Delete:
		return nil, db.Delete(instance, instance.ID())
	default:
		return nil, fmt.Errorf("Unknown action")
	}
}

// ProxyEndpointDBCommand is a DBCommand to modify ProxyEndpoint instances.
type ProxyEndpointDBCommand struct {
	DBCommand `json:"command"`
	Instance  model.ProxyEndpoint `json:"instance"`
}

// NewProxyEndpointDBCommand returns a new command to execute with the proxy endpoint.
func NewProxyEndpointDBCommand(action DBWriteAction, instance model.ProxyEndpoint) *ProxyEndpointDBCommand {
	return &ProxyEndpointDBCommand{DBCommand: DBCommand{Action: action}, Instance: instance}
}

// CommandName is the name of the command in the Raft log.
func (c *ProxyEndpointDBCommand) CommandName() string {
	return "ProxyEndpoint"
}

// Apply runs the DB action against the data store.
func (c *ProxyEndpointDBCommand) Apply(server goraft.Server) (interface{}, error) {
	return c.DBCommand.Apply(server, c.Instance)
}

// LibraryDBCommand is a DBCommand to modify Library instances.
type LibraryDBCommand struct {
	DBCommand `json:"command"`
	Instance  model.Library `json:"instance"`
}

// NewLibraryDBCommand returns a new command to execute with the proxy endpoint.
func NewLibraryDBCommand(action DBWriteAction, instance model.Library) *LibraryDBCommand {
	return &LibraryDBCommand{DBCommand: DBCommand{Action: action}, Instance: instance}
}

// CommandName is the name of the command in the Raft log.
func (c *LibraryDBCommand) CommandName() string {
	return "Library"
}

// Apply runs the DB action against the data store.
func (c *LibraryDBCommand) Apply(server goraft.Server) (interface{}, error) {
	return c.DBCommand.Apply(server, c.Instance)
}
