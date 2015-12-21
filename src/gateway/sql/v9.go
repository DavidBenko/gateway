package sql

func migrateToV9(db *DB) error {
	tx := db.MustBegin()
	tx.MustExec(db.SQL("v9/set_token_not_null"))
	tx.MustExec(`UPDATE schema SET version = 9;`)
	return tx.Commit()
}
