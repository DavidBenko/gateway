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

// RouteAPIs routes all the endpoints for api management by account admins.
func RouteAPIs(router aphttp.Router, db *sql.DB) {
	router.Handle("/apis",
		handlers.MethodHandler{
			"GET":  aphttp.ErrorCatchingHandler(ListAPIsHandler(db)),
			"POST": aphttp.ErrorCatchingHandler(CreateAPIHandler(db)),
		})
	router.Handle("/apis/{id}",
		handlers.HTTPMethodOverrideHandler(handlers.MethodHandler{
			"GET":    aphttp.ErrorCatchingHandler(ShowAPIHandler(db)),
			"PUT":    aphttp.ErrorCatchingHandler(UpdateAPIHandler(db)),
			"DELETE": aphttp.ErrorCatchingHandler(DeleteAPIHandler(db)),
		}))
}

// ListAPIsHandler returns a handler that lists the apis.
func ListAPIsHandler(db *sql.DB) aphttp.ErrorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		apis, err := model.AllAPIsForAccountID(db, accountIDFromSession(r))
		if err != nil {
			log.Printf("%s Error listing apis: %v", config.System, err)
			return aphttp.DefaultServerError()
		}

		return serializeAPIs(apis, w)
	}
}

// CreateAPIHandler returns a handler that creates the api.
func CreateAPIHandler(db *sql.DB) aphttp.ErrorReturningHandler {
	return insertOrUpdateAPIHandler(db, true)
}

// ShowAPIHandler returns a handler that shows the api.
func ShowAPIHandler(db *sql.DB) aphttp.ErrorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		id := instanceID(r)
		api, err := model.FindAPIForAccountID(db, id, accountIDFromSession(r))
		if err != nil {
			return aphttp.NewError(fmt.Errorf("No api with id %d in account", id), 404)
		}

		return serialize(wrappedAPI{api}, w)
	}
}

// UpdateAPIHandler returns a handler that updates the api.
func UpdateAPIHandler(db *sql.DB) aphttp.ErrorReturningHandler {
	return insertOrUpdateAPIHandler(db, false)
}

// DeleteAPIHandler returns a handler that deletes the api.
func DeleteAPIHandler(db *sql.DB) aphttp.ErrorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		err := performInTransaction(db, func(tx *sqlx.Tx) error {
			return model.DeleteAPIForAccountID(tx, instanceID(r), accountIDFromSession(r))
		})
		if err != nil {
			log.Printf("%s Error deleting api: %v", config.System, err)
			return aphttp.DefaultServerError()
		}

		w.WriteHeader(http.StatusOK)
		return nil
	}
}

func insertOrUpdateAPIHandler(db *sql.DB, isInsert bool) aphttp.ErrorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		api, err := readAPI(r)
		if err != nil {
			log.Printf("%s Error reading api: %v", config.System, err)
			return aphttp.DefaultServerError()
		}
		api.AccountID = accountIDFromSession(r)

		var method func(*sqlx.Tx) error
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

		err = performInTransaction(db, method)
		if err != nil {
			log.Printf("%s Error %s api: %v", config.System, desc, err)
			return aphttp.DefaultServerError()
		}

		return serialize(wrappedAPI{api}, w)
	}
}

type wrappedAPI struct {
	API *model.API `json:"api"`
}

func readAPI(r *http.Request) (*model.API, error) {
	var wrapped wrappedAPI
	if err := deserialize(&wrapped, r); err != nil {
		return nil, err
	}
	return wrapped.API, nil
}

func serializeAPIs(apis []*model.API, w http.ResponseWriter) aphttp.Error {
	wrappedAPIs := struct {
		APIs []*model.API `json:"apis"`
	}{apis}
	return serialize(wrappedAPIs, w)
}
