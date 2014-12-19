package sql

import (
	"fmt"
	"testing"
)

func v1TestDB() *DB {
	db, _ := setupFreshMemoryDB()
	setupSchemaTable(db)
	migrateToV1(db)
	return db
}

func TestAccountsPresence(t *testing.T) {
	db := v1TestDB()
	_, err := db.Query("SELECT COUNT(*) FROM `accounts`;")
	if err != nil {
		t.Errorf("Should not error counting accounts: %v", err)
	}
}

func TestAccountsNameNotNull(t *testing.T) {
	db := v1TestDB()
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
	db := v1TestDB()
	_, err := db.Exec("INSERT INTO `accounts` (`name`) VALUES ('Foo Corp');")
	if err != nil {
		t.Errorf("Should not error on first insertion: %v", err)
	}
	_, err = db.Exec("INSERT INTO `accounts` (`name`) VALUES ('Foo Corp');")
	if err == nil {
		t.Errorf("Should error on duplicate insertion")
	}
}

func v1AccountSeededDB() *DB {
	db := v1TestDB()
	db.Exec("INSERT INTO `accounts` (`id`, `name`) VALUES (1, 'Foo Corp');")
	return db
}

func v1UserInsert(name, email, pw string) string {
	return fmt.Sprintf("INSERT INTO `users` "+
		"(`account_id`, `name`, `email`, `password`) "+
		"VALUES (1, %s, %s, %s);", name, email, pw)
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
