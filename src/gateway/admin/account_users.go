package admin

import (
	"fmt"
	"log"
	"net/http"

	"gateway/config"
	aphttp "gateway/http"
	"gateway/model"
	sql "gateway/sql"

	"github.com/gorilla/handlers"
	"github.com/jmoiron/sqlx"
)

// RouteAccountUsers routes all the endpoints for account based user management
// by site admins.
func RouteAccountUsers(router aphttp.Router, db *sql.DB) {
	router.Handle("/accounts/{accountID}/users",
		handlers.MethodHandler{
			"GET":  aphttp.ErrorCatchingHandler(ListAccountUsersHandler(db)),
			"POST": aphttp.ErrorCatchingHandler(CreateAccountUserHandler(db)),
		})
	router.Handle("/accounts/{accountID}/users/{id}",
		handlers.HTTPMethodOverrideHandler(handlers.MethodHandler{
			"GET":    aphttp.ErrorCatchingHandler(ShowAccountUserHandler(db)),
			"PUT":    aphttp.ErrorCatchingHandler(UpdateAccountUserHandler(db)),
			"DELETE": aphttp.ErrorCatchingHandler(DeleteAccountUserHandler(db)),
		}))
}

// ListAccountUsersHandler returns a handler that lists the users.
func ListAccountUsersHandler(db *sql.DB) aphttp.ErrorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		users, err := model.AllUsersForAccountID(db, accountID(r))
		if err != nil {
			log.Printf("%s Error listing users: %v", config.System, err)
			return aphttp.DefaultServerError()
		}

		return serializeUsers(users, w)
	}
}

// CreateAccountUserHandler returns a handler that creates the user.
func CreateAccountUserHandler(db *sql.DB) aphttp.ErrorReturningHandler {
	return insertOrUpdateAccountUserHandler(db, true)
}

// ShowAccountUserHandler returns a handler that shows the user.
func ShowAccountUserHandler(db *sql.DB) aphttp.ErrorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		id := instanceID(r)
		user, err := model.FindUserForAccountID(db, id, accountID(r))
		if err != nil {
			return aphttp.NewError(fmt.Errorf("No user with id %d in account", id), 404)
		}

		return serialize(wrappedSanitizedUser{sanitizeUser(user)}, w)
	}
}

// UpdateAccountUserHandler returns a handler that updates the accountUser.
func UpdateAccountUserHandler(db *sql.DB) aphttp.ErrorReturningHandler {
	return insertOrUpdateAccountUserHandler(db, false)
}

// DeleteAccountUserHandler returns a handler that deletes the accountUser.
func DeleteAccountUserHandler(db *sql.DB) aphttp.ErrorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		err := performInTransaction(db, func(tx *sqlx.Tx) error {
			return model.DeleteUserForAccountID(tx, instanceID(r), accountID(r))
		})
		if err != nil {
			log.Printf("%s Error deleting user: %v", config.System, err)
			return aphttp.DefaultServerError()
		}

		w.WriteHeader(http.StatusOK)
		return nil
	}
}

func insertOrUpdateAccountUserHandler(db *sql.DB, isInsert bool) aphttp.ErrorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		user, err := readUser(r)
		if err != nil {
			log.Printf("%s Error reading user: %v", config.System, err)
			return aphttp.DefaultServerError()
		}

		var method func(*sqlx.Tx) error
		var desc string
		if isInsert {
			user.AccountID = accountID(r)
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

		err = performInTransaction(db, method)
		if err != nil {
			log.Printf("%s Error %s user: %v", config.System, desc, err)
			return aphttp.DefaultServerError()
		}

		return serialize(wrappedSanitizedUser{sanitizeUser(user)}, w)
	}
}
