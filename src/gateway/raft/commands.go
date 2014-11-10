package raft

import (
	"fmt"

	"gateway/db"
	"gateway/model"

	goraft "github.com/goraft/raft"
)

func init() {
	goraft.RegisterCommand(&UpdateRouterCommand{})

	goraft.RegisterCommand(&EndpointDBCommand{})
	goraft.RegisterCommand(&LibraryDBCommand{})
}

// UpdateRouterCommand is a DBCommand to modify Endpoint instances.
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

// EndpointDBCommand is a DBCommand to modify Endpoint instances.
type EndpointDBCommand struct {
	DBCommand `json:"command"`
	Instance  *model.Endpoint `json:"instance"`
}

// NewEndpointDBCommand returns a new command to execute with the proxy endpoint.
func NewEndpointDBCommand(action DBWriteAction, instance *model.Endpoint) *EndpointDBCommand {
	return &EndpointDBCommand{DBCommand: DBCommand{Action: action}, Instance: instance}
}

// CommandName is the name of the command in the Raft log.
func (c *EndpointDBCommand) CommandName() string {
	return "Endpoint"
}

// Apply runs the DB action against the data store.
func (c *EndpointDBCommand) Apply(server goraft.Server) (interface{}, error) {
	return c.DBCommand.Apply(server, c.Instance)
}

// LibraryDBCommand is a DBCommand to modify Library instances.
type LibraryDBCommand struct {
	DBCommand `json:"command"`
	Instance  *model.Library `json:"instance"`
}

// NewLibraryDBCommand returns a new command to execute with the proxy endpoint.
func NewLibraryDBCommand(action DBWriteAction, instance *model.Library) *LibraryDBCommand {
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
