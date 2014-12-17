package sql

import "testing"

func TestSchemaCreation(t *testing.T) {
	db, _ := setupFreshMemoryDB()
	if err := setupSchemaTable(db); err != nil {
		t.Errorf("Should not error setting up schema table: %v", err)
	}
	version, err := db.CurrentVersion()
	if err != nil {
		t.Errorf("Should not error getting version after setup: %v", err)
	}
	if version != 0 {
		t.Errorf("Expected version to be 0 after setup, got %v", version)
	}
}
