package model

import (
	"errors"
	"fmt"
	"strings"

	aperrors "gateway/errors"
	apsql "gateway/sql"

	"github.com/jmoiron/sqlx/types"
)

const (
	ProxyEndpointTypeHTTP = "http"
	ProxyEndpointTypeJob  = "job"
)

var ProxyEndpointNotifyTypeMap = map[string]string{
	ProxyEndpointTypeHTTP: "proxy_endpoints",
	ProxyEndpointTypeJob:  "jobs",
}

// ProxyEndpoint holds the data to power the proxy for a given API endpoint.
type ProxyEndpoint struct {
	AccountID       int64  `json:"-"`
	UserID          int64  `json:"-"`
	APIID           int64  `json:"api_id,omitempty" db:"api_id" path:"apiID"`
	EndpointGroupID *int64 `json:"endpoint_group_id,omitempty" db:"endpoint_group_id"`
	EnvironmentID   int64  `json:"environment_id,omitempty" db:"environment_id"`

	ID          int64          `json:"id,omitempty" path:"id"`
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Active      bool           `json:"active"`
	CORSEnabled bool           `json:"cors_enabled" db:"cors_enabled"`
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

type Job struct {
	ProxyEndpoint
}

func (e *ProxyEndpoint) GetType() string {
	return e.Type
}

func (e *ProxyEndpoint) SetType(t string) {
	e.Type = t
}

// Validate validates the model.
func (e *ProxyEndpoint) Validate(isInsert bool) aperrors.Errors {
	errors := make(aperrors.Errors)
	if e.Name == "" || strings.TrimSpace(e.Name) == "" {
		errors.Add("name", "must not be blank")
	}
	var isComponentInsert bool
	for i, c := range e.Components {
		isComponentInsert = isInsert || c.ID == 0
		cErrors := c.Validate(isComponentInsert)
		if !cErrors.Empty() {
			errors.Add("components", fmt.Sprintf("%d is invalid: %v", i, cErrors))
		}
	}
	if e.Type == ProxyEndpointTypeHTTP {
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
		for i, t := range e.Tests {
			tErrors := t.Validate()
			if !tErrors.Empty() {
				errors.Add("tests", fmt.Sprintf("%d is invalid: %v", i, tErrors))
			}
		}
	}
	return errors
}

// ValidateFromDatabaseError translates possible database constraint errors
// into validation errors.
func (e *ProxyEndpoint) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	if apsql.IsUniqueConstraint(err, "proxy_endpoints", "api_id", "type", "name") {
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

// All returns all proxyEndpoints on the Account's API in default order.
func (e *ProxyEndpoint) All(db *apsql.DB) ([]*ProxyEndpoint, error) {
	proxyEndpoints := []*ProxyEndpoint{}
	if e.APIID > 0 {
		if e.Type == "" {
			err := db.Select(&proxyEndpoints, db.SQL("proxy_endpoints/all"), e.APIID, e.AccountID)
			if err != nil {
				return nil, err
			}
		} else {
			err := db.Select(&proxyEndpoints, db.SQL("proxy_endpoints/all_for_type"), e.Type, e.APIID, e.AccountID)
			if err != nil {
				return nil, err
			}
		}
	} else {
		err := db.Select(&proxyEndpoints, db.SQL("proxy_endpoints/all_for_type_and_account"), e.Type, e.AccountID)
		if err != nil {
			return nil, err
		}
	}
	for _, endpoint := range proxyEndpoints {
		endpoint.AccountID = e.AccountID
		endpoint.UserID = e.UserID
		endpoint.APIID = e.APIID
	}
	return proxyEndpoints, nil
}

// AllProxyEndpointsForRouting returns all proxyEndpoints in an
// unspecified order, with enough data for routing.
func AllProxyEndpointsForRouting(db *apsql.DB) ([]*ProxyEndpoint, error) {
	proxyEndpoints := []*ProxyEndpoint{}
	err := db.Select(&proxyEndpoints, db.SQL("proxy_endpoints/all_routing"))
	return proxyEndpoints, err
}

// AllProxyEndpointsForRoutingForAPIID returns all proxyEndpoints for an
// api in an unspecified order, with enough data for routing.
func AllProxyEndpointsForRoutingForAPIID(db *apsql.DB, apiID int64) ([]*ProxyEndpoint, error) {
	proxyEndpoints := []*ProxyEndpoint{}
	err := db.Select(&proxyEndpoints, db.SQL("proxy_endpoints/all_routing_api"), apiID)
	return proxyEndpoints, err
}

// Find returns the proxyEndpoint with the id, api id, and account_id specified.
func (e *ProxyEndpoint) Find(db *apsql.DB) (*ProxyEndpoint, error) {
	proxyEndpoint := ProxyEndpoint{
		AccountID: e.AccountID,
		UserID:    e.UserID,
		APIID:     e.APIID,
	}
	if e.ID > 0 {
		err := db.Get(&proxyEndpoint, db.SQL("proxy_endpoints/find"), e.ID, e.Type, e.APIID, e.AccountID)
		if err != nil {
			return nil, err
		}
	} else if e.Name != "" {
		err := db.Get(&proxyEndpoint, db.SQL("proxy_endpoints/find_name"), e.Name, e.Type, e.APIID, e.AccountID)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("Not enough information for proxy endpoint find.")
	}

	var err error
	proxyEndpoint.Components, err = AllProxyEndpointComponentsForEnvironmentOnAPI(
		db, e.APIID, proxyEndpoint.EnvironmentID, e.ID,
	)
	if err != nil {
		return nil, err
	}

	if e.Type == ProxyEndpointTypeHTTP {
		proxyEndpoint.Tests, err = AllProxyEndpointTestsForEndpointID(db, e.ID)
		if err != nil {
			return nil, err
		}
	}

	return &proxyEndpoint, nil
}

// FindProxyEndpointForProxy returns the proxyEndpoint with the id specified;
// it includes all relationships.
func FindProxyEndpointForProxy(db *apsql.DB, id int64, endpointType string) (*ProxyEndpoint, error) {
	proxyEndpoint := ProxyEndpoint{}
	err := db.Get(&proxyEndpoint, db.SQL("proxy_endpoints/find_id"), id, endpointType)
	if err != nil {
		return nil, aperrors.NewWrapped("Finding proxy endpoint", err)
	}

	proxyEndpoint.Components, err = AllProxyEndpointComponentsForEnvironmentOnAPI(
		db, proxyEndpoint.APIID, proxyEndpoint.EnvironmentID, id,
	)
	if err != nil {
		return nil, aperrors.NewWrapped("Fetching components", err)
	}

	proxyEndpoint.Environment, err = FindEnvironmentForProxy(db, proxyEndpoint.EnvironmentID)
	if err != nil {
		return nil, aperrors.NewWrapped("Fetching environment", err)
	}

	proxyEndpoint.API, err = FindAPIForProxy(db, proxyEndpoint.APIID)
	if err != nil {
		return nil, aperrors.NewWrapped("Fetching API", err)
	}

	if endpointType == ProxyEndpointTypeHTTP {
		proxyEndpoint.Tests, err = AllProxyEndpointTestsForEndpointID(db, id)
		if err != nil {
			return nil, aperrors.NewWrapped("Fetching tests", err)
		}

		schemas, err := FindProxyEndpointSchemasForProxy(db, proxyEndpoint.ID, proxyEndpoint.APIID)
		if err != nil {
			return nil, aperrors.NewWrapped("Fetching Schema", err)
		}
		if len(schemas) > 0 {
			proxyEndpoint.Schema = schemas[0]
		}
	}

	return &proxyEndpoint, nil
}

// Delete deletes the proxyEndpoint with the id, api_id and account_id specified.
func (e *ProxyEndpoint) Delete(tx *apsql.Tx) error {
	err := tx.DeleteOne(tx.SQL("proxy_endpoints/delete"), e.ID, e.Type, e.APIID, e.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify(ProxyEndpointNotifyTypeMap[e.Type], e.AccountID, e.UserID, e.APIID, 0, e.ID, apsql.Delete)
}

// Insert inserts the proxyEndpoint into the database as a new row.
func (e *ProxyEndpoint) Insert(tx *apsql.Tx) error {
	routes, err := marshaledForStorage(e.Routes)
	if err != nil {
		return err
	}

	e.ID, err = tx.InsertOne(tx.SQL("proxy_endpoints/insert"),
		e.APIID, e.AccountID, e.Type, e.Name, e.Description, e.EndpointGroupID, e.APIID,
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

	if e.Type == ProxyEndpointTypeHTTP {
		for _, test := range e.Tests {
			err = test.Insert(tx, e.ID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Notify(ProxyEndpointNotifyTypeMap[e.Type], e.AccountID, e.UserID, e.APIID, 0, e.ID, apsql.Insert)
}

// Update updates the proxyEndpoint in the database.
func (e *ProxyEndpoint) Update(tx *apsql.Tx) error {
	routes, err := marshaledForStorage(e.Routes)
	if err != nil {
		return err
	}

	err = tx.UpdateOne(tx.SQL("proxy_endpoints/update"),
		e.Name, e.Description,
		e.EndpointGroupID, e.APIID,
		e.EnvironmentID, e.APIID,
		e.Active, e.CORSEnabled,
		routes,
		e.ID, e.Type, e.APIID, e.AccountID)
	if err != nil {
		return err
	}

	var validReferenceIDs []int64
	for position, component := range e.Components {
		if component.ID == 0 {
			err = component.Insert(tx, e.ID, e.APIID, position)
		} else {
			err = component.Update(tx, e.ID, e.APIID, position)
		}
		if err != nil {
			return err
		}
		validReferenceIDs = append(
			validReferenceIDs,
			*component.ProxyEndpointComponentReferenceID,
		)
	}
	err = DeleteProxyEndpointComponentsWithEndpointIDAndRefNotInSlice(
		tx, e.ID, validReferenceIDs,
	)
	if err != nil {
		return err
	}

	if e.Type == ProxyEndpointTypeHTTP {
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
	}

	return tx.Notify(ProxyEndpointNotifyTypeMap[e.Type], e.AccountID, e.UserID, e.APIID, 0, e.ID, apsql.Update)
}
