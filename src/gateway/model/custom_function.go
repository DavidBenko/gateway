package model

import (
	aperrors "gateway/errors"
	apsql "gateway/sql"
)

type CustomFunction struct {
	AccountID int64 `json:"-"`
	UserID    int64 `json:"-"`
	APIID     int64 `json:"api_id,omitempty" db:"api_id" path:"apiID"`

	ID          int64  `json:"id,omitempty" path:"id"`
	Name        string `json:"name"`
	Language    string `json:"language"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
}

func (c *CustomFunction) Validate(isInsert bool) aperrors.Errors {
	errors := make(aperrors.Errors)
	if c.Name == "" {
		errors.Add("name", "must have a name")
	}
	return errors
}

func (c *CustomFunction) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	if apsql.IsUniqueConstraint(err, "custom_functions", "api_id", "name") {
		errors.Add("name", "is already taken")
	}
	return errors
}

func (c *CustomFunction) All(db *apsql.DB) ([]*CustomFunction, error) {
	functions := []*CustomFunction{}
	err := db.Select(&functions, db.SQL("custom_functions/all"), c.APIID, c.AccountID)
	for _, function := range functions {
		function.AccountID = c.AccountID
		function.UserID = c.UserID
	}
	return functions, err
}

func (c *CustomFunction) Find(db *apsql.DB) (*CustomFunction, error) {
	function := CustomFunction{
		AccountID: c.AccountID,
		UserID:    c.UserID,
	}
	err := db.Get(&function, db.SQL("custom_functions/find"), c.ID, c.APIID, c.AccountID)
	return &function, err
}

func (c *CustomFunction) Delete(tx *apsql.Tx) error {
	err := tx.DeleteOne(tx.SQL("custom_functions/delete"), c.ID, c.APIID, c.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify("custom_functions", c.AccountID, c.UserID, c.APIID, 0, c.ID, apsql.Delete)
}

func (c *CustomFunction) Insert(tx *apsql.Tx) error {
	var err error
	c.ID, err = tx.InsertOne(tx.SQL("custom_functions/insert"),
		c.APIID, c.AccountID, c.Name, c.Description, c.Active)
	if err != nil {
		return err
	}

	return tx.Notify("custom_functions", c.AccountID, c.UserID, c.APIID, 0, c.ID, apsql.Insert)
}

func (c *CustomFunction) Update(tx *apsql.Tx) error {
	err := tx.UpdateOne(tx.SQL("custom_functions/update"),
		c.Name, c.Description, c.Active,
		c.ID, c.APIID, c.AccountID)
	if err != nil {
		return err
	}

	return tx.Notify("custom_functions", c.AccountID, c.UserID, c.APIID, 0, c.ID, apsql.Update)
}
