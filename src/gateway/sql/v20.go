package sql

func migrateToV20(db *DB) error {
	tx := db.MustBegin()
	tx.MustExec(db.SQL("v20/create_job_tests"))
	tx.MustExec(`UPDATE schema SET version = 20;`)
	return tx.Commit()
}
