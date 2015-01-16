package admin

import (
	"gateway/config"
	aphttp "gateway/http"
	apsql "gateway/sql"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
)

type transactional func(tx *sqlx.Tx) error

func performInTransaction(db *apsql.DB, method transactional) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	methodErr := method(tx)
	if methodErr != nil {
		err = tx.Rollback()
		if err != nil {
			log.Printf("%s Error rolling back transaction!", config.System)
		}
		return methodErr
	}

	return tx.Commit()
}

// DatabaseAwareHandler adds an apsql.DB to an aphttp.ErrorReturningHandler
type DatabaseAwareHandler func(w http.ResponseWriter,
	r *http.Request,
	db *apsql.DB) aphttp.Error

// DatabaseWrappedHandler passes along the db to the handler.
func DatabaseWrappedHandler(db *apsql.DB, handler DatabaseAwareHandler) aphttp.ErrorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		return handler(w, r, db)
	}
}

// TransactionAwareHandler adds an apsql.Tx to an aphttp.ErrorReturningHandler
type TransactionAwareHandler func(w http.ResponseWriter,
	r *http.Request,
	tx *apsql.Tx) aphttp.Error

// TransactionWrappedHandler executes the handler inside a transaction.
func TransactionWrappedHandler(db *apsql.DB, handler TransactionAwareHandler) aphttp.ErrorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		tx, err := db.Begin()
		if err != nil {
			log.Printf("%s Error beginning transaction: %v", config.System, err)
			return aphttp.DefaultServerError()
		}
		handlerError := handler(w, r, tx)
		if handlerError != nil {
			if err = tx.Rollback(); err != nil {
				log.Printf("%s Error rolling back transaction: %v", config.System, err)
			}
		} else {
			if err = tx.Commit(); err != nil {
				log.Printf("%s Error committing transaction: %v", config.System, err)
			}
		}
		return handlerError
	}
}
