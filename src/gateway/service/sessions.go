package service

import (
	"time"

	"gateway/config"
	"gateway/logreport"
	"gateway/sql"
)

func SessionDeletionService(conf config.Configuration, db *sql.DB) {
	if !conf.Jobs {
		return
	}

	deleteTicker := time.NewTicker(24 * time.Hour)
	go func() {
		for _ = range deleteTicker.C {
			err := db.DoInTransaction(func(tx *sql.Tx) error {
				_, err := tx.Exec(tx.SQL("sessions/delete_stale"), time.Now().Unix())
				return err
			})
			if err != nil {
				logreport.Printf("[sessions] %v", err)
			}
		}
	}()
}
