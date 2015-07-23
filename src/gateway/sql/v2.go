package sql

func migrateToV2(db *DB) error {
	tx := db.MustBegin()
	tx.MustExec(db.SQL("v2/create_proxy_endpoint_tests"))
	tx.MustExec(db.SQL("v2/create_proxy_endpoint_test_pairs"))
	tx.MustExec(`UPDATE schema SET version = 2;`)
	return tx.Commit()
}
