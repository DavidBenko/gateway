package sql

func migrateToV5(db *DB) error {
	tx := db.MustBegin()
	tx.MustExec(db.SQL("v5/add_show_javascript_errors"))
	tx.MustExec(`UPDATE schema SET version = 5;`)
	return tx.Commit()
}
