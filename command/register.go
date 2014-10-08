package command

import "github.com/goraft/raft"

// RegisterCommands registers all commands with the Raft implementation.
func RegisterCommands() {
	raft.RegisterCommand(&CreateProxyEndpointCommand{})
}
