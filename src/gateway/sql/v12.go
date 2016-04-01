package sql

func migrateToV12(db *DB) error {
	tx := db.MustBegin()
	tx.MustExec(db.SQL("v12/create_push_channels"))
	tx.MustExec(db.SQL("v12/create_push_devices"))
	tx.MustExec(db.SQL("v12/create_push_messages"))
	tx.MustExec(`UPDATE schema SET version = 12;`)
	return tx.Commit()
}
