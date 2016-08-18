package sql

func migrateToV15(db *DB) error {
	tx := db.MustBegin()
	tx.MustExec(db.SQL("v15/create_plans"))
	tx.MustExec(db.SQL("v15/update_account_with_stripe_details"))
	tx.MustExec(`UPDATE schema SET version = 15;`)
	return tx.Commit()
}
