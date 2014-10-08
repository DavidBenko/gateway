package main

import (
	"fmt"
	"log"

	"os"

	"github.com/AnyPresence/gateway/command"
	"github.com/AnyPresence/gateway/config"
	"github.com/AnyPresence/gateway/db"
	"github.com/AnyPresence/gateway/proxy"
	"github.com/AnyPresence/gateway/raft"
)

func main() {
	conf, err := config.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(fmt.Sprintf("Error parsing config file: %v", err))
	}

	log.Print("Registering Raft commands")
	command.RegisterCommands()

	log.Print("Starting Raft server")
	raft := raft.NewServer(conf.Raft, db.NewMemoryStore())
	raft.Setup()
	go raft.Run()

	log.Print("Starting proxy server")
	proxy := proxy.NewServer(conf.Proxy, raft.RaftServer)
	go proxy.Run()

	select {}
}
