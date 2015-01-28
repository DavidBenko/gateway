package sql

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
)

func (tx *Tx) InsertOne(baseQuery string, args ...interface{}) (id int64, err error) {
	if tx.Driver == Postgres {
		query := tx.Q(baseQuery + ` RETURNING "id";`)
		err = tx.Get(&id, query, args...)
		return
	}

	result, err := tx.Exec(tx.Q(baseQuery+";"), args...)
	if err != nil {
		return
	}
	return result.LastInsertId()
}

func (tx *Tx) UpdateOne(query string, args ...interface{}) error {
	result, err := tx.Exec(query, args...)
	if err != nil {
		return err
	}
	numRows, err := result.RowsAffected()
	if err != nil || numRows != 1 {
		return fmt.Errorf("Expected 1 row to be affected; got %d, error: %v", numRows, err)
	}
	return nil
}

func (tx *Tx) DeleteOne(query string, args ...interface{}) error {
	return tx.UpdateOne(query, args...)
}

var qrx = regexp.MustCompile(`\?`)

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

func (db *DB) Q(sql string) string {
	return q(sql, db.Driver)
}

func (db *DB) Get(dest interface{}, query string, args ...interface{}) error {
	return db.DB.Get(dest, db.Q(query), args...)
}

func (db *DB) Select(dest interface{}, query string, args ...interface{}) error {
	return db.DB.Select(dest, db.Q(query), args...)
}

func (tx *Tx) Q(sql string) string {
	return q(sql, tx.Driver)
}

func (tx *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tx.Tx.Exec(tx.Q(query), args...)
}
