package model

import (
	"errors"
	aperrors "gateway/errors"
	apsql "gateway/sql"
	"strings"

	"github.com/jmoiron/sqlx/types"
)

type PushDevice struct {
	AccountID int64 `json:"-"`
	UserID    int64 `json:"-"`

	ID               int64          `json:"id,omitempty" path:"id"`
	RemoteEndpointID int64          `json:"remote_endpoint_id" db:"remote_endpoint_id"`
	PushChannelID    int64          `json:"push_channel_id" path:"pushChannelID"`
	Expires          int64          `json:"expires"`
	Name             string         `json:"name"`
	Type             string         `json:"type"`
	Token            string         `json:"token"`
	Data             types.JsonText `json:"-" db:"data"`
}

type PushChannelPushDevice struct {
	ID            int64 `json:"id,omitempty" path:"id"`
	PushDevicelID int64 `json:"push_device_id" db:"push_device_id" path:"pushDeviceID"`
	PushChannelID int64 `json:"push_channel_id" db:"push_channel_id" path:"pushChannelID"`
	Expires       int64 `json:"expires"`
}

func (d *PushDevice) Validate(isInsert bool) aperrors.Errors {
	errors := make(aperrors.Errors)
	if d.Name == "" || strings.TrimSpace(d.Name) == "" {
		errors.Add("name", "must not be blank")
	}
	if d.Type == "" || strings.TrimSpace(d.Type) == "" {
		errors.Add("type", "must not be blank")
	}
	if d.Token == "" || strings.TrimSpace(d.Token) == "" {
		errors.Add("token", "must not be blank")
	}
	return errors
}

func (d *PushDevice) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	if apsql.IsUniqueConstraint(err, "push_devices", "remote_endpoint_id", "type", "token") {
		errors.Add("token", "is already taken")
	}
	return errors
}

func (d *PushDevice) All(db *apsql.DB) ([]*PushDevice, error) {
	devices := []*PushDevice{}
	var err error
	if d.PushChannelID > 0 {
		err = db.Select(&devices, db.SQL("push_devices/all_by_channel"),
			d.PushChannelID, d.AccountID)
	} else {
		err = db.Select(&devices, db.SQL("push_devices/all"), d.AccountID)
	}
	if err != nil {
		return nil, err
	}
	for _, device := range devices {
		device.AccountID = d.AccountID
		device.UserID = d.UserID
		if d.PushChannelID != 0 {
			device.PushChannelID = d.PushChannelID
		}
	}
	return devices, nil
}

func (d *PushDevice) Find(db *apsql.DB) (*PushDevice, error) {
	device := PushDevice{
		AccountID:        d.AccountID,
		UserID:           d.UserID,
		PushChannelID:    d.PushChannelID,
		RemoteEndpointID: d.RemoteEndpointID,
	}
	var err error
	if d.ID > 0 {
		if d.PushChannelID > 0 {
			err = db.Get(&device, db.SQL("push_devices/find_by_channel"), d.ID,
				d.PushChannelID, d.AccountID)
		} else {
			err = db.Get(&device, db.SQL("push_devices/find_by_id"), d.ID, d.AccountID)
		}
	} else {
		if d.PushChannelID > 0 {
			err = db.Get(&device, db.SQL("push_devices/find_by_token_and_channel"), d.Token,
				d.PushChannelID, d.AccountID)
		} else {
			err = db.Get(&device, db.SQL("push_devices/find_by_token_and_remote_endpoint"), d.Token, d.Type,
				d.RemoteEndpointID, d.AccountID)
		}
	}
	return &device, err
}

func (d *PushDevice) Delete(tx *apsql.Tx) error {
	var err error
	if d.PushChannelID > 0 {
		err = d.DeleteFromChannel(tx)
	} else {
		err = tx.DeleteOne(tx.SQL("push_devices/delete"), d.ID,
			d.RemoteEndpointID, d.AccountID)
	}
	if err != nil {
		return err
	}
	return tx.Notify("push_devices", d.AccountID, d.UserID, 0, 0, d.ID, apsql.Delete)
}

func (d *PushDevice) DeleteFromChannel(tx *apsql.Tx) error {
	err := tx.DeleteOne(tx.SQL("push_devices/delete_from_channel_mapping"), d.ID,
		d.PushChannelID, d.AccountID)
	if err != nil {
		return err
	}
	return nil
}

func (d *PushDevice) Insert(tx *apsql.Tx) error {
	pushChannel := PushChannel{ID: d.PushChannelID, AccountID: d.AccountID}
	err := tx.Get(&pushChannel, tx.SQL("push_channels/find"), pushChannel.ID,
		d.AccountID)
	if err == nil {
		d.RemoteEndpointID = pushChannel.RemoteEndpointID
		device := PushDevice{}
		err = tx.Get(&device, tx.SQL("push_devices/find_by_token_and_remote_endpoint"), d.Token, d.Type,
			d.RemoteEndpointID, d.AccountID)
		d.ID = device.ID
	}
	if err != nil {
		data, err := marshaledForStorage(d.Data)
		if err != nil {
			return err
		}
		d.ID, err = tx.InsertOne(tx.SQL("push_devices/insert"),
			d.PushChannelID, d.AccountID,
			d.Name, d.Type, d.Token, data)
		if err != nil {
			return err
		}
	} else {
		if !(d.PushChannelID > 0) {
			return errors.New("push_channel_id required")
		}
	}
	err = d.UpsertChannelMappings(tx)
	if err != nil {
		return err
	}
	return tx.Notify("push_devices", d.AccountID, d.UserID, 0, 0, d.ID, apsql.Insert)
}

func (d *PushDevice) Update(tx *apsql.Tx) error {
	data, err := marshaledForStorage(d.Data)
	if err != nil {
		return err
	}

	err = tx.UpdateOne(tx.SQL("push_devices/update"), d.Name, d.Type, d.Token, data,
		d.ID, d.PushChannelID, d.AccountID)
	if err != nil {
		return err
	}
	err = d.UpsertChannelMappings(tx)
	if err != nil {
		return err
	}
	return tx.Notify("push_devices", d.AccountID, d.UserID, 0, 0, d.ID, apsql.Update)
}

func (d *PushDevice) UpsertChannelMappings(tx *apsql.Tx) error {
	pushChannelPushDevice := PushChannelPushDevice{}
	err := tx.DB.Get(&pushChannelPushDevice, tx.DB.SQL("push_devices/find_channel_mapping"), d.ID,
		d.PushChannelID, d.AccountID)
	if err != nil {
		_, err = tx.InsertOne(tx.SQL("push_devices/insert_channel_mapping"), d.ID, d.PushChannelID, d.AccountID, d.Expires)
		if err != nil {
			return err
		}
	} else {
		err = tx.UpdateOne(tx.SQL("push_devices/update_channel_mapping"), d.Expires, d.ID, d.PushChannelID, d.AccountID)
		if err != nil {
			return err
		}
	}
	return nil
}
