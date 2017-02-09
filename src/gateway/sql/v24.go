package sql

func migrateToV24(db *DB) error {
	db.DisableSqliteTriggers()
	defer db.EnableSqliteTriggers()

	tx := db.MustBegin()
	tx.MustExec(db.SQL("v24/update_users"))
	tx.MustExec(`UPDATE schema SET version = 24;`)
	return tx.Commit()
}
