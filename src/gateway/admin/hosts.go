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

// RouteHosts routes all the endpoints for host management by account admins.
func RouteHosts(router aphttp.Router, db *sql.DB) {
	router.Handle("/apis/{apiID}/hosts",
		handlers.MethodHandler{
			"GET":  aphttp.ErrorCatchingHandler(ListHostsHandler(db)),
			"POST": aphttp.ErrorCatchingHandler(CreateHostHandler(db)),
		})
	router.Handle("/apis/{apiID}/hosts/{id}",
		handlers.HTTPMethodOverrideHandler(handlers.MethodHandler{
			"GET": aphttp.ErrorCatchingHandler(ShowHostHandler(db)),
			"PUT": aphttp.ErrorCatchingHandler(UpdateHostHandler(db)),
			// "DELETE": aphttp.ErrorCatchingHandler(DeleteHostHandler(db)),
		}))
}

// ListHostsHandler returns a handler that lists the hosts.
func ListHostsHandler(db *sql.DB) aphttp.ErrorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		hosts, err := model.AllHostsForAPIIDAndAccountID(db,
			apiIDFromPath(r), accountIDFromSession(r))
		if err != nil {
			log.Printf("%s Error listing hosts: %v", config.System, err)
			return aphttp.DefaultServerError()
		}

		return serializeHosts(hosts, w)
	}
}

// CreateHostHandler returns a handler that creates the host.
func CreateHostHandler(db *sql.DB) aphttp.ErrorReturningHandler {
	return insertOrUpdateHostHandler(db, true)
}

// ShowHostHandler returns a handler that shows the host.
func ShowHostHandler(db *sql.DB) aphttp.ErrorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		id := instanceID(r)
		host, err := model.FindHostForAPIIDAndAccountID(db, id,
			apiIDFromPath(r), accountIDFromSession(r))
		if err != nil {
			return aphttp.NewError(fmt.Errorf("No host with id %d in api", id), 404)
		}

		return serialize(wrappedHost{host}, w)
	}
}

// UpdateHostHandler returns a handler that updates the host.
func UpdateHostHandler(db *sql.DB) aphttp.ErrorReturningHandler {
	return insertOrUpdateHostHandler(db, false)
}

// DeleteHostHandler returns a handler that deletes the host.
func DeleteHostHandler(db *sql.DB) aphttp.ErrorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		err := performInTransaction(db, func(tx *sqlx.Tx) error {
			return model.DeleteHostForAPIIDAndAccountID(tx, instanceID(r),
				apiIDFromPath(r), accountIDFromSession(r))
		})
		if err != nil {
			log.Printf("%s Error deleting host: %v", config.System, err)
			return aphttp.DefaultServerError()
		}

		w.WriteHeader(http.StatusOK)
		return nil
	}
}

func insertOrUpdateHostHandler(db *sql.DB, isInsert bool) aphttp.ErrorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		host, err := readHost(r)
		if err != nil {
			log.Printf("%s Error reading host: %v", config.System, err)
			return aphttp.DefaultServerError()
		}
		host.APIID = apiIDFromPath(r)
		host.AccountID = accountIDFromSession(r)

		var method func(*sqlx.Tx) error
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

		err = performInTransaction(db, method)
		if err != nil {
			log.Printf("%s Error %s host: %v", config.System, desc, err)
			return aphttp.DefaultServerError()
		}

		return serialize(wrappedHost{host}, w)
	}
}

type wrappedHost struct {
	Host *model.Host `json:"host"`
}

func readHost(r *http.Request) (*model.Host, error) {
	var wrapped wrappedHost
	if err := deserialize(&wrapped, r); err != nil {
		return nil, err
	}
	return wrapped.Host, nil
}

func serializeHosts(hosts []*model.Host, w http.ResponseWriter) aphttp.Error {
	wrappedHosts := struct {
		Hosts []*model.Host `json:"hosts"`
	}{hosts}
	return serialize(wrappedHosts, w)
}
