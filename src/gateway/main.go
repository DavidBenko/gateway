package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"os"

	"gateway/config"
	"gateway/db"
	"gateway/proxy"
	"gateway/raft"
)

func main() {
	conf, err := config.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(fmt.Sprintf("Error parsing config file: %v", err))
	}

	// Each server name must be unique
	rand.Seed(time.Now().UnixNano())

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
