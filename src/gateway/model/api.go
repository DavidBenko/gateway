package model

import (
	"fmt"
	"gateway/config"
	apsql "gateway/sql"
	"log"
)

// API represents a top level grouping of endpoints accessible at a host.
type API struct {
	AccountID int64  `json:"-" db:"account_id"`
	ID        int64  `json:"id"`
	Name      string `json:"name"`
}

// Validate validates the model.
func (a *API) Validate() Errors {
	errors := make(Errors)
	if a.Name == "" {
		errors.add("name", "must not be blank")
	}
	return errors
}

// AllAPIsForAccountID returns all apis on the Account in default order.
func AllAPIsForAccountID(db *apsql.DB, accountID int64) ([]*API, error) {
	apis := []*API{}
	err := db.Select(&apis,
		"SELECT `id`, `name` FROM `apis` WHERE account_id = ? ORDER BY `name` ASC;",
		accountID)
	return apis, err
}

// FindAPIForAccountID returns the api with the id and account_id specified.
func FindAPIForAccountID(db *apsql.DB, id, accountID int64) (*API, error) {
	api := API{}
	err := db.Get(&api, "SELECT `id`, `name` FROM `apis` WHERE `id` = ? AND account_id = ?;",
		id, accountID)
	return &api, err
}

// DeleteAPIForAccountID deletes the api with the id and account_id specified.
func DeleteAPIForAccountID(tx *apsql.Tx, id, accountID int64) error {
	result, err := tx.Exec("DELETE FROM `apis` WHERE `id` = ? AND account_id = ?;",
		id, accountID)
	if err != nil {
		return err
	}

	numRows, err := result.RowsAffected()
	if err != nil || numRows != 1 {
		return fmt.Errorf("Expected 1 row to be affected; got %d, error: %v", numRows, err)
	}

	return nil
}

// Insert inserts the api into the database as a new row.
func (a *API) Insert(tx *apsql.Tx) error {
	result, err := tx.Exec("INSERT INTO `apis` (`account_id`, `name`) VALUES (?, ?);",
		a.AccountID, a.Name)
	if err != nil {
		return err
	}
	a.ID, err = result.LastInsertId()
	if err != nil {
		log.Printf("%s Error getting last insert ID for api: %v",
			config.System, err)
		return err
	}
	return nil
}

// Update updates the api in the database.
func (a *API) Update(tx *apsql.Tx) error {
	result, err := tx.Exec("UPDATE `apis` SET `name` = ? WHERE `id` = ? AND `account_id` = ?;",
		a.Name, a.ID, a.AccountID)
	if err != nil {
		return err
	}
	numRows, err := result.RowsAffected()
	if err != nil || numRows != 1 {
		return fmt.Errorf("Expected 1 row to be affected; got %d, error: %v", numRows, err)
	}
	return nil
}
