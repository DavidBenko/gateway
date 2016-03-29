package sql

func migrateToV11(db *DB) error {
	if db.Driver == Sqlite3 {
		db.DB.MustExec("PRAGMA foreign_keys=OFF;")
	}

	tx := db.MustBegin()
	tx.MustExec(db.SQL("v11/migrate"))
	tx.MustExec(`UPDATE schema SET version = 11;`)

	err := tx.Commit()

	if db.Driver == Sqlite3 {
		db.DB.MustExec("PRAGMA foreign_keys=ON;")
	}

	return err
}
