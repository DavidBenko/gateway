package sql

import (
	drvr "database/sql/driver"
	"errors"
	"fmt"
	"gateway/config"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
	sqlite3 "github.com/mattn/go-sqlite3"
)

// NotificationEventType determined what happened in a notified update.
type NotificationEventType int

const (
	// Insert means a row was inserted
	Insert NotificationEventType = iota

	// Update means a row was updated
	Update

	// Delete means a row was deleted
	Delete
)

const postgresNotifyChannel = "gateway"

const (
	NOTIFICATION_TAG_DEFAULT = "default"
	NOTIFICATION_TAG_AUTO    = "auto"
	NOTIFICATION_TAG_IMPORT  = "import"
)

// Notification is used to serialize information to pass with events
type Notification struct {
	Table     string
	AccountID int64
	UserID    int64
	APIID     int64
	ID        int64
	Event     NotificationEventType
	Tag       string
	Messages  []interface{}
}

// A Listener gets notified of notifications
type Listener interface {
	// Notify tells the listener a particular notification was fired
	Notify(*Notification)

	// Reconnect tells the listener that we may have been disconnected, but
	// have reconnected. They should update all state that could have changed.
	Reconnect()
}

// ErrZeroRowsAffected is an error returned from updates when 0 rows were changed.
var ErrZeroRowsAffected = errors.New("Zero rows affected")

var qrx = regexp.MustCompile(`\?`)

// Connect opens and returns a database connection.
func Connect(conf config.Database) (*DB, error) {
	var driver driverType
	switch conf.Driver {
	case "sqlite3", "postgres":
		driver = driverType(conf.Driver)
	default:
		return nil,
			fmt.Errorf("Database driver must be sqlite3 or postgres (got '%v')",
				conf.Driver)
	}

	log.Printf("%s Connecting to database", config.System)
	sqlxDB, err := sqlx.Open(conf.Driver, conf.ConnectionString)
	if err != nil {
		return nil, err
	}

	sqlxDB.SetMaxOpenConns(int(conf.MaxConnections))

	db := DB{sqlxDB, driver, []Listener{}, sync.RWMutex{}}

	switch conf.Driver {
	case Sqlite3:
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
		// Connection is established, proceed with listener
		// We use Postgres' NOTIFY to update state in a cluster
		db.startListening(conf)
	}

	return &db, nil
}

// IsUniqueConstraint returns whether or not the error looks like a unique constraint error
func IsUniqueConstraint(err error, table string, keys ...string) bool {
	errString := err.Error()
	pgString := fmt.Sprintf("pq: duplicate key value violates unique constraint \"%s_%s_key\"",
		table, strings.Join(keys, "_"))

	if strings.Contains(errString, pgString) {
		return true
	}

	fullKeys := []string{}
	for _, k := range keys {
		fullKeys = append(fullKeys, strings.Join([]string{table, k}, "."))
	}

	sqliteString := fmt.Sprintf("UNIQUE constraint failed: %s", strings.Join(fullKeys, ", "))
	return strings.Contains(errString, sqliteString)
}

// IsNotNullConstraint returns whether or not the error looks like a not null constraint error
func IsNotNullConstraint(err error, table, column string) bool {
	errString := err.Error()
	pgString := fmt.Sprintf("pq: null value in column \"%s\" violates not-null constraint",
		column)

	if strings.Contains(errString, pgString) {
		return true
	}

	sqliteString := fmt.Sprintf("NOT NULL constraint failed: %s.%s", table, column)
	return strings.Contains(errString, sqliteString)
}

// NQs returns n comma separated '?'s
func NQs(n int) string {
	return strings.Join(strings.Split(strings.Repeat("?", n), ""), ",")
}

// q converts "?" characters to $1, $2, $n on Postgres
func q(sql string, driver driverType) string {
	if driver == Sqlite3 {
		return sql
	}
	n := 0
	return qrx.ReplaceAllStringFunc(sql, func(string) string {
		n++
		return "$" + strconv.Itoa(n)
	})
}
