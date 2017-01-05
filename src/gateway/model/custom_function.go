package model

import (
	"fmt"
	aperrors "gateway/errors"
	apsql "gateway/sql"
)

const (
	CustomFunctionLanguageJava   = "java"
	CustomFunctionLanguageNode   = "node"
	CustomFunctionLanguageCSharp = "c#"
	CustomFunctionLanguagePython = "python"
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

func (c *CustomFunction) ImageName() string {
	return fmt.Sprintf("%v_%v/%v", c.AccountID, c.APIID, c.ID)
}

func (c *CustomFunction) Validate(isInsert bool) aperrors.Errors {
	errors := make(aperrors.Errors)
	if c.Name == "" {
		errors.Add("name", "must have a name")
	}
	if isInsert {
		switch c.Language {
		case CustomFunctionLanguageJava:
		case CustomFunctionLanguageNode:
		case CustomFunctionLanguageCSharp:
		case CustomFunctionLanguagePython:
		default:
			errors.Add("language", "invalid language")
		}
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
	var err error
	if c.ID > 0 {
		err = db.Get(&function, db.SQL("custom_functions/find"), c.ID, c.APIID, c.AccountID)
	} else {
		err = db.Get(&function, db.SQL("custom_functions/find_name"), c.Name, c.APIID, c.AccountID)
	}
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

func (c *CustomFunction) AfterInsert(tx *apsql.Tx) error {
	dir := "custom_function/" + c.Language
	files, err := AssetDir(dir)
	if err != nil {
		return err
	}

	for _, f := range files {
		file := CustomFunctionFile{
			AccountID:        c.AccountID,
			UserID:           c.UserID,
			APIID:            c.APIID,
			CustomFunctionID: c.ID,
			Name:             f,
			Body:             string(MustAsset(dir + "/" + f)),
		}
		err = file.Insert(tx)
		if err != nil {
			return err
		}
	}

	return nil
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
