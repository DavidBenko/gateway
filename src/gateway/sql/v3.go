package sql

func migrateToV3(db *DB) error {
	tx := db.MustBegin()
	tx.MustExec(db.SQL("v3/create_soap_remote_endpoints"))
	tx.MustExec(`UPDATE schema SET version = 3;`)
	return tx.Commit()
}
