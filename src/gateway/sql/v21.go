package sql

func migrateToV21(db *DB) error {
	db.DisableSqliteTriggers()
	defer db.EnableSqliteTriggers()

	tx := db.MustBegin()
	tx.MustExec(db.SQL("v21/update_soap_remote_endpoint"))
	tx.MustExec(`UPDATE schema SET version = 21;`)
	return tx.Commit()
}
