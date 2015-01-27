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

//go:generate ./serialize.rb Environment

// EnvironmentsController manages Environments.
type EnvironmentsController struct{}

// List lists the environments.
func (c *EnvironmentsController) List(w http.ResponseWriter, r *http.Request,
	db *apsql.DB) aphttp.Error {
	environments, err := model.AllEnvironmentsForAPIIDAndAccountID(db,
		apiIDFromPath(r), accountIDFromSession(r))
	if err != nil {
		log.Printf("%s Error listing environments: %v", config.System, err)
		return aphttp.DefaultServerError()
	}

	return c.serializeCollection(environments, w)
}

// Create creates the environment.
func (c *EnvironmentsController) Create(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {

	return c.insertOrUpdate(w, r, tx, true)
}

// Show shows the environment.
func (c *EnvironmentsController) Show(w http.ResponseWriter, r *http.Request,
	db *apsql.DB) aphttp.Error {
	id := instanceID(r)
	environment, err := model.FindEnvironmentForAPIIDAndAccountID(db, id,
		apiIDFromPath(r), accountIDFromSession(r))
	if err != nil {
		return aphttp.NewError(fmt.Errorf("No environment with id %d in environment", id), 404)
	}

	return c.serializeInstance(environment, w)
}

// Update updates the environment.
func (c *EnvironmentsController) Update(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {

	return c.insertOrUpdate(w, r, tx, false)
}

// Delete deletes the environment.
func (c *EnvironmentsController) Delete(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {

	err := model.DeleteEnvironmentForAPIIDAndAccountID(tx, instanceID(r),
		apiIDFromPath(r), accountIDFromSession(r))
	if err != nil {
		log.Printf("%s Error deleting environment: %v", config.System, err)
		return aphttp.DefaultServerError()
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (c *EnvironmentsController) insertOrUpdate(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx, isInsert bool) aphttp.Error {

	environment, httpErr := c.deserializeInstance(r)
	if httpErr != nil {
		return httpErr
	}
	environment.APIID = apiIDFromPath(r)
	environment.AccountID = accountIDFromSession(r)

	var method func(*apsql.Tx) error
	var desc string
	if isInsert {
		method = environment.Insert
		desc = "inserting"
	} else {
		environment.ID = instanceID(r)
		method = environment.Update
		desc = "updating"
	}

	validationErrors := environment.Validate()
	if !validationErrors.Empty() {
		return serialize(wrappedErrors{validationErrors}, w)
	}

	if err := method(tx); err != nil {
		log.Printf("%s Error %s environment: %v", config.System, desc, err)
		return aphttp.DefaultServerError()
	}

	return c.serializeInstance(environment, w)
}
