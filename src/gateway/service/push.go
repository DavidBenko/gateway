package service

import (
	"time"

	"gateway/config"
	"gateway/logreport"
	"gateway/sql"
)

func PushDeletionService(conf config.Configuration, db *sql.DB) {
	if !conf.Jobs {
		return
	}

	deleteTicker := time.NewTicker(24 * time.Hour)
	go func() {
		for _ = range deleteTicker.C {
			now := time.Now().Unix()
			err := db.DoInTransaction(func(tx *sql.Tx) error {
				_, err := tx.Exec(tx.SQL("push_channels/delete_stale"), now)
				if err != nil {
					return err
				}
				_, err = tx.Exec(tx.SQL("push_devices/delete_stale"), now)
				return err
			})
			if err != nil {
				logreport.Printf("[sessions] %v", err)
			}
		}
	}()
}
