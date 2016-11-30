package sql

func migrateToV21(db *DB) error {
	db.DisableSqliteTriggers()
	defer db.EnableSqliteTriggers()

	tx := db.MustBegin()
	tx.MustExec(db.SQL("v21/create_proxy_endpoint_channels"))
	tx.MustExec(db.SQL("v21/change_proxy_endpoint_tests"))
	tx.MustExec(`UPDATE schema SET version = 21;`)
	return tx.Commit()
}
