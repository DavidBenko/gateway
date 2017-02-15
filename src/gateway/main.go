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
	"gateway/core"
	"gateway/docker"
	"gateway/errors/report"
	"gateway/http"
	"gateway/logreport"
	"gateway/mail"
	"gateway/model"
	"gateway/osslicenses"
	"gateway/proxy"
	"gateway/push"
	"gateway/service"
	"gateway/soap"
	"gateway/sql"
	statssql "gateway/stats/sql"
	"gateway/version"

	stripe "github.com/stripe/stripe-go"
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

	if ossLicensesCheck() {
		fmt.Printf("%s", osslicenses.OssLicenseList)
		return
	}

	if exampleConfigCheck() {
		fmt.Printf("%s", config.ExampleConfigurationFile)
		return
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	// Setup logging
	log.SetFlags(log.Ldate | log.Lmicroseconds)
	log.SetOutput(admin.Interceptor)

	// Parse configuration
	conf, err := config.Parse(os.Args[1:])
	if err != nil {
		logreport.Fatalf("%s Error parsing config file: %v", config.System, err)
	}

	logreport.Printf("%s Running Gateway %s (%s)",
		config.System, version.Name(), version.Commit())

	// Set up error reporting
	if conf.Airbrake.APIKey != "" && conf.Airbrake.ProjectID != 0 && !conf.DevMode() {
		abEnv := "production"
		if conf.Airbrake.Environment != "" {
			abEnv = conf.Airbrake.Environment
		}
		report.RegisterReporter(report.ConfigureAirbrake(conf.Airbrake.APIKey, conf.Airbrake.ProjectID, abEnv))
	}

	// Setup the database
	db, err := sql.Connect(conf.Database)
	if err != nil {
		logreport.Fatalf("%s Error connecting to database: %v", config.System, err)
	}

	if !db.UpToDate() {
		if conf.Database.Migrate || conf.DevMode() {
			if err = db.Migrate(); err != nil {
				logreport.Fatalf("Error migrating database: %v", err)
			}
		} else {
			logreport.Fatalf("%s The database is not up to date. "+
				"Please migrate by invoking with the -db-migrate flag.",
				config.System)
		}
	}

	var statsDb *statssql.SQL

	if conf.Stats.Collect {
		// Setup the stats database
		statsDb, err = statssql.Connect(conf.Stats)
		if err != nil {
			logreport.Fatalf("%s Error connecting to stats database: %v", config.System, err)
		}
		// Make sure the stats db is up-to-date and migrated.
		if !statsDb.UpToDate() {
			if conf.Stats.Migrate || conf.DevMode() {
				if err = statsDb.Migrate(); err != nil {
					logreport.Fatalf("Error migrating stats database: %v", err)
				}
			} else {
				logreport.Fatalf("%s The stats database is not up to date. "+
					"Please migrate by invoking with the -stats-migrate flag.",
					config.System)
			}
		}
	}

	if conf.RemoteEndpoint.ScrubData {
		logreport.Printf("%s Scrubbing remote endpoint data...", config.System)
		err = model.ScrubExistingRemoteEndpointData(db)
		if err != nil {
			logreport.Fatal(err)
		}
		logreport.Printf("%s Remote endpoint data scrubbing complete.", config.System)
	}

	if !conf.DevMode() {
		// if Stripe API keys are set and we are in server mode, let's setup the Stripe client.
		if conf.Admin.StripeSecretKey != "" && conf.Admin.StripePublishableKey != "" {
			logreport.Printf("Setting up Stripe client.")
			stripe.Key = conf.Admin.StripeSecretKey
			if conf.Admin.StripeFallbackPlan == "" {
				logreport.Fatalf("A Stripe fallback plan is required when Stripe is configured.")
			}
			if conf.Admin.StripeMigrateAccounts {
				err = model.MigrateAccountsToStripe(db, conf.Admin.StripeFallbackPlan)
				if err != nil {
					logreport.Fatal(err)
				}
			}
		}
	}

	commands := config.Commands()
	if len(commands) > 0 {
		processCommands(commands, conf, db)
		return
	}

	// Set up dev mode account
	if conf.DevMode() {
		if _, err := model.FirstAccount(db); err != nil {
			logreport.Printf("%s Creating development account", config.System)
			if err := createDevAccount(db); err != nil {
				logreport.Fatalf("Could not create account: %v", err)
			}
		}
		if account, err := model.FirstAccount(db); err == nil {
			if users, _ := model.AllUsersForAccountID(db, account.ID); len(users) == 0 {
				logreport.Printf("%s Creating development user", config.System)
				if err := createDevUser(db); err != nil {
					logreport.Fatalf("Could not create account: %v", err)
				}
			}
		} else {
			logreport.Fatal("Dev account doesn't exist")
		}
	}

	warp := core.NewCore(conf, db, statsDb)

	service.ElasticLoggingService(conf)
	service.BleveLoggingService(conf.Bleve)
	service.PostgresLoggingService(conf.PostgresLogging)
	service.LogPublishingService(conf.Admin)
	service.SessionDeletionService(conf, db)
	service.PushDeletionService(conf, db)
	service.JobsService(conf, warp)

	model.InitializeRemoteEndpointTypes(conf.RemoteEndpoint)

	push.SetupMQTT(db, conf.Push, warp.ExecuteMQTT)

	// Write script remote endpoints to tmp fireLifecycleHooks
	err = model.WriteAllScriptFiles(db)
	if err != nil {
		logreport.Printf("%s Unable to write script files due to error: %v", config.System, err)
	}

	// Configure Docker
	if conf.RemoteEndpoint.DockerEnabled || conf.RemoteEndpoint.CustomFunctionEnabled {
		logreport.Printf("%s Configuring Docker remote endpoint support...", config.System)
		err = docker.ConfigureDockerClient(conf.Docker)
		if err != nil {
			conf.RemoteEndpoint.DockerEnabled = false
			logreport.Printf("%s Unable to configure Docker due to error: %v.  Docker remote endpoints will not be enabled.", config.System, err)
		} else {
			info, err := docker.DockerClientInfo()
			if err != nil {
				conf.RemoteEndpoint.DockerEnabled = false
				logreport.Printf("%s Unable to connect to Docker host to get info: %v.  Docker remote endpoints will not be enabled.", config.System, err)
			} else {
				logreport.Printf("%s Docker remote endpoint support configured:\n%v", config.System, info)
			}
		}
	}

	// Configure SOAP
	if conf.RemoteEndpoint.SoapEnabled {
		logreport.Printf("Configuring SOAP remote endpoint support...")
		err = soap.Configure(conf.Soap, conf.DevMode())
		if err != nil {
			conf.RemoteEndpoint.SoapEnabled = false
			logreport.Printf("%s Unable to configure SOAP due to error: %v.  SOAP services will not be available.", config.System, err)
		} else {
			logreport.Printf("SOAP remote endpoint support configured.")
		}
	}

	// Start up listeners for soap_remote_endpoints, so that we can keep the file system in sync with the DB
	model.StartSoapRemoteEndpointUpdateListener(db)

	if conf.RemoteEndpoint.DockerEnabled {
		// Listen to docker remote endpoint changes.
		model.StartDockerEndpointUpdateListener(db)
	}

	model.ConfigureDefaultAPIAccessScheme(conf.Admin.DefaultAPIAccessScheme)

	// Start the proxy
	logreport.Printf("%s Starting server", config.System)
	proxy := proxy.NewServer(conf, db, warp)
	go proxy.Run()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		err := soap.Shutdown(sig)
		if err != nil {
			logreport.Printf("Error shutting down SOAP service: %v", err)
		}

		service.StopJobsService()

		warp.Shutdown()

		done <- true
	}()

	<-done

	logreport.Println("Shutdown complete")
}

