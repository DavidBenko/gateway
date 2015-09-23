package sql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
)

// Tx wraps a *sql.Tx with the driver we're using
type Tx struct {
	*sqlx.Tx
	db            *DB
	notifications []*Notification
}

// Get wraps sqlx's Get with driver-specific query modifications.
func (tx *Tx) Get(dest interface{}, query string, args ...interface{}) error {
	return tx.Tx.Get(dest, tx.q(query), args...)
}

// Select wraps sqlx's Select with driver-specific query modifications.
func (tx *Tx) Select(dest interface{}, query string, args ...interface{}) error {
	return tx.Tx.Select(dest, tx.q(query), args...)
}

// Exec wraps sqlx's Exec with driver-specific query modifications
func (tx *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tx.Tx.Exec(tx.q(query), args...)
}

// InsertOne inserts a row into the DB and returns the ID of the new row.
func (tx *Tx) InsertOne(baseQuery string, args ...interface{}) (id int64, err error) {
	if strings.HasSuffix(baseQuery, ";") {
		log.Fatalf("InsertOne query must not end in ;: %s", baseQuery)
	}
	if tx.db.Driver == Postgres {
		query := tx.q(baseQuery + ` RETURNING "id";`)
		err = tx.Get(&id, query, args...)
		return
	}

	result, err := tx.Exec(tx.q(baseQuery+";"), args...)
	if err != nil {
		return
	}
	return result.LastInsertId()
}

// UpdateOne updates a row, returning success iff 1 row was affected
func (tx *Tx) UpdateOne(query string, args ...interface{}) error {
	result, err := tx.Exec(query, args...)
	if err != nil {
		return err
	}
	numRows, err := result.RowsAffected()
	if err == nil && numRows == 0 {
		return ErrZeroRowsAffected
	}
	if err != nil || numRows != 1 {
		return fmt.Errorf("Expected 1 row to be affected; got %d, error: %v", numRows, err)
	}
	return nil
}

// DeleteOne is an alias for UpdateOne, for better semantic flavor when used
// with DELETE sql queries
func (tx *Tx) DeleteOne(query string, args ...interface{}) error {
	return tx.UpdateOne(query, args...)
}

// Notify creates a notification and posts it against this transaction.
//
// The implementation of posting a notification varies with database driver.
// Sqlite tracks notifications manually and notifies them on transaction commit
// (in memory on a single single box, to be used for development only), whereas
//  Postgres uses its NOTIFY command and triggers on commit for database-based
// listeners.
func (tx *Tx) Notify(table string, accountID, apiID, id int64, event NotificationEventType, messages ...interface{}) error {
	n := Notification{
		Table:     table,
		AccountID: accountID,
		APIID:     apiID,
		ID:        id,
		Event:     event,
		Messages:  messages,
	}

	if tx.db.Driver == Sqlite3 {
		tx.notifications = append(tx.notifications, &n)
		return nil
	}

	if tx.db.Driver == Postgres {
		json, err := json.Marshal(&n)
		if err != nil {
			return err
		}
		_, err = tx.Exec(fmt.Sprintf("Notify \"%s\", '%s'",
			postgresNotifyChannel, string(json)))
		return err
	}

	return nil
}

// Commit commits the transaction and sends out pending notifications
func (tx *Tx) Commit() error {
	err := tx.Tx.Commit()
	if err == nil {
		for _, n := range tx.notifications {
			tx.db.notifyListeners(n)
		}
	}
	return err
}

// SQL returns a sql query from a static file, scoped to driver
func (tx *Tx) SQL(name string) string {
	return tx.db.SQL(name)
}

// does driver modifications to query
func (tx *Tx) q(sql string) string {
	return tx.db.q(sql)
}
