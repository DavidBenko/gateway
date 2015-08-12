package sqlserver

import (
	"bytes"
	"database/sql"
	"fmt"
	"regexp"
	"sort"

	"github.com/jmoiron/sqlx"

	"gateway/db"
)

// nonUniqueKeys is the set of keys to exclude from UniqueServer()
var nonUniqueKeys *regexp.Regexp

// init compiles non-unique keys when the package is loaded.
func init() {
	nonUniqueKeys = regexp.MustCompile("^connection timeout$")
}

// Conn defines MSSQL connection string parameters.
type Conn map[string]interface{}

// Spec is a Specifier for a MSSQL connection.
type Spec struct {
	sqlConn Conn
	maxIdle int
	maxOpen int
}

// Config sets up a SQL connection with the given Configurators as follows:
//
// import sqls "gateway/db/sqlserver"
//
// conn, err := pool.Connect(sqls.Config(
//         sqls.Connection(s spec),
//         sqls.MaxOpenIdle(10, 100),
// ))
//
// Usable configs:
//  - Connection(s Conn)
//  - MaxOpenIdle(open, idle int)
func Config(confs ...db.Configurator) (db.Specifier, error) {
	spec := &Spec{}
	for _, conf := range confs {
		err := conf(spec)
		if err != nil {
			return nil, err
		}
	}
	return spec, nil
}

// Connection is a Configurator for a SQL Server database which ensures the
// correct connection parameters exist and sets them up.  This should be used
// to create any SQL Server connection.
func Connection(s Conn) db.Configurator {
	return func(spec db.Specifier) error {
		// https://github.com/denisenkom/go-mssqldb#connection-parameters
		for _, k := range []string{
			"server",
			"port",
			"user id",
			"password",
			"database",
			"schema",
		} {
			if _, ok := s[k]; !ok {
				return fmt.Errorf("SQL Config missing %q key", k)
			}
		}

		switch sqls := spec.(type) {
		case *Spec:
			sqls.sqlConn = s
			return nil
		default:
			return fmt.Errorf("SQL Server Connection requires Conn, got %T", spec)
		}
	}
}

// MaxOpenIdle determines the number of maximum open and idle connections to
// the database.  A value of 0 for `open` signals unlimited open connections.
func MaxOpenIdle(open, idle int) db.Configurator {
	maxOpen, maxIdle := open, idle
	return func(spec db.Specifier) error {
		switch sqls := spec.(type) {
		case *Spec:
			sqls.maxIdle, sqls.maxOpen = maxIdle, maxOpen
			return nil
		default:
			return fmt.Errorf("SQL Server MaxOpenIdle requires sqlserver.Conn, got %T", spec)
		}
	}
}

// ConnectionString serializes the SQL Server connection parameters into a
// usable string.
func (s *Spec) ConnectionString() string {
	conn := s.sqlConn
	// https://github.com/denisenkom/go-mssqldb#connection-parameters
	// Get a sorted array of keys of s.
	keys := make([]string, len(conn))
	var i int
	for key := range conn {
		keys[i] = key
		i++
	}

	sort.Strings(keys)

	// Now iterate over the config map and get values.
	var buffer bytes.Buffer
	for _, key := range keys {
		buffer.WriteString(fmt.Sprintf("%s=%v;", key, conn[key]))
	}

	return buffer.String()
}

// UniqueServer serializes the unique parts of the given *Spec.
func (s *Spec) UniqueServer() string {
	conn := s.sqlConn

	// Get a sorted array of keys of s.
	keys := []string{}
	for key := range conn {
		if !nonUniqueKeys.Match([]byte(key)) {
			// our key must not be in the RE of non-unique keys
			// i.e. connection timeout.
			keys = append(keys, key)
		}
	}

	sort.Strings(keys)

	// Now iterate over the config map and get values.
	var buffer bytes.Buffer
	for _, key := range keys {
		buffer.WriteString(fmt.Sprintf("%s=%v;", key, conn[key]))
	}

	return buffer.String()
}

// NeedsUpdate returns true if the db.Specifier needs to be updated.
func (s *Spec) NeedsUpdate(spec db.Specifier) bool {
	sqls := spec.(*Spec)
	return s.maxIdle != sqls.maxIdle || s.maxOpen != sqls.maxOpen
}

// DB implements db.DB and wraps a *sqlx.DB
type DB struct {
	*sqlx.DB
	conf *Spec
}

// Spec returns the db.Specifier for the given DB.
func (d *DB) Spec() db.Specifier {
	return d.conf
}

// Update updates an existing DB with the given db.Specifier.
func (d *DB) Update(s db.Specifier) error {
	spec, ok := s.(*Spec)
	if !ok {
		return fmt.Errorf("can't update SQL Server database with \"%T\"", spec)
	}

	d.conf.maxIdle = spec.maxIdle
	d.conf.maxOpen = spec.maxOpen
	d.DB.SetMaxIdleConns(spec.maxIdle)
	d.DB.SetMaxOpenConns(spec.maxOpen)
	return nil
}

// NewDB creates a new *sqlx.DB, and wraps it with its config in a *DB.
func (s *Spec) NewDB() (db.DB, error) {
	sqlsDB, err := sql.Open("mssql", s.ConnectionString())
	if err != nil {
		return nil, err
	}

	db := sqlx.NewDb(sqlsDB, "mssql")

	db.SetMaxIdleConns(s.maxIdle)
	db.SetMaxOpenConns(s.maxOpen)

	return &DB{db, s}, nil
}
