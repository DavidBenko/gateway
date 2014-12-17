package sql

func migrateToV1(db *DB) error {
	tx := db.MustBegin()
	tx.MustExec(db.sql("v1/create_accounts"))
	tx.MustExec(`UPDATE schema SET version = 1;`)
	return tx.Commit()
}
