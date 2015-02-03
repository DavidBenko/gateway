package model

import apsql "gateway/sql"

// API represents a top level grouping of endpoints accessible at a host.
type API struct {
	AccountID   int64  `json:"-" db:"account_id"`
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CORSAllow   string `json:"cors_allow" db:"cors_allow"`
}

// Validate validates the model.
func (a *API) Validate() Errors {
	errors := make(Errors)
	if a.Name == "" {
		errors.add("name", "must not be blank")
	}
	if a.CORSAllow == "" {
		errors.add("cors_allow", "must not be blank (use '*' for everything)")
	}
	return errors
}

// ValidateFromDatabaseError translates possible database constraint errors
// into validation errors.
func (a *API) ValidateFromDatabaseError(err error) Errors {
	errors := make(Errors)
	if err.Error() == "UNIQUE constraint failed: apis.account_id, apis.name" ||
		err.Error() == `pq: duplicate key value violates unique constraint "apis_account_id_name_key"` {
		errors.add("name", "is already taken")
	}

	return errors
}

// AllAPIsForAccountID returns all apis on the Account in default order.
func AllAPIsForAccountID(db *apsql.DB, accountID int64) ([]*API, error) {
	apis := []*API{}
	err := db.Select(&apis,
		`SELECT id, name, description, cors_allow
		 FROM apis WHERE account_id = ?
		 ORDER BY name ASC;`,
		accountID)
	return apis, err
}

// FindAPIForAccountID returns the api with the id and account_id specified.
func FindAPIForAccountID(db *apsql.DB, id, accountID int64) (*API, error) {
	api := API{}
	err := db.Get(&api,
		`SELECT id, name, description, cors_allow
		 FROM apis
		 WHERE id = ? AND account_id = ?;`,
		id, accountID)
	return &api, err
}

// DeleteAPIForAccountID deletes the api with the id and account_id specified.
func DeleteAPIForAccountID(tx *apsql.Tx, id, accountID int64) error {
	return tx.DeleteOne(`DELETE FROM apis WHERE id = ? AND account_id = ?;`,
		id, accountID)
}

// Insert inserts the api into the database as a new row.
func (a *API) Insert(tx *apsql.Tx) (err error) {
	a.ID, err = tx.InsertOne(
		`INSERT INTO apis
		(account_id, name, description, cors_allow)
		VALUES (?, ?, ?, ?)`,
		a.AccountID, a.Name, a.Description, a.CORSAllow)
	return
}

// Update updates the api in the database.
func (a *API) Update(tx *apsql.Tx) error {
	return tx.UpdateOne(
		`UPDATE apis
		 SET name = ?, description = ?, cors_allow = ?
		 WHERE id = ? AND account_id = ?;`,
		a.Name, a.Description, a.CORSAllow, a.ID, a.AccountID)
}
