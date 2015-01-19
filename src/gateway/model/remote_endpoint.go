package model

import (
	"fmt"
	"gateway/config"
	apsql "gateway/sql"
	"log"
)

// RemoteEndpoint is an endpoint that a proxy endpoint delegates to.
type RemoteEndpoint struct {
	AccountID int64 `json:"-"`
	APIID     int64 `json:"-" db:"api_id"`

	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Validate validates the model.
func (e *RemoteEndpoint) Validate() Errors {
	errors := make(Errors)
	if e.Name == "" {
		errors.add("name", "must not be blank")
	}
	return errors
}

// AllRemoteEndpointsForAPIIDAndAccountID returns all remoteEndpoints on the Account's API in default order.
func AllRemoteEndpointsForAPIIDAndAccountID(db *apsql.DB, apiID, accountID int64) ([]*RemoteEndpoint, error) {
	remoteEndpoints := []*RemoteEndpoint{}
	err := db.Select(&remoteEndpoints,
		"SELECT "+
			"  `remote_endpoints`.`id` as `id`, "+
			"  `remote_endpoints`.`name` as `name`, "+
			"  `remote_endpoints`.`description` as `description` "+
			"FROM `remote_endpoints`, `apis` "+
			"WHERE `remote_endpoints`.`api_id` = ? "+
			"  AND `remote_endpoints`.`api_id` = `apis`.`id` "+
			"  AND `apis`.`account_id` = ? "+
			"ORDER BY `remote_endpoints`.`name` ASC;",
		apiID, accountID)
	return remoteEndpoints, err
}

// FindRemoteEndpointForAPIIDAndAccountID returns the remoteEndpoint with the id, api id, and account_id specified.
func FindRemoteEndpointForAPIIDAndAccountID(db *apsql.DB, id, apiID, accountID int64) (*RemoteEndpoint, error) {
	remoteEndpoint := RemoteEndpoint{}
	err := db.Get(&remoteEndpoint,
		"SELECT "+
			"  `remote_endpoints`.`id` as `id`, "+
			"  `remote_endpoints`.`name` as `name`, "+
			"  `remote_endpoints`.`description` as `description` "+
			"FROM `remote_endpoints`, `apis` "+
			"WHERE `remote_endpoints`.`id` = ? "+
			"  AND `remote_endpoints`.`api_id` = ? "+
			"  AND `remote_endpoints`.`api_id` = `apis`.`id` "+
			"  AND `apis`.`account_id` = ?;",
		id, apiID, accountID)
	return &remoteEndpoint, err
}

// DeleteRemoteEndpointForAPIIDAndAccountID deletes the remoteEndpoint with the id, api_id and account_id specified.
func DeleteRemoteEndpointForAPIIDAndAccountID(tx *apsql.Tx, id, apiID, accountID int64) error {
	result, err := tx.Exec(
		"DELETE FROM `remote_endpoints` "+
			"WHERE `remote_endpoints`.`id` = ? "+
			"  AND `remote_endpoints`.`api_id` IN "+
			"      (SELECT `id` FROM `apis` WHERE `id` = ? AND `account_id` = ?)",
		id, apiID, accountID)
	if err != nil {
		return err
	}

	numRows, err := result.RowsAffected()
	if err != nil || numRows != 1 {
		return fmt.Errorf("Expected 1 row to be affected; got %d, error: %v", numRows, err)
	}

	return nil
}

// Insert inserts the remoteEndpoint into the database as a new row.
func (e *RemoteEndpoint) Insert(tx *apsql.Tx) error {
	result, err := tx.Exec(
		"INSERT INTO `remote_endpoints` (`api_id`, `name`, `description`) "+
			"VALUES ( "+
			"  (SELECT `id` FROM `apis` WHERE `id` = ? AND `account_id` = ?), "+
			"  ?, ?);",
		e.APIID, e.AccountID, e.Name, e.Description)
	if err != nil {
		return err
	}
	e.ID, err = result.LastInsertId()
	if err != nil {
		log.Printf("%s Error getting last insert ID for remoteEndpoint: %v",
			config.System, err)
		return err
	}
	return nil
}

// Update updates the remoteEndpoint in the database.
func (e *RemoteEndpoint) Update(tx *apsql.Tx) error {
	result, err := tx.Exec(
		"UPDATE `remote_endpoints` "+
			"SET `name` = ?, `description` = ? "+
			"WHERE `remote_endpoints`.`id` = ? "+
			"  AND `remote_endpoints`.`api_id` IN "+
			"      (SELECT `id` FROM `apis` WHERE `id` = ? AND `account_id` = ?)",
		e.Name, e.Description, e.ID, e.APIID, e.AccountID)
	if err != nil {
		return err
	}
	numRows, err := result.RowsAffected()
	if err != nil || numRows != 1 {
		return fmt.Errorf("Expected 1 row to be affected; got %d, error: %v", numRows, err)
	}
	return nil
}
