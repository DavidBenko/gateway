package sql

import (
	"errors"
	"fmt"
	"gateway/config"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
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

// Notification is used to serialize information to pass with events
type Notification struct {
	Table string
	APIID int64
	Event NotificationEventType
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
	sqlxDB, err := sqlx.Connect(conf.Driver, conf.ConnectionString)
	if err != nil {
		return nil, err
	}

	db := DB{sqlxDB, driver, []Listener{}, sync.RWMutex{}}

	switch conf.Driver {
	case Sqlite3:
		// Foreign key support is disabled by default
		db.Exec("PRAGMA foreign_keys = ON;")

	case Postgres:
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

	if errString == pgString {
		return true
	}

	fullKeys := []string{}
	for _, k := range keys {
		fullKeys = append(fullKeys, strings.Join([]string{table, k}, "."))
	}

	sqliteString := fmt.Sprintf("UNIQUE constraint failed: %s", strings.Join(fullKeys, ", "))
	return errString == sqliteString
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
