package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"os"

	"gateway/config"
	"gateway/license"
	"gateway/proxy"
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

	log.Printf("%s Starting server", config.System)
	proxy := proxy.NewServer(conf.Proxy, conf.Admin)
	go proxy.Run()

	done := make(chan bool)
	<-done
}