func versionCheck() bool {
	return len(os.Args) >= 2 &&
		strings.ToLower(os.Args[1:2][0]) == "-version"
}

func ossLicensesCheck() bool {
	return len(os.Args) >= 2 &&
		strings.ToLower(os.Args[1:2][0]) == "-oss-licenses"
}

func exampleConfigCheck() bool {
	return len(os.Args) >= 2 &&
		strings.ToLower(os.Args[1:2][0]) == "-example-config"
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
		Email:       "developer@example.net",
		NewPassword: randomPassword(),
		Confirmed:   true,
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

type Process func(params map[string]string, conf config.Configuration, db *sql.DB)

type Parameter struct {
	Name     string
	Required bool
}

type Command struct {
	Parameters []Parameter
	Usage      string
	Process    Process
}

var commands = map[string]Command{
	"accounts": {
		[]Parameter{},
		"accounts",
		accounts,
	},
	"accounts:create": {
		[]Parameter{
			{"name", true},
			{"plan_id", false},
		},
		"accounts:create name:\"<name>\"",
		accountsCreate,
	},
	"accounts:update": {
		[]Parameter{
			{"id", true},
			{"name", false},
			{"plan_id", false},
		},
		"accounts:update <id> [name:\"<name>\"]",
		accountsUpdate,
	},
	"accounts:destroy": {
		[]Parameter{
			{"id", true},
		},
		"accounts:destroy <id>",
		accountsDestroy,
	},
	"users": {
		[]Parameter{
			{"id", true},
		},
		"users <account-id>",
		users,
	},
	"users:create": {
		[]Parameter{
			{"id", true},
			{"name", true},
			{"email", true},
			{"password", true},
			{"admin", true},
			{"confirmed", true},
		},
		"users:create <account-id> name:\"<name>\" email:<email> password:<password> admin:<true/false> confirmed:<true/false>",
		usersCreate,
	},
	"users:update": {
		[]Parameter{
			{"id", true},
			{"name", false},
			{"email", false},
			{"password", false},
			{"admin", false},
			{"confirmed", false},
		},
		"users:update <id> [name:\"<name>\"] [email:<email>] [password:<password>] [admin:<true/false>] [confirmed:<true/false>]",
		usersUpdate,
	},
	"users:destroy": {
		[]Parameter{
			{"id", true},
		},
		"users:destroy <id>",
		usersDestroy,
	},
	"plans": {
		[]Parameter{},
		"plans",
		plans,
	},
	"plans:create": {
		[]Parameter{
			{"name", true},
			{"stripe_name", true},
			{"max_users", true},
			{"javascript_timeout", true},
			{"job_timeout", true},
			{"price", true},
		},
		"plans:create name:\"<name>\" stripe_name:<stripe_name> max_users:<max_users> javascript_timeout:<javascript_timeout> job_timeout:<job_timeout> price:<price>",
		plansCreate,
	},
	"plans:update": {
		[]Parameter{
			{"id", true},
			{"name", false},
			{"stripe_name", false},
			{"max_users", false},
			{"javascript_timeout", false},
			{"job_timeout", false},
			{"price", false},
		},
		"plans:update <id> [name:\"<name>\"] [email:<email>] [stripe_name:<stripe_name>] [max_users:<max_users>] [javascript_timeout:<javascript_timeout>] [job_timeout:<job_timeout>] [price:<price>]",
		plansUpdate,
	},
	"plans:destroy": {
		[]Parameter{
			{"id", true},
		},
		"plans:destroy <id>",
		plansDestroy,
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

func getAccount(params map[string]string, db *sql.DB) *model.Account {
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		logreport.Fatal(err)
	}
	account, err := model.FindAccount(db, int64(id))
	if err != nil {
		logreport.Fatal(err)
	}
	return account
}

func getUser(params map[string]string, db *sql.DB) *model.User {
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		logreport.Fatal(err)
	}
	user, err := model.FindUserByID(db, int64(id))
	if err != nil {
		logreport.Fatal(err)
	}
	return user
}

func accounts(params map[string]string, conf config.Configuration, db *sql.DB) {
	accounts, err := model.AllAccounts(db)
	if err != nil {
		logreport.Fatal(err)
	}
	fmt.Println("=== Accounts")
	for _, account := range accounts {
		fmt.Printf("%v %v %v\n", account.ID, account.Name, account.PlanID.Int64)
	}
}

func accountsCreate(params map[string]string, conf config.Configuration, db *sql.DB) {
	account := &model.Account{
		Name: params["name"],
	}
	if params["plan_id"] != "" {
		planID, err := strconv.Atoi(params["plan_id"])
		if err != nil {
			logreport.Fatal(err)
		}
		account.PlanID = sql.MakeNullInt64(int64(planID))
	}
	err := db.DoInTransaction(func(tx *sql.Tx) error {
		return account.Insert(tx)
	})
	if err != nil {
		logreport.Fatal(err)
	}
	fmt.Printf("Create account %v %v\n", account.ID, account.Name)
}

func accountsUpdate(params map[string]string, conf config.Configuration, db *sql.DB) {
	account := getAccount(params, db)
	if params["name"] != "" {
		account.Name = params["name"]
	}
	if params["plan_id"] != "" {
		planID, err := strconv.Atoi(params["plan_id"])
		if err != nil {
			logreport.Fatal(err)
		}
		account.PlanID = sql.MakeNullInt64(int64(planID))
	}
	err := db.DoInTransaction(func(tx *sql.Tx) error {
		return account.Update(tx)
	})
	if err != nil {
		logreport.Fatal(err)
	}
	fmt.Printf("Updated account %v %v\n", account.ID, account.Name)
}

func accountsDestroy(params map[string]string, conf config.Configuration, db *sql.DB) {
	account := getAccount(params, db)
	fmt.Printf("delete account: %v %v?\n", account.ID, account.Name)
	fmt.Printf("enter id (%v):", account.ID)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		logreport.Fatal(err)
	}
	response = strings.Trim(response, "\n")
	enteredId, err := strconv.Atoi(response)
	if err != nil {
		logreport.Fatal(err)
	}
	if account.ID != int64(enteredId) {
		logreport.Fatal("the entered id doesn't match")
	}
	fmt.Printf("enter name (%v):", account.Name)
	response, err = reader.ReadString('\n')
	if err != nil {
		logreport.Fatal(err)
	}
	response = strings.Trim(response, "\n")
	if account.Name != response {
		logreport.Fatal("the entered name doesn't match")
	}
	err = db.DoInTransaction(func(tx *sql.Tx) error {
		return model.DeleteAccount(tx, account.ID)
	})
	if err != nil {
		logreport.Fatal(err)
	}
	fmt.Printf("Destroyed account %v %v\n", account.ID, account.Name)
}

