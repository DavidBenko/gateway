package sql

import (
	"gateway/config"
	"testing"
)

func TestConnectSqlite(t *testing.T) {
	conf := config.Database{Driver: "sqlite3", ConnectionString: ":memory:"}
	if _, err := Connect(conf); err != nil {
		t.Error(err)
	}
}

func TestConnectBogus(t *testing.T) {
	conf := config.Database{Driver: "notAdriver", ConnectionString: ":memory:"}
	if _, err := Connect(conf); err == nil {
		t.Error("Driver must be in our recognized list")
	}
}

func setupFreshMemoryDB() (*DB, error) {
	conf := config.Database{Driver: "sqlite3", ConnectionString: ":memory:"}
	return Connect(conf)
}

func TestCurrentVersionFresh(t *testing.T) {
	db, _ := setupFreshMemoryDB()
	if _, err := db.CurrentVersion(); err == nil {
		t.Error("Fresh database should not have a version")
	}
}

func TestUpToDateFresh(t *testing.T) {
	db, _ := setupFreshMemoryDB()
	if db.UpToDate() {
		t.Error("Fresh database should not be up to date")
	}
}

func TestMigrate(t *testing.T) {
	db, _ := setupFreshMemoryDB()
	db.Migrate()
	if !db.UpToDate() {
		t.Error("Migrated database should be up to date")
	}
}
