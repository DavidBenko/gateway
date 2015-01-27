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

//go:generate ./serialize.rb Host

// HostsController manages Hosts.
type HostsController struct{}

// List lists the hosts.
func (c *HostsController) List(w http.ResponseWriter, r *http.Request,
	db *apsql.DB) aphttp.Error {
	hosts, err := model.AllHostsForAPIIDAndAccountID(db,
		apiIDFromPath(r), accountIDFromSession(r))
	if err != nil {
		log.Printf("%s Error listing hosts: %v", config.System, err)
		return aphttp.DefaultServerError()
	}

	return c.serializeCollection(hosts, w)
}

// Create creates the host.
func (c *HostsController) Create(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {

	return c.insertOrUpdate(w, r, tx, true)
}

// Show shows the host.
func (c *HostsController) Show(w http.ResponseWriter, r *http.Request,
	db *apsql.DB) aphttp.Error {
	id := instanceID(r)
	host, err := model.FindHostForAPIIDAndAccountID(db, id,
		apiIDFromPath(r), accountIDFromSession(r))
	if err != nil {
		return aphttp.NewError(fmt.Errorf("No host with id %d in host", id), 404)
	}

	return c.serializeInstance(host, w)
}

// Update updates the host.
func (c *HostsController) Update(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {

	return c.insertOrUpdate(w, r, tx, false)
}

// Delete deletes the host.
func (c *HostsController) Delete(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {

	err := model.DeleteHostForAPIIDAndAccountID(tx, instanceID(r),
		apiIDFromPath(r), accountIDFromSession(r))
	if err != nil {
		log.Printf("%s Error deleting host: %v", config.System, err)
		return aphttp.DefaultServerError()
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (c *HostsController) insertOrUpdate(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx, isInsert bool) aphttp.Error {

	host, httpErr := c.deserializeInstance(r)
	if httpErr != nil {
		return httpErr
	}
	host.APIID = apiIDFromPath(r)
	host.AccountID = accountIDFromSession(r)

	var method func(*apsql.Tx) error
	var desc string
	if isInsert {
		method = host.Insert
		desc = "inserting"
	} else {
		host.ID = instanceID(r)
		method = host.Update
		desc = "updating"
	}

	validationErrors := host.Validate()
	if !validationErrors.Empty() {
		return serialize(wrappedErrors{validationErrors}, w)
	}

	if err := method(tx); err != nil {
		log.Printf("%s Error %s host: %v", config.System, desc, err)
		return aphttp.DefaultServerError()
	}

	return c.serializeInstance(host, w)
}
