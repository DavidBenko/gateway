package model

import (
	apsql "gateway/sql"

	"github.com/jmoiron/sqlx/types"
)

// Environment represents a environment the API is available on.
type Environment struct {
	AccountID int64 `json:"-"`
	APIID     int64 `json:"api_id" db:"api_id"`

	ID          int64          `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Data        types.JsonText `json:"data"`
}

// Validate validates the model.
func (e *Environment) Validate() Errors {
	errors := make(Errors)
	if e.Name == "" {
		errors.add("name", "must not be blank")
	}
	return errors
}

// ValidateFromDatabaseError translates possible database constraint errors
// into validation errors.
func (e *Environment) ValidateFromDatabaseError(err error) Errors {
	errors := make(Errors)
	if err.Error() == "UNIQUE constraint failed: environments.api_id, environments.name" ||
		err.Error() == `pq: duplicate key value violates unique constraint "environments_api_id_name_key"` {
		errors.add("name", "is already taken")
	}
	return errors
}

// AllEnvironmentsForAPIIDAndAccountID returns all environments on the Account's API in default order.
func AllEnvironmentsForAPIIDAndAccountID(db *apsql.DB, apiID, accountID int64) ([]*Environment, error) {
	environments := []*Environment{}
	err := db.Select(&environments, db.SQL("environments/all"), apiID, accountID)
	return environments, err
}

// FindEnvironmentForAPIIDAndAccountID returns the environment with the id, api id, and account_id specified.
func FindEnvironmentForAPIIDAndAccountID(db *apsql.DB, id, apiID, accountID int64) (*Environment, error) {
	environment := Environment{}
	err := db.Get(&environment, db.SQL("environments/find"), id, apiID, accountID)
	return &environment, err
}

// FindEnvironmentForProxy returns the environment with the id specified.
func FindEnvironmentForProxy(db *apsql.DB, id int64) (*Environment, error) {
	environment := Environment{}
	err := db.Get(&environment, db.SQL("environments/find_proxy"), id)
	return &environment, err
}

// DeleteEnvironmentForAPIIDAndAccountID deletes the environment with the id, api_id and account_id specified.
func DeleteEnvironmentForAPIIDAndAccountID(tx *apsql.Tx, id, apiID, accountID int64) error {
	return tx.DeleteOne(tx.SQL("environments/delete"), id, apiID, accountID)
}

// Insert inserts the environment into the database as a new row.
func (e *Environment) Insert(tx *apsql.Tx) error {
	data, err := e.Data.MarshalJSON()
	if err != nil {
		return err
	}
	e.ID, err = tx.InsertOne(tx.SQL("environments/insert"),
		e.APIID, e.AccountID, e.Name, e.Description, string(data))
	return err
}

// Update updates the environment in the database.
func (e *Environment) Update(tx *apsql.Tx) error {
	data, err := e.Data.MarshalJSON()
	if err != nil {
		return err
	}
	return tx.UpdateOne(tx.SQL("environments/update"),
		e.Name, e.Description, string(data), e.ID, e.APIID, e.AccountID)
}
