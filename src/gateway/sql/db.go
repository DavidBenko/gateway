package sql

import (
	"fmt"
	"gateway/config"
	"log"

	"github.com/jmoiron/sqlx"

	// Add sqlite3 driver
	_ "github.com/mattn/go-sqlite3"
)

const currentVersion = 1

type driverType int64

const (
	// Sqlite3 driver type
	Sqlite3 driverType = iota

	// Postgres driver type
	Postgres
)

// DB wraps a *sqlx.DB with some convenience methods and data
type DB struct {
	*sqlx.DB
	Driver driverType
}

// Connect opens and returns a database connection.
func Connect(conf config.Database) (*DB, error) {
	var driver driverType
	switch conf.Driver {
	case "sqlite3":
		driver = Sqlite3
	case "postgres":
		driver = Postgres
	default:
		return nil,
			fmt.Errorf("Database driver must be sqlite3 or postgres (got '%v')",
				conf.Driver)
	}

	log.Printf("%s Connecting to database", config.System)
	db, err := sqlx.Connect(conf.Driver, conf.ConnectionString)
	if err != nil {
		return nil, err
	}

	return &DB{db, driver}, nil
}

// CurrentVersion returns the current version of the database, or an error if
// it has not been initialized.
func (db *DB) CurrentVersion() (int64, error) {
	var version int64
	err := db.Get(&version, `SELECT version FROM schema LIMIT 1`)
	if err != nil {
		return 0, fmt.Errorf("Could not get latest schema version: %v", err)
	}
	return version, err
}

// UpToDate returns whether or not the database is up to date
// with the latest schema
func (db *DB) UpToDate() bool {
	version, err := db.CurrentVersion()
	return (err == nil) && (version == currentVersion)
}

// Migrate migrates the database to the current version
func (db *DB) Migrate() error {
	version, err := db.CurrentVersion()

	if err != nil {
		if err = setupSchemaTable(db); err != nil {
			return fmt.Errorf("Could not create schema table: %v", err)
		}
	}

	if version < 1 {
		if err = migrateToV1(db); err != nil {
			return fmt.Errorf("Could not migrate to schema v1: %v", err)
		}
	}

	return nil
}
