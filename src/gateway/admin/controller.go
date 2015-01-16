package admin

import (
	"gateway/config"
	aphttp "gateway/http"
	apsql "gateway/sql"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
)

type Controller interface {
	List(db *apsql.DB) aphttp.ErrorReturningHandler
	Create(db *apsql.DB) aphttp.ErrorReturningHandler
	Show(db *apsql.DB) aphttp.ErrorReturningHandler
	Update(db *apsql.DB) aphttp.ErrorReturningHandler
	Delete(db *apsql.DB) aphttp.ErrorReturningHandler
}

func RouteResource(controller Controller, path string, router aphttp.Router, db *apsql.DB) {
	router.Handle(path,
		handlers.MethodHandler{
			"GET":  aphttp.ErrorCatchingHandler(controller.List(db)),
			"POST": aphttp.ErrorCatchingHandler(controller.Create(db)),
		})
	router.Handle(path+"/{id}",
		handlers.HTTPMethodOverrideHandler(handlers.MethodHandler{
			"GET":    aphttp.ErrorCatchingHandler(controller.Show(db)),
			"PUT":    aphttp.ErrorCatchingHandler(controller.Update(db)),
			"DELETE": aphttp.ErrorCatchingHandler(controller.Delete(db)),
		}))
}

// TransactionAwareHandler TODO
type TransactionAwareHandler func(w http.ResponseWriter,
	r *http.Request,
	tx *apsql.Tx) aphttp.Error

// TransactionWrappedHandler catches an error a handler throws and responds with it.
func TransactionWrappedHandler(db *apsql.DB, handler TransactionAwareHandler) aphttp.ErrorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		tx, err := db.Begin()
		if err != nil {
			log.Printf("%s Error beginning transaction: %v", config.System, err)
			return aphttp.DefaultServerError()
		}
		handlerError := handler(w, r, tx)
		if handlerError != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
		return handlerError
	}
}
