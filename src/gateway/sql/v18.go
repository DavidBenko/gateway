package sql

func migrateToV18(db *DB) error {
	db.DisableSqliteTriggers()
	defer db.EnableSqliteTriggers()

	tx := db.MustBegin()
	tx.MustExec(db.SQL("v18/add_timestamps"))
	tx.MustExec(`UPDATE schema SET version = 18;`)
	return tx.Commit()
}
