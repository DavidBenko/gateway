package testing

import (
	"errors"
	"gateway/db"
)

// Spec implements db.Specifier with stubbed return values.
type Spec struct {
	Conn         string
	Unique       string
	UpdateNeeded func(db.Specifier, db.Specifier) bool
	NewDBSpec    db.Specifier
	NewDBFunc    DBMaker
	NewDBError   error
	UpdateFunc   Updater
	UpdateErr    error
}

func (d *Spec) ConnectionString() string {
	return d.Conn
}

func (d *Spec) UniqueServer() string {
	return d.Unique
}

func (d *Spec) NeedsUpdate(s db.Specifier) bool {
	return d.UpdateNeeded(d, s)
}

func (d *Spec) NewDB() (db.DB, error) {
	return d.NewDBFunc(d.NewDBSpec, d.UpdateFunc, d.UpdateErr, d.NewDBError)
}

// MakeSpec creates a Spec with the given values as the return values for its
// methods.
//
// newDBSpec is the db.Specifier which will be used as the Specifier stub in
// NewDB.
//
// newDBFunc(spec, update, error) is the function which will be used in NewDB to
// create the new DB using the given spec and update stub values, and returning
// the given error.
//
// testing.MakeDB may be used as newDBFunc.
//
// updateFunc func(*DB, db.Specifier) error is the function which will be used
// by the new DB's Update(Specifier) method.
//
// testing.BasicUpdate may be used as updateFunc.
func MakeSpec(
	conn, unique string,
	needsUpdate func(db.Specifier, db.Specifier) bool,
	newDBSpec db.Specifier,
	newDBFunc DBMaker,
	newDBError error,
	updateFunc Updater,
	updateErr error,
) db.Specifier {
	return &Spec{
		Conn:         conn,
		Unique:       unique,
		UpdateNeeded: needsUpdate,
		NewDBSpec:    newDBSpec,
		NewDBFunc:    newDBFunc,
		NewDBError:   newDBError,
		UpdateFunc:   updateFunc,
		UpdateErr:    updateErr,
	}
}

func CompareConn(a, b db.Specifier) bool {
	return a.ConnectionString() == b.ConnectionString()
}

type specBlueprint struct {
	Spec
	NewDBError  error
	UpdateError error
}

var testingSpecs = map[string]specBlueprint{
	"simple-ok": specBlueprint{Spec: Spec{
		Conn:         "simple-ok",
		Unique:       "simple-ok",
		UpdateNeeded: CompareConn,
		NewDBFunc:    MakeDB,
		UpdateFunc:   BasicUpdate,
	}},
	"simple-ok-1": specBlueprint{Spec: Spec{
		Conn:         "simple-ok-1",
		Unique:       "simple-ok-1",
		UpdateNeeded: CompareConn,
		NewDBFunc:    MakeDB,
		UpdateFunc:   BasicUpdate,
	}},
	"simple-ok-2": specBlueprint{Spec: Spec{
		Conn:         "simple-ok-2",
		Unique:       "simple-ok-2",
		UpdateNeeded: CompareConn,
		NewDBFunc:    MakeDB,
		UpdateFunc:   BasicUpdate,
	}},
	"simple-ok-3": specBlueprint{Spec: Spec{
		Conn:         "simple-ok-3",
		Unique:       "simple-ok-3",
		UpdateNeeded: CompareConn,
		NewDBFunc:    MakeDB,
		UpdateFunc:   BasicUpdate,
	}},
	"simple-newdb-err": specBlueprint{Spec: Spec{
		Conn:         "simple-newdb-err",
		Unique:       "simple-newdb-err",
		UpdateNeeded: CompareConn,
		NewDBFunc:    MakeDB,
		UpdateFunc:   BasicUpdate,
	},
		NewDBError: errors.New("NewDB error"),
	},
}

// Specs creates a new slice of db.Specifier using a blueprint.  We don't want
// external packages to edit that blueprint.  Note that their NewDB method will
// use the given spec to create the testing.DB.
func Specs() map[string]db.Specifier {
	specs := make(map[string]db.Specifier)
	for name, spec := range testingSpecs {
		s := MakeSpec(
			spec.Conn, spec.Unique,
			spec.UpdateNeeded,
			nil,
			spec.NewDBFunc,
			spec.NewDBError,
			spec.UpdateFunc,
			spec.UpdateError,
		)
		sp := s.(*Spec)
		sp.NewDBSpec = s
		specs[name] = s
	}
	return specs
}
