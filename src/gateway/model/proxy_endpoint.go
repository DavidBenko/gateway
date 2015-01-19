package model

import (
	"fmt"
	"gateway/config"
	apsql "gateway/sql"
	"log"
)

// ProxyEndpoint holds the data to power the proxy for a given API endpoint.
type ProxyEndpoint struct {
	AccountID int64 `json:"-"`
	APIID     int64 `json:"-" db:"api_id"`

	// Group       *EndpointGroup
	// Environment *Environment

	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Active      bool
	CORSEnabled bool
	CORSAllow   string
}

// Validate validates the model.
func (e *ProxyEndpoint) Validate() Errors {
	errors := make(Errors)
	if e.Name == "" {
		errors.add("name", "must not be blank")
	}
	return errors
}

// AllProxyEndpointsForAPIIDAndAccountID returns all proxyEndpoints on the Account's API in default order.
func AllProxyEndpointsForAPIIDAndAccountID(db *apsql.DB, apiID, accountID int64) ([]*ProxyEndpoint, error) {
	proxyEndpoints := []*ProxyEndpoint{}
	err := db.Select(&proxyEndpoints,
		"SELECT "+
			"  `proxy_endpoints`.`id` as `id`, "+
			"  `proxy_endpoints`.`name` as `name`, "+
			"  `proxy_endpoints`.`description` as `description` "+
			"FROM `proxy_endpoints`, `apis` "+
			"WHERE `proxy_endpoints`.`api_id` = ? "+
			"  AND `proxy_endpoints`.`api_id` = `apis`.`id` "+
			"  AND `apis`.`account_id` = ? "+
			"ORDER BY `proxy_endpoints`.`name` ASC;",
		apiID, accountID)
	return proxyEndpoints, err
}

// FindProxyEndpointForAPIIDAndAccountID returns the proxyEndpoint with the id, api id, and account_id specified.
func FindProxyEndpointForAPIIDAndAccountID(db *apsql.DB, id, apiID, accountID int64) (*ProxyEndpoint, error) {
	proxyEndpoint := ProxyEndpoint{}
	err := db.Get(&proxyEndpoint,
		"SELECT "+
			"  `proxy_endpoints`.`id` as `id`, "+
			"  `proxy_endpoints`.`name` as `name`, "+
			"  `proxy_endpoints`.`description` as `description` "+
			"FROM `proxy_endpoints`, `apis` "+
			"WHERE `proxy_endpoints`.`id` = ? "+
			"  AND `proxy_endpoints`.`api_id` = ? "+
			"  AND `proxy_endpoints`.`api_id` = `apis`.`id` "+
			"  AND `apis`.`account_id` = ?;",
		id, apiID, accountID)
	return &proxyEndpoint, err
}

// DeleteProxyEndpointForAPIIDAndAccountID deletes the proxyEndpoint with the id, api_id and account_id specified.
func DeleteProxyEndpointForAPIIDAndAccountID(tx *apsql.Tx, id, apiID, accountID int64) error {
	result, err := tx.Exec(
		"DELETE FROM `proxy_endpoints` "+
			"WHERE `proxy_endpoints`.`id` = ? "+
			"  AND `proxy_endpoints`.`api_id` IN "+
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

// Insert inserts the proxyEndpoint into the database as a new row.
func (e *ProxyEndpoint) Insert(tx *apsql.Tx) error {
	result, err := tx.Exec(
		"INSERT INTO `proxy_endpoints` (`api_id`, `name`, `description`) "+
			"VALUES ( "+
			"  (SELECT `id` FROM `apis` WHERE `id` = ? AND `account_id` = ?), "+
			"  ?, ?);",
		e.APIID, e.AccountID, e.Name, e.Description)
	if err != nil {
		return err
	}
	e.ID, err = result.LastInsertId()
	if err != nil {
		log.Printf("%s Error getting last insert ID for proxyEndpoint: %v",
			config.System, err)
		return err
	}
	return nil
}

// Update updates the proxyEndpoint in the database.
func (e *ProxyEndpoint) Update(tx *apsql.Tx) error {
	result, err := tx.Exec(
		"UPDATE `proxy_endpoints` "+
			"SET `name` = ?, `description` = ? "+
			"WHERE `proxy_endpoints`.`id` = ? "+
			"  AND `proxy_endpoints`.`api_id` IN "+
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
