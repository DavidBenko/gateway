package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"gateway/admin"
	"gateway/config"
	"gateway/license"
	"gateway/model"
	"gateway/proxy"
	"gateway/service"
	"gateway/soap"
	"gateway/sql"
	"gateway/version"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	if versionCheck() {
		fmt.Printf("Gateway %s (%s)\n",
			version.Name(), version.Commit())
		return
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	// Setup logging
	log.SetFlags(log.Ldate | log.Lmicroseconds)
	log.SetOutput(admin.Interceptor)

	// Parse configuration
	conf, err := config.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("%s Error parsing config file: %v", config.System, err)
	}

	log.Printf("%s Running Gateway %s (%s)",
		config.System, version.Name(), version.Commit())

	// Setup the database
	db, err := sql.Connect(conf.Database)
	if err != nil {
		log.Fatalf("%s Error connecting to database: %v", config.System, err)
	}

	// Require a valid license key
	license.ValidateForever(conf.License, time.Hour)

	//check for sneaky people
	if license.DeveloperVersion {
		log.Printf("%s Checking developer version license constraints", config.System)
		accounts, _ := model.AllAccounts(db)
		if len(accounts) > license.DeveloperVersionAccounts {
			log.Fatalf("Developer version allows %v account(s).", license.DeveloperVersionAccounts)
		}
		for _, account := range accounts {
			var count int
			db.Get(&count, db.SQL("users/count"), account.ID)
			if count > license.DeveloperVersionUsers {
				log.Fatalf("Developer version allows %v user(s).", license.DeveloperVersionUsers)
			}

			apis, _ := model.AllAPIsForAccountID(db, account.ID)
			if len(apis) > license.DeveloperVersionAPIs {
				log.Fatalf("Developer version allows %v api(s).", license.DeveloperVersionAPIs)
			}
			for _, api := range apis {
				var count int
				db.Get(&count, db.SQL("proxy_endpoints/count_active"), api.ID)
				if count > license.DeveloperVersionProxyEndpoints {
					log.Fatalf("Developer version allows %v active proxy endpoint(s).", license.DeveloperVersionProxyEndpoints)
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

	commands := config.Commands()
	if len(commands) > 0 {
		processCommands(commands, db)
		return
	}

	// Set up dev mode account
	if conf.DevMode() {
		if _, err := model.FirstAccount(db); err != nil {
			log.Printf("%s Creating development account", config.System)
			if err := createDevAccount(db); err != nil {
				log.Fatalf("Could not create account: %v", err)
			}
		}
		if account, err := model.FirstAccount(db); err == nil {
			if users, _ := model.AllUsersForAccountID(db, account.ID); len(users) == 0 {
				log.Printf("%s Creating development user", config.System)
				if err := createDevUser(db); err != nil {
					log.Fatalf("Could not create account: %v", err)
				}
			}
		} else {
			log.Fatal("Dev account doesn't exist")
		}
	}

	service.ElasticLoggingService(conf.Elastic)
	service.BleveLoggingService(conf.Bleve)
	service.LogPublishingService(conf.Admin)

	// Write script remote endpoints to tmp fireLifecycleHooks
	err = model.WriteAllScriptFiles(db)
	if err != nil {
		log.Printf("%s Unable to write script files due to error: %v", config.System, err)
	}

	// Configure SOAP
	err = soap.Configure(conf.Soap, conf.DevMode())
	if err != nil {
		log.Printf("%s Unable to configure SOAP due to error: %v.  SOAP services will not be available.", config.System, err)
	}

	// Start up listeners for soap_remote_endpoints, so that we can keep the file system in sync with the DB
	model.StartSoapRemoteEndpointUpdateListener(db)

	// Start the proxy
	log.Printf("%s Starting server", config.System)
	proxy := proxy.NewServer(conf, db)
	go proxy.Run()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		err := soap.Shutdown(sig)
		if err != nil {
			log.Printf("Error shutting down SOAP service: %v", err)
		}
		done <- true
	}()

	<-done

	log.Println("Shutdown complete")
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

var symbols = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randomPassword() string {
	password := make([]rune, 16)
	for i := range password {
		password[i] = symbols[rand.Intn(len(symbols))]
	}
	return string(password)
}

func createDevUser(db *sql.DB) error {
	account, err := model.FirstAccount(db)
	if err != nil {
		return err
	}
	user := &model.User{
		AccountID:   account.ID,
		Name:        "developer",
		Email:       "developer@justapis.com",
		NewPassword: randomPassword(),
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if err = user.Insert(tx); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

type Parameter struct {
	Name     string
	Required bool
}

type Command struct {
	Parameters []Parameter
	Usage      string
}

var commands = map[string]Command{
	"accounts": {
		[]Parameter{},
		"accounts",
	},
	"accounts:create": {
		[]Parameter{
			{"name", true},
		},
		"accounts:create name:\"<name>\"",
	},
	"accounts:update": {
		[]Parameter{
			{"id", true},
			{"name", false},
		},
		"accounts:update <id> name:\"<name>\"",
	},
	"accounts:destroy": {
		[]Parameter{
			{"id", true},
		},
		"accounts:destroy <id>",
	},
	"users": {
		[]Parameter{
			{"id", true},
		},
		"users <account-id>",
	},
	"users:create": {
		[]Parameter{
			{"id", true},
			{"name", true},
			{"email", true},
			{"password", true},
		},
		"users:create <account-id> name:\"<name>\" email:<email> password:<password>",
	},
	"users:update": {
		[]Parameter{
			{"id", true},
			{"name", false},
			{"email", false},
			{"password", false},
		},
		"users:update <id> name:\"<name>\" email:<email> password:<password>",
	},
	"users:destroy": {
		[]Parameter{
			{"id", true},
		},
		"users:destroy <id>",
	},
}

func getParameters(command string, params []string) map[string]string {
	cmd, ok := commands[command]
	if !ok {
		fmt.Println("invalid command")
		return nil
	}

	values := map[string]string{}
	for _, param := range params {
		value := strings.Split(param, ":")
		switch len(value) {
		case 1:
			values["id"] = value[0]
		case 2:
			values[value[0]] = value[1]
		}
	}
	for _, param := range cmd.Parameters {
		if param.Required && len(values[param.Name]) == 0 {
			fmt.Println(cmd.Usage)
			return nil
		}
	}
	return values
}

func processCommands(cmds []string, db *sql.DB) {
	if cmds[0] == "help" {
		if len(cmds) > 1 {
			fmt.Println(commands[cmds[1]].Usage)
			return
		}
		for command := range commands {
			fmt.Println(command)
		}
		return
	}
	params := getParameters(cmds[0], cmds[1:])
	if params == nil {
		return
	}
	getAccount := func() *model.Account {
		id, err := strconv.Atoi(params["id"])
		if err != nil {
			log.Fatal(err)
		}
		account, err := model.FindAccount(db, int64(id))
		if err != nil {
			log.Fatal(err)
		}
		return account
	}
	getUser := func() *model.User {
		id, err := strconv.Atoi(params["id"])
		if err != nil {
			log.Fatal(err)
		}
		user, err := model.FindUserByID(db, int64(id))
		if err != nil {
			log.Fatal(err)
		}
		return user
	}
	switch cmds[0] {
	case "accounts":
		accounts, err := model.AllAccounts(db)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("=== Accounts")
		for _, account := range accounts {
			fmt.Printf("%v %v\n", account.ID, account.Name)
		}
	case "accounts:create":
		account := &model.Account{
			Name: params["name"],
		}
		err := db.DoInTransaction(func(tx *sql.Tx) error {
			return account.Insert(tx)
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Create account %v %v\n", account.ID, account.Name)
	case "accounts:update":
		account := getAccount()
		if params["name"] != "" {
			account.Name = params["name"]
		}
		err := db.DoInTransaction(func(tx *sql.Tx) error {
			return account.Update(tx)
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Updated account %v %v\n", account.ID, account.Name)
	case "accounts:destroy":
		account := getAccount()
		fmt.Printf("delete account: %v %v?\n", account.ID, account.Name)
		fmt.Printf("enter id (%v):", account.ID)
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		response = strings.Trim(response, "\n")
		enteredId, err := strconv.Atoi(response)
		if err != nil {
			log.Fatal(err)
		}
		if account.ID != int64(enteredId) {
			log.Fatal("the entered id doesn't match")
		}
		fmt.Printf("enter name (%v):", account.Name)
		response, err = reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		response = strings.Trim(response, "\n")
		if account.Name != response {
			log.Fatal("the entered name doesn't match")
		}
		err = db.DoInTransaction(func(tx *sql.Tx) error {
			return model.DeleteAccount(tx, account.ID)
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Destroyed account %v %v\n", account.ID, account.Name)
	case "users":
		account := getAccount()
		users, err := model.AllUsersForAccountID(db, account.ID)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("=== Users for Account %v %v\n", account.ID, account.Name)
		for _, user := range users {
			fmt.Printf("%v %v %v\n", user.ID, user.Name, user.Email)
		}
	case "users:create":
		account := getAccount()
		user := &model.User{
			AccountID:   account.ID,
			Name:        params["name"],
			Email:       params["email"],
			NewPassword: params["password"],
		}
		err := db.DoInTransaction(func(tx *sql.Tx) error {
			return user.Insert(tx)
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Created user %v %v for account %v %v\n", user.ID, user.Email,
			account.ID, account.Name)
	case "users:update":
		user := getUser()
		if params["name"] != "" {
			user.Name = params["name"]
		}
		if params["email"] != "" {
			user.Email = params["email"]
		}
		if params["password"] != "" {
			user.NewPassword = params["password"]
		}
		err := db.DoInTransaction(func(tx *sql.Tx) error {
			return user.Update(tx)
		})
		if err != nil {
			log.Fatal(err)
		}
		account, err := model.FindAccount(db, user.AccountID)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Updated user %v %v for account %v %v\n", user.ID, user.Email,
			account.ID, account.Name)
	case "users:destroy":
		user := getUser()
		fmt.Printf("delete user: %v %v?\n", user.ID, user.Email)
		fmt.Printf("enter id (%v):", user.ID)
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		response = strings.Trim(response, "\n")
		enteredId, err := strconv.Atoi(response)
		if err != nil {
			log.Fatal(err)
		}
		if user.ID != int64(enteredId) {
			log.Fatal("the entered id doesn't match")
		}
		fmt.Printf("enter email (%v):", user.Email)
		response, err = reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		response = strings.Trim(response, "\n")
		if user.Email != response {
			log.Fatal("the entered email doesn't match")
		}
		err = db.DoInTransaction(func(tx *sql.Tx) error {
			err := model.CanDeleteUser(tx, user.ID)
			if err != nil {
				return err
			}
			return model.DeleteUserForAccountID(tx, user.ID, user.AccountID, 0)
		})
		if err != nil {
			log.Fatal(err)
		}
		account, err := model.FindAccount(db, user.AccountID)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Destroyed user %v %v on account %v %v\n", user.ID, user.Email,
			account.ID, account.Name)
	}
}
