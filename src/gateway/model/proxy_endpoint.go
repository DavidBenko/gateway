package model

import (
	"errors"
	"fmt"
	"strings"

	aperrors "gateway/errors"
	"gateway/license"
	apsql "gateway/sql"

	"github.com/jmoiron/sqlx/types"
)

// ProxyEndpoint holds the data to power the proxy for a given API endpoint.
type ProxyEndpoint struct {
	AccountID       int64  `json:"-"`
	UserID          int64  `json:"-"`
	APIID           int64  `json:"api_id,omitempty" db:"api_id"`
	EndpointGroupID *int64 `json:"endpoint_group_id,omitempty" db:"endpoint_group_id"`
	EnvironmentID   int64  `json:"environment_id,omitempty" db:"environment_id"`

	ID          int64          `json:"id,omitempty"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Active      bool           `json:"active"`
	CORSEnabled bool           `json:"cors_enabled,omitempty" db:"cors_enabled"`
	Routes      types.JsonText `json:"routes,omitempty"`

	Components []*ProxyEndpointComponent `json:"components,omitempty"`
	Tests      []*ProxyEndpointTest      `json:"tests,omitempty"`

	// Export Indices
	ExportEndpointGroupIndex int `json:"endpoint_group_index,omitempty"`
	ExportEnvironmentIndex   int `json:"environment_index,omitempty"`

	// Proxy Data Cache
	Environment *Environment         `json:"-"`
	API         *API                 `json:"-"`
	Schema      *ProxyEndpointSchema `json:"-"`
}

// Validate validates the model.
func (e *ProxyEndpoint) Validate(isInsert bool) aperrors.Errors {
	errors := make(aperrors.Errors)
	if e.Name == "" || strings.TrimSpace(e.Name) == "" {
		errors.Add("name", "must not be blank")
	}
	routes, err := e.GetRoutes()
	if err != nil {
		errors.Add("routes", "are invalid")
	} else {
		for i, r := range routes {
			rErrors := r.Validate()
			if !rErrors.Empty() {
				errors.Add("routes", fmt.Sprintf("%d is invalid: %v", i, rErrors))
			}
		}
	}
	for i, c := range e.Components {
		cErrors := c.Validate()
		if !cErrors.Empty() {
			errors.Add("components", fmt.Sprintf("%d is invalid: %v", i, cErrors))
		}
	}
	for i, t := range e.Tests {
		tErrors := t.Validate()
		if !tErrors.Empty() {
			errors.Add("tests", fmt.Sprintf("%d is invalid: %v", i, tErrors))
		}
	}
	return errors
}

// ValidateFromDatabaseError translates possible database constraint errors
// into validation errors.
func (e *ProxyEndpoint) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	if apsql.IsUniqueConstraint(err, "proxy_endpoints", "api_id", "name") {
		errors.Add("name", "is already taken")
	}
	if apsql.IsNotNullConstraint(err, "proxy_endpoints", "environment_id") {
		errors.Add("environment_id", "must be a valid environment in this API")
	}
	if apsql.IsNotNullConstraint(err, "proxy_endpoint_calls", "remote_endpoint_id") {
		errors.Add("components", "all calls must reference a valid remote endpoint in this API")
	}
	if apsql.IsUniqueConstraint(err, "proxy_endpoint_tests", "endpoint_id", "name") {
		errors.Add("tests", "name is already taken")
	}
	return errors
}

// AllProxyEndpointsForAPIIDAndAccountID returns all proxyEndpoints on the Account's API in default order.
func AllProxyEndpointsForAPIIDAndAccountID(db *apsql.DB, apiID, accountID int64) ([]*ProxyEndpoint, error) {
	proxyEndpoints := []*ProxyEndpoint{}
	err := db.Select(&proxyEndpoints, db.SQL("proxy_endpoints/all"), apiID, accountID)
	return proxyEndpoints, err
}

// AllActiveProxyEndpointsForRouting returns all proxyEndpoints in an
// unspecified order, with enough data for routing.
func AllProxyEndpointsForRouting(db *apsql.DB) ([]*ProxyEndpoint, error) {
	proxyEndpoints := []*ProxyEndpoint{}
	err := db.Select(&proxyEndpoints, db.SQL("proxy_endpoints/all_routing"))
	return proxyEndpoints, err
}

// AllActiveProxyEndpointsForRoutingForAPIID returns all proxyEndpoints for an
// api in an unspecified order, with enough data for routing.
func AllProxyEndpointsForRoutingForAPIID(db *apsql.DB, apiID int64) ([]*ProxyEndpoint, error) {
	proxyEndpoints := []*ProxyEndpoint{}
	err := db.Select(&proxyEndpoints, db.SQL("proxy_endpoints/all_routing_api"), apiID)
	return proxyEndpoints, err
}

// FindProxyEndpointForAPIIDAndAccountID returns the proxyEndpoint with the id, api id, and account_id specified.
func FindProxyEndpointForAPIIDAndAccountID(db *apsql.DB, id, apiID, accountID int64) (*ProxyEndpoint, error) {
	proxyEndpoint := ProxyEndpoint{}
	err := db.Get(&proxyEndpoint, db.SQL("proxy_endpoints/find"), id, apiID, accountID)
	if err != nil {
		return nil, err
	}

	proxyEndpoint.Components, err = AllProxyEndpointComponentsForEndpointID(db, id)
	if err != nil {
		return nil, err
	}

	proxyEndpoint.Tests, err = AllProxyEndpointTestsForEndpointID(db, id)
	return &proxyEndpoint, err
}

// FindProxyEndpointForProxy returns the proxyEndpoint with the id specified;
// it includes all relationships.
func FindProxyEndpointForProxy(db *apsql.DB, id int64) (*ProxyEndpoint, error) {
	proxyEndpoint := ProxyEndpoint{}
	err := db.Get(&proxyEndpoint, db.SQL("proxy_endpoints/find_id"), id)
	if err != nil {
		return nil, aperrors.NewWrapped("Finding proxy endpoint", err)
	}

	proxyEndpoint.Components, err = AllProxyEndpointComponentsForEndpointID(db, id)
	if err != nil {
		return nil, aperrors.NewWrapped("Fetching components", err)
	}

	proxyEndpoint.Tests, err = AllProxyEndpointTestsForEndpointID(db, id)
	if err != nil {
		return nil, aperrors.NewWrapped("Fetching tests", err)
	}

	var remoteEndpointIDs []int64
	callsByRemoteEndpointID := make(map[int64][]*ProxyEndpointCall)
	for _, component := range proxyEndpoint.Components {
		for _, call := range component.AllCalls() {
			remoteEndpointIDs = append(remoteEndpointIDs, call.RemoteEndpointID)
			callsByRemoteEndpointID[call.RemoteEndpointID] = append(callsByRemoteEndpointID[call.RemoteEndpointID], call)
		}
	}
	remoteEndpoints, err := AllRemoteEndpointsForIDsInEnvironment(db,
		remoteEndpointIDs, proxyEndpoint.EnvironmentID)
	if err != nil {
		return nil, aperrors.NewWrapped("Fetching remote endpoints", err)
	}
	for _, remoteEndpoint := range remoteEndpoints {
		for _, call := range callsByRemoteEndpointID[remoteEndpoint.ID] {
			call.RemoteEndpoint = remoteEndpoint
		}
	}

	proxyEndpoint.Environment, err = FindEnvironmentForProxy(db, proxyEndpoint.EnvironmentID)
	if err != nil {
		return nil, aperrors.NewWrapped("Fetching environment", err)
	}

	proxyEndpoint.API, err = FindAPIForProxy(db, proxyEndpoint.APIID)
	if err != nil {
		return nil, aperrors.NewWrapped("Fetching API", err)
	}

	schemas, err := FindProxyEndpointSchemasForProxy(db, proxyEndpoint.ID, proxyEndpoint.APIID)
	if err != nil {
		return nil, aperrors.NewWrapped("Fetching Schema", err)
	}
	if len(schemas) > 0 {
		proxyEndpoint.Schema = schemas[0]
	}

	return &proxyEndpoint, nil
}

// DeleteProxyEndpointForAPIIDAndAccountID deletes the proxyEndpoint with the id, api_id and account_id specified.
func DeleteProxyEndpointForAPIIDAndAccountID(tx *apsql.Tx, id, apiID, accountID, userID int64) error {
	err := tx.DeleteOne(tx.SQL("proxy_endpoints/delete"), id, apiID, accountID)
	if err != nil {
		return err
	}
	return tx.Notify("proxy_endpoints", accountID, userID, apiID, 0, id, apsql.Delete)
}

// Insert inserts the proxyEndpoint into the database as a new row.
func (e *ProxyEndpoint) Insert(tx *apsql.Tx) error {
	routes, err := marshaledForStorage(e.Routes)
	if err != nil {
		return err
	}

	if license.DeveloperVersion && e.Active {
		var count int
		tx.Get(&count, tx.SQL("proxy_endpoints/count_active"), e.APIID)
		if count >= license.DeveloperVersionProxyEndpoints {
			return errors.New(fmt.Sprintf("Developer version allows %v active proxy endpoint(s).", license.DeveloperVersionProxyEndpoints))
		}
	}

	e.ID, err = tx.InsertOne(tx.SQL("proxy_endpoints/insert"),
		e.APIID, e.AccountID, e.Name, e.Description, e.EndpointGroupID, e.APIID,
		e.EnvironmentID, e.APIID, e.Active, e.CORSEnabled, routes)
	if err != nil {
		return err
	}

	for position, component := range e.Components {
		err = component.Insert(tx, e.ID, e.APIID, position)
		if err != nil {
			return err
		}
	}

	for _, test := range e.Tests {
		err = test.Insert(tx, e.ID)
		if err != nil {
			return err
		}
	}

	return tx.Notify("proxy_endpoints", e.AccountID, e.UserID, e.APIID, 0, e.ID, apsql.Insert)
}

// Update updates the proxyEndpoint in the database.
func (e *ProxyEndpoint) Update(tx *apsql.Tx) error {
	routes, err := marshaledForStorage(e.Routes)
	if err != nil {
		return err
	}

	if license.DeveloperVersion {
		proxyEndpoint := ProxyEndpoint{}
		err := tx.Get(&proxyEndpoint, tx.SQL("proxy_endpoints/find_id"), e.ID)
		if err != nil {
			return aperrors.NewWrapped("Finding proxy endpoint", err)
		}

		if !proxyEndpoint.Active && e.Active {
			var count int
			tx.Get(&count, tx.SQL("proxy_endpoints/count_active"), e.APIID)
			if count >= license.DeveloperVersionProxyEndpoints {
				return errors.New(fmt.Sprintf("Developer version allows %v active proxy endpoint(s).", license.DeveloperVersionProxyEndpoints))
			}
		}
	}

	err = tx.UpdateOne(tx.SQL("proxy_endpoints/update"),
		e.Name, e.Description,
		e.EndpointGroupID, e.APIID,
		e.EnvironmentID, e.APIID,
		e.Active, e.CORSEnabled,
		routes,
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

	var validTestIDs []int64
	for _, test := range e.Tests {
		if test.ID == 0 {
			err = test.Insert(tx, e.ID)
			if err != nil {
				return err
			}
		} else {
			err = test.Update(tx, e.ID)
			if err != nil {
				return err
			}
		}
		validTestIDs = append(validTestIDs, test.ID)
	}
	err = DeleteProxyEndpointTestsWithEndpointIDAndNotInList(tx, e.ID, validTestIDs)
	if err != nil {
		return err
	}

	return tx.Notify("proxy_endpoints", e.AccountID, e.UserID, e.APIID, 0, e.ID, apsql.Update)
}
