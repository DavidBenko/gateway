package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"os"

	"gateway/config"
	"gateway/license"
	"gateway/proxy"
	"gateway/sql"
	"gateway/version"
)

func main() {
	if versionCheck() {
		fmt.Printf("Gateway %s (%s)\n",
			version.Name(), version.Commit())
		return
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	// Setup logging
	log.SetFlags(log.Ldate | log.Lmicroseconds)
	log.SetOutput(os.Stdout)

	// Parse configuration
	conf, err := config.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("%s Error parsing config file: %v", config.System, err)
	}

	log.Printf("%s Running Gateway %s (%s)",
		config.System, version.Name(), version.Commit())

	// Require a valid license key
	license.ValidateForever(conf.License, time.Hour)

	// Setup the database
	db, err := sql.Connect(conf.Database)
	if err != nil {
		log.Fatalf("%s Error connecting to database: %v", config.System, err)
	}
	if !db.UpToDate() {
		if conf.Database.Migrate {
			if err = db.Migrate(); err != nil {
				log.Fatalf("Error migrating database: %v", err)
			}
		} else {
			log.Fatalf("%s The database is not up to date. "+
				"Please migrate by invoking with the -db-migrate flag.",
				config.System)
		}
	}

	// Start the proxy
	log.Printf("%s Starting server", config.System)
	proxy := proxy.NewServer(conf.Proxy, conf.Admin, db)
	go proxy.Run()

	done := make(chan bool)
	<-done
}

func versionCheck() bool {
	return len(os.Args) >= 2 &&
		strings.ToLower(os.Args[1:2][0]) == "-version"
}
