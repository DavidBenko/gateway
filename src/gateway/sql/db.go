package sql

import (
	"encoding/json"
	"fmt"
	"gateway/config"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	// Add sqlite3 driver
	_ "github.com/mattn/go-sqlite3"

	// Add postgres driver
	"github.com/lib/pq"

	"gateway/logreport"
)

const currentVersion = 11

type driverType string

const (
	// Sqlite3 driver type
	Sqlite3 = "sqlite3"

	// Postgres driver type
	Postgres = "postgres"
)

// DB wraps a *sqlx.DB with some convenience methods and data
type DB struct {
	*sqlx.DB
	Driver         driverType
	listeners      []Listener
	listenersMutex sync.RWMutex
}

// CurrentVersion returns the current version of the database, or an error if
// it has not been initialized.
func (db *DB) CurrentVersion() (int64, error) {
	var version int64
	err := db.Get(&version, `SELECT version FROM schema LIMIT 1`)
	if err != nil {
		return 0, fmt.Errorf("Could not get latest schema version: %v", err)
	}
	return version, err
}

// UpToDate returns whether or not the database is up to date
// with the latest schema
func (db *DB) UpToDate() bool {
	version, err := db.CurrentVersion()
	return (err == nil) && (version == currentVersion)
}

// Migrate migrates the database to the current version
func (db *DB) Migrate() error {
	version, err := db.CurrentVersion()
	if err != nil {
		if err = setupSchemaTable(db); err != nil {
			return fmt.Errorf("Could not create schema table: %v", err)
		}
	}

	migrations := []func(*DB) error{
		migrateToV1,
		migrateToV2,
		migrateToV3,
		migrateToV4,
		migrateToV5,
		migrateToV6,
		migrateToV7,
		migrateToV8,
		migrateToV9,
		migrateToV10,
		migrateToV11,
	}

	for i := version; i < currentVersion; i++ {
		// Note that i = 3 is 4th migration, V4, etc.
		if err = migrations[i](db); err != nil {
			return fmt.Errorf("Could not migrate to schema v%d: %v", i+1, err)
		}
	}

	return nil
}

// Begin creates a new sqlx transaction wrapped in our own code
func (db *DB) Begin() (*Tx, error) {
	tx, err := db.DB.Beginx()
	return &Tx{
		Tx:                   tx,
		DB:                   db,
		tags:                 []string{NotificationTagDefault},
		notifications:        []*Notification{},
		postCommitHooks:      nil,
		postCommitHooksMutex: sync.RWMutex{},
	}, err
}

// Get wraps sqlx's Get with driver-specific query modifications.
func (db *DB) Get(dest interface{}, query string, args ...interface{}) error {
	return db.DB.Get(dest, db.q(query), args...)
}

// Select wraps sqlx's Select with driver-specific query modifications.
func (db *DB) Select(dest interface{}, query string, args ...interface{}) error {
	return db.DB.Select(dest, db.q(query), args...)
}

// Queryx wraps sqlx's Queryx with driver-specific query modifications.
func (db *DB) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	return db.DB.Queryx(db.q(query), args...)
}

// RegisterListener registers a listener with the database
func (db *DB) RegisterListener(l Listener) {
	defer db.listenersMutex.Unlock()
	db.listenersMutex.Lock()
	db.listeners = append(db.listeners, l)
}

func (db *DB) notifyListeners(n *Notification) {
	defer db.listenersMutex.RUnlock()
	db.listenersMutex.RLock()

	for _, listener := range db.listeners {
		listener.Notify(n)
	}
}

func (db *DB) notifyListenersOfReconnection() {
	defer db.listenersMutex.RUnlock()
	db.listenersMutex.RLock()

	for _, listener := range db.listeners {
		listener.Reconnect()
	}
}

func (db *DB) startListening(conf config.Database) error {
	listener := pq.NewListener(conf.ConnectionString,
		2*time.Second,
		time.Minute,
		db.listenerConnectionEvent)
	err := listener.Listen(postgresNotifyChannel)
	if err != nil {
		return err
	}
	go db.waitForNotification(listener)
	return nil
}

func (db *DB) waitForNotification(l *pq.Listener) {
	for {
		select {
		case pgNotification := <-l.Notify:
			if pgNotification.Channel == postgresNotifyChannel {
				var notification Notification
				err := json.Unmarshal([]byte(pgNotification.Extra), &notification)
				if err != nil {
					logreport.Printf("%s Error parsing notification '%s': %v",
						config.System, pgNotification.Extra, err)
					continue
				}
				db.notifyListeners(&notification)
			} else {
				db.notifyListenersOfReconnection()
			}
		case <-time.After(90 * time.Second):
			go func() {
				l.Ping()
			}()
		}
	}
}

func (db *DB) listenerConnectionEvent(ev pq.ListenerEventType, err error) {
	if err != nil {
		logreport.Printf("%s Database listener connection problem: %v", config.System, err)
	}
}

// SQL returns a sql query from a static file, scoped to driver
func (db *DB) SQL(name string) string {
	asset := fmt.Sprintf("ansi/%s.sql", name)
	bytes, err := Asset(asset)
	if err != nil {
		asset = fmt.Sprintf("%s/%s.sql", db.Driver, name)
		bytes, err = Asset(asset)
		if err != nil {
			logreport.Fatalf("%s could not find %s", config.System, asset)
		}
	}
	return string(bytes)
}

// does driver modifications to query
func (db *DB) q(sql string) string {
	return q(sql, db.Driver)
}

// DoInTransaction takes a function, which is executed
// inside a new database transaction.  Any transaction-related
// errors are reported to the caller via the returned error
func (db *DB) DoInTransaction(logic func(tx *Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("Unable to begin transaction due to error: %v", err)
	}

	//err = endpoint.update(tx, false, false)
	err = logic(tx)

	if err != nil {
		rbErr := tx.Rollback()
		if rbErr != nil {
			return fmt.Errorf(`Encountered an error and attempted to rollback, but received error: "%v".  Original error was "%v"`, rbErr, err)
		}
		return fmt.Errorf("Rolled back transaction due to the following error: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Failed to commit transaction due to the following error: %v", err)
	}

	return nil
}
