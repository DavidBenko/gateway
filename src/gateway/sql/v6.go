package sql

func migrateToV6(db *DB) error {
	tx := db.MustBegin()
	tx.MustExec(db.SQL("v6/create_admin_column"))
	tx.MustExec(db.SQL("v6/create_token_column"))
	tx.MustExec(db.SQL("v6/create_confirmed_column"))
	tx.MustExec(`UPDATE schema SET version = 6;`)
	return tx.Commit()
}
