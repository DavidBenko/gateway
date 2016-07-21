package sql

func migrateToV14(db *DB) error {
	tx := db.MustBegin()
	tx.MustExec(db.SQL("v14/create_qos_column"))
	tx.MustExec(db.SQL("v14/create_mqtt_sessions"))
	tx.MustExec(`UPDATE schema SET version = 14;`)
	return tx.Commit()
}
