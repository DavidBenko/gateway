package main

import (
	"fmt"
	"log"

	"os"

	"github.com/AnyPresence/gateway/config"
	"github.com/AnyPresence/gateway/proxy"
)

func main() {
	conf, err := config.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(fmt.Sprintf("Error parsing config file: %v", err))
	}

	log.Print("Starting proxy server")
	go proxy.Run(conf.Proxy)

	select {}
}
