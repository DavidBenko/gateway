package model

import (
	apsql "gateway/sql"

	"github.com/jmoiron/sqlx/types"
)

// Library represents a library the API is available on.
type Library struct {
	AccountID int64 `json:"-"`
	APIID     int64 `json:"-" db:"api_id"`

	ID          int64          `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Data        types.JsonText `json:"data"`
}

// Validate validates the model.
func (l *Library) Validate() Errors {
	errors := make(Errors)
	if l.Name == "" {
		errors.add("name", "must not be blank")
	}
	return errors
}

// ValidateFromDatabaseError translates possible database constraint errors
// into validation errors.
func (l *Library) ValidateFromDatabaseError(err error) Errors {
	errors := make(Errors)
	if err.Error() == "UNIQUE constraint failed: libraries.api_id, libraries.name" ||
		err.Error() == `pq: duplicate key value violates unique constraint "libraries_api_id_name_key"` {
		errors.add("name", "is already taken")
	}
	return errors
}

// AllLibrariesForAPIIDAndAccountID returns all libraries on the Account's API in default order.
func AllLibrariesForAPIIDAndAccountID(db *apsql.DB, apiID, accountID int64) ([]*Library, error) {
	libraries := []*Library{}
	err := db.Select(&libraries,
		`SELECT
			libraries.id as id,
			libraries.name as name,
			libraries.description as description,
			libraries.data as data
		FROM libraries, apis
		WHERE libraries.api_id = ?
			AND libraries.api_id = apis.id
			AND apis.account_id = ?
		ORDER BY libraries.name ASC;`,
		apiID, accountID)
	return libraries, err
}

// FindLibraryForAPIIDAndAccountID returns the library with the id, api id, and account_id specified.
func FindLibraryForAPIIDAndAccountID(db *apsql.DB, id, apiID, accountID int64) (*Library, error) {
	library := Library{}
	err := db.Get(&library,
		`SELECT
			libraries.id as id,
			libraries.name as name,
			libraries.description as description
			libraries.data as data
		FROM libraries, apis
		WHERE libraries.id = ?
			AND libraries.api_id = ?
			AND libraries.api_id = apis.id
			AND apis.account_id = ?;`,
		id, apiID, accountID)
	return &library, err
}

// DeleteLibraryForAPIIDAndAccountID deletes the library with the id, api_id and account_id specified.
func DeleteLibraryForAPIIDAndAccountID(tx *apsql.Tx, id, apiID, accountID int64) error {
	return tx.DeleteOne(
		`DELETE FROM libraries
		WHERE libraries.id = ?
			AND libraries.api_id IN
				(SELECT id FROM apis WHERE id = ? AND account_id = ?);`,
		id, apiID, accountID)
}

// Insert inserts the library into the database as a new row.
func (l *Library) Insert(tx *apsql.Tx) error {
	data, err := l.Data.MarshalJSON()
	if err != nil {
		return err
	}
	l.ID, err = tx.InsertOne(
		`INSERT INTO libraries (api_id, name, description, data)
		VALUES ((SELECT id FROM apis WHERE id = ? AND account_id = ?),?, ?, ?)`,
		l.APIID, l.AccountID, l.Name, l.Description, string(data))
	return err
}

// Update updates the library in the databasl.
func (l *Library) Update(tx *apsql.Tx) error {
	data, err := l.Data.MarshalJSON()
	if err != nil {
		return err
	}
	return tx.UpdateOne(
		`UPDATE libraries
		SET name = ?, description = ?, data = ?
		WHERE libraries.id = ?
			AND libraries.api_id IN
			(SELECT id FROM apis WHERE id = ? AND account_id = ?);`,
		l.Name, l.Description, string(data), l.ID, l.APIID, l.AccountID)
}
