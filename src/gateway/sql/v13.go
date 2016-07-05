package sql

func migrateToV13(db *DB) error {
	tx := db.MustBegin()
	tx.MustExec(db.SQL("v13/update_push_devices_and_messages"))
	tx.MustExec(`UPDATE schema SET version = 13;`)
	return tx.Commit()
}
