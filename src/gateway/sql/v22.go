package sql

func migrateToV22(db *DB) error {
	db.DisableSqliteTriggers()
	defer db.EnableSqliteTriggers()

	tx := db.MustBegin()
	tx.MustExec(db.SQL("v22/create_proxy_endpoint_channels"))
	tx.MustExec(db.SQL("v22/change_proxy_endpoint_tests"))
	tx.MustExec(`UPDATE schema SET version = 22;`)
	return tx.Commit()
}