func users(params map[string]string, conf config.Configuration, db *sql.DB) {
	account := getAccount(params, db)
	users, err := model.AllUsersForAccountID(db, account.ID)
	if err != nil {
		logreport.Fatal(err)
	}
	fmt.Printf("=== Users for Account %v %v\n", account.ID, account.Name)
	for _, user := range users {
		s := fmt.Sprintf("%v %v %v", user.ID, user.Name, user.Email)
		if user.Admin {
			s += " admin"
		}
		if user.Confirmed {
			s += " confirmed"
		}
		fmt.Printf("%v\n", s)
	}
}

func usersCreate(params map[string]string, conf config.Configuration, db *sql.DB) {
	account := getAccount(params, db)
	_admin := false
	if params["admin"] == "true" {
		_admin = true
	}
	confirmed := false
	if params["confirmed"] == "true" {
		confirmed = true
	}
	user := &model.User{
		AccountID:   account.ID,
		Name:        params["name"],
		Email:       params["email"],
		NewPassword: params["password"],
		Admin:       _admin,
		Confirmed:   confirmed,
	}
	err := db.DoInTransaction(func(tx *sql.Tx) error {
		err := user.Insert(tx)
		if err != nil {
			return err
		}
		if !confirmed {
			return mail.SendConfirmEmail(conf.SMTP, conf.Proxy, conf.Admin, user, tx, false)
		}
		return nil
	})
	if err != nil {
		logreport.Fatal(err)
	}
	fmt.Printf("Created user %v %v for account %v %v\n", user.ID, user.Email,
		account.ID, account.Name)
}

