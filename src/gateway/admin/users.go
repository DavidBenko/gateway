package admin

import (
	"fmt"
	"gateway/config"
	aphttp "gateway/http"
	"gateway/model"
	"gateway/sql"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/jmoiron/sqlx"
)

// RouteUsers routes all the endpoints for user management by account admins.
func RouteUsers(router aphttp.Router, db *sql.DB) {
	router.Handle("/users",
		handlers.MethodHandler{
			"GET":  aphttp.ErrorCatchingHandler(ListUsersHandler(db, accountIDFromSession)),
			"POST": aphttp.ErrorCatchingHandler(CreateUserHandler(db, accountIDFromSession)),
		})
	router.Handle("/users/{id}",
		handlers.HTTPMethodOverrideHandler(handlers.MethodHandler{
			"GET":    aphttp.ErrorCatchingHandler(ShowUserHandler(db, accountIDFromSession)),
			"PUT":    aphttp.ErrorCatchingHandler(UpdateUserHandler(db, accountIDFromSession)),
			"DELETE": aphttp.ErrorCatchingHandler(DeleteUserHandler(db, accountIDFromSession)),
		}))
}

// ListUsersHandler returns a handler that lists the users.
func ListUsersHandler(db *sql.DB, accountID func(r *http.Request) int64) aphttp.ErrorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		users, err := model.AllUsersForAccountID(db, accountID(r))
		if err != nil {
			log.Printf("%s Error listing users: %v", config.System, err)
			return aphttp.DefaultServerError()
		}

		return serializeUsers(users, w)
	}
}

// CreateUserHandler returns a handler that creates the user.
func CreateUserHandler(db *sql.DB, accountID func(r *http.Request) int64) aphttp.ErrorReturningHandler {
	return insertOrUpdateUserHandler(db, accountID, true)
}

// ShowUserHandler returns a handler that shows the user.
func ShowUserHandler(db *sql.DB, accountID func(r *http.Request) int64) aphttp.ErrorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		id := instanceID(r)
		user, err := model.FindUserForAccountID(db, id, accountID(r))
		if err != nil {
			return aphttp.NewError(fmt.Errorf("No user with id %d in account", id), 404)
		}

		return serialize(wrappedSanitizedUser{sanitizeUser(user)}, w)
	}
}

// UpdateUserHandler returns a handler that updates the user.
func UpdateUserHandler(db *sql.DB, accountID func(r *http.Request) int64) aphttp.ErrorReturningHandler {
	return insertOrUpdateUserHandler(db, accountID, false)
}

// DeleteUserHandler returns a handler that deletes the user.
func DeleteUserHandler(db *sql.DB, accountID func(r *http.Request) int64) aphttp.ErrorReturningHandler {
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

func insertOrUpdateUserHandler(db *sql.DB, accountID func(r *http.Request) int64, isInsert bool) aphttp.ErrorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		user, err := readUser(r)
		if err != nil {
			log.Printf("%s Error reading user: %v", config.System, err)
			return aphttp.DefaultServerError()
		}

		user.AccountID = accountID(r)
		var method func(*sqlx.Tx) error
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

		err = performInTransaction(db, method)
		if err != nil {
			log.Printf("%s Error %s user: %v", config.System, desc, err)
			return aphttp.DefaultServerError()
		}

		return serialize(wrappedSanitizedUser{sanitizeUser(user)}, w)
	}
}

type wrappedUser struct {
	User *model.User `json:"user"`
}

type sanitizedUser struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type wrappedSanitizedUser struct {
	User *sanitizedUser `json:"user"`
}

func sanitizeUser(user *model.User) *sanitizedUser {
	return &sanitizedUser{user.ID, user.Name, user.Email}
}

func readUser(r *http.Request) (*model.User, error) {
	var wrapped wrappedUser
	if err := deserialize(&wrapped, r); err != nil {
		return nil, err
	}
	return wrapped.User, nil
}

func serializeUsers(users []*model.User, w http.ResponseWriter) aphttp.Error {
	wrappedUsers := struct {
		Users []*sanitizedUser `json:"users"`
	}{}
	for _, user := range users {
		wrappedUsers.Users = append(wrappedUsers.Users, sanitizeUser(user))
	}
	return serialize(wrappedUsers, w)
}
