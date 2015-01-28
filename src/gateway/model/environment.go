package model

import (
	"encoding/json"
	apsql "gateway/sql"
)

// Environment represents a environment the API is available on.
type Environment struct {
	AccountID int64 `json:"-"`
	APIID     int64 `json:"-" db:"api_id"`

	ID          int64           `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Data        json.RawMessage `json:"data"`
}

// Validate validates the model.
func (e *Environment) Validate() Errors {
	errors := make(Errors)
	if e.Name == "" {
		errors.add("name", "must not be blank")
	}
	return errors
}

// AllEnvironmentsForAPIIDAndAccountID returns all environments on the Account's API in default order.
func AllEnvironmentsForAPIIDAndAccountID(db *apsql.DB, apiID, accountID int64) ([]*Environment, error) {
	environments := []*Environment{}
	err := db.Select(&environments,
		`SELECT
			environments.id as id,
			environments.name as name,
			environments.description as description,
			environments.data as data
		FROM environments, apis
		WHERE environments.api_id = ?
			AND environments.api_id = apis.id
			AND apis.account_id = ?
		ORDER BY environments.name ASC;`,
		apiID, accountID)
	return environments, err
}

// FindEnvironmentForAPIIDAndAccountID returns the environment with the id, api id, and account_id specified.
func FindEnvironmentForAPIIDAndAccountID(db *apsql.DB, id, apiID, accountID int64) (*Environment, error) {
	environment := Environment{}
	err := db.Get(&environment,
		`SELECT
			environments.id as id,
			environments.name as name,
			environments.description as description
			environments.data as data
		FROM environments, apis
		WHERE environments.id = ?
			AND environments.api_id = ?
			AND environments.api_id = apis.id
			AND apis.account_id = ?;`,
		id, apiID, accountID)
	return &environment, err
}

// DeleteEnvironmentForAPIIDAndAccountID deletes the environment with the id, api_id and account_id specified.
func DeleteEnvironmentForAPIIDAndAccountID(tx *apsql.Tx, id, apiID, accountID int64) error {
	return tx.DeleteOne(
		`DELETE FROM environments
		WHERE environments.id = ?
			AND environments.api_id IN
				(SELECT id FROM apis WHERE id = ? AND account_id = ?);`,
		id, apiID, accountID)
}

// Insert inserts the environment into the database as a new row.
func (e *Environment) Insert(tx *apsql.Tx) error {
	data, err := e.Data.MarshalJSON()
	if err != nil {
		return err
	}
	e.ID, err = tx.InsertOne(
		`INSERT INTO environments (api_id, name, description, data)
		 VALUES ((SELECT id FROM apis WHERE id = ? AND account_id = ?),?, ?, ?)`,
		e.APIID, e.AccountID, e.Name, e.Description, string(data))
	return err
}

// Update updates the environment in the database.
func (e *Environment) Update(tx *apsql.Tx) error {
	data, err := e.Data.MarshalJSON()
	if err != nil {
		return err
	}
	return tx.UpdateOne(
		`UPDATE environments
		SET name = ?, description = ?, data = ?
		WHERE environments.id = ?
			AND environments.api_id IN
			(SELECT id FROM apis WHERE id = ? AND account_id = ?);`,
		e.Name, e.Description, string(data), e.ID, e.APIID, e.AccountID)
}
