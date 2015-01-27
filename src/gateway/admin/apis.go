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

//go:generate ./serialize.rb API

// APIsController manages APIs.
type APIsController struct{}

// List lists the apis.
func (c *APIsController) List(w http.ResponseWriter, r *http.Request,
	db *apsql.DB) aphttp.Error {
	apis, err := model.AllAPIsForAccountID(db, accountIDFromSession(r))
	if err != nil {
		log.Printf("%s Error listing apis: %v", config.System, err)
		return aphttp.DefaultServerError()
	}

	return c.serializeCollection(apis, w)
}

// Create creates the api.
func (c *APIsController) Create(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {

	return c.insertOrUpdate(w, r, tx, true)
}

// Show shows the api.
func (c *APIsController) Show(w http.ResponseWriter, r *http.Request,
	db *apsql.DB) aphttp.Error {

	id := instanceID(r)
	api, err := model.FindAPIForAccountID(db, id, accountIDFromSession(r))
	if err != nil {
		return aphttp.NewError(fmt.Errorf("No api with id %d in account", id), 404)
	}

	return c.serializeInstance(api, w)
}

// Update updates the api.
func (c *APIsController) Update(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {

	return c.insertOrUpdate(w, r, tx, false)
}

// Delete deletes the api.
func (c *APIsController) Delete(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {

	err := model.DeleteAPIForAccountID(tx, instanceID(r), accountIDFromSession(r))
	if err != nil {
		log.Printf("%s Error deleting api: %v", config.System, err)
		return aphttp.DefaultServerError()
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (c *APIsController) insertOrUpdate(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx, isInsert bool) aphttp.Error {

	api, httpErr := c.deserializeInstance(r)
	if httpErr != nil {
		return httpErr
	}
	api.AccountID = accountIDFromSession(r)

	var method func(*apsql.Tx) error
	var desc string
	if isInsert {
		method = api.Insert
		desc = "inserting"
	} else {
		api.ID = instanceID(r)
		method = api.Update
		desc = "updating"
	}

	validationErrors := api.Validate()
	if !validationErrors.Empty() {
		return serialize(wrappedErrors{validationErrors}, w)
	}

	if err := method(tx); err != nil {
		log.Printf("%s Error %s api: %v", config.System, desc, err)
		return aphttp.DefaultServerError()
	}

	return c.serializeInstance(api, w)
}
