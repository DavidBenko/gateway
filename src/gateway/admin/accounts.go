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

// RouteAccounts routes all the endpoints for account management
func RouteAccounts(router aphttp.Router, db *sql.DB) {
	router.Handle("/accounts",
		handlers.MethodHandler{
			"GET":  aphttp.ErrorCatchingHandler(ListAccountsHandler(db)),
			"POST": aphttp.ErrorCatchingHandler(CreateAccountHandler(db)),
		})
	router.Handle("/accounts/{id}",
		handlers.HTTPMethodOverrideHandler(handlers.MethodHandler{
			"GET":    aphttp.ErrorCatchingHandler(ShowAccountHandler(db)),
			"PUT":    aphttp.ErrorCatchingHandler(UpdateAccountHandler(db)),
			"DELETE": aphttp.ErrorCatchingHandler(DeleteAccountHandler(db)),
		}))
}

// ListAccountsHandler returns a handler that lists the accounts.
func ListAccountsHandler(db *sql.DB) aphttp.ErrorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		accounts, err := model.AllAccounts(db)
		if err != nil {
			log.Printf("%s Error listing accounts: %v", config.System, err)
			return aphttp.DefaultServerError()
		}

		return serializeAccounts(accounts, w)
	}
}

// CreateAccountHandler returns a handler that creates the account.
func CreateAccountHandler(db *sql.DB) aphttp.ErrorReturningHandler {
	return insertOrUpdateAccountHandler(db, true)
}

// ShowAccountHandler returns a handler that shows the account.
func ShowAccountHandler(db *sql.DB) aphttp.ErrorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		id := instanceID(r)
		account, err := model.FindAccount(db, id)
		if err != nil {
			return aphttp.NewError(fmt.Errorf("No account with id %d", id), 404)
		}

		return serialize(wrappedAccount{account}, w)
	}
}

// UpdateAccountHandler returns a handler that updates the account.
func UpdateAccountHandler(db *sql.DB) aphttp.ErrorReturningHandler {
	return insertOrUpdateAccountHandler(db, false)
}

// DeleteAccountHandler returns a handler that deletes the account.
func DeleteAccountHandler(db *sql.DB) aphttp.ErrorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		err := performInTransaction(db, func(tx *sqlx.Tx) error {
			return model.DeleteAccount(tx, instanceID(r))
		})
		if err != nil {
			log.Printf("%s Error deleting account: %v", config.System, err)
			return aphttp.DefaultServerError()
		}

		w.WriteHeader(http.StatusOK)
		return nil
	}
}

func insertOrUpdateAccountHandler(db *sql.DB, isInsert bool) aphttp.ErrorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		account, err := readAccount(r)
		if err != nil {
			log.Printf("%s Error reading account: %v", config.System, err)
			return aphttp.DefaultServerError()
		}

		var method func(*sqlx.Tx) error
		var desc string
		if isInsert {
			method = account.Insert
			desc = "inserting"
		} else {
			account.ID = instanceID(r)
			method = account.Update
			desc = "updating"
		}

		validationErrors := account.Validate()
		if !validationErrors.Empty() {
			return serialize(wrappedErrors{validationErrors}, w)
		}

		err = performInTransaction(db, method)
		if err != nil {
			log.Printf("%s Error %s account: %v", config.System, desc, err)
			return aphttp.DefaultServerError()
		}

		return serialize(wrappedAccount{account}, w)
	}
}

type wrappedAccount struct {
	Account *model.Account `json:"account"`
}

func readAccount(r *http.Request) (*model.Account, error) {
	var wrapped wrappedAccount
	if err := deserialize(&wrapped, r); err != nil {
		return nil, err
	}
	return wrapped.Account, nil
}

func serializeAccounts(accounts []model.Account, w http.ResponseWriter) aphttp.Error {
	wrappedAccounts := struct {
		Accounts []model.Account `json:"accounts"`
	}{accounts}
	return serialize(wrappedAccounts, w)
}
