package model

import (
	"archive/tar"
	"bytes"
	"errors"
	aperrors "gateway/errors"
	apsql "gateway/sql"
	"time"
)

type CustomFunctionFile struct {
	AccountID        int64 `json:"-"`
	UserID           int64 `json:"-"`
	APIID            int64 `json:"-" db:"api_id" path:"apiID"`
	CustomFunctionID int64 `json:"custom_function_id,omitempty" db:"custom_function_id" path:"customFunctionID"`

	ID        int64      `json:"id,omitempty" path:"id"`
	CreatedAt *time.Time `json:"-" db:"created_at"`
	UpdatedAt *time.Time `json:"-" db:"updated_at"`
	Name      string     `json:"name"`
	Body      string     `json:"body"`

	// Export Indices
	ExportCustomFunctionIndex int `json:"custom_function_index,omitempty"`
}

type CustomFunctionFiles []*CustomFunctionFile

func (f CustomFunctionFiles) Tar() (*bytes.Buffer, error) {
	buffer := &bytes.Buffer{}
	image := tar.NewWriter(buffer)

	for _, file := range f {
		updated := file.UpdatedAt
		if updated == nil {
			updated = file.CreatedAt
		}
		header := &tar.Header{
			Name:    file.Name,
			Mode:    0600,
			Size:    int64(len(file.Body)),
			ModTime: *updated,
		}
		if err := image.WriteHeader(header); err != nil {
			return nil, err
		}
		if _, err := image.Write([]byte(file.Body)); err != nil {
			return nil, err
		}
	}

	if err := image.Close(); err != nil {
		return nil, err
	}

	return buffer, nil
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

func (f *CustomFunctionFile) All(db *apsql.DB) (CustomFunctionFiles, error) {
	files := []*CustomFunctionFile{}
	var err error
	if f.APIID > 0 && f.AccountID > 0 {
		if f.CustomFunctionID > 0 {
			err = db.Select(&files, db.SQL("custom_function_files/all"), f.CustomFunctionID, f.APIID, f.AccountID)
		} else {
			err = db.Select(&files, db.SQL("custom_function_files/all_api"), f.APIID, f.AccountID)
		}
	} else {
		err = errors.New("APIID and AccountID required for All")
	}
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		file.AccountID = f.AccountID
		file.UserID = f.UserID
	}
	return files, nil
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
