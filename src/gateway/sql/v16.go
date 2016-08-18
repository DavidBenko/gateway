package sql

func migrateToV16(db *DB) error {
	tx := db.MustBegin()
	tx.MustExec(db.SQL("v16/create_keys"))
	tx.MustExec(`UPDATE schema SET version = 16;`)
	return tx.Commit()
}
