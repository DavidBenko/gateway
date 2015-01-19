package admin

import (
	"fmt"
	"gateway/config"
	aphttp "gateway/http"
	"gateway/model"
	apsql "gateway/sql"
	"log"
	"net/http"
)

//go:generate ./serialize.rb ProxyEndpoint

// ProxyEndpointsController manages ProxyEndpoints.
type ProxyEndpointsController struct{}

// List lists the ProxyEndpoints.
func (c *ProxyEndpointsController) List(w http.ResponseWriter, r *http.Request,
	db *apsql.DB) aphttp.Error {
	proxyEndpoints, err := model.AllProxyEndpointsForAPIIDAndAccountID(db,
		apiIDFromPath(r), accountIDFromSession(r))
	if err != nil {
		log.Printf("%s Error listing proxy endpoints: %v", config.System, err)
		return aphttp.DefaultServerError()
	}

	return c.serializeCollection(proxyEndpoints, w)
}

// Create creates the proxyEndpoint.
func (c *ProxyEndpointsController) Create(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {

	return c.insertOrUpdate(w, r, tx, true)
}

// Show shows the proxyEndpoint.
func (c *ProxyEndpointsController) Show(w http.ResponseWriter, r *http.Request,
	db *apsql.DB) aphttp.Error {
	id := instanceID(r)
	proxyEndpoint, err := model.FindProxyEndpointForAPIIDAndAccountID(db, id,
		apiIDFromPath(r), accountIDFromSession(r))
	if err != nil {
		return aphttp.NewError(fmt.Errorf("No proxy endpoint with id %d in api", id), 404)
	}

	return c.serializeInstance(proxyEndpoint, w)
}

// Update updates the proxyEndpoint.
func (c *ProxyEndpointsController) Update(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {

	return c.insertOrUpdate(w, r, tx, false)
}

// Delete deletes the proxyEndpoint.
func (c *ProxyEndpointsController) Delete(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {

	err := model.DeleteProxyEndpointForAPIIDAndAccountID(tx, instanceID(r),
		apiIDFromPath(r), accountIDFromSession(r))
	if err != nil {
		log.Printf("%s Error deleting proxy endpoint: %v", config.System, err)
		return aphttp.DefaultServerError()
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (c *ProxyEndpointsController) insertOrUpdate(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx, isInsert bool) aphttp.Error {

	proxyEndpoint, err := c.deserializeInstance(r)
	if err != nil {
		log.Printf("%s Error reading proxy endpoint: %v", config.System, err)
		return aphttp.DefaultServerError()
	}
	proxyEndpoint.APIID = apiIDFromPath(r)
	proxyEndpoint.AccountID = accountIDFromSession(r)

	var method func(*apsql.Tx) error
	var desc string
	if isInsert {
		method = proxyEndpoint.Insert
		desc = "inserting"
	} else {
		proxyEndpoint.ID = instanceID(r)
		method = proxyEndpoint.Update
		desc = "updating"
	}

	validationErrors := proxyEndpoint.Validate()
	if !validationErrors.Empty() {
		return serialize(wrappedErrors{validationErrors}, w)
	}

	if err = method(tx); err != nil {
		log.Printf("%s Error %s proxy endpoint: %v", config.System, desc, err)
		return aphttp.DefaultServerError()
	}

	return c.serializeInstance(proxyEndpoint, w)
}
