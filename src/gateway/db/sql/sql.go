package sql

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/jmoiron/sqlx"

	"gateway/db"
)

type driver string

const (
	postgres driver = "pgx"
	mssql    driver = "mssql"
	mysql    driver = "mysql"
)

var knownDrivers *regexp.Regexp

func init() {
	var drivers string
	for _, drv := range []string{
		string(mysql),
		string(mssql),
		string(postgres),
	} {
		drivers += fmt.Sprintf("(%s)|", drv)
	}
	knownDrivers = regexp.MustCompile(drivers[:len(drivers)-1])
}

type sqlSpec interface {
	db.Specifier
	getMaxOpenIdle() (int, int)
	setMaxOpenIdle(int, int)
	driver() driver
	validate() error
}

// spec is a base type for a SQL connection.
type spec struct {
	maxIdle int
	maxOpen int
}

func Config(confs ...db.Configurator) (db.Specifier, error) {
	var spec sqlSpec
	var ok bool
	for _, conf := range confs {
		s, err := conf(spec)
		if err != nil {
			return nil, err
		}
		spec, ok = s.(sqlSpec)
		if !ok {
			return nil, fmt.Errorf("sql.Config requires SQL Specifier, got %T", s)
		}
	}
	if spec == nil {
		return nil, errors.New("nil Specifier generated, use sql.Connection")
	}
	return spec, nil
}

// Connection is a Configurator for a SQL database which ensures the correct
// connection parameters exist and sets them up.  This should be used to create
// any SQL connection.
func Connection(c sqlSpec) db.Configurator {
	return func(s db.Specifier) (db.Specifier, error) {
		if err := validateSqlSpec(c); err != nil {
			return nil, err
		}

		if _, ok := s.(sqlSpec); s != nil && !ok {
			return nil, fmt.Errorf("SQL Connection already created with %T", s)
		}

		return c, nil
	}
}

func (s *spec) getMaxOpenIdle() (int, int) {
	return s.maxOpen, s.maxIdle
}

func (s *spec) setMaxOpenIdle(maxO, maxI int) {
	s.maxOpen, s.maxIdle = maxO, maxI
}

// MaxOpenIdle determines the number of maximum open and idle connections to
// the database.  A value of 0 for `open` signals unlimited open connections.
func MaxOpenIdle(maxOpen, maxIdle int) db.Configurator {
	return func(s db.Specifier) (db.Specifier, error) {
		if tSpec, ok := s.(sqlSpec); ok {
			tSpec.setMaxOpenIdle(maxOpen, maxIdle)
			return tSpec, nil
		}
		return nil, fmt.Errorf("SQL MaxOpenIdle got bad Specifier %T", s)
	}
}

// NeedsUpdate returns true if the db.specifier needs to be updated.  This will
// cause a panic if we try to compare a SQL database to a non-SQL database.
func (s *spec) NeedsUpdate(dbSpec db.Specifier) bool {
	if dbSpec == nil {
		log.Panicf("tried to compare to nil Specifier!")
	}
	if tSpec, ok := dbSpec.(sqlSpec); ok {
		maxOpen, maxIdle := tSpec.getMaxOpenIdle()
		return s.maxIdle != maxIdle || s.maxOpen != maxOpen
	}
	// If this happened, we're in deep trouble!  Abandon ship!
	log.Panicf("tried to compare wrong database kinds: SQL and %T", dbSpec)
	return false
}

// DB implements db.DB and wraps a *sqlx.DB
type DB struct {
	*sqlx.DB
	conf sqlSpec
}

// Spec returns the db.Specifier for the given DB.
func (d *DB) Spec() db.Specifier {
	return d.conf
}

// Update updates an existing DB with the given db.Specifier.
func (d *DB) Update(s db.Specifier) error {
	if d.conf == nil {
		return errors.New("can't update a nil SQL database conf")
	}

	spec, ok := s.(sqlSpec)
	if !ok {
		return fmt.Errorf("can't update SQL database with \"%T\"", spec)
	}

	maxO, maxI := spec.getMaxOpenIdle()

	d.conf.setMaxOpenIdle(maxO, maxI)
	d.DB.SetMaxOpenConns(maxO)
	d.DB.SetMaxIdleConns(maxI)

	return nil
}

// newDB Opens a new *database/sql.DB, and calls wrapDB to wrap it with its config.
func newDB(s sqlSpec) (db.DB, error) {
	drv := s.driver()
	d := string(drv)
	if !knownDrivers.MatchString(d) {
		return nil, fmt.Errorf("unknown sql driver %q", d)
	}

	sqlDb, err := sql.Open(d, s.ConnectionString())
	if err != nil {
		return nil, err
	}

	return wrapDB(sqlDb, s)
}

// wrapDB wraps a *database/sql.DB with its config.
func wrapDB(sqlDb *sql.DB, s sqlSpec) (db.DB, error) {
	drv := s.driver()
	d := string(drv)
	if !knownDrivers.MatchString(d) {
		return nil, fmt.Errorf("unknown sql driver %q", d)
	}

	sqlxDb := sqlx.NewDb(sqlDb, d)

	maxO, maxI := s.getMaxOpenIdle()
	sqlxDb.SetMaxIdleConns(maxI)
	sqlxDb.SetMaxOpenConns(maxO)

	return &DB{sqlxDb, s}, nil
}

type validation struct {
	kw      string
	errCond bool
	val     interface{}
	err     error
}

func validateSqlSpec(s sqlSpec) error {
	switch tS := s.(type) {
	case *SQLServerSpec:
		if tS == nil {
			return errors.New("cannot create SQL Connection with nil Specifier")
		}
	case *PostgresSpec:
		if tS == nil {
			return errors.New("cannot create SQL Connection with nil Specifier")
		}
	case *MySQLSpec:
		if tS == nil {
			return errors.New("cannot create SQL Connection with nil Specifier")
		}
	default:
		return fmt.Errorf("cannot create SQL Connection with %T", tS)
	}

	return s.validate()
}

func validate(s sqlSpec, vs []validation) error {
	if s == nil {
		return errors.New("can't validate nil SQL Specifier")
	}
	if dr := s.driver(); !knownDrivers.MatchString(string(dr)) {
		return fmt.Errorf("unknown SQL driver %q", dr)
	}
	msg := string(s.driver()) + " config errors: "
	var count int
	for _, v := range vs {
		if v.errCond {
			count++
			var val string
			if str, ok := v.val.(string); ok {
				val = fmt.Sprintf("%q", str)
			} else {
				val = fmt.Sprintf("%v", v.val)
			}
			msg += fmt.Sprintf("bad value %s for %q", val, v.kw)
			if v.err != nil {
				msg += fmt.Sprintf(" (%s)", v.err.Error())
			}
			msg += "; "
		}
	}
	if count > 0 {
		return errors.New(msg[:len(msg)-2])
	}
	return nil
}
