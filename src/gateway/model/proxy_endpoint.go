package model

import (
	"encoding/json"
	"fmt"
	apsql "gateway/sql"
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
	CORSEnabled       bool            `json:"cors_enabled,omitempty" db:"cors_enabled"`
	CORSAllowOverride *string         `json:"cors_allow_override,omitempty" db:"cors_allow_override"`
	Routes            json.RawMessage `json:"routes,omitempty"`

	Components []*ProxyEndpointComponent `json:"components,omitempty"`
}

// Validate validates the model.
func (e *ProxyEndpoint) Validate() Errors {
	errors := make(Errors)
	if e.Name == "" {
		errors.add("name", "must not be blank")
	}
	routes, err := e.GetRoutes()
	if err != nil {
		errors.add("routes", "are invalid")
	} else {
		for i, r := range routes {
			rErrors := r.Validate()
			if !rErrors.Empty() {
				errors.add("routes", fmt.Sprintf("%d is invalid: %v", i, rErrors))
			}
		}
	}
	return errors
}

// ValidateFromDatabaseError translates possible database constraint errors
// into validation errors.
func (e *ProxyEndpoint) ValidateFromDatabaseError(err error) Errors {
	errors := make(Errors)
	return errors
}

// AllProxyEndpointsForAPIIDAndAccountID returns all proxyEndpoints on the Account's API in default order.
func AllProxyEndpointsForAPIIDAndAccountID(db *apsql.DB, apiID, accountID int64) ([]*ProxyEndpoint, error) {
	proxyEndpoints := []*ProxyEndpoint{}
	err := db.Select(&proxyEndpoints,
		`SELECT
			proxy_endpoints.id as id,
			proxy_endpoints.name as name,
			proxy_endpoints.description as description,
			proxy_endpoints.endpoint_group_id as endpoint_group_id,
			proxy_endpoints.environment_id as environment_id,
			proxy_endpoints.active as active
			FROM proxy_endpoints, apis
		WHERE proxy_endpoints.api_id = ?
			AND proxy_endpoints.api_id = apis.id
			AND apis.account_id = ?
		ORDER BY
			proxy_endpoints.name ASC,
			proxy_endpoints.id ASC;`,
		apiID, accountID)
	return proxyEndpoints, err
}

// FindProxyEndpointForAPIIDAndAccountID returns the proxyEndpoint with the id, api id, and account_id specified.
func FindProxyEndpointForAPIIDAndAccountID(db *apsql.DB, id, apiID, accountID int64) (*ProxyEndpoint, error) {
	proxyEndpoint := ProxyEndpoint{}
	err := db.Get(&proxyEndpoint,
		`SELECT
			proxy_endpoints.id as id,
			proxy_endpoints.name as name,
			proxy_endpoints.description as description,
			proxy_endpoints.endpoint_group_id as endpoint_group_id,
			proxy_endpoints.environment_id as environment_id,
			proxy_endpoints.active as active,
			proxy_endpoints.cors_enabled as cors_enabled,
			proxy_endpoints.cors_allow_override as cors_allow_override,
			proxy_endpoints.routes as routes
		FROM proxy_endpoints, apis
		WHERE proxy_endpoints.id = ?
			AND proxy_endpoints.api_id = ?
			AND proxy_endpoints.api_id = apis.id
			AND apis.account_id = ?;`,
		id, apiID, accountID)

	if err != nil {
		return nil, err
	}

	proxyEndpoint.Components, err = AllProxyEndpointComponentsForEndpointID(db, id)
	return &proxyEndpoint, err
}

// DeleteProxyEndpointForAPIIDAndAccountID deletes the proxyEndpoint with the id, api_id and account_id specified.
func DeleteProxyEndpointForAPIIDAndAccountID(tx *apsql.Tx, id, apiID, accountID int64) error {
	return tx.DeleteOne(
		`DELETE FROM proxy_endpoints
		WHERE proxy_endpoints.id = ?
			AND proxy_endpoints.api_id IN
					(SELECT id FROM apis WHERE id = ? AND account_id = ?);`,
		id, apiID, accountID)
}

// Insert inserts the proxyEndpoint into the database as a new row.
func (e *ProxyEndpoint) Insert(tx *apsql.Tx) error {
	routes, err := e.Routes.MarshalJSON()
	if err != nil {
		return err
	}
	e.ID, err = tx.InsertOne(
		`INSERT INTO proxy_endpoints
			(api_id, name, description, endpoint_group_id, environment_id,
			 active, cors_enabled, cors_allow_override, routes)
		VALUES ((SELECT id FROM apis WHERE id = ? AND account_id = ?),
			?, ?,(SELECT id FROM endpoint_groups WHERE id = ? AND api_id = ?),
			(SELECT id FROM environments WHERE id = ? AND api_id = ?),
			?, ?, ?, ?)`,
		e.APIID, e.AccountID, e.Name, e.Description, e.EndpointGroupID, e.APIID,
		e.EnvironmentID, e.APIID, e.Active, e.CORSEnabled, e.CORSAllowOverride, string(routes))
	if err != nil {
		return err
	}

	for position, component := range e.Components {
		err = component.Insert(tx, e.ID, e.APIID, position)
		if err != nil {
			return err
		}
	}

	return nil
}

// Update updates the proxyEndpoint in the database.
func (e *ProxyEndpoint) Update(tx *apsql.Tx) error {
	routes, err := e.Routes.MarshalJSON()
	if err != nil {
		return err
	}
	err = tx.UpdateOne(
		`UPDATE proxy_endpoints
		SET
			name = ?,
			description = ?,
			endpoint_group_id =
				(SELECT id FROM endpoint_groups WHERE id = ? AND api_id = ?),
			environment_id =
				(SELECT id FROM environments WHERE id = ? AND api_id = ?),
			active = ?,
			cors_enabled = ?,
			cors_allow_override = ?,
			routes = ?
		WHERE proxy_endpoints.id = ?
			AND proxy_endpoints.api_id IN
					(SELECT id FROM apis WHERE id = ? AND account_id = ?);`,
		e.Name, e.Description,
		e.EndpointGroupID, e.APIID,
		e.EnvironmentID, e.APIID,
		e.Active, e.CORSEnabled, e.CORSAllowOverride,
		string(routes),
		e.ID, e.APIID, e.AccountID)
	if err != nil {
		return err
	}

	var validComponentIDs []int64
	for position, component := range e.Components {
		if component.ID == 0 {
			err = component.Insert(tx, e.ID, e.APIID, position)
			if err != nil {
				return err
			}
		} else {
			err = component.Update(tx, e.ID, e.APIID, position)
			if err != nil {
				return err
			}
		}
		validComponentIDs = append(validComponentIDs, component.ID)
	}
	err = DeleteProxyEndpointComponentsWithEndpointIDAndNotInList(tx, e.ID, validComponentIDs)
	if err != nil {
		return err
	}

	return nil
}
