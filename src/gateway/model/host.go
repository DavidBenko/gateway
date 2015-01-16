package model

import (
	"fmt"
	"gateway/config"
	apsql "gateway/sql"
	"log"
)

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

// AllHostsForAPIIDAndAccountID returns all hosts on the Account's API in default order.
func AllHostsForAPIIDAndAccountID(db *apsql.DB, apiID, accountID int64) ([]*Host, error) {
	hosts := []*Host{}
	err := db.Select(&hosts,
		"SELECT `hosts`.`id` as `id`, `hosts`.`name` as `name` "+
			"FROM `hosts`, `apis` "+
			"WHERE `hosts`.`api_id` = ? "+
			"AND `hosts`.`api_id` = `apis`.`id` "+
			"AND `apis`.`account_id` = ? "+
			"ORDER BY `hosts`.`name` ASC;",
		apiID, accountID)
	return hosts, err
}

// FindHostForAPIIDAndAccountID returns the host with the id, api id, and account_id specified.
func FindHostForAPIIDAndAccountID(db *apsql.DB, id, apiID, accountID int64) (*Host, error) {
	host := Host{}
	err := db.Get(&host,
		"SELECT `hosts`.`id` as `id`, `hosts`.`name` as `name` "+
			"FROM `hosts`, `apis` "+
			"WHERE `hosts`.`api_id` = ? "+
			"AND `hosts`.`api_id` = `apis`.`id` "+
			"AND `apis`.`account_id` = ? ",
		id, apiID, accountID)
	return &host, err
}

// DeleteHostForAPIIDAndAccountID deletes the host with the id, api_id and account_id specified.
func DeleteHostForAPIIDAndAccountID(tx *apsql.Tx, id, apiID, accountID int64) error {
	result, err := tx.Exec(
		"DELETE FROM `hosts`"+
			"INNER JOIN `apis` ON `hosts`.`api_id` = `api`.`id`"+
			"WHERE `hosts`.`id` = ? AND `apis`.`id` = ? AND `apis`.`account_id` = ?;",
		id, apiID, accountID)
	if err != nil {
		return err
	}

	numRows, err := result.RowsAffected()
	if err != nil || numRows != 1 {
		return fmt.Errorf("Expected 1 row to be affected; got %d, error: %v", numRows, err)
	}

	return nil
}

// Insert inserts the host into the database as a new row.
func (h *Host) Insert(tx *apsql.Tx) error {
	result, err := tx.Exec("INSERT INTO `hosts` (`api_id`, `name`) VALUES (?, ?);",
		h.APIID, h.Name)
	if err != nil {
		return err
	}
	h.ID, err = result.LastInsertId()
	if err != nil {
		log.Printf("%s Error getting last insert ID for host: %v",
			config.System, err)
		return err
	}
	return nil
}

// Update updates the host in the database.
func (h *Host) Update(tx *apsql.Tx) error {
	result, err := tx.Exec(
		"UPDATE `hosts` "+
			"INNER JOIN `apis` ON `hosts`.`api_id` = `api`.`id`"+
			"SET `name` = ? "+
			"WHERE `hosts`.`id` = ? AND `apis`.`id` = ? AND `apis`.`account_id` = ?;",
		h.Name, h.ID, h.APIID, h.AccountID)
	if err != nil {
		return err
	}
	numRows, err := result.RowsAffected()
	if err != nil || numRows != 1 {
		return fmt.Errorf("Expected 1 row to be affected; got %d, error: %v", numRows, err)
	}
	return nil
}
