package sql

import (
	"fmt"
	"testing"
)

////
/// Test Data
//

func v1AccountSeededDB() *DB {
	db := testDB(1)
	db.Exec("INSERT INTO `accounts` (`id`, `name`) VALUES (1, 'Foo Corp');")
	return db
}

////
/// Accounts
//

func TestAccountsPresence(t *testing.T) {
	db := testDB(1)
	_, err := db.Query("SELECT COUNT(*) FROM `accounts`;")
	if err != nil {
		t.Errorf("Should not error counting accounts: %v", err)
	}
}

func TestAccountsNameNotNull(t *testing.T) {
	db := testDB(1)
	_, err := db.Exec("INSERT INTO `accounts`;")
	if err == nil {
		t.Errorf("Should error without name")
	}
	_, err = db.Exec("INSERT INTO `accounts` (`name`) VALUES ('Foo Corp');")
	if err != nil {
		t.Errorf("Should not error with name: %v", err)
	}
}

func TestAccountsNameUnique(t *testing.T) {
	db := testDB(1)
	_, err := db.Exec("INSERT INTO `accounts` (`name`) VALUES ('Foo Corp');")
	if err != nil {
		t.Errorf("Should not error on first insertion: %v", err)
	}
	_, err = db.Exec("INSERT INTO `accounts` (`name`) VALUES ('Foo Corp');")
	if err == nil {
		t.Errorf("Should error on duplicate insertion")
	}
}

////
/// APIs
//

func TestAPIsPresence(t *testing.T) {
	db := testDB(1)
	_, err := db.Query("SELECT COUNT(*) FROM `apis`;")
	if err != nil {
		t.Errorf("Should not error counting apis: %v", err)
	}
}

////
/// Endpoint Groups
//

func TestEndpointGroupsPresence(t *testing.T) {
	db := testDB(1)
	_, err := db.Query("SELECT COUNT(*) FROM `endpoint_groups`;")
	if err != nil {
		t.Errorf("Should not error counting endpoint groups: %v", err)
	}
}

////
/// Environments
//

func TestEnvironmentsPresence(t *testing.T) {
	db := testDB(1)
	_, err := db.Query("SELECT COUNT(*) FROM `environments`;")
	if err != nil {
		t.Errorf("Should not error counting environments: %v", err)
	}
}

////
/// Hosts
//

func TestHostsPresence(t *testing.T) {
	db := testDB(1)
	_, err := db.Query("SELECT COUNT(*) FROM `hosts`;")
	if err != nil {
		t.Errorf("Should not error counting hosts: %v", err)
	}
}

////
/// Proxy Endpoint Calls
//

func TestProxyEndpointCallsPresence(t *testing.T) {
	db := testDB(1)
	_, err := db.Query("SELECT COUNT(*) FROM `proxy_endpoint_calls`;")
	if err != nil {
		t.Errorf("Should not error counting proxy endpoint calls: %v", err)
	}
}

////
/// Proxy Endpoint Components
//

func TestProxyEndpointComponentsPresence(t *testing.T) {
	db := testDB(1)
	_, err := db.Query("SELECT COUNT(*) FROM `proxy_endpoint_components`;")
	if err != nil {
		t.Errorf("Should not error counting proxy endpoint components: %v", err)
	}
}

////
/// Proxy Endpoint Transformations
//

func TestProxyEndpointTransformationsPresence(t *testing.T) {
	db := testDB(1)
	_, err := db.Query("SELECT COUNT(*) FROM `proxy_endpoint_transformations`;")
	if err != nil {
		t.Errorf("Should not error counting proxy endpoint transformations: %v", err)
	}
}

////
/// Proxy Endpoints
//

func TestProxyEndpointsPresence(t *testing.T) {
	db := testDB(1)
	_, err := db.Query("SELECT COUNT(*) FROM `proxy_endpoints`;")
	if err != nil {
		t.Errorf("Should not error counting proxy endpoints: %v", err)
	}
}

////
/// Remote Endpoint Environment Data
//

func TestRemoteEndpointEnvironmentDataPresence(t *testing.T) {
	db := testDB(1)
	_, err := db.Query("SELECT COUNT(*) FROM `remote_endpoint_environment_data`;")
	if err != nil {
		t.Errorf("Should not error counting remote endpoint environment data: %v", err)
	}
}

////
/// Remote Endpoints
//

func TestRemoteEndpointsPresence(t *testing.T) {
	db := testDB(1)
	_, err := db.Query("SELECT COUNT(*) FROM `remote_endpoints`;")
	if err != nil {
		t.Errorf("Should not error counting remote endpoints: %v", err)
	}
}

////
/// Users
//

func v1UserInsert(name, email, pw string) string {
	return v1UserInsertWithAccount(name, email, pw, 1)
}

func v1UserInsertWithAccount(name, email, pw string, acct int64) string {
	return fmt.Sprintf("INSERT INTO `users` "+
		"(`account_id`, `name`, `email`, `hashed_password`) "+
		"VALUES (%d, %s, %s, %s);", acct, name, email, pw)
}

func TestUsersPresence(t *testing.T) {
	db := testDB(1)
	_, err := db.Query("SELECT COUNT(*) FROM `users`;")
	if err != nil {
		t.Errorf("Should not error counting users: %v", err)
	}
}

func TestUsersNameNotNull(t *testing.T) {
	db := v1AccountSeededDB()
	_, err := db.Exec(v1UserInsert("NULL", "'geff@foo.com'", "'secure'"))
	if err == nil {
		t.Errorf("Should error without name")
	}
	_, err = db.Exec(v1UserInsert("'Geff'", "'geff@foo.com'", "'secure'"))
	if err != nil {
		t.Errorf("Should not error with name: %v", err)
	}
}

func TestUsersEmailNotNull(t *testing.T) {
	db := v1AccountSeededDB()
	_, err := db.Exec(v1UserInsert("'Geff'", "NULL", "secure"))
	if err == nil {
		t.Errorf("Should error without email")
	}
	_, err = db.Exec(v1UserInsert("'Geff'", "'geff@foo.com'", "'secure'"))
	if err != nil {
		t.Errorf("Should not error with email: %v", err)
	}
}

func TestUsersEmailUnique(t *testing.T) {
	db := v1AccountSeededDB()
	_, err := db.Exec(v1UserInsert("'Geff'", "'geff@foo.com'", "'secure'"))
	if err != nil {
		t.Errorf("Should not error on first insertion: %v", err)
	}
	_, err = db.Exec(v1UserInsert("'Geff'", "'geff@foo.com'", "'secure'"))
	if err == nil {
		t.Errorf("Should error on duplicate insertion")
	}
}

func TestUsersPasswordNotNull(t *testing.T) {
	db := v1AccountSeededDB()
	_, err := db.Exec(v1UserInsert("'Geff'", "'geff@foo.com'", "NULL"))
	if err == nil {
		t.Errorf("Should error without password")
	}
	_, err = db.Exec(v1UserInsert("'Geff'", "'geff@foo.com'", "'secure'"))
	if err != nil {
		t.Errorf("Should not error with password: %v", err)
	}
}
