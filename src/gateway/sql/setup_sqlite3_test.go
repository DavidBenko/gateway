package sql

import "gateway/config"

func setupFreshDB() (*DB, error) {
	conf := config.Database{Driver: "sqlite3", ConnectionString: ":memory:"}
	return Connect(conf)
}

func testDB(version int) *DB {
	db, _ := setupFreshDB()
	setupSchemaTable(db)
	if version >= 1 {
		migrateToV1(db)
	}
	return db
}
