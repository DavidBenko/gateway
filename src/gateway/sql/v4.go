package sql

func migrateToV4(db *DB) error {
	tx := db.MustBegin()
	tx.MustExec(db.SQL("v4/create_admin_column"))
	tx.MustExec(`UPDATE schema SET version = 4;`)
	return tx.Commit()
}
