package sql

func migrateToV1(db *DB) error {
	tx := db.MustBegin()
	tx.MustExec(db.sql("v1/create_accounts"))
	tx.MustExec(db.sql("v1/create_users"))
	tx.MustExec(db.sql("v1/create_apis"))
	tx.MustExec(db.sql("v1/create_hosts"))
	tx.MustExec(db.sql("v1/create_environments"))
	tx.MustExec(db.sql("v1/create_remote_endpoints"))
	tx.MustExec(db.sql("v1/create_remote_endpoint_environment_data"))
	tx.MustExec(db.sql("v1/create_endpoint_groups"))
	tx.MustExec(db.sql("v1/create_proxy_endpoints"))
	tx.MustExec(db.sql("v1/create_proxy_endpoint_components"))
	tx.MustExec(db.sql("v1/create_proxy_endpoint_calls"))
	tx.MustExec(db.sql("v1/create_proxy_endpoint_transformations"))
	tx.MustExec(db.sql("v1/create_routes"))
	tx.MustExec(`UPDATE schema SET version = 1;`)
	return tx.Commit()
}
