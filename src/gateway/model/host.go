package model

import apsql "gateway/sql"

// Host represents a host the API is available on.
type Host struct {
	AccountID int64 `json:"-"`
	APIID     int64 `json:"api_id" db:"api_id"`

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
	if err.Error() == "UNIQUE constraint failed: hosts.name" ||
		err.Error() == `pq: duplicate key value violates unique constraint "hosts_name_key"` {
		errors.add("name", "is already taken")
	}
	return errors
}

// AllHostsForAPIIDAndAccountID returns all hosts on the Account's API in default order.
func AllHostsForAPIIDAndAccountID(db *apsql.DB, apiID, accountID int64) ([]*Host, error) {
	hosts := []*Host{}
	err := db.Select(&hosts, db.SQL("hosts/all"), apiID, accountID)
	return hosts, err
}

// AllHosts returns all hosts in an unspecified order.
func AllHosts(db *apsql.DB) ([]*Host, error) {
	hosts := []*Host{}
	err := db.Select(&hosts, db.SQL("hosts/all_routing"))
	return hosts, err
}

// FindHostForAPIIDAndAccountID returns the host with the id, api id, and account_id specified.
func FindHostForAPIIDAndAccountID(db *apsql.DB, id, apiID, accountID int64) (*Host, error) {
	host := Host{}
	err := db.Get(&host, db.SQL("hosts/find"), id, apiID, accountID)
	return &host, err
}

// DeleteHostForAPIIDAndAccountID deletes the host with the id, api_id and account_id specified.
func DeleteHostForAPIIDAndAccountID(tx *apsql.Tx, id, apiID, accountID int64) error {
	err := tx.DeleteOne(tx.SQL("hosts/delete"), id, apiID, accountID)
	if err != nil {
		return err
	}
	return tx.Notify("hosts", apiID, apsql.Delete)
}

// Insert inserts the host into the database as a new row.
func (h *Host) Insert(tx *apsql.Tx) (err error) {
	h.ID, err = tx.InsertOne(tx.SQL("hosts/insert"),
		h.APIID, h.AccountID, h.Name)
	if err != nil {
		return err
	}
	return tx.Notify("hosts", h.APIID, apsql.Insert)
}

// Update updates the host in the database.
func (h *Host) Update(tx *apsql.Tx) error {
	err := tx.UpdateOne(tx.SQL("hosts/update"),
		h.Name, h.ID, h.APIID, h.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify("hosts", h.APIID, apsql.Update)
}
