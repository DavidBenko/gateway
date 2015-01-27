package admin

import (
	"fmt"
	"log"
	"net/http"

	"gateway/config"
	aphttp "gateway/http"
	"gateway/model"
	apsql "gateway/sql"
)

//go:generate ./serialize.rb Account

// AccountsController manages accounts
type AccountsController struct{}

// List returns a handler that lists the accounts.
func (c *AccountsController) List(w http.ResponseWriter, r *http.Request,
	db *apsql.DB) aphttp.Error {

	accounts, err := model.AllAccounts(db)
	if err != nil {
		log.Printf("%s Error listing accounts: %v", config.System, err)
		return aphttp.DefaultServerError()
	}

	return c.serializeCollection(accounts, w)
}

// Create returns a handler that creates the account.
func (c *AccountsController) Create(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {

	return c.insertOrUpdate(w, r, tx, true)
}

// Show returns a handler that shows the account.
func (c *AccountsController) Show(w http.ResponseWriter, r *http.Request,
	db *apsql.DB) aphttp.Error {

	id := instanceID(r)
	account, err := model.FindAccount(db, id)
	if err != nil {
		return aphttp.NewError(fmt.Errorf("No account with id %d", id), 404)
	}

	return c.serializeInstance(account, w)
}

// Update returns a handler that updates the account.
func (c *AccountsController) Update(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {

	return c.insertOrUpdate(w, r, tx, false)
}

// Delete is a handler that deletes the account.
func (c *AccountsController) Delete(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {

	err := model.DeleteAccount(tx, instanceID(r))
	if err != nil {
		log.Printf("%s Error deleting account: %v", config.System, err)
		return aphttp.DefaultServerError()
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (c *AccountsController) insertOrUpdate(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx, isInsert bool) aphttp.Error {
	account, httpErr := c.deserializeInstance(r)
	if httpErr != nil {
		return httpErr
	}

	var method func(*apsql.Tx) error
	var desc string
	if isInsert {
		method = account.Insert
		desc = "inserting"
	} else {
		account.ID = instanceID(r)
		method = account.Update
		desc = "updating"
	}

	validationErrors := account.Validate()
	if !validationErrors.Empty() {
		return SerializableValidationErrors{validationErrors}
	}

	if err := method(tx); err != nil {
		validationErrors = account.ValidateFromDatabaseError(err)
		if !validationErrors.Empty() {
			return SerializableValidationErrors{validationErrors}
		}
		log.Printf("%s Error %s account: %v", config.System, desc, err)
		return aphttp.DefaultServerError()
	}

	return c.serializeInstance(account, w)
}
