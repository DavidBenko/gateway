package sql

func migrateToV11(db *DB) error {
	tx := db.MustBegin()
	tx.MustExec(db.SQL("v11/migrate"))
	tx.MustExec(`UPDATE schema SET version = 11;`)
	return tx.Commit()
}
