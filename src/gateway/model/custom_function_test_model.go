package model

import (
	"errors"
	aperrors "gateway/errors"
	apsql "gateway/sql"

	"github.com/jmoiron/sqlx/types"
)

type CustomFunctionTest struct {
	AccountID        int64 `json:"-"`
	UserID           int64 `json:"-"`
	APIID            int64 `json:"-" db:"api_id" path:"apiID"`
	CustomFunctionID int64 `json:"custom_function_id,omitempty" db:"custom_function_id" path:"customFunctionID"`

	ID    int64          `json:"id,omitempty" path:"id"`
	Name  string         `json:"name"`
	Input types.JsonText `json:"input,omitempty"`

	// Export Indices
	ExportCustomFunctionIndex int `json:"custom_function_index,omitempty"`
}

func (t *CustomFunctionTest) Validate(isInsert bool) aperrors.Errors {
	errors := make(aperrors.Errors)
	if t.Name == "" {
		errors.Add("name", "must have a name")
	}
	return errors
}

func (t *CustomFunctionTest) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	if apsql.IsUniqueConstraint(err, "custom_function_tests", "custom_function_id", "name") {
		errors.Add("name", "is already taken")
	}
	return errors
}

func (t *CustomFunctionTest) All(db *apsql.DB) ([]*CustomFunctionTest, error) {
	tests := []*CustomFunctionTest{}
	var err error
	if t.APIID > 0 && t.AccountID > 0 {
		if t.CustomFunctionID > 0 {
			err = db.Select(&tests, db.SQL("custom_function_tests/all"), t.CustomFunctionID, t.APIID, t.AccountID)
		} else {
			err = db.Select(&tests, db.SQL("custom_function_tests/all_api"), t.APIID, t.AccountID)
		}
	} else {
		err = errors.New("APIID and AccountID required for All")
	}
	if err != nil {
		return nil, err
	}
	for _, test := range tests {
		test.AccountID = t.AccountID
		test.UserID = t.UserID
	}
	return tests, nil
}

func (t *CustomFunctionTest) Find(db *apsql.DB) (*CustomFunctionTest, error) {
	test := CustomFunctionTest{
		AccountID: t.AccountID,
		UserID:    t.UserID,
	}
	err := db.Get(&test, db.SQL("custom_function_tests/find"), t.ID, t.CustomFunctionID, t.APIID, t.AccountID)
	return &test, err
}

func (t *CustomFunctionTest) Delete(tx *apsql.Tx) error {
	err := tx.DeleteOne(tx.SQL("custom_function_tests/delete"), t.ID, t.CustomFunctionID, t.APIID, t.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify("custom_function_tests", t.AccountID, t.UserID, t.APIID, 0, t.ID, apsql.Delete)
}

func (t *CustomFunctionTest) Insert(tx *apsql.Tx) error {
	input, err := marshaledForStorage(t.Input)
	if err != nil {
		return err
	}

	t.ID, err = tx.InsertOne(tx.SQL("custom_function_tests/insert"), t.CustomFunctionID,
		t.APIID, t.AccountID, t.Name, input)
	if err != nil {
		return err
	}
	return tx.Notify("custom_function_tests", t.AccountID, t.UserID, t.APIID, 0, t.ID, apsql.Insert)
}

func (t *CustomFunctionTest) Update(tx *apsql.Tx) error {
	input, err := marshaledForStorage(t.Input)
	if err != nil {
		return err
	}

	err = tx.UpdateOne(tx.SQL("custom_function_tests/update"), t.Name, input,
		t.ID, t.CustomFunctionID, t.APIID, t.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify("custom_function_tests", t.AccountID, t.UserID, t.APIID, 0, t.ID, apsql.Update)
}
