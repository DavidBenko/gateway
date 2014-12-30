package sql

func migrateToV1(db *DB) error {
	tx := db.MustBegin()
	tx.MustExec(db.sql("v1/create_accounts"))
	tx.MustExec(db.sql("v1/create_users"))
	tx.MustExec(db.sql("v1/create_apis"))
	tx.MustExec(`UPDATE schema SET version = 1;`)
	return tx.Commit()
}
