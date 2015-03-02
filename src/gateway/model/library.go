package model

import (
	apsql "gateway/sql"

	"github.com/jmoiron/sqlx/types"
)

// Library represents a library the API is available on.
type Library struct {
	AccountID int64 `json:"-"`
	APIID     int64 `json:"api_id,omitempty" db:"api_id"`

	ID          int64          `json:"id,omitempty"`
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
	err := db.Select(&libraries, db.SQL("libraries/all"), apiID, accountID)
	return libraries, err
}

// AllLibrariesForProxy returns all libraries on the API in default order.
func AllLibrariesForProxy(db *apsql.DB, apiID int64) ([]*Library, error) {
	libraries := []*Library{}
	err := db.Select(&libraries, db.SQL("libraries/all_proxy"), apiID)
	return libraries, err
}

// FindLibraryForAPIIDAndAccountID returns the library with the id, api id, and account_id specified.
func FindLibraryForAPIIDAndAccountID(db *apsql.DB, id, apiID, accountID int64) (*Library, error) {
	library := Library{}
	err := db.Get(&library, db.SQL("libraries/find"), id, apiID, accountID)
	return &library, err
}

// DeleteLibraryForAPIIDAndAccountID deletes the library with the id, api_id and account_id specified.
func DeleteLibraryForAPIIDAndAccountID(tx *apsql.Tx, id, apiID, accountID int64) error {
	return tx.DeleteOne(tx.SQL("libraries/delete"), id, apiID, accountID)
}

// Insert inserts the library into the database as a new row.
func (l *Library) Insert(tx *apsql.Tx) error {
	data, err := marshaledForStorage(l.Data)
	if err != nil {
		return err
	}
	l.ID, err = tx.InsertOne(tx.SQL("libraries/insert"),
		l.APIID, l.AccountID, l.Name, l.Description, data)
	return err
}

// Update updates the library in the databasl.
func (l *Library) Update(tx *apsql.Tx) error {
	data, err := marshaledForStorage(l.Data)
	if err != nil {
		return err
	}
	return tx.UpdateOne(tx.SQL("libraries/update"),
		l.Name, l.Description, data, l.ID, l.APIID, l.AccountID)
}
