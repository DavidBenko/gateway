package sql

func migrateToV8(db *DB) error {
	if db.Driver == Sqlite3 {
		if _, err := db.Exec("PRAGMA foreign_keys=OFF;"); err != nil {
			return err
		}
	}
	tx := db.MustBegin()
	tx.MustExec(db.SQL("v8/drop_account_name_unique_constraint"))
	tx.MustExec(`UPDATE schema SET version = 8;`)
	err := tx.Commit()

	if db.Driver == Sqlite3 {
		if _, fkErr := db.Exec("PRAGMA foreign_keys=ON;"); fkErr != nil {
			if err != nil {
				return err
			}
			return fkErr
		}

		if _, fkErr := db.Queryx("PRAGMA foreign_key_check;"); fkErr != nil {
			if err != nil {
				return err
			}
			return fkErr
		}
	}

	return err
}
