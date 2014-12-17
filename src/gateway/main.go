package main

import (
	"log"
	"time"

	"os"

	"gateway/config"
	"gateway/license"
	"gateway/proxy"
	"gateway/sql"
)

func main() {
	log.SetFlags(log.Ldate | log.Lmicroseconds)
	log.SetOutput(os.Stdout)

	conf, err := config.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}

	license.ValidateForever(conf.License, time.Hour)

	db, err := sql.Connect(conf.Database)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	if !db.UpToDate() {
		if conf.Database.Migrate {
			if err = db.Migrate(); err != nil {
				log.Fatalf("Error migrating database: %v", err)
			}
		} else {
			message := "The database is not up to date.\n"
			message += "Please migrate by invoking with the -db-migrate flag."
			log.Fatal(message)
		}
	}

	log.Printf("%s Starting server", config.System)
	proxy := proxy.NewServer(conf.Proxy, conf.Admin)
	go proxy.Run()

	done := make(chan bool)
	<-done
}
