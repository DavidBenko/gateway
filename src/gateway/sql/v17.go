package sql

func migrateToV17(db *DB) error {
	db.MustExec(`PRAGMA foreign_keys = OFF;`)
	tx := db.MustBegin()
	tx.MustExec(db.SQL("v17/create_job_fields"))
	tx.MustExec(`UPDATE schema SET version = 17;`)
	err := tx.Commit()
	if err != nil {
		return err
	}
	db.MustExec(`PRAGMA foreign_keys = ON;`)
	return nil
}
