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

//go:generate ./serialize.rb RemoteEndpoint

// RemoteEndpointsController manages RemoteEndpoints.
type RemoteEndpointsController struct{}

// List lists the RemoteEndpoints.
func (c *RemoteEndpointsController) List(w http.ResponseWriter, r *http.Request,
	db *apsql.DB) aphttp.Error {
	remoteEndpoints, err := model.AllRemoteEndpointsForAPIIDAndAccountID(db,
		apiIDFromPath(r), accountIDFromSession(r))
	if err != nil {
		log.Printf("%s Error listing remote endpoints: %v", config.System, err)
		return aphttp.DefaultServerError()
	}

	return c.serializeCollection(remoteEndpoints, w)
}

// Create creates the remoteEndpoint.
func (c *RemoteEndpointsController) Create(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {

	return c.insertOrUpdate(w, r, tx, true)
}

// Show shows the remoteEndpoint.
func (c *RemoteEndpointsController) Show(w http.ResponseWriter, r *http.Request,
	db *apsql.DB) aphttp.Error {
	id := instanceID(r)
	remoteEndpoint, err := model.FindRemoteEndpointForAPIIDAndAccountID(db, id,
		apiIDFromPath(r), accountIDFromSession(r))
	if err != nil {
		return aphttp.NewError(fmt.Errorf("No remote endpoint with id %d in api", id), 404)
	}

	return c.serializeInstance(remoteEndpoint, w)
}

// Update updates the remoteEndpoint.
func (c *RemoteEndpointsController) Update(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {

	return c.insertOrUpdate(w, r, tx, false)
}

// Delete deletes the remoteEndpoint.
func (c *RemoteEndpointsController) Delete(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {

	err := model.DeleteRemoteEndpointForAPIIDAndAccountID(tx, instanceID(r),
		apiIDFromPath(r), accountIDFromSession(r))
	if err != nil {
		log.Printf("%s Error deleting remote endpoint: %v", config.System, err)
		return aphttp.DefaultServerError()
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (c *RemoteEndpointsController) insertOrUpdate(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx, isInsert bool) aphttp.Error {

	remoteEndpoint, err := c.deserializeInstance(r)
	if err != nil {
		log.Printf("%s Error reading remote endpoint: %v", config.System, err)
		return aphttp.DefaultServerError()
	}
	remoteEndpoint.APIID = apiIDFromPath(r)
	remoteEndpoint.AccountID = accountIDFromSession(r)

	var method func(*apsql.Tx) error
	var desc string
	if isInsert {
		method = remoteEndpoint.Insert
		desc = "inserting"
	} else {
		remoteEndpoint.ID = instanceID(r)
		method = remoteEndpoint.Update
		desc = "updating"
	}

	validationErrors := remoteEndpoint.Validate()
	if !validationErrors.Empty() {
		return serialize(wrappedErrors{validationErrors}, w)
	}

	if err = method(tx); err != nil {
		log.Printf("%s Error %s remote endpoint: %v", config.System, desc, err)
		return aphttp.DefaultServerError()
	}

	return c.serializeInstance(remoteEndpoint, w)
}
