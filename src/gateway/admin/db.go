package admin

import (
	"gateway/config"
	"gateway/sql"
	"log"

	"github.com/jmoiron/sqlx"
)

type transactional func(tx *sqlx.Tx) error

func performInTransaction(db *sql.DB, method transactional) error {
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
