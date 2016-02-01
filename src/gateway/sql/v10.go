package sql

func migrateToV10(db *DB) error {
	tx := db.MustBegin()
	tx.MustExec(db.SQL("v10/add_id_to_environment_data"))
	tx.MustExec(`UPDATE schema SET version = 10;`)
	return tx.Commit()
}
