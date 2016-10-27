package model

import (
	"errors"
	aperrors "gateway/errors"
	apsql "gateway/sql"
	"strings"
	"time"

	"github.com/jmoiron/sqlx/types"
)

type ScratchPad struct {
	AccountID        int64 `json:"-"`
	UserID           int64 `json:"-"`
	APIID            int64 `json:"-" path:"apiID"`
	RemoteEndpointID int64 `json:"-" path:"endpointID"`

	ID                int64          `json:"id,omitempty" path:"id"`
	EnvironmentDataID int64          `json:"environment_datum_id" db:"environment_data_id" path:"environmentDataID"`
	Name              string         `json:"name"`
	Code              string         `json:"code"`
	Data              types.JsonText `json:"-" db:"data"`

	// Export Indices
	ExportEnvironmentDataIndex int `json:"environment_data_index,omitempty"`
}

func (s *ScratchPad) Validate(isInsert bool) aperrors.Errors {
	errors := make(aperrors.Errors)
	if s.Name == "" || strings.TrimSpace(s.Name) == "" {
		errors.Add("name", "must not be blank")
	}
	return errors
}

func (s *ScratchPad) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	if apsql.IsUniqueConstraint(err, "scratch_pads", "environment_data_id", "name") {
		errors.Add("name", "is already taken")
	}
	return errors
}

func (s *ScratchPad) All(db *apsql.DB) ([]*ScratchPad, error) {
	pads := []*ScratchPad{}
	var err error
	if s.APIID > 0 && s.AccountID > 0 {
		if s.EnvironmentDataID > 0 && s.RemoteEndpointID > 0 {
			err = db.Select(&pads, db.SQL("scratch_pads/all"),
				s.EnvironmentDataID, s.RemoteEndpointID, s.APIID, s.AccountID)
		} else {
			err = db.Select(&pads, db.SQL("scratch_pads/all_api"), s.APIID, s.AccountID)
		}
		for _, pad := range pads {
			pad.AccountID = s.AccountID
			pad.UserID = s.UserID
			pad.APIID = s.APIID
			pad.RemoteEndpointID = s.RemoteEndpointID
		}
	} else {
		err = errors.New("APIID and AccountID required for All")
	}
	return pads, err
}

func (s *ScratchPad) Find(db *apsql.DB) (*ScratchPad, error) {
	pad := ScratchPad{
		AccountID:        s.AccountID,
		UserID:           s.UserID,
		APIID:            s.APIID,
		RemoteEndpointID: s.RemoteEndpointID,
	}
	err := db.Get(&pad, db.SQL("scratch_pads/find"), s.ID,
		s.EnvironmentDataID, s.RemoteEndpointID, s.APIID, s.AccountID)
	return &pad, err
}

func (s *ScratchPad) Delete(tx *apsql.Tx) error {
	err := tx.DeleteOne(tx.SQL("scratch_pads/delete"), s.ID,
		s.EnvironmentDataID, s.RemoteEndpointID, s.APIID, s.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify("scratch_pads", s.AccountID, s.UserID, s.APIID, 0, s.ID, apsql.Delete)
}

func (s *ScratchPad) Insert(tx *apsql.Tx) error {
	data, err := marshaledForStorage(s.Data)
	if err != nil {
		return err
	}

	s.ID, err = tx.InsertOne(tx.SQL("scratch_pads/insert"),
		s.EnvironmentDataID, s.RemoteEndpointID, s.APIID, s.AccountID,
		s.Name, s.Code, data, time.Now().UTC())
	if err != nil {
		return err
	}
	return tx.Notify("scratch_pads", s.AccountID, s.UserID, s.APIID, 0, s.ID, apsql.Insert)
}

func (s *ScratchPad) Update(tx *apsql.Tx) error {
	data, err := marshaledForStorage(s.Data)
	if err != nil {
		return err
	}

	err = tx.UpdateOne(tx.SQL("scratch_pads/update"), s.Name, s.Code, data, time.Now().UTC(), s.ID,
		s.EnvironmentDataID, s.RemoteEndpointID, s.APIID, s.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify("scratch_pads", s.AccountID, s.UserID, s.APIID, 0, s.ID, apsql.Update)
}
