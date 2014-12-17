package sql

func setupSchemaTable(db *DB) error {
	tx := db.MustBegin()
	tx.MustExec(`
    CREATE TABLE IF NOT EXISTS schema (
      version integer
    );
  `)
	tx.MustExec(`INSERT INTO schema VALUES (0);`)
	return tx.Commit()
}
