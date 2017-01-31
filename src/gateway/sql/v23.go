package sql

func migrateToV23(db *DB) error {
	db.DisableSqliteTriggers()
	defer db.EnableSqliteTriggers()

	tx := db.MustBegin()
	tx.MustExec(db.SQL("v23/update_users"))
	tx.MustExec(`UPDATE schema SET version = 23;`)
	return tx.Commit()
}
