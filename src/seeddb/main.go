package main

import (
	"fmt"
	"gateway/model"
	apsql "gateway/sql"
	"log"

	"github.com/jmoiron/sqlx"
)

const numAccounts = 25000
const numUsersPerAccount = 2
const numApisPerAccount = 5
const numProxyEndpointsPerEnvironment = 5

var db *apsql.DB
var tx *apsql.Tx

func init() {
	xdb, err := sqlx.Connect("postgres", "user=rsnyder dbname=gateway_dev sslmode=disable")
	if err != nil {
		log.Fatalf("Uh-oh: couldn't connect to db")
	}

	db = &apsql.DB{DB: xdb, Driver: apsql.Postgres}
	_, err = model.AllAccounts(db)
	if err != nil {
		log.Fatalf("Uh-oh: couldn't get accounts due to %v", err)
	}
}

func main() {
	var err error
	tx, err = db.Begin()
	if err != nil {
		log.Fatalf("Uh-oh: couldn't begin transaction due to %v", err)
	}
	for i := 0; i < numAccounts; i++ {
		createAccount(i)
		if i%1000 == 0 {
			log.Printf("Inserted %d records", i)
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Fatalf("Uh-oh: couldn't commit transaction due to %v", err)
	}
}

func createAccount(accountNum int) {
	//log.Println("Creating account ...")
	account := &model.Account{Name: fmt.Sprintf("Account %d", accountNum)}
	err := account.Insert(tx)
	if err != nil {
		log.Fatalf("Uh-oh: Unable to insert account: %v", err)
	}

	createUsersForAccount(accountNum, account)
	createApisForAccount(accountNum, account)
}

func createUsersForAccount(accountNum int, account *model.Account) {
	for i := 0; i < numUsersPerAccount; i++ {
		_, err := tx.InsertOne(
			"INSERT INTO users(account_id, name, email, admin, confirmed, hashed_password) VALUES (?, ?, ?, ?, ?, ?)",
			account.ID,
			fmt.Sprintf("User_%d_%d", accountNum, i),
			fmt.Sprintf("user_%d_%d@example.com", accountNum, i),
			(i%numUsersPerAccount == 0),
			true,
			fmt.Sprintf("$2a$10$Rsj4BIPDKarA2yktRtUBOOL6h0RzqFVxAbPvMorb2YDdYjK/8rJUK%d", i),
		)
		if err != nil {
			log.Fatalf("Uh-oh: Unable to insert user for account: %v", err)
		}
	}
}

func createApisForAccount(accountNum int, account *model.Account) {
	for i := 0; i < numApisPerAccount; i++ {
		id, err := tx.InsertOne(
			"INSERT INTO apis(account_id, name, description, cors_allow_origin, cors_allow_headers, cors_allow_credentials, cors_request_headers, cors_max_age, enable_swagger) VALUES (?,?,?,?,?,?,?,?,?)",
			account.ID,
			fmt.Sprintf("API_%d_%d", accountNum, i),
			fmt.Sprintf("Description %d %d", accountNum, i),
			"*",
			"content-type, accept",
			true,
			"*",
			600,
			1,
		)
		if err != nil {
			log.Fatalf("Uh-oh: Unable to insert API for account: %v", err)
		}

		createLibraryForAPI(accountNum, i, id)
		createHostForAPI(accountNum, i, id)
		endpointGroupID := createEndpointGroupForAPI(accountNum, i, id)
		createEnvironmentForAPI(accountNum, i, id, endpointGroupID)
	}
}

func createLibraryForAPI(accountNum int, apiNum int, apiID int64) {
	_, err := tx.InsertOne(
		"INSERT INTO libraries(api_id, name, description, data) VALUES(?, ?, ?, ?)",
		apiID,
		fmt.Sprintf("Lib%d", apiID),
		fmt.Sprintf("Library for API %d", apiID),
		fmt.Sprintf(`"console.log(\"ZOMG %d\")"`, apiID),
	)
	if err != nil {
		log.Fatalf("Uh-oh: Unable to insert library: %v", err)
	}
}

func createHostForAPI(accountNum int, apiNum int, apiID int64) {
	_, err := tx.InsertOne(
		"INSERT INTO hosts(api_id, name, hostname) VALUES(?, ?, ?)",
		apiID,
		fmt.Sprintf("foo%d", apiID),
		fmt.Sprintf("foo%d.lvh.me", apiID),
	)
	if err != nil {
		log.Fatalf("Uh-oh: Unable to insert host: %v", err)
	}
}

func createEnvironmentForAPI(accountNum int, apiNum int, apiID int64, endpointGroupID int64) {
	envID, err := tx.InsertOne(
		"INSERT INTO environments(api_id, name, description, data, session_name, session_auth_key, session_encryption_key, session_auth_key_rotate, session_encryption_key_rotate, show_javascript_errors) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		apiID,
		fmt.Sprintf("Environment_%d_%d", apiID, 1),
		fmt.Sprintf("Environment %d for API ID %d", 1, apiID),
		fmt.Sprintf("{   \"xxx\": \"%d\"   }", apiID),
		fmt.Sprintf("SESSION_NAME%d", apiID),
		fmt.Sprintf("SESSION_AUTH_KEY%d", apiID),
		fmt.Sprintf("SESSION_ENCRYPTION_KEY%d", apiID),
		fmt.Sprintf("SESSION_AUTH_KEY_ROTATE%d", apiID),
		fmt.Sprintf("SESSION_ENCRYPTION_KEY_ROTATE%d", apiID),
		true,
	)
	if err != nil {
		log.Fatalf("Uh-oh: Unable to insert environment: %v", err)
	}

	createProxyEndpointsForEnvironment(accountNum, apiNum, apiID, envID, endpointGroupID)
}

func createProxyEndpointsForEnvironment(accountNum int, apiNum int, apiID int64, environmentID int64, endpointGroupID int64) {
	for i := 0; i < numProxyEndpointsPerEnvironment; i++ {
		proxyEndpointID, err := tx.InsertOne(
			`INSERT INTO proxy_endpoints(api_id, endpoint_group_id, environment_id,
       name, description, active, cors_enabled, routes)
       VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
			apiID,
			environmentID,
			endpointGroupID,
			fmt.Sprintf("ProxyEndpoint%d", i),
			fmt.Sprintf("Proxy Endpoint %d for API ID %d", i, apiID),
			true,
			true,
			fmt.Sprintf(`[{"id":"%d%d","path":"/chargepoint","get_method":true,"post_method":false,"put_method":false,"delete_method":false,"proxy_endpoint_id":1,"methods":["GET"]}]`, apiID, i),
		)
		if err != nil {
			log.Fatalf("Uh-oh: unable to insert proxy endpoint: %v", err)
		}

		remoteEndpointID := createRemoteEndpointForProxyEndpoint(proxyEndpointID, apiID, environmentID)

		createProxyEndpointChildren(proxyEndpointID, remoteEndpointID)
	}
}

func createProxyEndpointChildren(proxyEndpointID int64, remoteEndpointID int64) {
	componentID, err := tx.InsertOne(
		`INSERT INTO proxy_endpoint_components(endpoint_id, conditional, conditional_positive, position, type, data) VALUES(?,?,?,?,?,?)`,
		proxyEndpointID,
		"",
		true,
		0,
		"single",
		`""`,
	)
	if err != nil {
		log.Fatalf("Uh-oh: unable to insert proxy endpoint component: %v", err)
	}

	callID, err := tx.InsertOne(
		`INSERT INTO proxy_endpoint_calls(component_id, remote_endpoint_id, endpoint_name_override, conditional, conditional_positive, position) VALUES(?,?,?,?,?,?)`,
		componentID,
		remoteEndpointID,
		"",
		fmt.Sprintf(`log("Stuff %d")`, proxyEndpointID),
		true,
		0,
	)
	if err != nil {
		log.Fatalf("Uh-oh: unable to insert proxy endpoint call: %v", err)
	}

	sql := `INSERT INTO proxy_endpoint_transformations(component_id, call_id, before, position, type, data) VALUES(?,?,?,?,?,?)`
	_, err = tx.InsertOne(sql, componentID, callID, true, 0, "js", `"console.log(\"Oh hey.\");"`)
	if err != nil {
		log.Fatalf("Uh-oh: unable to insert proxy endpoint transformation for before: %v", err)
	}

	_, err = tx.InsertOne(sql, componentID, callID, false, 0, "js", `"console.log(\"Oh hey.\");"`)
	if err != nil {
		log.Fatalf("Uh-oh: unable to insert proxy endpoint transformation for after: %v", err)
	}

	testID, err := tx.InsertOne(
		`INSERT INTO proxy_endpoint_tests(endpoint_id, name, methods, route, body, data) VALUES(?,?,?,?,?,?)`,
		proxyEndpointID,
		fmt.Sprintf("TestForEndpoint%d", proxyEndpointID),
		`["GET"]`,
		"chargepoint",
		"",
		"null",
	)
	if err != nil {
		log.Fatalf("Uh-oh: unable to insert proxy endpoint test: %v", err)
	}

	_, err = tx.InsertOne(
		`INSERT INTO proxy_endpoint_test_pairs(test_id, type, key, value) VALUES(?,?,?,?)`,
		testID,
		"get",
		"a",
		"b",
	)
	if err != nil {
		log.Fatalf("Uh-oh: unable to insert proxy endpoint test pair: %v", err)
	}

	_, err = tx.InsertOne(
		`INSERT INTO proxy_endpoint_schemas(endpoint_id, name, request_type, request_schema, response_same_as_request, response_type, response_schema, data) VALUES (?,?,?,?,?,?,?,?)`,
		proxyEndpointID,
		fmt.Sprintf("Schema for endpoint %d", proxyEndpointID),
		"json_schema",
		"{}",
		false,
		"json_schema",
		"{}",
		"null",
	)
	if err != nil {
		log.Fatalf("Uh-oh: unable to insert proxy endpoint schema: %v", err)
	}
}

func createRemoteEndpointForProxyEndpoint(proxyEndpointID int64, apiID int64, environmentID int64) int64 {
	remoteEndpointID, err := tx.InsertOne(
		`INSERT INTO remote_endpoints(api_id, name, codename, description, type, data) VALUES(?, ?, ?, ?, ?, ?)`,
		apiID,
		fmt.Sprintf("RemoteEndpoint%d", proxyEndpointID),
		"endpoint",
		"Description goes here",
		"soap",
		`{"headers":{},"query":{},"serviceName":"chargepointservices","url":null,"wssePasswordCredentials":{"password":"217eb8c9c92b91c52f25c9354cc773e2","username":"e39e34c98def1db64229f0554306133b53fd057661b271409090934"}}`,
	)
	if err != nil {
		log.Fatalf("Uh-oh: unable to insert remote endpoint: %v", err)
	}

	_, err = tx.InsertOne(
		`INSERT INTO soap_remote_endpoints(remote_endpoint_id, wsdl, generated_jar, generated_jar_thumbprint) VALUES (?,?,?,?)`,
		remoteEndpointID,
		"XXXXXXXXXXX",
		[]byte("XXXXXXXXXXX_JAR_BYTES"),
		"GENERATED_JAR_THUMBPRINT",
	)
	if err != nil {
		log.Fatalf("Uh-oh: unable to insert soap remote endpoint: %v", err)
	}

	_, err = tx.Exec(
		`INSERT INTO remote_endpoint_environment_data(remote_endpoint_id, environment_id, data) VALUES(?, ?, ?)`,
		remoteEndpointID,
		environmentID,
		`{"headers":{},"query":{},"serviceName":"chargepointservices","url":null,"wssePasswordCredentials":{"password":"217eb8c9c92b91c52f25c9354cc773e2","username":"e39e34c98def1db64229f0554306133b53fd057661b271409090934"}}`,
	)
	if err != nil {
		log.Fatalf("Uh-oh: unable to insert remote endpoint environment data: %v", err)
	}

	return remoteEndpointID
}

func createEndpointGroupForAPI(accountNum int, apiNum int, apiID int64) int64 {
	id, err := tx.InsertOne(
		"INSERT INTO endpoint_groups(api_id, name, description) VALUES (?,?,?)",
		apiID,
		fmt.Sprintf("Endpoint_Group_%d", apiID),
		fmt.Sprintf("Endpoint group for api ID %d", apiID),
	)
	if err != nil {
		log.Fatalf("Uh-oh: Unable to insert endpoint group: %v", err)
	}
	return id
}
