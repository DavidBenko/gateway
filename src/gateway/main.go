package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"os"

	"gateway/config"
	"gateway/db"
	"gateway/license"
	"gateway/proxy"
	"gateway/raft"
)

func main() {
	log.SetFlags(log.Ldate | log.Lmicroseconds)
	log.SetOutput(os.Stdout)

	conf, err := config.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(fmt.Sprintf("Error parsing config file: %v", err))
	}

	license.ValidateForever(conf.License, time.Hour)

	// Each server name must be unique
	rand.Seed(time.Now().UnixNano())

	db := db.NewMemoryStore()

	log.Printf("%s Starting Raft server", config.System)
	rServer := raft.NewServer(conf.Raft)
	rServer.Setup(db)
	go rServer.Run()

	log.Printf("%s Starting proxy server", config.System)
	proxy := proxy.NewServer(conf.Proxy, conf.Admin, raft.NewRaftDB(db, rServer.RaftServer))
	go proxy.Run()

	select {}
}
