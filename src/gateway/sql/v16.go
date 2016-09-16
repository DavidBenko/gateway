package sql

func migrateToV16(db *DB) error {
	db.DisableSqliteTriggers()
	defer db.EnableSqliteTriggers()

	tx := db.MustBegin()

	// Migrate dev mode account from JustAPIs to generic email. Can apply to both sqlite and postgres since either can run in dev-mode.
	tx.MustExec(db.q(`UPDATE users SET email = ? WHERE email = ?`), `developer@example.net`, `developer@justapis.com`)
	tx.MustExec(`UPDATE schema SET version = 16;`)
	return tx.Commit()
}
