package model

import (
	"encoding/json"
	"fmt"
	"gateway/config"
	apsql "gateway/sql"
	"log"
)

// ProxyEndpoint holds the data to power the proxy for a given API endpoint.
type ProxyEndpoint struct {
	AccountID       int64  `json:"-"`
	APIID           int64  `json:"-" db:"api_id"`
	EndpointGroupID *int64 `json:"endpoint_group_id" db:"endpoint_group_id"`
	EnvironmentID   int64  `json:"environment_id" db:"environment_id"`

	ID                int64           `json:"id"`
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	Active            bool            `json:"active"`
	CORSEnabled       bool            `json:"cors_enabled" db:"cors_enabled"`
	CORSAllowOverride *string         `json:"cors_allow_override" db:"cors_allow_override"`
	Routes            json.RawMessage `json:"routes"`
}

// ProxyEndpointRoute is a route on which the endpoint should be accessible.
type ProxyEndpointRoute struct {
	Path    string   `json:"path"`
	Methods []string `json:"methods"`
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
	return _proxyEndpoints(db, 0, apiID, accountID)
}

// FindProxyEndpointForAPIIDAndAccountID returns the proxyEndpoint with the id, api id, and account_id specified.
func FindProxyEndpointForAPIIDAndAccountID(db *apsql.DB, id, apiID, accountID int64) (*ProxyEndpoint, error) {
	endpoints, err := _proxyEndpoints(db, id, apiID, accountID)
	if err != nil {
		return nil, err
	}
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("No endpoint with id %d found", id)
	}
	return endpoints[0], nil
}

func _proxyEndpoints(db *apsql.DB, id, apiID, accountID int64) ([]*ProxyEndpoint, error) {
	args := []interface{}{}
	query := "SELECT " +
		"  `proxy_endpoints`.`id` as `id`, " +
		"  `proxy_endpoints`.`name` as `name`, " +
		"  `proxy_endpoints`.`description` as `description`, " +
		"  `proxy_endpoints`.`endpoint_group_id` as `endpoint_group_id`, " +
		"  `proxy_endpoints`.`environment_id` as `environment_id`, " +
		"  `proxy_endpoints`.`active` as `active`, " +
		"  `proxy_endpoints`.`cors_enabled` as `cors_enabled`, " +
		"  `proxy_endpoints`.`cors_allow_override` as `cors_allow_override`, " +
		"  `proxy_endpoints`.`routes` as `routes` " +
		"FROM `proxy_endpoints`, `apis` " +
		"WHERE "
	if id != 0 {
		query = query + "`proxy_endpoints`.`id` = ? AND "
		args = append(args, id)
	}
	query = query + "`proxy_endpoints`.`api_id` = ? " +
		"  AND `proxy_endpoints`.`api_id` = `apis`.`id` " +
		"  AND `apis`.`account_id` = ? " +
		"ORDER BY `proxy_endpoints`.`name` ASC, " +
		"  `proxy_endpoints`.`id` ASC;"

	args = append(args, apiID, accountID)
	proxyEndpoints := []*ProxyEndpoint{}
	err := db.Select(&proxyEndpoints, query, args...)
	return proxyEndpoints, err
}

// DeleteProxyEndpointForAPIIDAndAccountID deletes the proxyEndpoint with the id, api_id and account_id specified.
func DeleteProxyEndpointForAPIIDAndAccountID(tx *apsql.Tx, id, apiID, accountID int64) error {
	result, err := tx.Exec(
		"DELETE FROM `proxy_endpoints` "+
			"WHERE `proxy_endpoints`.`id` = ? "+
			"  AND `proxy_endpoints`.`api_id` IN "+
			"      (SELECT `id` FROM `apis` WHERE `id` = ? AND `account_id` = ?);",
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
	routes, err := e.Routes.MarshalJSON()
	if err != nil {
		return err
	}
	result, err := tx.Exec(
		"INSERT INTO `proxy_endpoints` "+
			"(`api_id`, `name`, `description`, "+
			"`endpoint_group_id`, `environment_id`, "+
			"`active`, `cors_enabled`, `cors_allow_override`, "+
			"`routes`) "+
			"VALUES ("+
			"  (SELECT `id` FROM `apis` WHERE `id` = ? AND `account_id` = ?), "+
			"  ?, ?, "+
			"  (SELECT `id` FROM `endpoint_groups` WHERE `id` = ? AND `api_id` = ?), "+
			"  (SELECT `id` FROM `environments` WHERE `id` = ? AND `api_id` = ?), "+
			"  ?, ?, ?, ?);",
		e.APIID, e.AccountID, e.Name, e.Description, e.EndpointGroupID, e.APIID,
		e.EnvironmentID, e.APIID, e.Active, e.CORSEnabled, e.CORSAllowOverride, string(routes))
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
	routes, err := e.Routes.MarshalJSON()
	if err != nil {
		return err
	}
	result, err := tx.Exec(
		"UPDATE `proxy_endpoints` "+
			"SET `name` = ?, "+
			"  `description` = ?, "+
			"  `endpoint_group_id` = "+
			"     (SELECT `id` FROM `endpoint_groups` WHERE `id` = ? AND `api_id` = ?), "+
			"  `environment_id` = "+
			"     (SELECT `id` FROM `environments` WHERE `id` = ? AND `api_id` = ?), "+
			"  `active` = ?, "+
			"  `cors_enabled` = ?, "+
			"  `cors_allow_override` = ?, "+
			"  `routes` = ? "+
			"WHERE `proxy_endpoints`.`id` = ? "+
			"  AND `proxy_endpoints`.`api_id` IN "+
			"      (SELECT `id` FROM `apis` WHERE `id` = ? AND `account_id` = ?)",
		e.Name, e.Description,
		e.EndpointGroupID, e.APIID,
		e.EnvironmentID, e.APIID,
		e.Active, e.CORSEnabled, e.CORSAllowOverride,
		string(routes),
		e.ID, e.APIID, e.AccountID)
	if err != nil {
		return err
	}
	numRows, err := result.RowsAffected()
	if err != nil || numRows != 1 {
		return fmt.Errorf("Expected 1 row to be affected; got %d, error: %v", numRows, err)
	}
	return nil
}
