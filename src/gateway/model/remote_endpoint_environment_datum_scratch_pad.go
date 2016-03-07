package model

import (
	"errors"
	"strings"

	aperrors "gateway/errors"
	apsql "gateway/sql"

	"github.com/jmoiron/sqlx/types"
)

type RemoteEndpointEnvironmentDatumScratchPad struct {
	AccountID        int64 `json:"-"`
	UserID           int64 `json:"-"`
	APIID            int64 `json:"-" path:"apiID"`
	RemoteEndpointID int64 `json:"-" path:"endpointID"`

	ID                int64          `json:"id,omitempty" path:"id"`
	EnvironmentDataID int64          `json:"environment_datum_id" db:"remote_endpoint_environment_data_id" path:"environmentDataID"`
	Name              string         `json:"name"`
	Code              string         `json:"code"`
	Data              types.JsonText `json:"-" db:"data"`

	// Export Indices
	ExportEnvironmentDataIndex int `json:"environment_data_index,omitempty"`
}

func (s *RemoteEndpointEnvironmentDatumScratchPad) Validate(isInsert bool) aperrors.Errors {
	errors := make(aperrors.Errors)
	if s.Name == "" || strings.TrimSpace(s.Name) == "" {
		errors.Add("name", "must not be blank")
	}
	return errors
}

func (s *RemoteEndpointEnvironmentDatumScratchPad) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	if apsql.IsUniqueConstraint(err, "scratch_pads", "remote_endpoint_environment_data_id", "name") {
		errors.Add("name", "is already taken")
	}
	return errors
}

func (s *RemoteEndpointEnvironmentDatumScratchPad) All(db *apsql.DB) ([]*RemoteEndpointEnvironmentDatumScratchPad, error) {
	pads := []*RemoteEndpointEnvironmentDatumScratchPad{}
	var err error
	if s.APIID > 0 && s.AccountID > 0 {
		if s.EnvironmentDataID > 0 && s.RemoteEndpointID > 0 {
			err = db.Select(&pads, db.SQL("scratch_pads/all"),
				s.EnvironmentDataID, s.RemoteEndpointID, s.APIID, s.AccountID)
		} else {
			err = db.Select(&pads, db.SQL("scratch_pads/all_api"), s.APIID, s.AccountID)
		}
	} else {
		err = errors.New("APIID and AccountID required for All")
	}
	return pads, err
}

func (s *RemoteEndpointEnvironmentDatumScratchPad) Find(db *apsql.DB) (*RemoteEndpointEnvironmentDatumScratchPad, error) {
	pad := RemoteEndpointEnvironmentDatumScratchPad{}
	err := db.Get(&pad, db.SQL("scratch_pads/find"), s.ID,
		s.EnvironmentDataID, s.RemoteEndpointID, s.APIID, s.AccountID)
	return &pad, err
}

func (s *RemoteEndpointEnvironmentDatumScratchPad) Delete(tx *apsql.Tx) error {
	err := tx.DeleteOne(tx.SQL("scratch_pads/delete"), s.ID,
		s.EnvironmentDataID, s.RemoteEndpointID, s.APIID, s.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify("scratch_pads", s.AccountID, s.UserID, s.APIID, 0, s.ID, apsql.Delete)
}

func (s *RemoteEndpointEnvironmentDatumScratchPad) Insert(tx *apsql.Tx) error {
	data, err := marshaledForStorage(s.Data)
	if err != nil {
		return err
	}

	s.ID, err = tx.InsertOne(tx.SQL("scratch_pads/insert"),
		s.EnvironmentDataID, s.RemoteEndpointID, s.APIID, s.AccountID,
		s.Name, s.Code, data)
	if err != nil {
		return err
	}
	return tx.Notify("scratch_pads", s.AccountID, s.UserID, s.APIID, 0, s.ID, apsql.Insert)
}

func (s *RemoteEndpointEnvironmentDatumScratchPad) Update(tx *apsql.Tx) error {
	data, err := marshaledForStorage(s.Data)
	if err != nil {
		return err
	}

	err = tx.UpdateOne(tx.SQL("scratch_pads/update"), s.Name, s.Code, data, s.ID,
		s.EnvironmentDataID, s.RemoteEndpointID, s.APIID, s.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify("scratch_pads", s.AccountID, s.UserID, s.APIID, 0, s.ID, apsql.Update)
}