func usersUpdate(params map[string]string, conf config.Configuration, db *sql.DB) {
	user := getUser(params, db)
	if params["name"] != "" {
		user.Name = params["name"]
	}
	if params["email"] != "" {
		user.Email = params["email"]
	}
	if params["password"] != "" {
		user.NewPassword = params["password"]
	}
	if params["admin"] != "" {
		if params["admin"] == "true" {
			user.Admin = true
		} else {
			user.Admin = false
		}
	}
	if params["confirmed"] != "" {
		if params["confirmed"] == "true" {
			user.Confirmed = true
		} else {
			user.Confirmed = false
		}
	}
	err := db.DoInTransaction(func(tx *sql.Tx) error {
		return user.Update(tx)
	})
	if err != nil {
		logreport.Fatal(err)
	}
	account, err := model.FindAccount(db, user.AccountID)
	if err != nil {
		logreport.Fatal(err)
	}
	fmt.Printf("Updated user %v %v for account %v %v\n", user.ID, user.Email,
		account.ID, account.Name)
}

func usersDestroy(params map[string]string, conf config.Configuration, db *sql.DB) {
	user := getUser(params, db)
	fmt.Printf("delete user: %v %v?\n", user.ID, user.Email)
	fmt.Printf("enter id (%v):", user.ID)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		logreport.Fatal(err)
	}
	response = strings.Trim(response, "\n")
	enteredId, err := strconv.Atoi(response)
	if err != nil {
		logreport.Fatal(err)
	}
	if user.ID != int64(enteredId) {
		logreport.Fatal("the entered id doesn't match")
	}
	fmt.Printf("enter email (%v):", user.Email)
	response, err = reader.ReadString('\n')
	if err != nil {
		logreport.Fatal(err)
	}
	response = strings.Trim(response, "\n")
	if user.Email != response {
		logreport.Fatal("the entered email doesn't match")
	}
	err = db.DoInTransaction(func(tx *sql.Tx) error {
		err := model.CanDeleteUser(tx, user.ID, user.AccountID, http.AuthTypeAdmin)
		if err != nil {
			return err
		}
		return model.DeleteUserForAccountID(tx, user.ID, user.AccountID, 0)
	})
	if err != nil {
		logreport.Fatal(err)
	}
	account, err := model.FindAccount(db, user.AccountID)
	if err != nil {
		logreport.Fatal(err)
	}
	fmt.Printf("Destroyed user %v %v on account %v %v\n", user.ID, user.Email,
		account.ID, account.Name)
}

