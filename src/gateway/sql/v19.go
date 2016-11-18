package sql

func migrateToV19(db *DB) error {
	db.DisableSqliteTriggers()
	defer db.EnableSqliteTriggers()

	tx := db.MustBegin()
	tx.MustExec(db.SQL("v19/change_timers_unique"))
	tx.MustExec(`UPDATE schema SET version = 19;`)
	return tx.Commit()
}
