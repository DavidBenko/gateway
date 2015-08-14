package postgres

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
var spaces *regexp.Regexp
var escapeChars *regexp.Regexp

// init compiles non-unique keys when the package is loaded.
func init() {
	nonUniqueKeys = regexp.MustCompile("^connect_timeout$")
	spaces = regexp.MustCompile(" ")
	escapeChars = regexp.MustCompile("'")
}

// Conn defines Postgres connection string parameters.
type Conn map[string]interface{}

// Spec is a Specifier for a Postgres connection.
type Spec struct {
	pqConn  Conn
	maxIdle int
	maxOpen int
}

// Config sets up a Postgres connection with the given Configurators as follows:
//
// import pq "gateway/db/postgres"
//
// conn, err := pool.Connect(pq.Config(
//         pq.Connection(s spec),
//         pq.MaxOpenIdle(10, 100),
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

// Connection is a Configurator for a Postgres database which ensures the
// correct connection parameters exist and sets them up.  This should be used
// to create any Postgres connection.
func Connection(s Conn) db.Configurator {
	return func(spec db.Specifier) error {
		// http://godoc.org/github.com/lib/pq#hdr-Connection_String_Parameters
		for _, k := range []string{
			"port",
			"user",
			"password",
			"dbname",
			"host",
		} {
			if _, ok := s[k]; !ok {
				return fmt.Errorf("Postgres Config missing %q key", k)
			}
		}

		if len(s) == 6 {
			if _, ok := s["connect_timeout"]; !ok {
				return fmt.Errorf("unexpected key in Postgres connection %+v", s)
			}
		}

		if len(s) > 6 {
			return fmt.Errorf("bad Postgres connection map %+v", s)
		}

		switch pqSpec := spec.(type) {
		case *Spec:
			pqSpec.pqConn = s
			return nil
		default:
			return fmt.Errorf("Postgres Connection requires postgres.Conn, got %T", spec)
		}
	}
}

// MaxOpenIdle determines the number of maximum open and idle connections to
// the database.  A value of 0 for `open` signals unlimited open connections.
func MaxOpenIdle(open, idle int) db.Configurator {
	maxOpen, maxIdle := open, idle
	return func(spec db.Specifier) error {
		switch pqSpec := spec.(type) {
		case *Spec:
			pqSpec.maxIdle, pqSpec.maxOpen = maxIdle, maxOpen
			return nil
		default:
			return fmt.Errorf("Postgres MaxOpenIdle requires postgres.Conn, got %T", spec)
		}
	}
}

// ConnectionString serializes the Postgres connection parameters into a usable
// string.
func (s *Spec) ConnectionString() string {
	conn := s.pqConn

	buf := bytes.NewBufferString("postgres://")

	if username, ok := conn["username"]; ok {
		if usernameStr, ok := username.(string); ok {
			buf.WriteString(usernameStr)
			if password, ok := conn["password"]; ok {
				if passwordStr, ok := password.(string); ok {
					buf.WriteString(":")
					buf.WriteString(passwordStr)
				}
			}
			buf.WriteString("@")
		}
	}

	if host, ok := conn["host"]; ok {
		if hostStr, ok := host.(string); ok {
			buf.WriteString(hostStr)
		}
	}

	if port, ok := conn["port"]; ok {
		if portStr, ok := port.(string); ok {
			buf.WriteString(":")
			buf.WriteString(portStr)
		}
	}

	if dbname, ok := conn["dbname"]; ok {
		if dbnameStr, ok := dbname.(string); ok {
			buf.WriteString("/")
			buf.WriteString(dbnameStr)
		}
	}

	return buf.String()
}

// UniqueServer serializes the unique parts of the given *Spec.
func (s *Spec) UniqueServer() string {
	conn := s.pqConn

	// Get a sorted array of keys of s.
	keys := []string{}
	for key := range conn {
		if !nonUniqueKeys.MatchString(key) {
			// our key must not be in the RE of non-unique keys
			// i.e. connection timeout.
			keys = append(keys, key)
		}
	}

	return conn.serializeFor(keys)
}

func (c Conn) serializeFor(keys []string) string {
	sort.Strings(keys)

	// Iterate over the config map and get values.  Escape and quote
	// as needed.
	var buffer bytes.Buffer
	for _, key := range keys {
		val := fmt.Sprintf("%v", c[key])
		if escapeChars.MatchString(val) {
			val = escapeChars.ReplaceAllString(val, `\$0`)
		}
		if spaces.MatchString(val) {
			val = "'" + val + "'"
		}
		buffer.WriteString(fmt.Sprintf("%s=%v ", key, val))
	}

	s := buffer.String()
	return s[:len(s)-1]
}

// NeedsUpdate returns true if the db.Specifier needs to be updated.
func (s *Spec) NeedsUpdate(spec db.Specifier) bool {
	pqSpec := spec.(*Spec)
	return s.maxIdle != pqSpec.maxIdle || s.maxOpen != pqSpec.maxOpen
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
		return fmt.Errorf("can't update Postgres database with \"%T\"", spec)
	}

	d.conf.maxIdle = spec.maxIdle
	d.conf.maxOpen = spec.maxOpen
	d.DB.SetMaxIdleConns(spec.maxIdle)
	d.DB.SetMaxOpenConns(spec.maxOpen)
	return nil
}

// NewDB creates a new *sqlx.DB, and wraps it with its config in a *DB.
func (s *Spec) NewDB() (db.DB, error) {
	pqDB, err := sql.Open("pgx", s.ConnectionString())
	if err != nil {
		return nil, err
	}

	db := sqlx.NewDb(pqDB, "pgx")

	db.SetMaxIdleConns(s.maxIdle)
	db.SetMaxOpenConns(s.maxOpen)

	return &DB{db, s}, nil
}