func getPlan(params map[string]string, db *sql.DB) *model.Plan {
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		logreport.Fatal(err)
	}
	plan, err := model.FindPlan(db, int64(id))
	if err != nil {
		logreport.Fatal(err)
	}
	return plan
}

func plans(params map[string]string, conf config.Configuration, db *sql.DB) {
	plans, err := model.AllPlans(db)
	if err != nil {
		logreport.Fatal(err)
	}
	fmt.Println("=== Plans")
	for _, plan := range plans {
		fmt.Printf("%v %v %v %v %v %v\n", plan.ID, plan.Name, plan.StripeName, plan.MaxUsers, plan.JavascriptTimeout, plan.Price)
	}
}

func plansCreate(params map[string]string, conf config.Configuration, db *sql.DB) {
	maxUsers, err := strconv.Atoi(params["max_users"])
	if err != nil {
		logreport.Fatal(err)
	}
	javascriptTimeout, err := strconv.Atoi(params["javascript_timeout"])
	if err != nil {
		logreport.Fatal(err)
	}
	jobTimeout, err := strconv.Atoi(params["job_timeout"])
	if err != nil {
		logreport.Fatal(err)
	}
	price, err := strconv.Atoi(params["price"])
	if err != nil {
		logreport.Fatal(err)
	}
	plan := &model.Plan{
		Name:              params["name"],
		StripeName:        params["stripe_name"],
		MaxUsers:          int64(maxUsers),
		JavascriptTimeout: int64(javascriptTimeout),
		JobTimeout:        int64(jobTimeout),
		Price:             int64(price),
	}
	err = db.DoInTransaction(func(tx *sql.Tx) error {
		return plan.Insert(tx)
	})
	if err != nil {
		logreport.Fatal(err)
	}
	fmt.Printf("Create plan %v %v %v %v %v %v\n", plan.ID, plan.Name, plan.StripeName, plan.MaxUsers, plan.JavascriptTimeout, plan.Price)
}

