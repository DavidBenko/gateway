package sql

import (
	"fmt"
)

// CurrentVersion returns the current version of the database, or an error if
// it has not been initialized.
func (db *SQL) CurrentVersion() (int64, error) {
	var version int64
	err := db.Get(&version, `SELECT version FROM stats_schema LIMIT 1`)
	if err != nil {
		return 0, fmt.Errorf("Could not get latest stats schema version: %v", err)
	}
	return version, err
}

// UpToDate returns whether or not the database is up to date
// with the latest schema
func (db *SQL) UpToDate() bool {
	version, err := db.CurrentVersion()
	return (err == nil) && (version == Version)
}

// Migrate migrates the given *sqlx.DB assuming a `version` table exists
// specifying the version of the stats schema.  If no version is given, the
// schema will be created.  This function will panic if no migration script
// exists for the given version migration.
func (db *SQL) Migrate() (e error) {
	defer func() {
		if r := recover(); r != nil {
			if tR, ok := r.(error); ok {
				e = tR
			} else {
				panic(r)
			}
		}
	}()

	version, err := db.CurrentVersion()
	if err != nil {
		if err = setupStatsSchemaTable(db); err != nil {
			return fmt.Errorf("Could not create schema table: %v", err)
		}
	}

	for v := version; v < Version; v++ {
		tx := db.MustBegin()
		tx.MustExec(string(MustAsset(fmt.Sprintf(
			"static/migrations/%s/v%d.sql", db.DriverName(), v+1,
		))))
		tx.MustExec(fmt.Sprintf(
			"UPDATE stats_schema SET version = %d;", v+1,
		))
		tx.Commit()
	}
	return nil
}

func setupStatsSchemaTable(db *SQL) error {
	tx := db.MustBegin()
	tx.MustExec(`CREATE TABLE IF NOT EXISTS stats_schema (version INTEGER);`)
	tx.MustExec(`INSERT INTO stats_schema VALUES (0);`)
	return tx.Commit()
}
