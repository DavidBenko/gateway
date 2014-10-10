package raft

import goraft "github.com/goraft/raft"

// RegisterCommands registers all commands with the Raft implementation.
func RegisterCommands() {
	goraft.RegisterCommand(&createProxyEndpointCommand{})
	goraft.RegisterCommand(&updateProxyEndpointCommand{})
	goraft.RegisterCommand(&deleteProxyEndpointByNameCommand{})
}
