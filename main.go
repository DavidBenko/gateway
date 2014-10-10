package main

import (
	"fmt"
	"log"

	"os"

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
	raft.RegisterCommands()

	db := db.NewMemoryStore()

	log.Print("Starting Raft server")
	rServer := raft.NewServer(conf.Raft)
	rServer.Setup(db)
	go rServer.Run()

	log.Print("Starting proxy server")
	proxy := proxy.NewServer(conf.Proxy, raft.NewRaftDB(db, rServer.RaftServer))
	go proxy.Run()

	select {}
}
