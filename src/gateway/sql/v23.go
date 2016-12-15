package sql

func migrateToV23(db *DB) error {
	tx := db.MustBegin()
	tx.MustExec(db.SQL("v23/create_custom_functions"))
	tx.MustExec(db.SQL("v23/create_custom_function_files"))
	tx.MustExec(`UPDATE schema SET version = 23;`)
	return tx.Commit()
}
