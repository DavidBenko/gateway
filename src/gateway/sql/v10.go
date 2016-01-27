package sql

func migrateToV10(db *DB) error {
	tx := db.MustBegin()
	tx.MustExec(db.SQL("v10/create_sessions"))
	tx.MustExec(db.SQL("v10/create_session_columns"))
	tx.MustExec(`UPDATE schema SET version = 10;`)
	return tx.Commit()
}
