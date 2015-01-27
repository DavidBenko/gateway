package model

import (
	"fmt"
	"gateway/config"
	"gateway/sql"
	"log"
)

// Account represents a single tenant in multi-tenant deployment.
type Account struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Validate validates the model.
func (a *Account) Validate() Errors {
	errors := make(Errors)
	if a.Name == "" {
		errors.add("name", "must not be blank")
	}
	return errors
}

// ValidationErrorsFromDatabase translates possible database constraint errors
// into validation errors.
func (a *Account) ValidateFromDatabaseError(err error) Errors {
	errors := make(Errors)
	if err.Error() == "UNIQUE constraint failed: accounts.name" {
		errors.add("name", "is already taken")
	}
	return errors
}

// AllAccounts returns all accounts in default order.
func AllAccounts(db *sql.DB) ([]*Account, error) {
	accounts := []*Account{}
	err := db.Select(&accounts,
		"SELECT `id`, `name` FROM `accounts` ORDER BY `name` ASC;")
	return accounts, err
}

// FindAccount returns the account with the id specified.
func FindAccount(db *sql.DB, id int64) (*Account, error) {
	account := Account{}
	err := db.Get(&account, "SELECT `id`, `name` FROM `accounts` WHERE `id` = ?;", id)
	return &account, err
}

// DeleteAccount deletes the account with the id specified.
func DeleteAccount(tx *sql.Tx, id int64) error {
	result, err := tx.Exec("DELETE FROM `accounts` WHERE `id` = ?;", id)
	if err != nil {
		return err
	}

	numRows, err := result.RowsAffected()
	if err != nil || numRows != 1 {
		return fmt.Errorf("Expected 1 row to be affected; got %d, error: %v", numRows, err)
	}

	return nil
}

// Insert inserts the account into the database as a new row.
func (a *Account) Insert(tx *sql.Tx) error {
	result, err := tx.Exec("INSERT INTO `accounts` (`name`) VALUES (?);",
		a.Name)
	if err != nil {
		return err
	}
	a.ID, err = result.LastInsertId()
	if err != nil {
		log.Printf("%s Error getting last insert ID for account: %v",
			config.System, err)
		return err
	}
	return nil
}

// Update updates the account in the database.
func (a *Account) Update(tx *sql.Tx) error {
	result, err := tx.Exec("UPDATE `accounts` SET `name` = ? WHERE `id` = ?;",
		a.Name, a.ID)
	if err != nil {
		return err
	}
	numRows, err := result.RowsAffected()
	if err != nil || numRows != 1 {
		return fmt.Errorf("Expected 1 row to be affected; got %d, error: %v", numRows, err)
	}
	return nil
}
