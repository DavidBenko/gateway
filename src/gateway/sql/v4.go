package sql

func migrateToV4(db *DB) error {
	tx := db.MustBegin()
	tx.MustExec(db.SQL("v4/create_schemas"))
	tx.MustExec(db.SQL("v4/create_proxy_endpoint_schemas"))
	tx.MustExec(`UPDATE schema SET version = 4;`)
	return tx.Commit()
}
