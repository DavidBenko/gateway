package model

import (
	"errors"
	"fmt"
	aperrors "gateway/errors"
	"gateway/license"
	apsql "gateway/sql"

	"github.com/jmoiron/sqlx/types"
)

// ProxyEndpoint holds the data to power the proxy for a given API endpoint.
type ProxyEndpoint struct {
	AccountID       int64  `json:"-"`
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

	// Export Indices
	ExportEndpointGroupIndex int `json:"endpoint_group_index,omitempty"`
	ExportEnvironmentIndex   int `json:"environment_index,omitempty"`

	// Proxy Data Cache
	Environment *Environment `json:"-"`
	API         *API         `json:"-"`
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
	for i, c := range e.Components {
		cErrors := c.Validate()
		if !cErrors.Empty() {
			errors.add("components", fmt.Sprintf("%d is invalid: %v", i, cErrors))
		}
	}
	return errors
}

// ValidateFromDatabaseError translates possible database constraint errors
// into validation errors.
func (e *ProxyEndpoint) ValidateFromDatabaseError(err error) Errors {
	errors := make(Errors)
	if apsql.IsUniqueConstraint(err, "proxy_endpoints", "api_id", "name") {
		errors.add("name", "is already taken")
	}
	if apsql.IsNotNullConstraint(err, "proxy_endpoints", "environment_id") {
		errors.add("environment_id", "must be a valid environment in this API")
	}
	if apsql.IsNotNullConstraint(err, "proxy_endpoint_calls", "remote_endpoint_id") {
		errors.add("components", "all calls must reference a valid remote endpoint in this API")
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
func AllActiveProxyEndpointsForRouting(db *apsql.DB) ([]*ProxyEndpoint, error) {
	proxyEndpoints := []*ProxyEndpoint{}
	err := db.Select(&proxyEndpoints, db.SQL("proxy_endpoints/all_routing"), true)
	return proxyEndpoints, err
}

// AllActiveProxyEndpointsForRoutingForAPIID returns all proxyEndpoints for an
// api in an unspecified order, with enough data for routing.
func AllActiveProxyEndpointsForRoutingForAPIID(db *apsql.DB, apiID int64) ([]*ProxyEndpoint, error) {
	proxyEndpoints := []*ProxyEndpoint{}
	err := db.Select(&proxyEndpoints, db.SQL("proxy_endpoints/all_routing_api"), true, apiID)
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

	return &proxyEndpoint, nil
}

// DeleteProxyEndpointForAPIIDAndAccountID deletes the proxyEndpoint with the id, api_id and account_id specified.
func DeleteProxyEndpointForAPIIDAndAccountID(tx *apsql.Tx, id, apiID, accountID int64) error {
	err := tx.DeleteOne(tx.SQL("proxy_endpoints/delete"), id, apiID, accountID)
	if err != nil {
		return err
	}
	return tx.Notify("proxy_endpoints", apiID, apsql.Delete)
}

// Insert inserts the proxyEndpoint into the database as a new row.
func (e *ProxyEndpoint) Insert(tx *apsql.Tx) error {
	routes, err := marshaledForStorage(e.Routes)
	if err != nil {
		return err
	}

	if license.DeveloperVersion {
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

	return tx.Notify("proxy_endpoints", e.APIID, apsql.Insert)
}

// Update updates the proxyEndpoint in the database.
func (e *ProxyEndpoint) Update(tx *apsql.Tx) error {
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

	return tx.Notify("proxy_endpoints", e.APIID, apsql.Update)
}
