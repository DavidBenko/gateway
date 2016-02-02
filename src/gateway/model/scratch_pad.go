package model

import (
	aperrors "gateway/errors"
	apsql "gateway/sql"

	"github.com/jmoiron/sqlx/types"
)

type ScratchPad struct {
	AccountID        int64 `json:"-"`
	UserID           int64 `json:"-"`
	APIID            int64 `json:"-" path:"apiID"`
	RemoteEndpointID int64 `json:"-" path:"endpointID"`

	ID                              int64          `json:"id,omitempty" path:"id"`
	RemoteEndpointEnvironmentDataID int64          `json:"remote_endpoint_environment_data_id" db:"remote_endpoint_environment_data_id" path:"environmentDataID"`
	Name                            string         `json:"name"`
	Code                            string         `json:"code"`
	Data                            types.JsonText `json:"data" db:"data"`
}

func (s *ScratchPad) Validate(isInsert bool) aperrors.Errors {
	errors := make(aperrors.Errors)
	return errors
}

func (s *ScratchPad) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	return errors
}

func (s *ScratchPad) All(db *apsql.DB) ([]*ScratchPad, error) {
	pads := []*ScratchPad{}
	err := db.Select(&pads, db.SQL("scratch_pads/all"),
		s.RemoteEndpointEnvironmentDataID, s.RemoteEndpointID, s.APIID, s.AccountID)
	return pads, err
}

func (s *ScratchPad) Find(db *apsql.DB) (*ScratchPad, error) {
	pad := ScratchPad{}
	return &pad, nil
}

func (s *ScratchPad) Delete(tx *apsql.Tx) error {
	return nil
}

func (s *ScratchPad) Insert(tx *apsql.Tx) error {
	return nil
}

func (s *ScratchPad) Update(tx *apsql.Tx) error {
	return nil
}
