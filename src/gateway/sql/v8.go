package sql

func migrateToV8(db *DB) error {
	if db.Driver == Sqlite3 {
		db.DB.MustExec("PRAGMA foreign_keys=OFF;")
	}

	tx := db.MustBegin()
	tx.MustExec(db.SQL("v8/drop_account_name_unique_constraint"))
	tx.MustExec(`UPDATE schema SET version = 8;`)
	err := tx.Commit()

	if db.Driver == Sqlite3 {
		db.DB.MustExec("PRAGMA foreign_keys=ON;")
	}

	return err
}
