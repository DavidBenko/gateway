package model

import (
	"errors"
	"fmt"

	"gateway/license"
	"gateway/sql"
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

// ValidateFromDatabaseError translates possible database constraint errors
// into validation errors.
func (a *Account) ValidateFromDatabaseError(err error) Errors {
	errors := make(Errors)
	if sql.IsUniqueConstraint(err, "accounts", "name") {
		errors.add("name", "is already taken")
	}
	return errors
}

// AllAccounts returns all accounts in default order.
func AllAccounts(db *sql.DB) ([]*Account, error) {
	accounts := []*Account{}
	err := db.Select(&accounts, db.SQL("accounts/all"))
	return accounts, err
}

// FirstAccount returns the first account found.
func FirstAccount(db *sql.DB) (*Account, error) {
	account := Account{}
	err := db.Get(&account, db.SQL("accounts/first"))
	return &account, err
}

// FindAccount returns the account with the id specified.
func FindAccount(db *sql.DB, id int64) (*Account, error) {
	account := Account{}
	err := db.Get(&account, db.SQL("accounts/find"), id)
	return &account, err
}

// DeleteAccount deletes the account with the id specified.
func DeleteAccount(tx *sql.Tx, id int64) error {
	err := tx.DeleteOne(tx.SQL("accounts/delete"), id)
	if err != nil {
		return err
	}
	return tx.Notify("accounts", id, 0, sql.Delete)
}

// Insert inserts the account into the database as a new row.
func (a *Account) Insert(tx *sql.Tx) (err error) {
	if license.DeveloperVersion {
		var count int
		tx.Get(&count, tx.SQL("accounts/count"))
		if count >= license.DeveloperVersionAccounts {
			return errors.New(fmt.Sprintf("Developer version allows %v account(s).", license.DeveloperVersionAccounts))
		}
	}

	a.ID, err = tx.InsertOne(tx.SQL("accounts/insert"), a.Name)
	return err
}

// Update updates the account in the database.
func (a *Account) Update(tx *sql.Tx) error {
	return tx.UpdateOne(tx.SQL("accounts/update"), a.Name, a.ID)
}
