package sql

func migrateToV7(db *DB) error {
	tx := db.MustBegin()
	tx.MustExec(db.SQL("v7/create_indexes"))
	tx.MustExec(`UPDATE schema SET version = 7;`)
	return tx.Commit()
}
