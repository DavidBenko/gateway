package model

import (
	"strings"

	aperrors "gateway/errors"
	apsql "gateway/sql"

	"github.com/jmoiron/sqlx/types"
)

type PushDevice struct {
	AccountID        int64 `json:"-"`
	UserID           int64 `json:"-"`
	APIID            int64 `json:"-" path:"apiID"`
	RemoteEndpointID int64 `json:"-" path:"endpointID"`

	ID            int64          `json:"id,omitempty" path:"id"`
	PushChannelID int64          `json:"push_channel_id" db:"push_channel_id" path:"pushChannelID"`
	Name          string         `json:"name"`
	Type          string         `json:"type"`
	Token         string         `json:"token"`
	Expires       int64          `json:"expires"`
	Data          types.JsonText `json:"-" db:"data"`
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
	if apsql.IsUniqueConstraint(err, "push_devices", "push_device_id", "name") {
		errors.Add("name", "is already taken")
	}
	return errors
}

func (d *PushDevice) All(db *apsql.DB) ([]*PushDevice, error) {
	devices := []*PushDevice{}
	err := db.Select(&devices, db.SQL("push_devices/all"),
		d.PushChannelID, d.RemoteEndpointID, d.APIID, d.AccountID)
	if err != nil {
		return nil, err
	}
	for _, device := range devices {
		device.AccountID = d.AccountID
		device.UserID = d.UserID
		device.APIID = d.APIID
		device.RemoteEndpointID = d.RemoteEndpointID
	}
	return devices, nil
}

func (d *PushDevice) Find(db *apsql.DB) (*PushDevice, error) {
	device := PushDevice{
		AccountID:        d.AccountID,
		UserID:           d.UserID,
		APIID:            d.APIID,
		RemoteEndpointID: d.RemoteEndpointID,
	}
	err := db.Get(&device, db.SQL("push_devices/find"), d.ID,
		d.PushChannelID, d.RemoteEndpointID, d.APIID, d.AccountID)
	return &device, err
}

func (d *PushDevice) Delete(tx *apsql.Tx) error {
	err := tx.DeleteOne(tx.SQL("push_devices/delete"), d.ID,
		d.PushChannelID, d.RemoteEndpointID, d.APIID, d.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify("push_devices", d.AccountID, d.UserID, d.APIID, 0, d.ID, apsql.Delete)
}

func (d *PushDevice) Insert(tx *apsql.Tx) error {
	data, err := marshaledForStorage(d.Data)
	if err != nil {
		return err
	}

	d.ID, err = tx.InsertOne(tx.SQL("push_devices/insert"),
		d.PushChannelID, d.RemoteEndpointID, d.APIID, d.AccountID,
		d.Name, d.Type, d.Token, d.Expires, data)
	if err != nil {
		return err
	}
	return tx.Notify("push_devices", d.AccountID, d.UserID, d.APIID, 0, d.ID, apsql.Insert)
}

func (d *PushDevice) Update(tx *apsql.Tx) error {
	data, err := marshaledForStorage(d.Data)
	if err != nil {
		return err
	}

	err = tx.UpdateOne(tx.SQL("push_devices/update"), d.Name, d.Type, d.Token, d.Expires, data, d.ID,
		d.PushChannelID, d.RemoteEndpointID, d.APIID, d.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify("push_devices", d.AccountID, d.UserID, d.APIID, 0, d.ID, apsql.Update)
}
