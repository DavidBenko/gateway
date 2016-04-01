package model

import (
	"strings"

	aperrors "gateway/errors"
	apsql "gateway/sql"

	"github.com/jmoiron/sqlx/types"
)

type PushChannel struct {
	AccountID int64 `json:"-"`
	UserID    int64 `json:"-"`
	APIID     int64 `json:"-" path:"apiID"`

	ID               int64          `json:"id,omitempty" path:"id"`
	RemoteEndpointID int64          `json:"remote_endpoint_id" db:"remote_endpoint_id" path:"endpointID"`
	Name             string         `json:"name"`
	Expires          int64          `json:"expires"`
	Data             types.JsonText `json:"-" db:"data"`
}

func (c *PushChannel) Validate(isInsert bool) aperrors.Errors {
	errors := make(aperrors.Errors)
	if c.Name == "" || strings.TrimSpace(c.Name) == "" {
		errors.Add("name", "must not be blank")
	}
	return errors
}

func (c *PushChannel) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	if apsql.IsUniqueConstraint(err, "push_channels", "remote_endpoint_id", "name") {
		errors.Add("name", "is already taken")
	}
	return errors
}

func (c *PushChannel) All(db *apsql.DB) ([]*PushChannel, error) {
	channels := []*PushChannel{}
	err := db.Select(&channels, db.SQL("push_channels/all"),
		c.RemoteEndpointID, c.APIID, c.AccountID)
	if err != nil {
		return nil, err
	}
	for _, channel := range channels {
		channel.AccountID = c.AccountID
		channel.UserID = c.UserID
		channel.APIID = c.APIID
	}
	return channels, nil
}

func (c *PushChannel) Find(db *apsql.DB) (*PushChannel, error) {
	channel := PushChannel{
		AccountID: c.AccountID,
		UserID:    c.UserID,
		APIID:     c.APIID,
	}
	err := db.Get(&channel, db.SQL("push_channels/find"), c.ID,
		c.RemoteEndpointID, c.APIID, c.AccountID)
	return &channel, err
}

func (c *PushChannel) Delete(tx *apsql.Tx) error {
	err := tx.DeleteOne(tx.SQL("push_channels/delete"), c.ID,
		c.RemoteEndpointID, c.APIID, c.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify("push_channels", c.AccountID, c.UserID, c.APIID, 0, c.ID, apsql.Delete)
}

func (c *PushChannel) Insert(tx *apsql.Tx) error {
	data, err := marshaledForStorage(c.Data)
	if err != nil {
		return err
	}

	c.ID, err = tx.InsertOne(tx.SQL("push_channels/insert"),
		c.RemoteEndpointID, c.APIID, c.AccountID,
		c.Name, c.Expires, data)
	if err != nil {
		return err
	}
	return tx.Notify("push_channels", c.AccountID, c.UserID, c.APIID, 0, c.ID, apsql.Insert)
}

func (c *PushChannel) Update(tx *apsql.Tx) error {
	data, err := marshaledForStorage(c.Data)
	if err != nil {
		return err
	}

	err = tx.UpdateOne(tx.SQL("push_channels/update"), c.Name, c.Expires, data, c.ID,
		c.RemoteEndpointID, c.APIID, c.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify("push_channels", c.AccountID, c.UserID, c.APIID, 0, c.ID, apsql.Update)
}
