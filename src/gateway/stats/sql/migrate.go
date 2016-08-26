package sql

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Migrate migrates the given *sqlx.DB assuming a `version` table exists
// specifying the version of the stats schema.  If no version is given, the
// schema will be created.  This function will panic if no migration script
// exists for the given version migration.
func Migrate(s *sqlx.DB, driver Driver) (e error) {
	defer func() {
		if r := recover(); r != nil {
			if tR, ok := r.(error); ok {
				e = tR
			} else {
				panic(r)
			}
		}
	}()

	var version int64

	tx := s.MustBegin()
	tx.MustExec(`CREATE TABLE IF NOT EXISTS stats_schema (version INTEGER);`)
	tx.Commit()

	tx = s.MustBegin()
	for v := version; v < Version; v++ {
		tx.MustExec(string(MustAsset(fmt.Sprintf(
			"static/migrations/%s/v%d.sql", driver, v+1,
		))))
		tx.MustExec(fmt.Sprintf(
			"UPDATE stats_schema SET version = %d;", v,
		))
	}
	tx.Commit()
	return nil
}
