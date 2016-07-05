package model

import (
	aperrors "gateway/errors"
	apsql "gateway/sql"

	"github.com/jmoiron/sqlx/types"
)

type PushChannelMessage struct {
	AccountID int64 `json:"-"`
	UserID    int64 `json:"-"`

	ID            int64          `json:"id,omitempty" path:"id"`
	PushChannelID int64          `json:"push_channel_id" db:"push_channel_id" path:"pushChannelID"`
	Stamp         int64          `json:"stamp"`
	Data          types.JsonText `json:"data" db:"data"`
}

func (d *PushChannelMessage) Validate(isInsert bool) aperrors.Errors {
	errors := make(aperrors.Errors)
	return errors
}

func (d *PushChannelMessage) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	return errors
}

func (m *PushChannelMessage) All(db *apsql.DB) ([]*PushChannelMessage, error) {
	messages := []*PushChannelMessage{}
	var err error
	if m.PushChannelID > 0 {
		err = db.Select(&messages, db.SQL("push_channel_messages/all_by_channel"), m.PushChannelID, m.AccountID)
	} else {
		err = db.Select(&messages, db.SQL("push_channel_messages/all"), m.AccountID)
	}
	if err != nil {
		return nil, err
	}
	for _, message := range messages {
		message.AccountID = m.AccountID
		message.UserID = m.UserID
	}
	return messages, nil
}

func (m *PushChannelMessage) Find(db *apsql.DB) (*PushChannelMessage, error) {
	message := PushChannelMessage{
		AccountID:     m.AccountID,
		UserID:        m.UserID,
		PushChannelID: m.PushChannelID,
	}
	err := db.Get(&message, db.SQL("push_channel_messages/find"), m.ID, m.PushChannelID, m.AccountID)
	return &message, err
}

func (m *PushChannelMessage) Delete(tx *apsql.Tx) error {
	err := tx.DeleteOne(tx.SQL("push_channel_messages/delete"), m.ID, m.PushChannelID, m.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify("push_channel_messages", m.AccountID, m.UserID, 0, 0, m.ID, apsql.Delete)
}

func (m *PushChannelMessage) Insert(tx *apsql.Tx) error {
	data, err := marshaledForStorage(m.Data)
	if err != nil {
		return err
	}

	m.ID, err = tx.InsertOne(tx.SQL("push_channel_messages/insert"), m.PushChannelID, m.AccountID,
		m.Stamp, data)
	if err != nil {
		return err
	}
	return tx.Notify("push_channel_messages", m.AccountID, m.UserID, 0, 0, m.ID, apsql.Insert)
}

func (m *PushChannelMessage) Update(tx *apsql.Tx) error {
	data, err := marshaledForStorage(m.Data)
	if err != nil {
		return err
	}

	err = tx.UpdateOne(tx.SQL("push_channel_messages/update"), m.Stamp, data, m.ID, m.PushChannelID, m.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify("push_channel_messages", m.AccountID, m.UserID, 0, 0, m.ID, apsql.Update)
}
