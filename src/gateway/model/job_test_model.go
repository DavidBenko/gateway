package model

import (
	"errors"
	aperrors "gateway/errors"
	apsql "gateway/sql"

	"github.com/jmoiron/sqlx/types"
)

type JobTest struct {
	AccountID int64 `json:"-"`
	UserID    int64 `json:"-"`
	APIID     int64 `json:"-" db:"api_id" path:"apiID"`
	JobID     int64 `json:"job_id,omitempty" db:"job_id" path:"jobID"`

	ID         int64          `json:"id,omitempty" path:"id"`
	Name       string         `json:"name"`
	Parameters types.JsonText `json:"parameters,omitempty"`

	// Export Indices
	ExportJobIndex int `json:"job_index,omitempty"`
}

func (t *JobTest) Validate(isInsert bool) aperrors.Errors {
	errors := make(aperrors.Errors)
	if t.Name == "" {
		errors.Add("name", "must have a name")
	}
	return errors
}

func (t *JobTest) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	if apsql.IsUniqueConstraint(err, "job_tests", "job_id", "name") {
		errors.Add("name", "is already taken")
	}
	return errors
}

func (t *JobTest) All(db *apsql.DB) ([]*JobTest, error) {
	tests := []*JobTest{}
	var err error
	if t.APIID > 0 && t.AccountID > 0 {
		if t.JobID > 0 {
			err = db.Select(&tests, db.SQL("job_tests/all"), t.JobID, t.APIID, t.AccountID)
		} else {
			err = db.Select(&tests, db.SQL("job_tests/all_api"), t.APIID, t.AccountID)
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

func (t *JobTest) Find(db *apsql.DB) (*JobTest, error) {
	test := JobTest{
		AccountID: t.AccountID,
		UserID:    t.UserID,
	}
	err := db.Get(&test, db.SQL("job_tests/find"), t.ID, t.JobID, t.APIID, t.AccountID)
	return &test, err
}

func (t *JobTest) Delete(tx *apsql.Tx) error {
	err := tx.DeleteOne(tx.SQL("job_tests/delete"), t.ID, t.JobID, t.APIID, t.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify("job_tests", t.AccountID, t.UserID, t.APIID, t.JobID, t.ID, apsql.Delete)
}

func (t *JobTest) Insert(tx *apsql.Tx) error {
	parameters, err := marshaledForStorage(t.Parameters)
	if err != nil {
		return err
	}

	t.ID, err = tx.InsertOne(tx.SQL("job_tests/insert"), t.JobID,
		t.APIID, t.AccountID, t.Name, parameters)
	if err != nil {
		return err
	}
	return tx.Notify("job_tests", t.AccountID, t.UserID, t.APIID, t.JobID, t.ID, apsql.Insert)
}

func (t *JobTest) Update(tx *apsql.Tx) error {
	parameters, err := marshaledForStorage(t.Parameters)
	if err != nil {
		return err
	}

	err = tx.UpdateOne(tx.SQL("job_tests/update"), t.Name, parameters,
		t.ID, t.JobID, t.APIID, t.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify("job_tests", t.AccountID, t.UserID, t.APIID, t.JobID, t.ID, apsql.Update)
}