func plansUpdate(params map[string]string, conf config.Configuration, db *sql.DB) {
	plan := getPlan(params, db)
	if params["name"] != "" {
		plan.Name = params["name"]
	}
	if params["stripe_name"] != "" {
		plan.StripeName = params["stripe_name"]
	}
	if params["max_users"] != "" {
		maxUsers, err := strconv.Atoi(params["max_users"])
		if err != nil {
			logreport.Fatal(err)
		}
		plan.MaxUsers = int64(maxUsers)
	}
	if params["javascript_timeout"] != "" {
		javascripTimeout, err := strconv.Atoi(params["javascript_timeout"])
		if err != nil {
			logreport.Fatal(err)
		}
		plan.JavascriptTimeout = int64(javascripTimeout)
	}
	if params["job_timeout"] != "" {
		jobTimeout, err := strconv.Atoi(params["job_timeout"])
		if err != nil {
			logreport.Fatal(err)
		}
		plan.JobTimeout = int64(jobTimeout)
	}
	if params["price"] != "" {
		price, err := strconv.Atoi(params["price"])
		if err != nil {
			logreport.Fatal(err)
		}
		plan.Price = int64(price)
	}
	err := db.DoInTransaction(func(tx *sql.Tx) error {
		return plan.Update(tx)
	})
	if err != nil {
		logreport.Fatal(err)
	}
	fmt.Printf("Updated plan %v %v %v %v %v %v\n", plan.ID, plan.Name, plan.StripeName, plan.MaxUsers, plan.JavascriptTimeout, plan.Price)
}

func plansDestroy(params map[string]string, conf config.Configuration, db *sql.DB) {
	plan := getPlan(params, db)
	fmt.Printf("delete plan: %v %v?\n", plan.ID, plan.Name)
	fmt.Printf("enter id (%v):", plan.ID)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		logreport.Fatal(err)
	}
	response = strings.Trim(response, "\n")
	enteredId, err := strconv.Atoi(response)
	if err != nil {
		logreport.Fatal(err)
	}
	if plan.ID != int64(enteredId) {
		logreport.Fatal("the entered id doesn't match")
	}
	fmt.Printf("enter name (%v):", plan.Name)
	response, err = reader.ReadString('\n')
	if err != nil {
		logreport.Fatal(err)
	}
	response = strings.Trim(response, "\n")
	if plan.Name != response {
		logreport.Fatal("the entered name doesn't match")
	}
	err = db.DoInTransaction(func(tx *sql.Tx) error {
		return model.DeletePlan(tx, plan.ID)
	})
	if err != nil {
		logreport.Fatal(err)
	}
	fmt.Printf("Destroyed plan %v %v %v %v %v %v\n", plan.ID, plan.Name, plan.StripeName, plan.MaxUsers, plan.JavascriptTimeout, plan.Price)
}

func processCommands(cmds []string, conf config.Configuration, db *sql.DB) {
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
	commands[cmds[0]].Process(params, conf, db)
}
