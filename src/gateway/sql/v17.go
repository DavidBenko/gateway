package sql

func migrateToV17(db *DB) error {
	db.DisableSqliteTriggers()
	defer db.EnableSqliteTriggers()

	tx := db.MustBegin()
	tx.MustExec(db.SQL("v17/create_job_fields"))
	tx.MustExec(db.SQL("v17/create_timers"))
	tx.MustExec(`UPDATE schema SET version = 17;`)
	return tx.Commit()
}
