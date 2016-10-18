package sql

//go:generate go-bindata -o sql_gen.go -nocompress -pkg sql static/...

import (
	drvr "database/sql/driver"
	"fmt"
	"gateway/config"
	"gateway/logreport"
	"github.com/jmoiron/sqlx"
	sqlite3 "github.com/mattn/go-sqlite3"
	"time"
)

// Driver is the driver to be used for the given stats logger / sampler.  This
// must be one of the given constants.
type Driver string

const (
	// Version is the version of the Gateway stats schema.
	Version = 1

	// SQLite3 is the SQLite3 Driver.
	SQLite3 Driver = "sqlite3"

	// Postgres is the Postgres Driver.
	Postgres Driver = "postgres"
)

// Given a time, dayMillis returns the number of milliseconds into the day.
func dayMillis(t time.Time) int64 {
	return int64(t.Hour()*1000*60*60 +
		t.Minute()*1000*60 +
		t.Second()*1000)
}

// SQL implements stats.Logger and stats.Sampler on a SQL backend.
type SQL struct {
	// NAME is the name of the given node.
	NAME string
	*sqlx.DB
}

// Connect opens and returns a database connection.
func Connect(conf config.Stats) (*SQL, error) {
	var driver Driver
	switch conf.Driver {
	case "sqlite3", "postgres":
		driver = Driver(conf.Driver)
	default:
		return nil,
			fmt.Errorf("Database driver must be sqlite3 or postgres (got '%v')",
				conf.Driver)
	}

	logreport.Printf("%s Connecting to database", config.System)
	sqlxDB, err := sqlx.Open(conf.Driver, conf.ConnectionString)
	if err != nil {
		return nil, err
	}

	sqlxDB.SetMaxOpenConns(int(conf.MaxConnections))

	db := SQL{"global", sqlxDB}

	switch driver {
	case SQLite3:
		// Foreign key support is disabled by default for each new sqlite connection.
		// Before actually establishing the connection (by calling 'Ping'),
		// add a ConnectHook that turns foreign_keys on.  This ensures that any
		// new sqlite connection that is added to the connection pool has foreign
		// keys enabled.
		if sqliteDriver, ok := db.DB.Driver().(*sqlite3.SQLiteDriver); ok {
			sqliteDriver.ConnectHook = func(conn *sqlite3.SQLiteConn) error {
				_, err := conn.Exec("PRAGMA foreign_keys=ON;", []drvr.Value{})
				return err
			}
		}
		// ConnectHook is set up successfully, so Ping to establish the connection
		err := db.Ping()
		if err != nil {
			return nil, err
		}
	case Postgres:
		// Establish the connection by executing the Ping first
		err := db.Ping()
		if err != nil {
			return nil, err
		}
	}

	return &db, nil
}

// quoteCol quotes a column name correctly depending on driver.
func (s *SQL) quoteCol(str string) string {
	if Driver(s.DriverName()) == SQLite3 {
		return fmt.Sprintf("`%s`", str)
	}
	return fmt.Sprintf(`"%s"`, str)
}

// Parameters returns the correct number of ?'s or $n's as a slice, depending on
// driver.
func (s *SQL) Parameters(n int) []string {
	if n < 1 {
		return nil
	}

	result := make([]string, n)

	if Driver(s.DriverName()) == SQLite3 {
		for i := 0; i < n; i++ {
			result[i] = "?"
		}
		return result
	}

	for i := 1; i <= n; i++ {
		result[i-1] = fmt.Sprintf(`$%d`, i)
	}

	return result
}
