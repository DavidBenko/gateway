package sql

func migrateToV19(db *DB) error {
	tx := db.MustBegin()
	tx.MustExec(db.SQL("v19/create_job_tests"))
	tx.MustExec(`UPDATE schema SET version = 19;`)
	return tx.Commit()
}
