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

//go:generate ./serialize.rb User c.sanitize sanitizedUser

// UsersController manages users.
type UsersController struct {
	accountID func(r *http.Request) int64
}

// List lists the users.
func (c *UsersController) List(w http.ResponseWriter, r *http.Request,
	db *apsql.DB) aphttp.Error {

	users, err := model.AllUsersForAccountID(db, c.accountID(r))
	if err != nil {
		log.Printf("%s Error listing users: %v", config.System, err)
		return aphttp.DefaultServerError()
	}

	return c.serializeCollection(users, w)
}

// Create creates the user.
func (c *UsersController) Create(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {

	return c.insertOrUpdate(w, r, tx, true)
}

// Show shows the user.
func (c *UsersController) Show(w http.ResponseWriter, r *http.Request,
	db *apsql.DB) aphttp.Error {

	id := instanceID(r)
	user, err := model.FindUserForAccountID(db, id, c.accountID(r))
	if err != nil {
		return aphttp.NewError(fmt.Errorf("No user with id %d in account", id), 404)
	}

	return c.serializeInstance(user, w)
}

// Update updates the user.
func (c *UsersController) Update(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {

	return c.insertOrUpdate(w, r, tx, false)
}

// Delete deletes the user.
func (c *UsersController) Delete(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx) aphttp.Error {

	err := model.DeleteUserForAccountID(tx, instanceID(r), c.accountID(r))
	if err != nil {
		log.Printf("%s Error deleting user: %v", config.System, err)
		return aphttp.DefaultServerError()
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (c *UsersController) insertOrUpdate(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx, isInsert bool) aphttp.Error {

	user, httpErr := c.deserializeInstance(r)
	if httpErr != nil {
		return httpErr
	}

	user.AccountID = c.accountID(r)
	var method func(*apsql.Tx) error
	var desc string
	if isInsert {
		method = user.Insert
		desc = "inserting"
	} else {
		user.ID = instanceID(r)
		method = user.Update
		desc = "updating"
	}

	validationErrors := user.Validate()
	if !validationErrors.Empty() {
		return serialize(wrappedErrors{validationErrors}, w)
	}

	if err := method(tx); err != nil {
		log.Printf("%s Error %s user: %v", config.System, desc, err)
		return aphttp.DefaultServerError()
	}

	return c.serializeInstance(user, w)
}

type sanitizedUser struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (c *UsersController) sanitize(user *model.User) *sanitizedUser {
	return &sanitizedUser{user.ID, user.Name, user.Email}
}
