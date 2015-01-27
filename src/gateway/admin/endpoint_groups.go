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

//go:generate ./serialize.rb EndpointGroup

// EndpointGroupsController manages EndpointGroups.
type EndpointGroupsController struct{}

// List lists the endpointGroups.
func (c *EndpointGroupsController) List(w http.ResponseWriter, r *http.Request,
	db *apsql.DB) aphttp.Error {
	endpointGroups, err := model.AllEndpointGroupsForAPIIDAndAccountID(db,
		apiIDFromPath(r), accountIDFromSession(r))
	if err != nil {
		log.Printf("%s Error listing endpoint groups: %v", config.System, err)
		return aphttp.DefaultServerError()
	}

	return c.serializeCollection(endpointGroups, w)
}

// Create creates the endpointGroup.
func (c *EndpointGroupsController) Create(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {

	return c.insertOrUpdate(w, r, tx, true)
}

// Show shows the endpointGroup.
func (c *EndpointGroupsController) Show(w http.ResponseWriter, r *http.Request,
	db *apsql.DB) aphttp.Error {
	id := instanceID(r)
	endpointGroup, err := model.FindEndpointGroupForAPIIDAndAccountID(db, id,
		apiIDFromPath(r), accountIDFromSession(r))
	if err != nil {
		return aphttp.NewError(fmt.Errorf("No endpoint group with id %d in api", id), 404)
	}

	return c.serializeInstance(endpointGroup, w)
}

// Update updates the endpointGroup.
func (c *EndpointGroupsController) Update(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {

	return c.insertOrUpdate(w, r, tx, false)
}

// Delete deletes the endpointGroup.
func (c *EndpointGroupsController) Delete(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {

	err := model.DeleteEndpointGroupForAPIIDAndAccountID(tx, instanceID(r),
		apiIDFromPath(r), accountIDFromSession(r))
	if err != nil {
		log.Printf("%s Error deleting endpoint group: %v", config.System, err)
		return aphttp.DefaultServerError()
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (c *EndpointGroupsController) insertOrUpdate(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx, isInsert bool) aphttp.Error {

	endpointGroup, httpErr := c.deserializeInstance(r)
	if httpErr != nil {
		return httpErr
	}
	endpointGroup.APIID = apiIDFromPath(r)
	endpointGroup.AccountID = accountIDFromSession(r)

	var method func(*apsql.Tx) error
	var desc string
	if isInsert {
		method = endpointGroup.Insert
		desc = "inserting"
	} else {
		endpointGroup.ID = instanceID(r)
		method = endpointGroup.Update
		desc = "updating"
	}

	validationErrors := endpointGroup.Validate()
	if !validationErrors.Empty() {
		return serialize(wrappedErrors{validationErrors}, w)
	}

	if err := method(tx); err != nil {
		log.Printf("%s Error %s endpoint group: %v", config.System, desc, err)
		return aphttp.DefaultServerError()
	}

	return c.serializeInstance(endpointGroup, w)
}
