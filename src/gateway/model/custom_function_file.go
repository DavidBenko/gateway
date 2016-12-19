package model

import (
	aperrors "gateway/errors"
	apsql "gateway/sql"
)

type CustomFunctionFile struct {
	AccountID        int64 `json:"-"`
	UserID           int64 `json:"-"`
	APIID            int64 `json:"-" db:"api_id" path:"apiID"`
	CustomFunctionID int64 `json:"custom_function_id,omitempty" db:"custom_function_id" path:"customFunctionID"`

	ID   int64  `json:"id,omitempty" path:"id"`
	Name string `json:"name"`
	Body string `json:"body"`

	// Export Indices
	ExportCustomFunctionIndex int `json:"custom_function_index,omitempty"`
}

func (f *CustomFunctionFile) Validate(isInsert bool) aperrors.Errors {
	errors := make(aperrors.Errors)
	if f.Name == "" {
		errors.Add("name", "must have a name")
	}
	return errors
}

func (f *CustomFunctionFile) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	if apsql.IsUniqueConstraint(err, "custom_function_files", "custom_function_id", "name") {
		errors.Add("name", "is already taken")
	}
	return errors
}

func (f *CustomFunctionFile) All(db *apsql.DB) ([]*CustomFunctionFile, error) {
	files := []*CustomFunctionFile{}
	err := db.Select(&files, db.SQL("custom_function_files/all"), f.CustomFunctionID, f.APIID, f.AccountID)
	for _, file := range files {
		file.AccountID = f.AccountID
		file.UserID = f.UserID
	}
	return files, err
}

func (f *CustomFunctionFile) Find(db *apsql.DB) (*CustomFunctionFile, error) {
	file := CustomFunctionFile{
		AccountID: f.AccountID,
		UserID:    f.UserID,
	}
	err := db.Get(&file, db.SQL("custom_function_files/find"), f.ID, f.CustomFunctionID, f.APIID, f.AccountID)
	return &file, err
}

func (f *CustomFunctionFile) Delete(tx *apsql.Tx) error {
	err := tx.DeleteOne(tx.SQL("custom_function_files/delete"), f.ID, f.CustomFunctionID, f.APIID, f.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify("custom_function_files", f.AccountID, f.UserID, f.APIID, 0, f.ID, apsql.Delete)
}

func (f *CustomFunctionFile) Insert(tx *apsql.Tx) error {
	var err error
	f.ID, err = tx.InsertOne(tx.SQL("custom_function_files/insert"),
		f.CustomFunctionID, f.APIID, f.AccountID, f.Name, f.Body)
	if err != nil {
		return err
	}

	return tx.Notify("custom_function_files", f.AccountID, f.UserID, f.APIID, 0, f.ID, apsql.Insert)
}

func (f *CustomFunctionFile) Update(tx *apsql.Tx) error {
	err := tx.UpdateOne(tx.SQL("custom_function_files/update"),
		f.Name, f.Body,
		f.ID, f.CustomFunctionID, f.APIID, f.AccountID)
	if err != nil {
		return err
	}

	return tx.Notify("custom_function_files", f.AccountID, f.UserID, f.APIID, 0, f.ID, apsql.Update)
}
