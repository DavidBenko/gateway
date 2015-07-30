package testing

import (
	"fmt"
	"gateway/db"
)

// DB implements db.DB with stubbed return values for Spec and Update.
type DB struct {
	spec        db.Specifier
	update      Updater
	updateError error
}

func (d *DB) Spec() db.Specifier {
	return d.spec
}

func (d *DB) Update(s db.Specifier) error {
	return d.update(d, s)
}

// Updater takes a db.DB and a db.Specifier and returns an error.
type Updater func(d db.DB, s db.Specifier) error

// BasicUpdate expects a *DB and returns its updateError.  It can function as
// an Updater to be used in MakeSpec.
func BasicUpdate(d db.DB, s db.Specifier) error {
	if tdb, ok := d.(*DB); ok {
		if tdb.updateError == nil {
			tdb.spec = s
		}
		return tdb.updateError
	}
	return fmt.Errorf("BasicUpdate needs *testing.DB, received %T", d)
}

// DBMaker creates a DB using the given Specifer, Updater, and update error.
// It returns the makeError it is given.  See MakeDB for an example.
type DBMaker func(db.Specifier, Updater, error, error) (db.DB, error)

// MakeDB returns a db.DB which uses the given spec and update as the values
// for its Spec() db.Specifier and Update(db.Specifier) methods.  It can
// function as a DBMaker in MakeSpec.
func MakeDB(spec db.Specifier, update Updater, updateErr, makeErr error) (db.DB, error) {
	return &DB{
		spec:        spec,
		update:      update,
		updateError: updateErr,
	}, makeErr
}
