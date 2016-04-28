package model

import (
	"strings"

	aperrors "gateway/errors"
	apsql "gateway/sql"

	"github.com/jmoiron/sqlx/types"
)

type PushChannel struct {
	UserID int64 `json:"-"`

	ID               int64          `json:"id,omitempty" path:"id"`
	AccountID        int64          `json:"account_id" db:"account_id"`
	APIID            int64          `json:"api_id" db:"api_id"`
	RemoteEndpointID int64          `json:"remote_endpoint_id" db:"remote_endpoint_id"`
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
	err := db.Select(&channels, db.SQL("push_channels/all"), c.AccountID)
	if err != nil {
		return nil, err
	}
	for _, channel := range channels {
		channel.AccountID = c.AccountID
		channel.UserID = c.UserID
	}
	return channels, nil
}

func (c *PushChannel) Find(db *apsql.DB) (*PushChannel, error) {
	channel := PushChannel{
		AccountID: c.AccountID,
		UserID:    c.UserID,
	}
	var err error
	if c.ID == 0 {
		err = db.Get(&channel, db.SQL("push_channels/find_name"), c.Name,
			c.APIID, c.RemoteEndpointID, c.AccountID)
	} else {
		err = db.Get(&channel, db.SQL("push_channels/find"), c.ID,
			c.AccountID)
	}
	return &channel, err
}

func (c *PushChannel) Delete(tx *apsql.Tx) error {
	err := tx.DeleteOne(tx.SQL("push_channels/delete"), c.ID,
		c.AccountID)
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
		c.AccountID,
		c.APIID, c.AccountID,
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

	err = tx.UpdateOne(tx.SQL("push_channels/update"), c.APIID, c.AccountID,
		c.RemoteEndpointID, c.APIID, c.AccountID,
		c.Name, c.Expires, data,
		c.ID, c.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify("push_channels", c.AccountID, c.UserID, c.APIID, 0, c.ID, apsql.Update)
}
