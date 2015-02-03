package model

import apsql "gateway/sql"

// Host represents a host the API is available on.
type Host struct {
	AccountID int64 `json:"-"`
	APIID     int64 `json:"-" db:"api_id"`

	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Validate validates the model.
func (h *Host) Validate() Errors {
	errors := make(Errors)
	if h.Name == "" {
		errors.add("name", "must not be blank")
	}
	return errors
}

// ValidateFromDatabaseError translates possible database constraint errors
// into validation errors.
func (h *Host) ValidateFromDatabaseError(err error) Errors {
	errors := make(Errors)
	return errors
}

// AllHostsForAPIIDAndAccountID returns all hosts on the Account's API in default order.
func AllHostsForAPIIDAndAccountID(db *apsql.DB, apiID, accountID int64) ([]*Host, error) {
	hosts := []*Host{}
	err := db.Select(&hosts,
		`SELECT
			hosts.id as id,
			hosts.name as name
		FROM hosts, apis
		WHERE hosts.api_id = ?
			AND hosts.api_id = apis.id
			AND apis.account_id = ?
		ORDER BY hosts.name ASC;`,
		apiID, accountID)
	return hosts, err
}

// FindHostForAPIIDAndAccountID returns the host with the id, api id, and account_id specified.
func FindHostForAPIIDAndAccountID(db *apsql.DB, id, apiID, accountID int64) (*Host, error) {
	host := Host{}
	err := db.Get(&host,
		`SELECT
			hosts.id as id,
			hosts.name as name
		FROM hosts, apis
		WHERE hosts.id = ?
			AND hosts.api_id = ?
			AND hosts.api_id = apis.id
			AND apis.account_id = ?;`,
		id, apiID, accountID)
	return &host, err
}

// DeleteHostForAPIIDAndAccountID deletes the host with the id, api_id and account_id specified.
func DeleteHostForAPIIDAndAccountID(tx *apsql.Tx, id, apiID, accountID int64) error {
	return tx.DeleteOne(
		`DELETE FROM hosts
		WHERE hosts.id = ?
			AND hosts.api_id IN
				(SELECT id FROM apis WHERE id = ? AND account_id = ?)`,
		id, apiID, accountID)
}

// Insert inserts the host into the database as a new row.
func (h *Host) Insert(tx *apsql.Tx) (err error) {
	h.ID, err = tx.InsertOne(
		`INSERT INTO hosts (api_id, name)
		VALUES ((SELECT id FROM apis WHERE id = ? AND account_id = ?),?)`,
		h.APIID, h.AccountID, h.Name)
	return
}

// Update updates the host in the database.
func (h *Host) Update(tx *apsql.Tx) error {
	return tx.UpdateOne(
		`UPDATE hosts
		SET name = ?
		WHERE hosts.id = ?
			AND hosts.api_id IN
				(SELECT id FROM apis WHERE id = ? AND account_id = ?)`,
		h.Name, h.ID, h.APIID, h.AccountID)
}
