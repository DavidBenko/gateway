package sql

func migrateToV1(db *DB) error {
	tx := db.MustBegin()
	tx.MustExec(`
    CREATE TABLE IF NOT EXISTS accounts (
      id integer PRIMARY KEY,
      name varchar(255)
      );
      `)
	tx.MustExec(`UPDATE schema SET version = 1;`)
	return tx.Commit()
}
