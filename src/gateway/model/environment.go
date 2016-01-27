package model

import (
	"errors"
	aperrors "gateway/errors"
	aphttp "gateway/http"
	apsql "gateway/sql"

	"github.com/jmoiron/sqlx/types"
)

const (
	SessionTypeClient = "client"
	SessionTypeServer = "server"
)

// Environment represents a environment the API is available on.
type Environment struct {
	AccountID int64 `json:"-"`
	UserID    int64 `json:"-"`
	APIID     int64 `json:"api_id,omitempty" db:"api_id"`

	ID          int64          `json:"id,omitempty"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Data        types.JsonText `json:"data"`

	SessionType                string `json:"session_type" db:"session_type"`
	SessionHeader              string `json:"session_header" db:"session_header"`
	SessionName                string `json:"session_name" db:"session_name"`
	SessionAuthKey             string `json:"session_auth_key" db:"session_auth_key"`
	SessionEncryptionKey       string `json:"session_encryption_key" db:"session_encryption_key"`
	SessionAuthKeyRotate       string `json:"session_auth_key_rotate" db:"session_auth_key_rotate"`
	SessionEncryptionKeyRotate string `json:"session_encryption_key_rotate" db:"session_encryption_key_rotate"`

	ShowJavascriptErrors bool `json:"show_javascript_errors" db:"show_javascript_errors"`
}

// Validate validates the model.
func (e *Environment) Validate(isInsert bool) aperrors.Errors {
	errors := make(aperrors.Errors)
	if e.Name == "" {
		errors.Add("name", "must not be blank")
	}
	return errors
}

// ValidateFromDatabaseError translates possible database constraint errors
// into validation errors.
func (e *Environment) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	if err.Error() == "UNIQUE constraint failed: environments.api_id, environments.name" ||
		err.Error() == `pq: duplicate key value violates unique constraint "environments_api_id_name_key"` {
		errors.Add("name", "is already taken")
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

// CanDeleteEnvironment checks whether deleting would violate any constraints
func CanDeleteEnvironment(tx *apsql.Tx, id, accountID int64, auth aphttp.AuthType) error {
	var count int64
	if err := tx.Get(&count,
		`SELECT COUNT(id) FROM proxy_endpoints
		 WHERE environment_id = ?;`, id); err != nil {
		return errors.New("Could not check if environment could be deleted.")
	}

	if count > 0 {
		return errors.New("There are proxy endpoints that reference this environment.")
	}

	return nil
}

// DeleteEnvironmentForAPIIDAndAccountID deletes the environment with the id, api_id and account_id specified.
func DeleteEnvironmentForAPIIDAndAccountID(tx *apsql.Tx, id, apiID, accountID, userID int64) error {
	err := tx.DeleteOne(tx.SQL("environments/delete"), id, apiID, accountID)
	if err != nil {
		return err
	}
	return tx.Notify("environments", accountID, userID, apiID, 0, id, apsql.Delete)
}

// Insert inserts the environment into the database as a new row.
func (e *Environment) Insert(tx *apsql.Tx) error {
	data, err := marshaledForStorage(e.Data)
	if err != nil {
		return err
	}
	e.ID, err = tx.InsertOne(tx.SQL("environments/insert"),
		e.APIID, e.AccountID, e.Name, e.Description, data,
		e.SessionType, e.SessionHeader,
		e.SessionName, e.SessionAuthKey, e.SessionEncryptionKey,
		e.SessionAuthKeyRotate, e.SessionEncryptionKeyRotate, e.ShowJavascriptErrors)
	if err != nil {
		return err
	}
	return tx.Notify("environments", e.AccountID, e.UserID, e.APIID, 0, e.ID, apsql.Insert)
}

// Update updates the environment in the database.
func (e *Environment) Update(tx *apsql.Tx) error {
	data, err := marshaledForStorage(e.Data)
	if err != nil {
		return err
	}
	err = tx.UpdateOne(tx.SQL("environments/update"),
		e.Name, e.Description, data,
		e.SessionType, e.SessionHeader,
		e.SessionName, e.SessionAuthKey, e.SessionEncryptionKey,
		e.SessionAuthKeyRotate, e.SessionEncryptionKeyRotate, e.ShowJavascriptErrors,
		e.ID, e.APIID, e.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify("environments", e.AccountID, e.UserID, e.APIID, 0, e.ID, apsql.Update)
}
