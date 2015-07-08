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
	"gateway/model"
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

	//check for sneaky people
	if conf.License == "" {
		accounts, _ := model.AllAccounts(db)
		if len(accounts) > license.DeveloperVersionAccounts {
			log.Fatalf("Developer version allows %v account(s).", license.DeveloperVersionAccounts)
		}
		for _, account := range accounts {
			var count int
			db.Get(&count, db.SQL("users/count"), account.ID)
			if count >= license.DeveloperVersionUsers {
				log.Fatalf("Developer version allows %v user(s).", license.DeveloperVersionUsers)
			}

			apis, _ := model.AllAPIsForAccountID(db, account.ID)
			if len(apis) > license.DeveloperVersionAPIs {
				log.Fatalf("Developer version allows %v api(s).", license.DeveloperVersionAPIs)
			}
			for _, api := range apis {
				var count int
				db.Get(&count, db.SQL("proxy_endpoints/count"), api.ID)
				if count > license.DeveloperVersionProxyEndpoints {
					log.Fatalf("Developer version allows %v proxy endpoint(s).", license.DeveloperVersionProxyEndpoints)
				}
			}
		}
	}

	if !db.UpToDate() {
		if conf.Database.Migrate || conf.DevMode() {
			if err = db.Migrate(); err != nil {
				log.Fatalf("Error migrating database: %v", err)
			}
		} else {
			log.Fatalf("%s The database is not up to date. "+
				"Please migrate by invoking with the -db-migrate flag.",
				config.System)
		}
	}

	// Set up dev mode account
	if conf.DevMode() {
		if _, err := model.FirstAccount(db); err != nil {
			log.Printf("%s Creating development account", config.System)
			if err := createDevAccount(db); err != nil {
				log.Fatalf("Could not create account: %v", err)
			}
		}
	}
	// Start the proxy
	log.Printf("%s Starting server", config.System)
	proxy := proxy.NewServer(conf, db)
	go proxy.Run()

	done := make(chan bool)
	<-done
}

func versionCheck() bool {
	return len(os.Args) >= 2 &&
		strings.ToLower(os.Args[1:2][0]) == "-version"
}

func createDevAccount(db *sql.DB) error {
	devAccount := &model.Account{Name: "Dev Account"}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if err = devAccount.Insert(tx); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}
