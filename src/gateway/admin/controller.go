package admin

import (
	"database/sql"
	"gateway/config"
	aphttp "gateway/http"
	apsql "gateway/sql"
	"log"
	"net/http"
)

// requestID := context.Get(r, aphttp.ContextRequestIDKey).(string)

// TransactionAwareHandler TODO
type TransactionAwareHandler func(w http.ResponseWriter,
	r *http.Request,
	tx *sql.Tx) aphttp.Error

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
